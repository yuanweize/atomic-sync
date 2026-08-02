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
	return []byte(`[
	  {"Path":"Movie","ModTime":"2026-01-01T00:00:00Z","Size":-1,"IsDir":true},
	  {"Path":"Movie/Extras","ModTime":"2026-01-01T00:00:00Z","Size":-1,"IsDir":true},
	  {"Path":"Movie/file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":100,"IsDir":false}
	]`)
}

func listingFor(args []string) []byte {
	if len(args) > 1 && args[0] == "lsjson" && (args[1] == "/sources/media/Movie" || args[1] == "GD:data/media/Movie") {
		return []byte(`[
		  {"Path":"Extras","ModTime":"2026-01-01T00:00:00Z","Size":-1,"IsDir":true},
		  {"Path":"file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":100,"IsDir":false}
		]`)
	}
	return listing()
}

func TestRunnerDryRunDoesNotWrite(t *testing.T) {
	database := runnerStore(t)
	if err := database.SaveAnalysis(context.Background(), model.Analysis{JobID: "job_runner", State: "completed", Summary: map[string]int{}}); err != nil {
		t.Fatal(err)
	}
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
		case "lsf":
			return []byte("directory not found"), errors.New("exit status 3")
		case "copy":
			if !contains(args, "--dry-run") {
				t.Fatalf("copy dry run omitted --dry-run: %#v", args)
			}
			return nil, nil
		default:
			t.Fatalf("unexpected dry-run command %q", args[0])
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 2)
	runner.execute = fake.execute
	job := runnerJob()
	job.DryRun = true
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	runs, err := database.Runs(context.Background(), 10)
	if err != nil || len(runs) != 1 || runs[0].State != "completed" || !strings.Contains(runs[0].Message, "no source or destination files changed") {
		t.Fatalf("unexpected dry run: %#v err=%v", runs, err)
	}
	if calls := fake.snapshot(); countCommand(calls, "copy") != 1 || countCommand(calls, "lsf") != 1 || countCommand(calls, "lsjson") != 2 {
		t.Fatalf("dry run did not exercise the real rclone plan: %#v", calls)
	}
	if _, err = database.Analysis(context.Background(), job.ID); err != nil {
		t.Fatalf("dry run invalidated branch analysis: %v", err)
	}
}

func TestRunnerMoveDryRunDoesNotWrite(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
		case "lsf":
			return []byte("directory not found"), errors.New("exit status 3")
		case "move":
			if !contains(args, "--dry-run") || !contains(args, "--size-only") {
				t.Fatalf("move dry run omitted required flags: %#v", args)
			}
			return nil, nil
		default:
			t.Fatalf("unexpected move dry-run command %q", args[0])
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.Verify = "size"
	job.DryRun = true
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if calls := fake.snapshot(); countCommand(calls, "move") != 1 || countCommand(calls, "lsf") != 1 || countCommand(calls, "lsjson") != 2 {
		t.Fatalf("move dry run did not exercise the real rclone plan: %#v", calls)
	}
}

func TestRunnerCopyPreflightsEmptyDestinationAndUsesRcloneCopy(t *testing.T) {
	database := runnerStore(t)
	if err := database.SaveAnalysis(context.Background(), model.Analysis{JobID: "job_runner", State: "completed", Summary: map[string]int{}}); err != nil {
		t.Fatal(err)
	}
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
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
	if countCommand(calls, "lsf") != 1 || countCommand(calls, "copy") != 1 || countCommand(calls, "check") != 0 || countCommand(calls, "move") != 0 {
		t.Fatalf("unexpected copy sequence: %#v", calls)
	}
	for _, call := range calls {
		if call[0] != "copy" {
			continue
		}
		if len(call) < 3 || call[1] != "/sources/media/Movie" || call[2] != "GD:data/media/Movie" {
			t.Fatalf("copy did not target the assigned final destination: %#v", call)
		}
		for _, flag := range []string{"--immutable", "--checksum", "--create-empty-src-dirs"} {
			if !contains(call, flag) {
				t.Fatalf("copy omitted %s: %#v", flag, call)
			}
		}
	}
	if _, err := database.Analysis(context.Background(), job.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("write run left stale branch analysis: %v", err)
	}
}

func TestRunnerMoveUsesRcloneMoveAndDeletesSource(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
		case "move":
			for _, flag := range []string{"--immutable", "--checksum", "--ignore-existing", "--delete-empty-src-dirs"} {
				if !contains(args, flag) {
					t.Fatalf("rclone move omitted %s: %#v", flag, args)
				}
			}
			if args[1] != "/sources/media/Movie" || args[2] != "GD:data/media/Movie" {
				t.Fatalf("rclone move was not configured for verified resumable deletion: %#v", args)
			}
			if contains(args, "--check-first") {
				t.Fatalf("rclone move must stream checks and transfers for throughput: %#v", args)
			}
			return nil, nil
		case "lsf":
			return []byte("directory not found"), errors.New("exit status 3")
		default:
			t.Fatalf("unexpected command: %#v", args)
			return nil, nil
		}
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.ConflictPolicy = model.ConflictMergeImmutable
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "move") != 1 || countCommand(calls, "copy") != 0 || countCommand(calls, "check") != 0 || countCommand(calls, "lsf") != 1 {
		t.Fatalf("unexpected move sequence: %#v", calls)
	}
	runs, err := database.Runs(context.Background(), 10)
	if err != nil || len(runs) != 1 || runs[0].State != "completed" || !strings.Contains(runs[0].Message, "source files removed by rclone") {
		t.Fatalf("move completion was not audited: %#v err=%v", runs, err)
	}
}

func TestRunnerInterruptedMoveCanResumeWithImmutableMerge(t *testing.T) {
	database := runnerStore(t)
	moveAttempts := 0
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
		case "move":
			moveAttempts++
			for _, flag := range []string{"--immutable", "--checksum", "--ignore-existing", "--delete-empty-src-dirs"} {
				if !contains(args, flag) {
					t.Fatalf("resumable move omitted %s: %#v", flag, args)
				}
			}
			if moveAttempts == 1 {
				return nil, errors.New("interrupted after partial transfer")
			}
			return nil, nil
		case "lsf":
			return []byte("directory not found"), errors.New("exit status 3")
		default:
			t.Fatalf("unexpected command: %#v", args)
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.ConflictPolicy = model.ConflictMergeImmutable
	err := runner.Run(context.Background(), job)
	if err == nil || !strings.Contains(err.Error(), "source and destination may each contain part") {
		t.Fatalf("move failure did not explain partial-state recovery: %v", err)
	}
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatalf("immutable move did not resume after interruption: %v", err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "move") != 2 || countCommand(calls, "copy") != 0 || countCommand(calls, "check") != 0 || countCommand(calls, "lsf") != 1 {
		t.Fatalf("interrupted move did not retry through the same merge path: %#v", calls)
	}
	runs, runsErr := database.Runs(context.Background(), 10)
	if runsErr != nil || len(runs) != 2 {
		t.Fatalf("move attempts were not persisted: %#v err=%v", runs, runsErr)
	}
	states := map[string]int{}
	for _, run := range runs {
		states[run.State]++
	}
	if states["failed"] != 1 || states["completed"] != 1 {
		t.Fatalf("interrupted and resumed move states were not audited: %#v", runs)
	}
}

func TestRunnerConflictFailsClosedAndPreservesSource(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listingFor(args), nil
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
	if countCommand(calls, "lsf") != 1 || countCommand(calls, "copy") != 0 || countCommand(calls, "move") != 0 || countCommand(calls, "check") != 0 {
		t.Fatalf("conflict changed data or skipped the preflight: %#v", calls)
	}
	runs, _ := database.Runs(context.Background(), 10)
	if len(runs) != 1 || runs[0].State != "failed" {
		t.Fatalf("conflict was not audited: %#v", runs)
	}
}

func TestRunnerImmutableMergeCopiesDirectlyAndPreservesSource(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listingFor(args), nil
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
	if countCommand(calls, "copy") != 1 || countCommand(calls, "check") != 0 || countCommand(calls, "move") != 0 || countCommand(calls, "lsf") != 0 {
		t.Fatalf("unexpected immutable merge sequence: %#v", calls)
	}
	var immutable bool
	for _, call := range calls {
		if call[0] == "copy" && contains(call, "--immutable") {
			immutable = true
			if call[1] != "/sources/media/Movie" || call[2] != "GD:data/media/Movie" {
				t.Fatalf("immutable merge did not copy directly to the final destination: %#v", call)
			}
		}
		if call[0] == "copy" {
			for _, flag := range []string{"--immutable", "--checksum", "--create-empty-src-dirs"} {
				if !contains(call, flag) {
					t.Fatalf("immutable copy omitted %s: %#v", flag, call)
				}
			}
		}
	}
	if !immutable {
		t.Fatalf("immutable merge flag missing: %#v", calls)
	}
}

func TestRunnerTransferFailuresAreAudited(t *testing.T) {
	tests := []struct {
		name           string
		mode           string
		policy         string
		failCommand    string
		wantCopy       int
		wantMove       int
		wantErrorMatch string
	}{
		{name: "copy transfer", mode: model.ModeCopy, policy: model.ConflictFail, failCommand: "copy", wantCopy: 1, wantErrorMatch: "rclone copy failed; source preserved"},
		{name: "immutable copy transfer", mode: model.ModeCopy, policy: model.ConflictMergeImmutable, failCommand: "copy", wantCopy: 1, wantErrorMatch: "rclone copy failed; source preserved"},
		{name: "move transfer", mode: model.ModeMove, policy: model.ConflictMergeImmutable, failCommand: "move", wantMove: 1, wantErrorMatch: "rclone move failed; source and destination may each contain part"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database := runnerStore(t)
			fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
				switch args[0] {
				case "lsjson":
					return listingFor(args), nil
				case "lsf":
					return []byte("directory not found"), errors.New("exit status 3")
				case "copy", "move":
					if args[0] == test.failCommand {
						return nil, errors.New("injected transfer failure")
					}
				}
				return nil, nil
			}}
			runner := New(database, "rclone", 1)
			runner.execute = fake.execute
			job := runnerJob()
			job.Mode = test.mode
			job.DeleteSource = test.mode == model.ModeMove
			job.ConflictPolicy = test.policy
			err := runner.Run(context.Background(), job)
			if err == nil || !strings.Contains(err.Error(), test.wantErrorMatch) {
				t.Fatalf("unexpected failure result: %v", err)
			}
			calls := fake.snapshot()
			if countCommand(calls, "copy") != test.wantCopy || countCommand(calls, "move") != test.wantMove || countCommand(calls, "check") != 0 {
				t.Fatalf("unexpected commands after failure: %#v", calls)
			}
			runs, runsErr := database.Runs(context.Background(), 10)
			if runsErr != nil || len(runs) != 1 || runs[0].State != "failed" {
				t.Fatalf("failed run was not persisted: %#v err=%v", runs, runsErr)
			}
		})
	}
}

func TestRunnerMoveFailPolicyUsesRcloneSizeOnly(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listingFor(args), nil
		}
		if args[0] == "lsf" {
			return []byte("directory not found"), errors.New("exit status 3")
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.Verify = "size"
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "move") != 1 || countCommand(calls, "lsf") != 2 {
		t.Fatalf("fail-closed size verification did not preflight and move once: %#v", calls)
	}
	for _, call := range calls {
		if call[0] == "move" && (!contains(call, "--size-only") || !contains(call, "--ignore-existing") || contains(call, "--checksum")) {
			t.Fatalf("size verification flags are misleading: %#v", call)
		}
	}
}

func TestRunnerMovePreservesAndReportsExistingDestinationPaths(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
		case "move":
			if !contains(args, "--ignore-existing") {
				t.Fatalf("move could delete a source overlap: %#v", args)
			}
			return nil, nil
		case "lsf":
			if len(args) < 2 || args[1] != "/sources/media/Movie" || !contains(args, "--files-only") {
				t.Fatalf("remaining-source check used the wrong path or flags: %#v", args)
			}
			return []byte("file.mkv\n"), nil
		default:
			t.Fatalf("unexpected command: %#v", args)
			return nil, nil
		}
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.ConflictPolicy = model.ConflictMergeImmutable
	err := runner.Run(context.Background(), job)
	if err == nil || !strings.Contains(err.Error(), "preserved source files") || !strings.Contains(err.Error(), "unit is partial") {
		t.Fatalf("source overlap was not reported as partial: %v", err)
	}
	runs, runsErr := database.Runs(context.Background(), 10)
	if runsErr != nil || len(runs) != 1 || runs[0].State != "failed" {
		t.Fatalf("partial move was not persisted as failed: %#v err=%v", runs, runsErr)
	}
}

func TestRunnerMoveCannotCompleteWithMissingDestinationFile(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			if len(args) > 1 && args[1] == "GD:data/media/Movie" {
				return []byte(`[]`), nil
			}
			return listingFor(args), nil
		case "move":
			return nil, nil
		case "lsf":
			t.Fatalf("remaining-source check ran before destination completeness was proven: %#v", args)
		default:
			t.Fatalf("unexpected command: %#v", args)
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.ConflictPolicy = model.ConflictMergeImmutable
	err := runner.Run(context.Background(), job)
	if err == nil || !strings.Contains(err.Error(), `destination is missing discovered file "file.mkv"`) || !strings.Contains(err.Error(), "unit as partial") {
		t.Fatalf("incomplete destination was not reported as partial: %v", err)
	}
	runs, runsErr := database.Runs(context.Background(), 10)
	if runsErr != nil || len(runs) != 1 || runs[0].State != "failed" {
		t.Fatalf("incomplete destination was not persisted as failed: %#v err=%v", runs, runsErr)
	}
}

func TestRunnerChecksumVerificationUsesBackendHashes(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] == "lsjson" {
			return listingFor(args), nil
		}
		if args[0] == "lsf" {
			return []byte("directory not found"), errors.New("exit status 3")
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	if err := runner.Run(context.Background(), runnerJob()); err != nil {
		t.Fatal(err)
	}
	calls := fake.snapshot()
	if countCommand(calls, "copy") != 1 {
		t.Fatalf("checksum verification did not execute one copy: %#v", calls)
	}
	for _, call := range calls {
		if call[0] == "copy" && (!contains(call, "--checksum") || contains(call, "--size-only")) {
			t.Fatalf("checksum verification bypassed rclone's native backend hashes: %#v", call)
		}
	}
}

func TestRunnerPinsTransferManifestAndStableAge(t *testing.T) {
	database := runnerStore(t)
	var manifest string
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			return listingFor(args), nil
		case "copy":
			index := indexOf(args, "--files-from-raw")
			if index < 0 || index+1 >= len(args) {
				t.Fatalf("copy did not pin the discovered file set: %#v", args)
			}
			manifest = args[index+1]
			contents, err := os.ReadFile(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if string(contents) != "file.mkv\n" {
				t.Fatalf("transfer manifest = %q, want one discovered file", contents)
			}
			age := indexOf(args, "--min-age")
			if age < 0 || age+1 >= len(args) || args[age+1] != "86400s" {
				t.Fatalf("copy did not enforce the stable window at transfer time: %#v", args)
			}
			return nil, nil
		default:
			t.Fatalf("unexpected command: %#v", args)
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.ConflictPolicy = model.ConflictMergeImmutable
	job.SettleSeconds = 24 * 60 * 60
	if err := runner.Run(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if manifest == "" {
		t.Fatal("copy never exposed a transfer manifest")
	}
	if _, err := os.Stat(manifest); !os.IsNotExist(err) {
		t.Fatalf("transfer manifest was not removed: %v", err)
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
	if len(calls) != 1 || !containsPair(calls[0], "--transfers", "3") || !containsPair(calls[0], "--checkers", "4") || !containsPair(calls[0], "--tpslimit", "5") || !containsPair(calls[0], "--tpslimit-burst", "5") {
		t.Fatalf("rclone limits missing from argv: %#v", calls)
	}
}

func TestRunnerCanDisableTPSLimit(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, _ []string) ([]byte, error) { return nil, nil }}
	runner := NewWithLimits(database, "rclone", 1, 3, 4, 0)
	runner.execute = fake.execute
	if _, err := runner.command(context.Background(), "version"); err != nil {
		t.Fatal(err)
	}
	call := fake.snapshot()[0]
	if contains(call, "--tpslimit") || contains(call, "--tpslimit-burst") {
		t.Fatalf("disabled TPS limit remained in argv: %#v", call)
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

	if err := runner.Run(context.Background(), job); err == nil || !strings.Contains(err.Error(), "requires deleteSource=true") {
		t.Fatalf("synchronous run accepted invalid persisted job: %v", err)
	}
	if err := runner.Start(job); err == nil || !strings.Contains(err.Error(), "requires deleteSource=true") {
		t.Fatalf("asynchronous start accepted invalid persisted job: %v", err)
	}
	if err := runner.StartAnalysis(job); err == nil || !strings.Contains(err.Error(), "requires deleteSource=true") {
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

func TestRunnerRevalidatesStableWindowImmediatelyBeforeTransfer(t *testing.T) {
	database := runnerStore(t)
	listings := 0
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		switch args[0] {
		case "lsjson":
			listings++
			if listings == 1 {
				modified := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339Nano)
				return []byte(`[{"Path":"Movie/file.mkv","ModTime":"` + modified + `","IsDir":false}]`), nil
			}
			modified := time.Now().UTC().Format(time.RFC3339Nano)
			return []byte(`[{"Path":"new-episode.mkv","ModTime":"` + modified + `","IsDir":false}]`), nil
		case "lsf":
			return []byte("directory not found"), errors.New("exit status 3")
		default:
			t.Fatalf("unstable unit reached transfer command: %#v", args)
		}
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.SettleSeconds = 24 * 60 * 60
	if err := runner.Run(context.Background(), job); err == nil || !strings.Contains(err.Error(), "has not satisfied the stable window") {
		t.Fatalf("new file was not rejected before transfer: %v", err)
	}
	if calls := fake.snapshot(); countCommand(calls, "copy") != 0 || countCommand(calls, "move") != 0 {
		t.Fatalf("unstable unit reached rclone transfer: %#v", calls)
	}
}

func TestRunnerRejectsUnitFingerprintChangesBeforeTransfer(t *testing.T) {
	const initial = `[{"Path":"Movie/file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":100,"IsDir":false}]`
	tests := []struct {
		name          string
		current       func() string
		settleSeconds int
		want          string
	}{
		{
			name: "added path",
			current: func() string {
				return `[
				  {"Path":"file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":100,"IsDir":false},
				  {"Path":"extra.mkv","ModTime":"2026-01-01T00:00:00Z","Size":200,"IsDir":false}
				]`
			},
			want: `added path "extra.mkv"`,
		},
		{
			name:    "deleted path",
			current: func() string { return `[]` },
			want:    `deleted path "file.mkv"`,
		},
		{
			name: "size changed",
			current: func() string {
				return `[{"Path":"file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":101,"IsDir":false}]`
			},
			want: `size changed for path "file.mkv"`,
		},
		{
			name: "modification time changed",
			current: func() string {
				return `[{"Path":"file.mkv","ModTime":"2026-01-02T00:00:00Z","Size":100,"IsDir":false}]`
			},
			want: `modification time changed for path "file.mkv"`,
		},
		{
			name: "type changed",
			current: func() string {
				return `[{"Path":"file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":-1,"IsDir":true}]`
			},
			want: `type changed for path "file.mkv"`,
		},
		{
			name: "added young file",
			current: func() string {
				modified := time.Now().UTC().Format(time.RFC3339Nano)
				return `[
				  {"Path":"file.mkv","ModTime":"2026-01-01T00:00:00Z","Size":100,"IsDir":false},
				  {"Path":"downloading.mkv","ModTime":"` + modified + `","Size":200,"IsDir":false}
				]`
			},
			settleSeconds: 24 * 60 * 60,
			want:          `has not satisfied the stable window`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database := runnerStore(t)
			listings := 0
			fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
				if args[0] == "lsf" {
					return []byte("directory not found"), errors.New("exit status 3")
				}
				if args[0] != "lsjson" {
					t.Fatalf("changed unit reached non-listing rclone command: %#v", args)
				}
				listings++
				if listings == 1 {
					return []byte(initial), nil
				}
				return []byte(test.current()), nil
			}}
			runner := New(database, "rclone", 1)
			runner.execute = fake.execute
			job := runnerJob()
			job.Mode = model.ModeMove
			job.DeleteSource = true
			job.Verify = "size"
			job.SettleSeconds = test.settleSeconds
			err := runner.Run(context.Background(), job)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("fingerprint change returned %v, want %q", err, test.want)
			}
			calls := fake.snapshot()
			if countCommand(calls, "lsjson") != 2 || countCommand(calls, "copy") != 0 || countCommand(calls, "move") != 0 || countCommand(calls, "lsf") != 1 {
				t.Fatalf("changed unit reached transfer or destination preflight: %#v", calls)
			}
		})
	}
}

func TestRunnerRejectsReservedSourceNamespace(t *testing.T) {
	database := runnerStore(t)
	modified := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339Nano)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		if args[0] != "lsjson" {
			t.Fatalf("reserved source reached command: %#v", args)
		}
		return []byte(`[{"Path":".atomic-sync-staging/file.mkv","ModTime":"` + modified + `","IsDir":false}]`), nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	if err := runner.Run(context.Background(), runnerJob()); err == nil || !strings.Contains(err.Error(), "reserved Atomic Sync namespace") {
		t.Fatalf("reserved source namespace was silently ignored: %v", err)
	}
}

func TestRunnerRejectsReservedSourceEndpointBeforeListing(t *testing.T) {
	database := runnerStore(t)
	fake := &fakeExecutor{fn: func(_ context.Context, args []string) ([]byte, error) {
		t.Fatalf("reserved source endpoint reached rclone: %#v", args)
		return nil, nil
	}}
	runner := New(database, "rclone", 1)
	runner.execute = fake.execute
	job := runnerJob()
	job.Source = "/sources/media/.atomic-sync-staging/recovery"
	if err := runner.Run(context.Background(), job); err == nil || !strings.Contains(err.Error(), "reserved .atomic-sync-staging namespace") {
		t.Fatalf("reserved source endpoint was accepted: %v", err)
	}
	if calls := fake.snapshot(); len(calls) != 0 {
		t.Fatalf("reserved source endpoint invoked rclone: %#v", calls)
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
if [ "$1" = "lsjson" ] && [ "$2" = "/sources/shared-drive" ]; then
  printf '[{"Path":"Movie/file.mkv"}]'
  printf '%s\n' "2026/07/31 23:37:38 NOTICE: GD: ` + sharedDriveClientIDNotice + `" >&2
  exit 0
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
	out, err = runner.execRclone(context.Background(), "lsjson", "/sources/shared-drive")
	if err != nil || string(out) != `[{"Path":"Movie/file.mkv"}]` {
		t.Fatalf("shared Drive client_id notice blocked trusted inventory: output=%q err=%v", out, err)
	}
	if onlySharedDriveClientIDNotices("2026/07/31 23:37:38 NOTICE: GD: " + sharedDriveClientIDNotice + "\ndiagnostic warning") {
		t.Fatal("mixed inventory diagnostics were incorrectly trusted")
	}
	if onlySharedDriveClientIDNotices("ERROR fake NOTICE: GD: " + sharedDriveClientIDNotice) {
		t.Fatal("non-rclone log prefix was incorrectly trusted")
	}
	if onlySharedDriveClientIDNotices("2026/07/31 23:37:38 NOTICE: ERROR: quota: " + sharedDriveClientIDNotice) {
		t.Fatal("same-line diagnostic prefix was incorrectly trusted as a remote label")
	}
	if onlySharedDriveClientIDNotices("2026/07/31 23:37:38 NOTICE: ERROR quota: " + sharedDriveClientIDNotice) {
		t.Fatal("unsafe remote-like diagnostic prefix was incorrectly trusted")
	}
	truncatedPrefix := strings.Repeat("unexpected diagnostic\n", 2000)
	if onlySharedDriveClientIDNotices(truncatedPrefix + "2026/07/31 23:37:38 NOTICE: GD: " + sharedDriveClientIDNotice) {
		t.Fatal("an unexpected diagnostic before the bounded log tail was incorrectly trusted")
	}
	overlongNotices := strings.Repeat("2026/07/31 23:37:38 NOTICE: GD: "+sharedDriveClientIDNotice+"\n", 300)
	if onlySharedDriveClientIDNotices(overlongNotices) {
		t.Fatal("overlong inventory diagnostics were incorrectly trusted")
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

func indexOf(values []string, wanted string) int {
	for index, value := range values {
		if value == wanted {
			return index
		}
	}
	return -1
}

func containsPair(values []string, first, second string) bool {
	for index := 0; index+1 < len(values); index++ {
		if values[index] == first && values[index+1] == second {
			return true
		}
	}
	return false
}
