package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yuanweize/atomic-sync/internal/engine"
	"github.com/yuanweize/atomic-sync/internal/model"
	"github.com/yuanweize/atomic-sync/internal/store"
)

type apiHarness struct {
	handler http.Handler
	store   *store.Store
	runner  *engine.Runner
}

func newHarness(t *testing.T, token string) apiHarness {
	return newHarnessWithRclone(t, token, "rclone")
}

func newHarnessWithRclone(t *testing.T, token, rclone string) apiHarness {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	runner := engine.New(database, rclone, 1)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = runner.Shutdown(ctx)
		_ = database.Close()
	})
	return apiHarness{handler: New(database, runner, token).Handler(), store: database, runner: runner}
}

func validAPIJob(id string) model.Job {
	return model.Job{
		ID: id, Name: "Archive", Source: "/sources/media",
		Destinations:   []model.Destination{{Name: "gd", Path: "GD:data/media", Weight: 1}},
		Mode:           "copy",
		Grouping:       "folder",
		Concurrency:    1,
		Verify:         "checksum",
		ConflictPolicy: model.ConflictFail,
		DryRun:         true,
	}
}

func perform(handler http.Handler, method, target, token string, body io.Reader) *httptest.ResponseRecorder {
	authorization := ""
	if token != "" {
		authorization = "Bearer " + token
	}
	return performWithAuthorization(handler, method, target, authorization, body)
}

func performWithAuthorization(handler http.Handler, method, target, authorization string, body io.Reader) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, body)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestPublicUIAndHealthRemainAvailableWithToken(t *testing.T) {
	harness := newHarness(t, "correct-horse-battery-staple")
	for _, target := range []string{"/", "/app.js", "/favicon.svg"} {
		response := perform(harness.handler, http.MethodGet, target, "", nil)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d", target, response.Code)
		}
		if response.Header().Get("Content-Security-Policy") == "" {
			t.Fatalf("%s lacks CSP", target)
		}
	}
	health := perform(harness.handler, http.MethodGet, "/api/health", "", nil)
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"authRequired":true`) {
		t.Fatalf("unexpected health response: %d %s", health.Code, health.Body.String())
	}
	ready := perform(harness.handler, http.MethodGet, "/api/ready", "", nil)
	if ready.Code != http.StatusOK {
		t.Fatalf("ready returned %d", ready.Code)
	}
}

func TestProtectedAPIRequiresBearerToken(t *testing.T) {
	harness := newHarness(t, "secret-token")
	for _, token := range []string{"", "wrong-token"} {
		response := perform(harness.handler, http.MethodGet, "/api/jobs", token, nil)
		if response.Code != http.StatusUnauthorized || response.Header().Get("WWW-Authenticate") == "" {
			t.Fatalf("token %q returned %d", token, response.Code)
		}
	}
	for _, authorization := range []string{
		"secret-token",
		"Basic secret-token",
		"bearer secret-token",
		"Bearer",
		"Bearer  secret-token",
		"Bearer secret-token extra",
	} {
		response := performWithAuthorization(harness.handler, http.MethodGet, "/api/jobs", authorization, nil)
		if response.Code != http.StatusUnauthorized || response.Header().Get("WWW-Authenticate") == "" {
			t.Fatalf("authorization %q returned %d", authorization, response.Code)
		}
	}
	response := perform(harness.handler, http.MethodGet, "/api/jobs", "secret-token", nil)
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("authorized response: %d %s", response.Code, response.Body.String())
	}
	notFound := perform(harness.handler, http.MethodGet, "/api/not-a-route", "secret-token", nil)
	if notFound.Code != http.StatusNotFound || !strings.Contains(notFound.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("unknown API route leaked SPA: %d %s", notFound.Code, notFound.Body.String())
	}
}

func TestCreateJobGeneratesSafeIDAndDefaultsToDryRun(t *testing.T) {
	harness := newHarness(t, "secret-token")
	payload := `{
      "id":"x'><script>alert(1)</script>",
      "name":"Archive movies",
      "source":"/sources/media/movies",
      "destinations":[{"name":"gd-primary","path":"GD:data/media/movies","weight":1}],
      "mode":"copy","grouping":"folder","concurrency":2,"verify":"checksum","conflictPolicy":"fail"
    }`
	response := perform(harness.handler, http.MethodPost, "/api/jobs", "secret-token", strings.NewReader(payload))
	if response.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	var job model.Job
	if err := json.Unmarshal(response.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(job.ID, "job_") || strings.Contains(job.ID, "script") || !job.DryRun {
		t.Fatalf("unsafe create result: %#v", job)
	}
	stored, err := harness.store.Job(context.Background(), job.ID)
	if err != nil || stored.ID != job.ID || stored.CreatedAt.IsZero() {
		t.Fatalf("job not persisted: %#v %v", stored, err)
	}
}

func TestCreateRejectsUnknownFieldsAndMultipleDocuments(t *testing.T) {
	harness := newHarness(t, "token")
	base := `{"name":"Archive","source":"/sources/media","destinations":[{"name":"gd","path":"GD:data/media","weight":1}],"mode":"copy","grouping":"folder","concurrency":1,"verify":"checksum","conflictPolicy":"fail"}`
	var decoded map[string]any
	if err := json.Unmarshal([]byte(base), &decoded); err != nil {
		t.Fatal(err)
	}
	decoded["unexpected"] = true
	unknown, _ := json.Marshal(decoded)
	response := perform(harness.handler, http.MethodPost, "/api/jobs", "token", bytes.NewReader(unknown))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unknown field returned %d", response.Code)
	}
	response = perform(harness.handler, http.MethodPost, "/api/jobs", "token", strings.NewReader(base+base))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("multiple JSON documents returned %d", response.Code)
	}
}

func TestCreateAndUpdateRejectCrossJobPathOverlap(t *testing.T) {
	harness := newHarness(t, "token")
	existing := validAPIJob("job_movies")
	existing.Name = "Movies"
	existing.Source = "/sources/media/movies"
	existing.Destinations[0].Path = "GD:data/media/movies"
	if err := harness.store.SaveJob(context.Background(), existing); err != nil {
		t.Fatal(err)
	}

	nested := validAPIJob("")
	nested.Name = "Nested"
	nested.Source = "/sources/media/movies/new"
	nested.Destinations[0].Path = "GD:data/media/new"
	payload, _ := json.Marshal(nested)
	response := perform(harness.handler, http.MethodPost, "/api/jobs", "token", bytes.NewReader(payload))
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "overlap") {
		t.Fatalf("overlapping create returned %d: %s", response.Code, response.Body.String())
	}

	tv := validAPIJob("job_tv")
	tv.Name = "TV"
	tv.Source = "/sources/media/tvseries"
	tv.Destinations[0].Path = "GD:data/media/tvseries"
	if err := harness.store.SaveJob(context.Background(), tv); err != nil {
		t.Fatal(err)
	}
	tv.Destinations[0].Path = "GD:data/media/movies/archive"
	payload, _ = json.Marshal(tv)
	response = perform(harness.handler, http.MethodPut, "/api/jobs/"+tv.ID, "token", bytes.NewReader(payload))
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "overlap") {
		t.Fatalf("overlapping update returned %d: %s", response.Code, response.Body.String())
	}

	legacy := validAPIJob("job_legacy")
	legacy.Name = "Legacy overlap"
	legacy.Source = "/sources/media/movies/legacy"
	legacy.Destinations[0].Path = "GD:data/media/legacy"
	if err := harness.store.SaveJob(context.Background(), legacy); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"/api/jobs/job_legacy/run", "/api/jobs/job_legacy/analysis"} {
		response = perform(harness.handler, http.MethodPost, target, "token", nil)
		if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "overlap") {
			t.Fatalf("legacy overlapping job %s returned %d: %s", target, response.Code, response.Body.String())
		}
	}
}

func TestPausedJobCannotRun(t *testing.T) {
	harness := newHarness(t, "token")
	job := model.Job{
		ID: "job_paused", Name: "Paused", Source: "/sources/media",
		Destinations: []model.Destination{{Name: "gd", Path: "GD:data/media", Weight: 1}},
		Mode:         "copy", Grouping: "folder", Concurrency: 1, Verify: "checksum",
		ConflictPolicy: model.ConflictFail, DryRun: true, Paused: true,
	}
	if err := harness.store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	response := perform(harness.handler, http.MethodPost, "/api/jobs/job_paused/run", "token", nil)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "paused") {
		t.Fatalf("paused run returned %d %s", response.Code, response.Body.String())
	}
}

func TestLegacyDestructiveJobReturnsBadRequest(t *testing.T) {
	harness := newHarness(t, "token")
	job := validAPIJob("job_legacy_move")
	job.Mode = "move"
	if err := harness.store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"/api/jobs/job_legacy_move/run", "/api/jobs/job_legacy_move/analysis"} {
		response := perform(harness.handler, http.MethodPost, target, "token", nil)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "mode must be copy") {
			t.Fatalf("legacy destructive job %s returned %d: %s", target, response.Code, response.Body.String())
		}
	}
}

func TestAnalysisRoutesReturnPersistedBranchStatus(t *testing.T) {
	harness := newHarness(t, "token")
	analysis := model.Analysis{
		JobID: "job_archive", State: "completed", StartedAt: time.Now().UTC(),
		Summary: map[string]int{model.ArchivePartial: 1},
		Units:   []model.UnitAnalysis{{Unit: "Show", Destination: "gd", Status: model.ArchivePartial, SourceFiles: 2, MatchingFiles: 1, MissingFiles: 1, Coverage: 50}},
	}
	if err := harness.store.SaveAnalysis(context.Background(), analysis); err != nil {
		t.Fatal(err)
	}
	response := perform(harness.handler, http.MethodGet, "/api/analyses", "token", nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"partial":1`) || strings.Contains(response.Body.String(), `"units"`) {
		t.Fatalf("analyses returned %d %s", response.Code, response.Body.String())
	}
	response = perform(harness.handler, http.MethodGet, "/api/jobs/job_archive/analysis", "token", nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"coverage":50`) {
		t.Fatalf("analysis detail returned %d %s", response.Code, response.Body.String())
	}
}

func TestJobUpdateInvalidatesAnalysisAndLocksAssignedPlacement(t *testing.T) {
	harness := newHarness(t, "token")
	job := validAPIJob("job_update")
	if err := harness.store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	analysis := model.Analysis{JobID: job.ID, State: "completed", Summary: map[string]int{}, StartedAt: time.Now().UTC()}
	if err := harness.store.SaveAnalysis(context.Background(), analysis); err != nil {
		t.Fatal(err)
	}
	job.Name = "Renamed"
	payload, _ := json.Marshal(job)
	response := perform(harness.handler, http.MethodPut, "/api/jobs/"+job.ID, "token", bytes.NewReader(payload))
	if response.Code != http.StatusOK {
		t.Fatalf("safe update returned %d: %s", response.Code, response.Body.String())
	}
	if _, err := harness.store.Analysis(context.Background(), job.ID); err == nil {
		t.Fatal("job update left stale analysis")
	}
	if _, err := harness.store.Assign(context.Background(), job.ID, "Movie", "gd"); err != nil {
		t.Fatal(err)
	}
	job.Source = "/sources/other"
	payload, _ = json.Marshal(job)
	response = perform(harness.handler, http.MethodPut, "/api/jobs/"+job.ID, "token", bytes.NewReader(payload))
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "placement") {
		t.Fatalf("assigned placement change returned %d: %s", response.Code, response.Body.String())
	}
}

func TestRunningJobCannotBeUpdated(t *testing.T) {
	script := filepath.Join(t.TempDir(), "blocking-rclone")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexec sleep 30\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	harness := newHarnessWithRclone(t, "token", script)
	job := validAPIJob("job_active")
	if err := harness.store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if err := harness.runner.Start(job); err != nil {
		t.Fatal(err)
	}
	job.Name = "Changed during run"
	payload, _ := json.Marshal(job)
	response := perform(harness.handler, http.MethodPut, "/api/jobs/"+job.ID, "token", bytes.NewReader(payload))
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "running") {
		t.Fatalf("running update returned %d: %s", response.Code, response.Body.String())
	}
	response = perform(harness.handler, http.MethodDelete, "/api/jobs/"+job.ID, "token", nil)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "running") {
		t.Fatalf("running delete returned %d: %s", response.Code, response.Body.String())
	}
}
