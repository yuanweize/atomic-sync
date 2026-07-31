package engine

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yuanweize/atomic-sync/internal/model"
	"github.com/yuanweize/atomic-sync/internal/store"
)

type fakeExecutor struct {
	mu    sync.Mutex
	calls [][]string
	fn    func(context.Context, []string) ([]byte, error)
}

func (fake *fakeExecutor) execute(ctx context.Context, args ...string) ([]byte, error) {
	fake.mu.Lock()
	fake.calls = append(fake.calls, append([]string(nil), args...))
	fake.mu.Unlock()
	return fake.fn(ctx, args)
}

func (fake *fakeExecutor) snapshot() [][]string {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	result := make([][]string, len(fake.calls))
	for i := range fake.calls {
		result[i] = append([]string(nil), fake.calls[i]...)
	}
	return result
}

func runnerStore(t *testing.T) *store.Store {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "runner.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func runnerJob() model.Job {
	return model.Job{
		ID: "job_runner", Name: "Archive", Source: "/sources/media",
		Destinations: []model.Destination{{Name: "gd", Path: "GD:data/media", Weight: 1}},
		Mode:         "copy", Grouping: "folder", SettleSeconds: 0, Concurrency: 2,
		Verify: "checksum", ConflictPolicy: model.ConflictFail,
	}
}

func listing() []byte {
	modified := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339Nano)
	return []byte(`[{"Path":"Movie/file.mkv","ModTime":"` + modified + `","IsDir":false}]`)
}

func TestRunnerDryRunDoesNotWrite(t *testing.T) {
	database := runnerStore(t)
	if err := database.SaveAnalysis(context.Background(), model.Analysis{JobID: "job_runner", State: "completed", Summary: map[string]int{}}); err != nil {
		t.Fatal(err)
	}
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] != "lsjson" {
			t.Fatalf("dry run executed %q", args[0])
		}
		return listing(), nil
	}}
	runner := New(database, "rclone", 2)
	runner.execute = fake.execute
	job := runnerJob()
	job.DryRun = true
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	runs, err := database.Runs(context.Background(), 10)
	if err != nil || len(runs) != 1 || runs[0].State != "completed" || !strings.Contains(runs[0].Message, "no changes") {
		t.Fatalf("unexpected dry run: %#v err=%v", runs, err)
	}
	if len(fake.snapshot()) != 1 {
		t.Fatalf("unexpected commands: %#v", fake.snapshot())
	}
	if _, err = database.Analysis(context.Background(), job.ID); err != nil {
		t.Fatalf("dry run invalidated branch analysis: %v", err)
	}
}

func TestRunnerCopyPublishesOnlyToEmptyDestination(t *testing.T) {
	database := runnerStore(t)
	if err := database.SaveAnalysis(context.Background(), model.Analysis{JobID: "job_runner", State: "completed", Summary: map[string]int{}}); err != nil {
		t.Fatal(err)
	}
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listing(), nil
		case "lsf":
			return []byte("directory not found"), errors.New("exit status 3")
		default:
			return nil, nil
		}
	}}
	runner := New(database, "rclone", 2)
	runner.execute = fake.execute
	job := runnerJob()
	job.Include = []string{"*.mkv"}
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "copy") != 1 || countCommand(calls, "check") != 2 || countCommand(calls, "moveto") != 1 {
		t.Fatalf("unexpected publish sequence: %#v", calls)
	}
	if countCommand(calls, "purge") != 0 {
		t.Fatalf("copy mode purged data: %#v", calls)
	}
	for _, call := range calls {
		if (call[0] == "copy" || call[0] == "check") && !containsPair(call, "--include", "*.mkv") {
			t.Fatalf("filters missing from %q: %#v", call[0], call)
		}
	}
	if _, err := database.Analysis(context.Background(), job.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("write run left stale branch analysis: %v", err)
	}
}

func TestRunnerConflictFailsClosedAndPreservesSource(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listing(), nil
		}
		if args[0] == "lsf" {
			return []byte("existing.mkv\n"), nil
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	err := runner.Run(context.Background(), runnerJob())
	if err == nil || !strings.Contains(err.Error(), "destination already exists") {
		t.Fatalf("expected destination conflict, got %v", err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "moveto") != 0 || countCommand(calls, "purge") != 0 {
		t.Fatalf("conflict changed data: %#v", calls)
	}
	runs, _ := database.Runs(context.Background(), 10)
	if len(runs) != 1 || runs[0].State != "failed" {
		t.Fatalf("conflict was not audited: %#v", runs)
	}
}

func TestRunnerImmutableMergeVerifiesBeforeSourcePurge(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listing(), nil
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = "move"
	job.ConflictPolicy = model.ConflictMergeImmutable
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "copy") != 2 || countCommand(calls, "check") != 3 || countCommand(calls, "purge") != 2 {
		t.Fatalf("unexpected immutable merge sequence: %#v", calls)
	}
	var immutable bool
	var stagePurge, sourcePurge int
	for index, call := range calls {
		if call[0] == "copy" && contains(call, "--immutable") {
			immutable = true
		}
		if call[0] == "purge" && strings.Contains(call[1], ".atomic-sync-staging") {
			stagePurge = index
		}
		if call[0] == "purge" && call[1] == "/sources/media/Movie" {
			sourcePurge = index
		}
	}
	if !immutable || sourcePurge <= stagePurge || sourcePurge == 0 {
		t.Fatalf("source purge was not last and immutable: %#v", calls)
	}
}

func TestRunnerStartIsExclusiveAndShutdownCancels(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(ctx context.Context, _ []string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	if err := runner.Start(job); err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(job); !errors.Is(err, ErrJobActive) {
		t.Fatalf("duplicate start returned %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if runner.IsActive(job.ID) {
		t.Fatal("job remained active after shutdown")
	}
	job.Paused = true
	if err := runner.Run(context.Background(), job); !errors.Is(err, ErrJobPaused) {
		t.Fatalf("paused job returned %v", err)
	}
}

func TestArchiveAnalysisLooksInsideOverlappingFolders(t *testing.T) {
	database := runnerStore(t)
	sourceListing := []byte(`[
      {"Path":"Pending/movie.mkv","Size":100,"IsDir":false},
      {"Path":"Partial/a.mkv","Size":100,"IsDir":false},
      {"Path":"Partial/b.mkv","Size":200,"IsDir":false},
      {"Path":"Ready/a.mkv","Size":100,"IsDir":false},
      {"Path":"Conflict/a.mkv","Size":100,"IsDir":false},
	  {"Path":"ArchivedShell","Size":-1,"IsDir":true},
	  {"Path":"TypeCollision","Size":100,"IsDir":false},
      {"Path":"Empty","Size":-1,"IsDir":true}
    ]`)
	destinationListing := []byte(`[
      {"Path":"Partial/a.mkv","Size":100,"IsDir":false},
      {"Path":"Ready/a.mkv","Size":100,"IsDir":false},
      {"Path":"Ready/destination-extra.nfo","Size":50,"IsDir":false},
      {"Path":"Conflict/a.mkv","Size":999,"IsDir":false},
	  {"Path":"ArchivedShell","Size":-1,"IsDir":true},
	  {"Path":"ArchivedShell/movie.mkv","Size":300,"IsDir":false},
	  {"Path":"TypeCollision","Size":-1,"IsDir":true},
      {"Path":"Archived/movie.mkv","Size":300,"IsDir":false}
    ]`)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] != "lsjson" {
			t.Fatalf("analysis executed write command: %#v", args)
		}
		if args[1] == "/sources/media" {
			return sourceListing, nil
		}
		return destinationListing, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	statuses := map[string]model.UnitAnalysis{}
	for _, unit := range units {
		statuses[unit.Unit] = unit
	}
	wants := map[string]string{
		"Archived": model.ArchiveArchived, "ArchivedShell": model.ArchiveArchived,
		"Conflict": model.ArchiveConflict, "TypeCollision": model.ArchiveConflict,
		"Empty": model.ArchiveEmpty, "Partial": model.ArchivePartial,
		"Pending": model.ArchivePending, "Ready": model.ArchiveReadyToVerify,
	}
	for unit, want := range wants {
		if got := statuses[unit].Status; got != want {
			t.Errorf("%s status=%q want=%q; analysis=%#v", unit, got, want, statuses[unit])
		}
	}
	if statuses["Partial"].Coverage != 50 || statuses["Partial"].MissingFiles != 1 {
		t.Fatalf("partial coverage incorrect: %#v", statuses["Partial"])
	}
	if statuses["Conflict"].ConflictingFiles != 1 || len(statuses["Conflict"].ConflictSamples) != 1 {
		t.Fatalf("conflict evidence missing: %#v", statuses["Conflict"])
	}
	if !statuses["ArchivedShell"].SourcePresent || !statuses["ArchivedShell"].DestinationPresent {
		t.Fatalf("empty source shell presence was lost: %#v", statuses["ArchivedShell"])
	}
	if statuses["TypeCollision"].ConflictingFiles != 1 {
		t.Fatalf("file/directory collision was not detected: %#v", statuses["TypeCollision"])
	}
	if countCommand(fake.snapshot(), "lsjson") != 2 {
		t.Fatalf("analysis should list each branch once: %#v", fake.snapshot())
	}
}

func TestArchiveAnalysisFailsClosedWhenSourceScanFails(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[1] == "/sources/media" {
			return []byte("directory not found"), errors.New("exit status 3")
		}
		return []byte(`[{"Path":"Archived/movie.mkv","Size":300,"IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	if _, err := runner.analyze(context.Background(), runnerJob()); err == nil || !strings.Contains(err.Error(), "scan source inventory") {
		t.Fatalf("source outage was treated as an empty branch: %v", err)
	}
}

func TestArchiveAnalysisReportsFullyArchivedEmptySource(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[1] == "/sources/media" {
			return []byte(`[]`), nil
		}
		return []byte(`[{"Path":"Archived/movie.mkv","Size":300,"IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	units, err := runner.analyze(context.Background(), runnerJob())
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Status != model.ArchiveArchived || units[0].SourcePresent || !units[0].DestinationPresent {
		t.Fatalf("fully archived source was misclassified: %#v", units)
	}
}

func TestSeasonAnalysisDetectsShallowFileDirectoryConflict(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[1] == "/sources/media" {
			return []byte(`[{"Path":"Show","Size":100,"IsDir":false}]`), nil
		}
		return []byte(`[{"Path":"Show","Size":-1,"IsDir":true}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Grouping = "season"
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Status != model.ArchiveConflict || units[0].ConflictingFiles != 1 {
		t.Fatalf("shallow season type collision was missed: %#v", units)
	}
}

func TestArchiveAnalysisDoesNotDuplicateSourceUnitAcrossDestinations(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		return []byte(`[{"Path":"Shared/file.mkv","Size":300,"IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Destinations = append(job.Destinations, model.Destination{Name: "gd-secondary", Path: "GD2:data/media", Weight: 1})
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Unit != "Shared" || units[0].Status != model.ArchiveReadyToVerify || !units[0].SourcePresent {
		t.Fatalf("source unit was duplicated or misclassified across destinations: %#v", units)
	}
}

func TestInventoryRejectsNegativeFileSize(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
		return []byte(`[{"Path":"Movie/file.mkv","Size":-1,"IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	if _, err := runner.inventory(context.Background(), "/sources/media", runnerJob()); err == nil || !strings.Contains(err.Error(), "negative size") {
		t.Fatalf("negative file size was accepted: %v", err)
	}
}

func TestNotFoundDetectionDoesNotHideRemoteFailures(t *testing.T) {
	if !isNotFound([]byte("directory not found"), errors.New("exit status 3")) {
		t.Fatal("precise missing-directory error was not recognized")
	}
	if isNotFound([]byte("failed to find root: Google Drive quota exceeded; file not found"), errors.New("exit status 1")) {
		t.Fatal("remote outage was treated as an absent destination")
	}
}

func TestStartAnalysisPersistsSummary(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[1] == "/sources/media" {
			return []byte(`[{"Path":"Movie/file.mkv","Size":100,"IsDir":false}]`), nil
		}
		return []byte(`[]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	if err := runner.StartAnalysis(job); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		analysis, err := database.Analysis(context.Background(), job.ID)
		if err == nil && analysis.State == "completed" {
			if analysis.Summary[model.ArchivePending] != 1 || len(analysis.Units) != 1 {
				t.Fatalf("unexpected analysis: %#v", analysis)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("analysis did not complete")
}

func countCommand(calls [][]string, command string) int {
	count := 0
	for _, call := range calls {
		if len(call) > 0 && call[0] == command {
			count++
		}
	}
	return count
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func containsPair(values []string, first, second string) bool {
	for index := 0; index+1 < len(values); index++ {
		if values[index] == first && values[index+1] == second {
			return true
		}
	}
	return false
}
