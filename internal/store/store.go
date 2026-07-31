package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/yuanweize/atomic-sync/internal/model"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;
CREATE TABLE IF NOT EXISTS jobs (id TEXT PRIMARY KEY, data BLOB NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS runs (id TEXT PRIMARY KEY, job_id TEXT NOT NULL, unit TEXT NOT NULL, destination TEXT NOT NULL, state TEXT NOT NULL, message TEXT NOT NULL DEFAULT '', started_at TEXT NOT NULL, finished_at TEXT);
CREATE INDEX IF NOT EXISTS idx_runs_job_started ON runs(job_id, started_at DESC);
CREATE TABLE IF NOT EXISTS assignments (job_id TEXT NOT NULL, unit TEXT NOT NULL, destination TEXT NOT NULL, PRIMARY KEY(job_id, unit));
CREATE TABLE IF NOT EXISTS analyses (job_id TEXT PRIMARY KEY, data BLOB NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error                   { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *Store) SaveJob(ctx context.Context, j model.Job) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if j.CreatedAt.IsZero() {
		var created string
		if err := tx.QueryRowContext(ctx, `SELECT created_at FROM jobs WHERE id=?`, j.ID).Scan(&created); err == nil {
			j.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		}
		if j.CreatedAt.IsZero() {
			j.CreatedAt = now
		}
	}
	j.UpdatedAt = now
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jobs(id,data,created_at,updated_at) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET data=excluded.data,updated_at=excluded.updated_at`, j.ID, b, j.CreatedAt.Format(time.RFC3339Nano), j.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM analyses WHERE job_id=?`, j.ID); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) Jobs(ctx context.Context) ([]model.Job, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT data FROM jobs ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Job{}
	for rows.Next() {
		var b []byte
		var j model.Job
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &j); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
func (s *Store) Job(ctx context.Context, id string) (model.Job, error) {
	var b []byte
	err := s.db.QueryRowContext(ctx, `SELECT data FROM jobs WHERE id=?`, id).Scan(&b)
	var j model.Job
	if err == nil {
		err = json.Unmarshal(b, &j)
	}
	return j, err
}
func (s *Store) DeleteJob(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM assignments WHERE job_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM analyses WHERE job_id=?`, id); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM jobs WHERE id=?`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}
func (s *Store) Assignment(ctx context.Context, job, unit string) (string, error) {
	var d string
	err := s.db.QueryRowContext(ctx, `SELECT destination FROM assignments WHERE job_id=? AND unit=?`, job, unit).Scan(&d)
	return d, err
}

func (s *Store) HasAssignments(ctx context.Context, job string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM assignments WHERE job_id=?)`, job).Scan(&exists)
	return exists, err
}
func (s *Store) Assign(ctx context.Context, job, unit, dest string) (string, error) {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO assignments(job_id,unit,destination) VALUES(?,?,?)`, job, unit, dest)
	if err != nil {
		return "", err
	}
	return s.Assignment(ctx, job, unit)
}
func (s *Store) CreateRun(ctx context.Context, r model.Run) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO runs(id,job_id,unit,destination,state,message,started_at) VALUES(?,?,?,?,?,?,?)`, r.ID, r.JobID, r.Unit, r.Destination, r.State, r.Message, r.StartedAt.Format(time.RFC3339Nano))
	return err
}
func (s *Store) Transition(ctx context.Context, id, to, msg string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var from string
	if err = tx.QueryRowContext(ctx, `SELECT state FROM runs WHERE id=?`, id).Scan(&from); err != nil {
		return err
	}
	if !model.CanTransition(from, to) {
		return errors.New("invalid state transition")
	}
	var done any
	if to == "completed" || to == "failed" {
		done = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE runs SET state=?,message=?,finished_at=? WHERE id=?`, to, msg, done, id); err != nil {
		return err
	}
	return tx.Commit()
}

// FailInterruptedRuns reconciles records left non-terminal by a previous
// process crash or forced shutdown. Staging data is deliberately preserved.
func (s *Store) FailInterruptedRuns(ctx context.Context, message string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx, `UPDATE runs SET state='failed',message=?,finished_at=? WHERE state NOT IN ('completed','failed')`, message, now)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) FailInterruptedAnalyses(ctx context.Context, message string) (int, error) {
	analyses, err := s.Analyses(ctx)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, analysis := range analyses {
		if analysis.State != "running" {
			continue
		}
		finished := time.Now().UTC()
		analysis.State = "failed"
		analysis.Message = message
		analysis.FinishedAt = &finished
		analysis.Units = nil
		if err = s.SaveAnalysis(ctx, analysis); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Store) SaveAnalysis(ctx context.Context, analysis model.Analysis) error {
	payload, err := json.Marshal(analysis)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO analyses(job_id,data,updated_at) VALUES(?,?,?) ON CONFLICT(job_id) DO UPDATE SET data=excluded.data,updated_at=excluded.updated_at`, analysis.JobID, payload, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) DeleteAnalysis(ctx context.Context, jobID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM analyses WHERE job_id=?`, jobID)
	return err
}

func (s *Store) Analysis(ctx context.Context, jobID string) (model.Analysis, error) {
	var payload []byte
	err := s.db.QueryRowContext(ctx, `SELECT data FROM analyses WHERE job_id=?`, jobID).Scan(&payload)
	var analysis model.Analysis
	if err == nil {
		err = json.Unmarshal(payload, &analysis)
	}
	return analysis, err
}

func (s *Store) Analyses(ctx context.Context) ([]model.Analysis, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT data FROM analyses ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	analyses := []model.Analysis{}
	for rows.Next() {
		var payload []byte
		var analysis model.Analysis
		if err = rows.Scan(&payload); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(payload, &analysis); err != nil {
			return nil, err
		}
		analyses = append(analyses, analysis)
	}
	return analyses, rows.Err()
}
func (s *Store) Runs(ctx context.Context, limit int) ([]model.Run, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,job_id,unit,destination,state,message,started_at,finished_at FROM runs ORDER BY started_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Run{}
	for rows.Next() {
		var r model.Run
		var start string
		var finish sql.NullString
		if err = rows.Scan(&r.ID, &r.JobID, &r.Unit, &r.Destination, &r.State, &r.Message, &start, &finish); err != nil {
			return nil, err
		}
		r.StartedAt, _ = time.Parse(time.RFC3339Nano, start)
		if finish.Valid {
			t, _ := time.Parse(time.RFC3339Nano, finish.String)
			r.FinishedAt = &t
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (s *Store) Stats(ctx context.Context) map[string]int {
	out := map[string]int{"jobs": 0, "running": 0, "completed": 0, "failed": 0}
	var jobs int
	s.db.QueryRowContext(ctx, `SELECT count(*) FROM jobs`).Scan(&jobs)
	out["jobs"] = jobs
	rows, err := s.db.QueryContext(ctx, `SELECT state,count(*) FROM runs GROUP BY state`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var k string
			var n int
			rows.Scan(&k, &n)
			if k == "completed" || k == "failed" {
				out[k] = n
			} else if k != "completed" && k != "failed" {
				out["running"] += n
			}
		}
	}
	return out
}
