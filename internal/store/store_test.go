package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/yuanweize/atomic-sync/internal/model"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	database, err := Open(filepath.Join(t.TempDir(), "atomic.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func testJob() model.Job {
	return model.Job{
		ID: "job_test", Name: "Test archive", Source: "/sources/media",
		Destinations: []model.Destination{{Name: "gd", Path: "GD:data/media", Weight: 1}},
		Mode:         "copy", Grouping: "folder", Concurrency: 1, Verify: "checksum", ConflictPolicy: model.ConflictFail, DryRun: true,
	}
}

func TestStoreLifecycle(t *testing.T) {
	ctx := context.Background()
	database := openTestStore(t)
	if err := database.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	job := testJob()
	if assigned, err := database.HasAssignments(ctx, job.ID); err != nil || assigned {
		t.Fatalf("new job unexpectedly has assignments: %v %v", assigned, err)
	}
	if err := database.SaveJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	stored, err := database.Job(ctx, job.ID)
	if err != nil || stored.CreatedAt.IsZero() {
		t.Fatalf("stored job invalid: %#v %v", stored, err)
	}
	created := stored.CreatedAt
	stored.Name = "Updated"
	time.Sleep(time.Millisecond)
	if err = database.SaveJob(ctx, stored); err != nil {
		t.Fatal(err)
	}
	stored, _ = database.Job(ctx, job.ID)
	if !stored.CreatedAt.Equal(created) || stored.Name != "Updated" || !stored.UpdatedAt.After(created) {
		t.Fatalf("timestamps or update incorrect: %#v", stored)
	}
	jobs, err := database.Jobs(ctx)
	if err != nil || len(jobs) != 1 {
		t.Fatalf("jobs=%d err=%v", len(jobs), err)
	}

	destination, err := database.Assign(ctx, job.ID, "Movie", "gd")
	if err != nil || destination != "gd" {
		t.Fatalf("assignment=%q err=%v", destination, err)
	}
	if destination, err = database.Assign(ctx, job.ID, "Movie", "other"); err != nil || destination != "gd" {
		t.Fatalf("assignment was not pinned: %q %v", destination, err)
	}
	if assigned, err := database.HasAssignments(ctx, job.ID); err != nil || !assigned {
		t.Fatalf("stored assignment was not detected: %v %v", assigned, err)
	}

	run := model.Run{ID: "run_dry", JobID: job.ID, Unit: "Movie", Destination: "gd", State: "discovered", StartedAt: time.Now().UTC()}
	if err = database.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err = database.Transition(ctx, run.ID, "transferring", ""); err != nil {
		t.Fatal(err)
	}
	if err = database.Transition(ctx, run.ID, "completed", "dry run"); err != nil {
		t.Fatal(err)
	}
	if err = database.Transition(ctx, run.ID, "transferring", ""); err == nil {
		t.Fatal("invalid transition was accepted")
	}
	runs, err := database.Runs(ctx, 10)
	if err != nil || len(runs) != 1 || runs[0].State != "completed" || runs[0].FinishedAt == nil {
		t.Fatalf("runs=%#v err=%v", runs, err)
	}
	stats := database.Stats(ctx)
	if stats["jobs"] != 1 || stats["completed"] != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	analysis := model.Analysis{JobID: job.ID, State: "completed", Summary: map[string]int{model.ArchivePending: 1}, StartedAt: time.Now().UTC()}
	if err = database.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatal(err)
	}
	storedAnalysis, err := database.Analysis(ctx, job.ID)
	if err != nil || storedAnalysis.Summary[model.ArchivePending] != 1 {
		t.Fatalf("analysis not stored: %#v %v", storedAnalysis, err)
	}
	analyses, err := database.Analyses(ctx)
	if err != nil || len(analyses) != 1 {
		t.Fatalf("analyses=%#v err=%v", analyses, err)
	}
	if err = database.DeleteAnalysis(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Analysis(ctx, job.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("explicit analysis delete failed: %v", err)
	}
	if err = database.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatal(err)
	}
	stored.Name = "Updated again"
	if err = database.SaveJob(ctx, stored); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Analysis(ctx, job.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("job update left stale analysis: %v", err)
	}
	if err = database.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatal(err)
	}
	if err = database.DeleteJob(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Assignment(ctx, job.ID, "Movie"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("assignment survived deletion: %v", err)
	}
	if _, err = database.Analysis(ctx, job.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("analysis survived deletion: %v", err)
	}
	if err = database.DeleteJob(ctx, job.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing delete should return no rows: %v", err)
	}
}

func TestFailInterruptedRuns(t *testing.T) {
	ctx := context.Background()
	database := openTestStore(t)
	run := model.Run{ID: "run_interrupted", JobID: "job", Unit: "Show", Destination: "gd", State: "discovered", StartedAt: time.Now().UTC()}
	if err := database.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err := database.Transition(ctx, run.ID, "transferring", ""); err != nil {
		t.Fatal(err)
	}
	count, err := database.FailInterruptedRuns(ctx, "recovered")
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	runs, _ := database.Runs(ctx, 10)
	if runs[0].State != "failed" || runs[0].Message != "recovered" || runs[0].FinishedAt == nil {
		t.Fatalf("run not reconciled: %#v", runs[0])
	}
}

func TestFailInterruptedAnalyses(t *testing.T) {
	ctx := context.Background()
	database := openTestStore(t)
	running := model.Analysis{JobID: "running", State: "running", Summary: map[string]int{}, StartedAt: time.Now().UTC()}
	completed := model.Analysis{JobID: "completed", State: "completed", Summary: map[string]int{}, StartedAt: time.Now().UTC()}
	if err := database.SaveAnalysis(ctx, running); err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAnalysis(ctx, completed); err != nil {
		t.Fatal(err)
	}
	count, err := database.FailInterruptedAnalyses(ctx, "recovered")
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	analysis, err := database.Analysis(ctx, running.JobID)
	if err != nil || analysis.State != "failed" || analysis.Message != "recovered" || analysis.FinishedAt == nil {
		t.Fatalf("analysis not reconciled: %#v err=%v", analysis, err)
	}
	analysis, err = database.Analysis(ctx, completed.JobID)
	if err != nil || analysis.State != "completed" {
		t.Fatalf("completed analysis changed: %#v err=%v", analysis, err)
	}
}
