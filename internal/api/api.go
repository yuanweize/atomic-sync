package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yuanweize/atomic-sync/internal/buildinfo"
	"github.com/yuanweize/atomic-sync/internal/engine"
	"github.com/yuanweize/atomic-sync/internal/model"
	"github.com/yuanweize/atomic-sync/internal/store"
)

const maxJSONBody = 1 << 20

const destructiveConfirmationHeader = "X-Atomic-Confirm-Job"
const defaultSettleSeconds = 30 * 24 * 60 * 60

type jobFieldPresence struct {
	dryRun       bool
	settleWindow bool
}

var (
	errPlacementLocked = errors.New("placement settings cannot change after units have been assigned; create a new job")
	errJobOverlap      = errors.New("job paths overlap another configured job")
)

//go:embed ui/*
var ui embed.FS

type API struct {
	store  *store.Store
	runner *engine.Runner
	token  string
	mux    *http.ServeMux
	// jobsMu keeps a persisted job snapshot stable while a run or analysis
	// becomes active, and serializes cross-job overlap checks with mutations.
	jobsMu sync.Mutex
}

func New(s *store.Store, runner *engine.Runner, token string) *API {
	a := &API{store: s, runner: runner, token: token, mux: http.NewServeMux()}
	a.routes()
	return a
}

func (a *API) Handler() http.Handler { return a.securityHeaders(a.auth(a.mux)) }

func (a *API) routes() {
	a.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok", "version": buildinfo.Version, "authRequired": a.token != "",
		})
	})
	a.mux.HandleFunc("GET /api/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := a.store.Ping(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	a.mux.HandleFunc("GET /api/system", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"version": buildinfo.Version, "commit": buildinfo.Commit, "builtAt": buildinfo.Date,
			"activeJobs": a.runner.ActiveCount(),
		})
	})
	a.mux.HandleFunc("GET /api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, a.store.Stats(r.Context()))
	})
	a.mux.HandleFunc("GET /api/jobs", a.jobs)
	a.mux.HandleFunc("POST /api/jobs", a.jobs)
	a.mux.HandleFunc("GET /api/jobs/{id}", a.job)
	a.mux.HandleFunc("PUT /api/jobs/{id}", a.job)
	a.mux.HandleFunc("DELETE /api/jobs/{id}", a.job)
	a.mux.HandleFunc("POST /api/jobs/{id}/run", a.run)
	a.mux.HandleFunc("GET /api/jobs/{id}/analysis", a.analysis)
	a.mux.HandleFunc("POST /api/jobs/{id}/analysis", a.analysis)
	a.mux.HandleFunc("GET /api/analyses", a.analyses)
	a.mux.HandleFunc("GET /api/runs", a.runs)
	a.mux.HandleFunc("GET /api/events", a.events)
	a.mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	sub, _ := fs.Sub(ui, "ui")
	a.mux.Handle("/", spa(http.FS(sub)))
}

func (a *API) jobs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		jobs, err := a.store.Jobs(r.Context())
		respond(w, jobs, err)
		return
	}
	job, provided, err := decodeJob(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	job.ID = newJobID()
	job.Normalize()
	if !provided.dryRun {
		job.DryRun = true
	}
	if !provided.settleWindow {
		job.SettleSeconds = defaultSettleSeconds
	}
	if err = job.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.jobsMu.Lock()
	defer a.jobsMu.Unlock()
	if err = a.validateNoJobOverlap(r.Context(), job); err != nil {
		respond(w, nil, err)
		return
	}
	if err = a.store.SaveJob(r.Context(), job); err != nil {
		respond(w, nil, err)
		return
	}
	saved, err := a.store.Job(r.Context(), job.ID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	writeJSON(w, http.StatusCreated, saved)
}

func (a *API) job(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	switch r.Method {
	case http.MethodDelete:
		a.deleteJob(w, r, id)
	case http.MethodGet:
		job, err := a.store.Job(r.Context(), id)
		respond(w, job, err)
	case http.MethodPut:
		a.updateJob(w, r, id)
	}
}

func (a *API) deleteJob(w http.ResponseWriter, r *http.Request, id string) {
	a.jobsMu.Lock()
	defer a.jobsMu.Unlock()
	if a.runner.IsActive(id) {
		writeError(w, http.StatusConflict, "cannot delete a running job")
		return
	}
	respond(w, map[string]bool{"deleted": true}, a.store.DeleteJob(r.Context(), id))
}

func (a *API) updateJob(w http.ResponseWriter, r *http.Request, id string) {
	job, provided, err := decodeJob(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.jobsMu.Lock()
	defer a.jobsMu.Unlock()
	if a.runner.IsActive(id) {
		writeError(w, http.StatusConflict, "cannot update a running job")
		return
	}
	current, err := a.store.Job(r.Context(), id)
	if err != nil {
		respond(w, nil, err)
		return
	}
	job.ID, job.CreatedAt = id, current.CreatedAt
	job.Normalize()
	if !provided.dryRun {
		job.DryRun = current.DryRun
	}
	if !provided.settleWindow {
		job.SettleSeconds = current.SettleSeconds
	}
	if err = job.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err = a.validatePlacementUpdate(r.Context(), current, job); err != nil {
		respond(w, nil, err)
		return
	}
	if err = a.validateNoJobOverlap(r.Context(), job); err != nil {
		respond(w, nil, err)
		return
	}
	if err = a.store.SaveJob(r.Context(), job); err != nil {
		respond(w, nil, err)
		return
	}
	saved, err := a.store.Job(r.Context(), id)
	respond(w, saved, err)
}

func (a *API) validateNoJobOverlap(ctx context.Context, candidate model.Job) error {
	jobs, err := a.store.Jobs(ctx)
	if err != nil {
		return err
	}
	for _, existing := range jobs {
		if existing.ID == candidate.ID {
			continue
		}
		if candidate.PlacementOverlaps(existing) {
			return fmt.Errorf("%w: %s", errJobOverlap, existing.Name)
		}
	}
	return nil
}

func (a *API) validatePlacementUpdate(ctx context.Context, current, next model.Job) error {
	if current.SamePlacement(next) {
		return nil
	}
	locked, err := a.store.HasAssignments(ctx, current.ID)
	if err != nil {
		return err
	}
	if locked {
		return errPlacementLocked
	}
	return nil
}

func (a *API) run(w http.ResponseWriter, r *http.Request) {
	a.jobsMu.Lock()
	defer a.jobsMu.Unlock()
	job, err := a.store.Job(r.Context(), r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	if job.Mode == model.ModeMove && !job.DryRun && !confirmedJobName(r.Header.Get(destructiveConfirmationHeader), job.Name) {
		writeError(w, http.StatusBadRequest, "move run requires the exact job name in X-Atomic-Confirm-Job")
		return
	}
	if err = a.validateNoJobOverlap(r.Context(), job); err != nil {
		respond(w, nil, err)
		return
	}
	if err = a.runner.Start(job); err != nil {
		respond(w, nil, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func confirmedJobName(provided, expected string) bool {
	return provided != "" && len(provided) == len(expected) && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func (a *API) analyses(w http.ResponseWriter, r *http.Request) {
	analyses, err := a.store.Analyses(r.Context())
	for index := range analyses {
		analyses[index].Units = nil
	}
	respond(w, analyses, err)
}

func (a *API) analysis(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if r.Method == http.MethodGet {
		analysis, err := a.store.Analysis(r.Context(), jobID)
		respond(w, analysis, err)
		return
	}
	a.jobsMu.Lock()
	defer a.jobsMu.Unlock()
	job, err := a.store.Job(r.Context(), jobID)
	if err != nil {
		respond(w, nil, err)
		return
	}
	if err = a.validateNoJobOverlap(r.Context(), job); err != nil {
		respond(w, nil, err)
		return
	}
	if err = a.runner.StartAnalysis(job); err != nil {
		respond(w, nil, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "analysis-started"})
}

func (a *API) runs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}
	runs, err := a.store.Runs(r.Context(), limit)
	respond(w, runs, err)
}

func (a *API) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	// Commit the stream immediately so the browser can distinguish a live
	// authenticated connection from a reconnect before the first heartbeat.
	flusher.Flush()
	ch, closeSubscription := a.runner.Subscribe()
	defer closeSubscription()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case event, open := <-ch:
			if !open {
				return
			}
			payload, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = io.WriteString(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func decodeJob(w http.ResponseWriter, r *http.Request) (model.Job, jobFieldPresence, error) {
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			return model.Job{}, jobFieldPresence{}, errors.New("content type must be application/json")
		}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxJSONBody))
	if err != nil {
		return model.Job{}, jobFieldPresence{}, errors.New("request body is too large or unreadable")
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(body, &fields); err != nil {
		return model.Job{}, jobFieldPresence{}, errors.New("invalid JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var job model.Job
	if err = decoder.Decode(&job); err != nil {
		return model.Job{}, jobFieldPresence{}, fmt.Errorf("invalid job: %w", err)
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return model.Job{}, jobFieldPresence{}, errors.New("invalid JSON: multiple values")
	}
	_, dryRunProvided := fields["dryRun"]
	_, settleWindowProvided := fields["settleSeconds"]
	return job, jobFieldPresence{dryRun: dryRunProvided, settleWindow: settleWindowProvided}, nil
}

func (a *API) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicAPI := r.URL.Path == "/api/health" || r.URL.Path == "/api/ready"
		if a.token == "" || !strings.HasPrefix(r.URL.Path, "/api/") || publicAPI {
			next.ServeHTTP(w, r)
			return
		}
		provided, hasBearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		valid := hasBearer && provided != "" && !strings.ContainsAny(provided, " \t") && len(provided) == len(a.token) && subtle.ConstantTimeCompare([]byte(provided), []byte(a.token)) == 1
		if !valid {
			w.Header().Set("WWW-Authenticate", `Bearer realm="Atomic Sync"`)
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func respond(w http.ResponseWriter, value any, err error) {
	if err == nil {
		writeJSON(w, http.StatusOK, value)
		return
	}
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, sql.ErrNoRows):
		status, message = http.StatusNotFound, "not found"
	case errors.Is(err, model.ErrInvalidJob):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, engine.ErrJobActive), errors.Is(err, engine.ErrJobPaused):
		status, message = http.StatusConflict, err.Error()
	case errors.Is(err, errPlacementLocked), errors.Is(err, errJobOverlap):
		status, message = http.StatusConflict, err.Error()
	case errors.Is(err, context.Canceled):
		status, message = http.StatusServiceUnavailable, "service is shutting down"
	default:
		slog.Error("API request failed", "error", err)
	}
	writeError(w, status, message)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func newJobID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err == nil {
		return "job_" + hex.EncodeToString(value)
	}
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func spa(root http.FileSystem) http.Handler {
	files := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + r.URL.Path)
		asset := strings.TrimPrefix(clean, "/")
		if asset == "" || asset == "." {
			asset = "index.html"
		}
		if file, err := root.Open(asset); err == nil {
			_ = file.Close()
			if asset == "index.html" {
				w.Header().Set("Cache-Control", "no-cache")
			}
			files.ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/"
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	})
}
