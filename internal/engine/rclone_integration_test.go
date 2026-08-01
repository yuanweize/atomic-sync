package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yuanweize/atomic-sync/internal/model"
)

func TestLocalRcloneDryRunAndMoveSemantics(t *testing.T) {
	t.Setenv("RCLONE_CONFIG", os.DevNull)
	rclone := os.Getenv("ATOMIC_TEST_RCLONE_BIN")
	if rclone == "" {
		var err error
		rclone, err = exec.LookPath("rclone")
		if err != nil {
			t.Skip("rclone is not installed")
		}
	} else if info, err := os.Stat(rclone); err != nil || info.IsDir() {
		t.Fatalf("ATOMIC_TEST_RCLONE_BIN %q is not an executable file: %v", rclone, err)
	}

	t.Run("dry run leaves both sides unchanged", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "movie.mkv"), "source")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictFail, "size", true)
		if err := runner.transferUnit(context.Background(), execution); err != nil {
			t.Fatal(err)
		}
		assertLocalFile(t, filepath.Join(source, "movie.mkv"), "source")
		if _, err := os.Stat(destination); !os.IsNotExist(err) {
			t.Fatalf("dry run created destination: %v", err)
		}
	})

	t.Run("move removes source after transfer", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "movie.mkv"), "source")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictFail, "size", false)
		if err := runner.transferUnit(context.Background(), execution); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(source, "movie.mkv")); !os.IsNotExist(err) {
			t.Fatalf("successful move retained source file: %v", err)
		}
		assertLocalFile(t, filepath.Join(destination, "movie.mkv"), "source")
	})

	t.Run("fail policy preserves an existing destination", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "movie.mkv"), "source")
		writeLocalTestFile(t, filepath.Join(destination, "movie.mkv"), "target")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictFail, "size", false)
		if err := runner.transferUnit(context.Background(), execution); err == nil {
			t.Fatal("existing destination was accepted by fail policy")
		}
		assertLocalFile(t, filepath.Join(source, "movie.mkv"), "source")
		assertLocalFile(t, filepath.Join(destination, "movie.mkv"), "target")
	})

	t.Run("immutable move merge preserves existing source paths", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "already.mkv"), "same")
		writeLocalTestFile(t, filepath.Join(source, "new.mkv"), "new")
		writeLocalTestFile(t, filepath.Join(destination, "already.mkv"), "same")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictMergeImmutable, "checksum", false)
		if err := runner.transferUnit(context.Background(), execution); err == nil {
			t.Fatal("move merge silently treated an existing destination path as removable")
		}
		assertLocalFile(t, filepath.Join(source, "already.mkv"), "same")
		if _, err := os.Stat(filepath.Join(source, "new.mkv")); !os.IsNotExist(err) {
			t.Fatalf("missing destination file was not moved: %v", err)
		}
		assertLocalFile(t, filepath.Join(destination, "already.mkv"), "same")
		assertLocalFile(t, filepath.Join(destination, "new.mkv"), "new")
	})

	t.Run("immutable checksum merge preserves different content", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "movie.mkv"), "source")
		writeLocalTestFile(t, filepath.Join(destination, "movie.mkv"), "target")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictMergeImmutable, "checksum", false)
		if err := runner.transferUnit(context.Background(), execution); err == nil {
			t.Fatal("different immutable destination was accepted")
		}
		assertLocalFile(t, filepath.Join(source, "movie.mkv"), "source")
		assertLocalFile(t, filepath.Join(destination, "movie.mkv"), "target")
	})

	t.Run("size-only move never deletes a same-size overlap", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "overlap.mkv"), "AAAA")
		writeLocalTestFile(t, filepath.Join(source, "missing.mkv"), "new")
		writeLocalTestFile(t, filepath.Join(destination, "overlap.mkv"), "BBBB")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictMergeImmutable, "size", false)
		if err := runner.transferUnit(context.Background(), execution); err == nil {
			t.Fatal("same-size overlap was reported as a completed move")
		}
		assertLocalFile(t, filepath.Join(source, "overlap.mkv"), "AAAA")
		assertLocalFile(t, filepath.Join(destination, "overlap.mkv"), "BBBB")
		assertLocalFile(t, filepath.Join(destination, "missing.mkv"), "new")
	})

	t.Run("move never absorbs a file arriving after revalidation", func(t *testing.T) {
		source, destination := localTransferPaths(t)
		writeLocalTestFile(t, filepath.Join(source, "discovered.mkv"), "discovered")
		execution, runner := localExecution(t, rclone, source, destination, model.ConflictMergeImmutable, "size", false)
		injected := false
		runner.execute = func(ctx context.Context, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "move" && !injected {
				injected = true
				writeLocalTestFile(t, filepath.Join(source, "late.mkv"), "late")
			}
			return runner.execRclone(ctx, args...)
		}
		if err := runner.transferUnit(context.Background(), execution); err == nil || !strings.Contains(err.Error(), "preserved source files") {
			t.Fatalf("late arrival was not retained and reported as partial: %v", err)
		}
		if !injected {
			t.Fatal("test did not inject a late source file")
		}
		assertLocalFile(t, filepath.Join(source, "late.mkv"), "late")
		if _, err := os.Stat(filepath.Join(source, "discovered.mkv")); !os.IsNotExist(err) {
			t.Fatalf("discovered file was not moved: %v", err)
		}
		assertLocalFile(t, filepath.Join(destination, "discovered.mkv"), "discovered")
		if _, err := os.Stat(filepath.Join(destination, "late.mkv")); !os.IsNotExist(err) {
			t.Fatalf("late file escaped the pinned manifest: %v", err)
		}
	})
}

func localTransferPaths(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	return filepath.Join(root, "source"), filepath.Join(root, "destination")
}

func localExecution(t *testing.T, rclone, source, destination, conflict, verify string, dryRun bool) (*unitExecution, *Runner) {
	t.Helper()
	database := runnerStore(t)
	job := runnerJob()
	job.ID = "job_local"
	job.Source = source
	job.Destinations = []model.Destination{{Name: "local", Path: destination, Weight: 1}}
	job.Mode = model.ModeMove
	job.DeleteSource = true
	job.ConflictPolicy = conflict
	job.Verify = verify
	job.DryRun = dryRun
	if err := database.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	run := model.Run{
		ID: "run_local", JobID: job.ID, Unit: "unit", Destination: "local",
		State: "discovered", StartedAt: time.Now().UTC(),
	}
	if err := database.CreateRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	runner := NewWithLimits(database, rclone, 1, 1, 1, 0)
	fingerprint, err := runner.scanUnitFingerprint(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	return &unitExecution{job: job, unit: "unit", fingerprint: fingerprint, destination: job.Destinations[0], destinationName: "local", run: run, source: source, final: destination}, runner
}

func writeLocalTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertLocalFile(t *testing.T, path, expected string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != expected {
		t.Fatalf("%s contains %q, want %q", path, contents, expected)
	}
}
