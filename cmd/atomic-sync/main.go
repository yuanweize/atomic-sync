package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yuanweize/atomic-sync/internal/api"
	"github.com/yuanweize/atomic-sync/internal/buildinfo"
	"github.com/yuanweize/atomic-sync/internal/config"
	"github.com/yuanweize/atomic-sync/internal/engine"
	"github.com/yuanweize/atomic-sync/internal/store"
)

var listen = net.Listen

func main() {
	if err := run(); err != nil {
		slog.Error("atomic-sync stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	if err := os.MkdirAll(cfg.DataDir, 0750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	db, err := store.Open(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("open database %s: %w", filepath.Clean(cfg.DBPath()), err)
	}
	defer db.Close()
	if recovered, err := db.FailInterruptedRuns(context.Background(), "interrupted by a previous process exit; inspect source and destination before retrying"); err != nil {
		return fmt.Errorf("reconcile interrupted runs: %w", err)
	} else if recovered > 0 {
		slog.Warn("reconciled interrupted runs", "count", recovered)
	}
	if recovered, err := db.FailInterruptedAnalyses(context.Background(), "interrupted by a previous process exit; run analysis again"); err != nil {
		return fmt.Errorf("reconcile interrupted analyses: %w", err)
	} else if recovered > 0 {
		slog.Warn("reconciled interrupted analyses", "count", recovered)
	}
	if cfg.APIToken == "" {
		slog.Warn("ATOMIC_API_TOKEN is empty; bind only to a trusted loopback interface")
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runner := engine.NewWithLimits(
		db, cfg.RcloneBin, cfg.MaxConcurrency,
		cfg.RcloneTransfers, cfg.RcloneCheckers, cfg.RcloneTPSLimit,
	)
	srv := &http.Server{
		Addr: cfg.Listen, Handler: api.New(db, runner, cfg.APIToken).Handler(),
		ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	listener, err := listen("tcp", cfg.Listen)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.Listen, err)
	}
	defer listener.Close()
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("atomic-sync listening", "address", cfg.Listen, "version", buildinfo.Version, "commit", buildinfo.Commit)
		err := srv.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
	}()
	var fatalErr error
	select {
	case <-ctx.Done():
	case fatalErr = <-serveErr:
		stop()
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		fatalErr = errors.Join(fatalErr, fmt.Errorf("HTTP shutdown: %w", err))
		_ = srv.Close()
	}
	if err := runner.Shutdown(shutdown); err != nil {
		fatalErr = errors.Join(fatalErr, fmt.Errorf("runner shutdown: %w", err))
	}
	return fatalErr
}
