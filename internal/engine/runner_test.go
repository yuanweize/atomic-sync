package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
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

func TestRunnerImmutableMergeRetainsSourceAndStaging(t *testing.T) {
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
	job.ConflictPolicy = model.ConflictMergeImmutable
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "copy") != 2 || countCommand(calls, "check") != 3 || countCommand(calls, "purge") != 0 {
		t.Fatalf("unexpected immutable merge sequence: %#v", calls)
	}
	var immutable bool
	checkIndex := 0
	for _, call := range calls {
		if call[0] == "copy" && contains(call, "--immutable") {
			immutable = true
		}
		if call[0] == "check" {
			if !contains(call, "--quiet") {
				t.Fatalf("routine successful check notices were not suppressed: %#v", call)
			}
			if !contains(call, "--download") {
				t.Fatalf("content verification did not use --download: %#v", call)
			}
			if checkIndex == 0 && contains(call, "--one-way") {
				t.Fatalf("fresh staging verification allowed extra files: %#v", call)
			}
			if checkIndex > 0 && !contains(call, "--one-way") {
				t.Fatalf("immutable destination verification rejected permitted extras: %#v", call)
			}
			checkIndex++
		}
	}
	if !immutable {
		t.Fatalf("immutable merge flag missing: %#v", calls)
	}
}

func TestRunnerVerificationFailuresPreserveSource(t *testing.T) {
	tests := []struct {
		name           string
		policy         string
		failCheck      int
		wantMoveTo     int
		wantCopy       int
		wantErrorMatch string
	}{
		{name: "fresh staging", policy: model.ConflictFail, failCheck: 1, wantCopy: 1, wantErrorMatch: "injected verification failure"},
		{name: "new destination final", policy: model.ConflictFail, failCheck: 2, wantMoveTo: 1, wantCopy: 1, wantErrorMatch: "source preserved"},
		{name: "immutable merge", policy: model.ConflictMergeImmutable, failCheck: 2, wantCopy: 2, wantErrorMatch: "immutable merge verification failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database := runnerStore(t)
			checks := 0
			fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
				switch args[0] {
				case "lsjson":
					return listing(), nil
				case "lsf":
					return []byte("directory not found"), errors.New("exit status 3")
				case "check":
					checks++
					if checks == test.failCheck {
						return nil, errors.New("injected verification failure")
					}
				}
				return nil, nil
			}}
			runner := New(database, "rclone", 1)
			runner.execute = fake.execute
			job := runnerJob()
			job.ConflictPolicy = test.policy
			err := runner.Run(context.Background(), job)
			if err == nil || !strings.Contains(err.Error(), test.wantErrorMatch) {
				t.Fatalf("unexpected failure result: %v", err)
			}
			calls := fake.snapshot()
			if countCommand(calls, "copy") != test.wantCopy || countCommand(calls, "moveto") != test.wantMoveTo {
				t.Fatalf("unexpected commands after failure: %#v", calls)
			}
			if countCommand(calls, "purge") != 0 {
				t.Fatalf("verification failure triggered deletion: %#v", calls)
			}
			runs, runsErr := database.Runs(context.Background(), 10)
			if runsErr != nil || len(runs) != 1 || runs[0].State != "failed" {
				t.Fatalf("failed run was not persisted: %#v err=%v", runs, runsErr)
			}
		})
	}
}

func TestRunnerSizeVerificationDoesNotClaimContentCheck(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listing(), nil
		}
		if args[0] == "lsf" {
			return []byte("directory not found"), errors.New("exit status 3")
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Verify = "size"
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	for _, call := range fake.snapshot() {
		if call[0] != "check" {
			continue
		}
		if !contains(call, "--size-only") || contains(call, "--download") {
			t.Fatalf("size verification flags are misleading: %#v", call)
		}
	}
}

func TestRunnerAppliesConservativeRcloneLimits(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) { return nil, nil }}
	runner := NewWithLimits(database, "rclone", 1, 3, 4, 5)
	runner.execute = fake.execute
	if _, err := runner.command(context.Background(), "version"); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if len(calls) != 1 || !containsPair(calls[0], "--transfers", "3") || !containsPair(calls[0], "--checkers", "4") || !containsPair(calls[0], "--tpslimit", "5") || !containsPair(calls[0], "--tpslimit-burst", "1") {
		t.Fatalf("rclone limits missing from argv: %#v", calls)
	}
}

func TestRunnerGlobalLimitGuardsEveryRcloneProcess(t *testing.T) {
	database := runnerStore(t)
	var mu sync.Mutex
	active, maximum := 0, 0
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
		mu.Lock()
		active++
		maximum = max(maximum, active)
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		active--
		mu.Unlock()
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = runner.command(context.Background(), "version")
		}()
	}
	wg.Wait()
	if maximum != 1 {
		t.Fatalf("global process limit was exceeded: max=%d", maximum)
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

func TestRunnerRejectsInvalidPersistedJobAtExecutionBoundary(t *testing.T) {
	database := runnerStore(t)
	runner := New(database, "rclone", 1)
	job := runnerJob()
	job.Mode = "move"

	if err := runner.Run(context.Background(), job); err == nil || !strings.Contains(err.Error(), "mode must be copy") {
		t.Fatalf("synchronous run accepted invalid persisted job: %v", err)
	}
	if err := runner.Start(job); err == nil || !strings.Contains(err.Error(), "mode must be copy") {
		t.Fatalf("asynchronous start accepted invalid persisted job: %v", err)
	}
	if err := runner.StartAnalysis(job); err == nil || !strings.Contains(err.Error(), "mode must be copy") {
		t.Fatalf("analysis accepted invalid persisted job: %v", err)
	}
	if runner.ActiveCount() != 0 {
		t.Fatalf("invalid job remained active: %d", runner.ActiveCount())
	}
}

func TestRunnerRejectsShallowAndOverlappingUnitsBeforeWrites(t *testing.T) {
	database := runnerStore(t)
	modified := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339Nano)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] != "lsjson" {
			t.Fatalf("invalid units reached write command: %#v", args)
		}
		return []byte(`[{"Path":"movie.mkv","ModTime":"` + modified + `","IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	if err := runner.Run(context.Background(), runnerJob()); err == nil || !strings.Contains(err.Error(), "valid folder directory unit") {
		t.Fatalf("shallow file was accepted: %v", err)
	}
	if err := validateIndependentUnits([]string{"Show", "Show/Season 01"}); err == nil || !strings.Contains(err.Error(), "overlapping migration units") {
		t.Fatalf("ancestor and descendant units were accepted: %v", err)
	}
	if err := validateIndependentUnits([]string{".atomic-sync-staging"}); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("reserved staging unit was accepted: %v", err)
	}
}

func TestDiscoverRequiresModTimeForStableWindow(t *testing.T) {
	for _, modTime := range []string{`""`, "null"} {
		t.Run(modTime, func(t *testing.T) {
			database := runnerStore(t)
			fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
				return []byte(`[{"Path":"Movie/file.mkv","ModTime":` + modTime + `,"IsDir":false}]`), nil
			}}
			runner := New(database, "rclone", 1)
			runner.execute = fake.execute
			job := runnerJob()
			job.SettleSeconds = 60
			if _, err := runner.discover(context.Background(), job); err == nil || !strings.Contains(err.Error(), "eligibility cannot be proven") {
				t.Fatalf("unknown modification time was treated as stable: %v", err)
			}
		})
	}
}

func TestDiscoverRejectsMalformedModTime(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
		return []byte(`[{"Path":"Movie/file.mkv","ModTime":"not-a-time","IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute

	if _, err := runner.discover(context.Background(), runnerJob()); err == nil || !strings.Contains(err.Error(), "decode rclone listing") {
		t.Fatalf("malformed modification time was accepted: %v", err)
	}
}

func TestArchiveAnalysisLooksInsideOverlappingFolders(t *testing.T) {
	database := runnerStore(t)
	sourceListing := []byte(`[
      {"Path":"Pending/movie.mkv","ModTime":"","Size":100,"IsDir":false},
	  {"Path":"PendingShell","Size":-1,"IsDir":true},
	  {"Path":"PendingShell/movie.mkv","Size":100,"IsDir":false},
      {"Path":"Partial/a.mkv","Size":100,"IsDir":false},
      {"Path":"Partial/b.mkv","Size":200,"IsDir":false},
	  {"Path":"Complementary/source.mkv","Size":125,"IsDir":false},
      {"Path":"Ready/a.mkv","Size":100,"IsDir":false},
      {"Path":"Conflict/a.mkv","Size":100,"IsDir":false},
	  {"Path":"ArchivedShell","Size":-1,"IsDir":true},
	  {"Path":"TypeCollision","Size":100,"IsDir":false},
      {"Path":"Empty","Size":-1,"IsDir":true}
    ]`)
	destinationListing := []byte(`[
	  {"Path":"PendingShell","Size":-1,"IsDir":true},
      {"Path":"Partial/a.mkv","Size":100,"IsDir":false},
	  {"Path":"Complementary/destination.mkv","Size":75,"IsDir":false},
      {"Path":"Ready/a.mkv","Size":100,"IsDir":false},
      {"Path":"Ready/destination-extra.nfo","Size":50,"IsDir":false},
      {"Path":"Conflict/a.mkv","Size":999,"IsDir":false},
	  {"Path":"ArchivedShell","Size":-1,"IsDir":true},
	  {"Path":"ArchivedShell/movie.mkv","Size":300,"IsDir":false},
	  {"Path":"TypeCollision","Size":-1,"IsDir":true},
	  {"Path":"Empty","Size":-1,"IsDir":true},
	  {"Path":"Archived/movie.mkv","Size":300,"IsDir":false},
	  {"Path":".atomic-sync-staging/job/run/Ghost/file.mkv","Size":999,"IsDir":false}
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
		"Complementary": model.ArchivePartial,
		"Conflict":      model.ArchiveConflict, "TypeCollision": model.ArchiveConflict,
		"Empty": model.ArchiveEmpty, "Partial": model.ArchivePartial,
		"Pending": model.ArchivePending, "PendingShell": model.ArchivePending,
		"Ready": model.ArchiveReadyToVerify,
	}
	for unit, want := range wants {
		if got := statuses[unit].Status; got != want {
			t.Errorf("%s status=%q want=%q; analysis=%#v", unit, got, want, statuses[unit])
		}
	}
	if statuses["Partial"].Coverage != 50 || statuses["Partial"].MissingFiles != 1 {
		t.Fatalf("partial coverage incorrect: %#v", statuses["Partial"])
	}
	if statuses["Complementary"].Coverage != 0 || statuses["Complementary"].MissingFiles != 1 || statuses["Complementary"].DestinationOnlyFiles != 1 {
		t.Fatalf("complementary branch contents were not reported as partial: %#v", statuses["Complementary"])
	}
	if statuses["Ready"].DestinationOnlyFiles != 1 || len(statuses["Ready"].DestinationOnlySamples) != 1 {
		t.Fatalf("destination-only evidence missing: %#v", statuses["Ready"])
	}
	if statuses["Conflict"].ConflictingFiles != 1 || len(statuses["Conflict"].ConflictSamples) != 1 {
		t.Fatalf("conflict evidence missing: %#v", statuses["Conflict"])
	}
	if !statuses["ArchivedShell"].SourcePresent || !statuses["ArchivedShell"].DestinationPresent {
		t.Fatalf("empty source shell presence was lost: %#v", statuses["ArchivedShell"])
	}
	if !statuses["Empty"].SourcePresent || !statuses["Empty"].DestinationPresent {
		t.Fatalf("two physical empty shells were not preserved: %#v", statuses["Empty"])
	}
	if !statuses["PendingShell"].DestinationPresent || statuses["PendingShell"].DestinationFiles != 0 {
		t.Fatalf("destination shell should remain pending, not absent or archived: %#v", statuses["PendingShell"])
	}
	if statuses["TypeCollision"].ConflictingFiles != 1 {
		t.Fatalf("file/directory collision was not detected: %#v", statuses["TypeCollision"])
	}
	if countCommand(fake.snapshot(), "lsjson") != 2 {
		t.Fatalf("analysis should list each branch once: %#v", fake.snapshot())
	}
	if _, exists := statuses[".atomic-sync-staging"]; exists {
		t.Fatalf("private staging namespace leaked into archive results: %#v", statuses[".atomic-sync-staging"])
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

func TestArchiveAnalysisRejectsMissingDestination(t *testing.T) {
	database := runnerStore(t)
	runner := New(database, "rclone", 1)
	job := runnerJob()
	job.Destinations = nil
	if _, err := runner.analyze(context.Background(), job); err == nil || !strings.Contains(err.Error(), "at least one destination") {
		t.Fatalf("invalid analysis panicked or proceeded without a destination: %v", err)
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

func TestSeasonAnalysisSuppressesCrossBranchEmptyAncestor(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[1] == "/sources/media" {
			return []byte(`[{"Path":"A","Size":-1,"IsDir":true}]`), nil
		}
		return []byte(`[
		  {"Path":"A (2020)/Season 01/E01.mkv","Size":200,"IsDir":false},
		  {"Path":"A/Season 01/E01.mkv","Size":300,"IsDir":false}
		]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Grouping = "season"
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].Unit != "A (2020)/Season 01" || units[1].Unit != "A/Season 01" || units[1].Status != model.ArchiveArchived {
		t.Fatalf("empty ancestor duplicated cross-branch season state: %#v", units)
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

func TestArchiveAnalysisFlagsCopiesOutsideAssignedDestination(t *testing.T) {
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
	if len(units) != 1 || units[0].Unit != "Shared" || units[0].Status != model.ArchiveConflict || units[0].UnexpectedDestinationFiles != 1 {
		t.Fatalf("copy outside the assigned destination was hidden or duplicated: %#v", units)
	}
}

func TestArchiveAnalysisIgnoresEmptyShellOnSecondaryDestination(t *testing.T) {
	database := runnerStore(t)
	if _, err := database.Assign(context.Background(), "job_runner", "Shared", "gd"); err != nil {
		t.Fatal(err)
	}
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[1] {
		case "/sources/media", "GD:data/media":
			return []byte(`[{"Path":"Shared/file.mkv","Size":300,"IsDir":false}]`), nil
		default:
			return []byte(`[{"Path":"Shared","Size":-1,"IsDir":true}]`), nil
		}
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Destinations = append(job.Destinations, model.Destination{Name: "gd-secondary", Path: "GD2:data/media", Weight: 1})
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Status != model.ArchiveReadyToVerify || units[0].UnexpectedDestinationFiles != 0 {
		t.Fatalf("empty mergerfs-style shell became a duplicate content conflict: %#v", units)
	}
	if len(units[0].UnexpectedDestinations) != 1 || units[0].UnexpectedDestinations[0] != "gd-secondary" {
		t.Fatalf("secondary shell presence was not retained as metadata: %#v", units[0])
	}
}

func TestArchiveAnalysisConsolidatesDestinationOnlyUnits(t *testing.T) {
	tests := []struct {
		name       string
		secondary  []byte
		wantStatus string
		wantOther  int
	}{
		{name: "one physical archive branch", secondary: []byte(`[]`), wantStatus: model.ArchiveArchived},
		{name: "duplicate physical archive branches", secondary: []byte(`[{"Path":"Archived/file.mkv","Size":300,"IsDir":false}]`), wantStatus: model.ArchiveConflict, wantOther: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database := runnerStore(t)
			fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
				switch args[1] {
				case "/sources/media":
					return []byte(`[]`), nil
				case "GD:data/media":
					return []byte(`[{"Path":"Archived/file.mkv","Size":300,"IsDir":false}]`), nil
				default:
					return test.secondary, nil
				}
			}}
			runner := New(database, "rclone", 1)
			runner.execute = fake.execute
			job := runnerJob()
			job.Destinations = append(job.Destinations, model.Destination{Name: "gd-secondary", Path: "GD2:data/media", Weight: 1})
			units, err := runner.analyze(context.Background(), job)
			if err != nil {
				t.Fatal(err)
			}
			if len(units) != 1 || units[0].Status != test.wantStatus || units[0].UnexpectedDestinationFiles != test.wantOther {
				t.Fatalf("destination-only unit was duplicated or misclassified: %#v", units)
			}
		})
	}
}

func TestSeasonAnalysisDetectsNestedSeasonPathMismatchAndEmptyShow(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[1] == "/sources/media" {
			return []byte(`[
			  {"Path":"Show/Season 03/Season 03/E01.mkv","Size":300,"IsDir":false},
			  {"Path":"Empty Show","Size":-1,"IsDir":true}
			]`), nil
		}
		return []byte(`[
		  {"Path":"Show/Season 03/E01.mkv","Size":300,"IsDir":false},
		  {"Path":"Empty Show","Size":-1,"IsDir":true}
		]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Grouping = "season"
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	statuses := make(map[string]model.UnitAnalysis, len(units))
	for _, unit := range units {
		statuses[unit.Unit] = unit
	}
	season := statuses["Show/Season 03"]
	if season.Status != model.ArchivePartial || season.Coverage != 0 || season.MissingFiles != 1 || season.DestinationOnlyFiles != 1 {
		t.Fatalf("nested season mismatch was hidden by equal folder names: %#v", season)
	}
	empty := statuses["Empty Show"]
	if empty.Status != model.ArchiveEmpty || !empty.SourcePresent || !empty.DestinationPresent {
		t.Fatalf("empty top-level show disappeared under season grouping: %#v", empty)
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

func TestInventoryRejectsReservedSourceNamespace(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
		return []byte(`[{"Path":".atomic-sync-staging/secret.mkv","Size":100,"IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	if _, err := runner.inventory(context.Background(), "/sources/media", runnerJob()); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("reserved source namespace was accepted: %v", err)
	}
}

func TestInventoryRejectsAmbiguousDuplicateFile(t *testing.T) {
	database := runnerStore(t)
	for _, secondSize := range []int{100, 200} {
		fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
			return []byte(fmt.Sprintf(`[
			  {"Path":"Movie/file.mkv","Size":100,"IsDir":false},
			  {"Path":"Movie/file.mkv","Size":%d,"IsDir":false}
			]`, secondSize)), nil
		}}
		runner := New(database, "rclone", 1)
		runner.execute = fake.execute
		if _, err := runner.inventory(context.Background(), "/sources/media", runnerJob()); err == nil || !strings.Contains(err.Error(), "ambiguous duplicate path") {
			t.Fatalf("ambiguous duplicate listing with size %d was accepted: %v", secondSize, err)
		}
	}
}

func TestDepthAnalysisKeepsSiblingEmptyUnits(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) {
		return []byte(`[
		  {"Path":"Show","Size":-1,"IsDir":true},
		  {"Path":"Show/Season 01","Size":-1,"IsDir":true},
		  {"Path":"Show/Season 02","Size":-1,"IsDir":true}
		]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Grouping = "depth"
	job.Depth = 3
	units, err := runner.analyze(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].Unit != "Show/Season 01" || units[1].Unit != "Show/Season 02" {
		t.Fatalf("sibling empty directories collapsed or disappeared: %#v", units)
	}
	for _, unit := range units {
		if unit.Status != model.ArchiveEmpty || !unit.SourcePresent || !unit.DestinationPresent {
			t.Fatalf("empty depth unit presence was lost: %#v", unit)
		}
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

func TestExecRcloneSeparatesDiagnosticsFromStructuredOutput(t *testing.T) {
	script := filepath.Join(t.TempDir(), "fake-rclone")
	contents := `#!/bin/sh
if [ "$1" = "lsf" ]; then
  printf 'directory not found\n' >&2
  exit 3
fi
printf '[{"Path":"Movie/file.mkv"}]'
printf 'diagnostic warning\n' >&2
`
	if err := os.WriteFile(script, []byte(contents), 0o700); err != nil {
		t.Fatal(err)
	}
	runner := New(runnerStore(t), script, 1)
	out, err := runner.execRclone(context.Background(), "version")
	if err != nil || string(out) != `[{"Path":"Movie/file.mkv"}]` {
		t.Fatalf("successful stdout was corrupted by stderr: output=%q err=%v", out, err)
	}
	if _, err = runner.execRclone(context.Background(), "lsjson", "/sources/media"); err == nil || !strings.Contains(err.Error(), "inventory is not trusted") {
		t.Fatalf("structured listing diagnostic was ignored: %v", err)
	}
	out, err = runner.execRclone(context.Background(), "lsf", "GD:missing")
	if err == nil || !isNotFound(out, err) {
		t.Fatalf("stderr-only not-found signal was lost: output=%q err=%v", out, err)
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
