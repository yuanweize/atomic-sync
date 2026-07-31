package main

import (
	"context"
	"errors"
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

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	if err := os.MkdirAll(cfg.DataDir, 0750); err != nil {
		slog.Error("create data directory", "error", err)
		os.Exit(1)
	}
	db, err := store.Open(cfg.DBPath())
	if err != nil {
		slog.Error("open database", "error", err, "path", filepath.Clean(cfg.DBPath()))
		os.Exit(1)
	}
	defer db.Close()
	if recovered, err := db.FailInterruptedRuns(context.Background(), "interrupted by a previous process exit; staging preserved"); err != nil {
		slog.Error("reconcile interrupted runs", "error", err)
		os.Exit(1)
	} else if recovered > 0 {
		slog.Warn("reconciled interrupted runs", "count", recovered)
	}
	if recovered, err := db.FailInterruptedAnalyses(context.Background(), "interrupted by a previous process exit; run analysis again"); err != nil {
		slog.Error("reconcile interrupted analyses", "error", err)
		os.Exit(1)
	} else if recovered > 0 {
		slog.Warn("reconciled interrupted analyses", "count", recovered)
	}
	if cfg.APIToken == "" {
		slog.Warn("ATOMIC_API_TOKEN is empty; bind only to a trusted loopback interface")
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runner := engine.New(db, cfg.RcloneBin, cfg.MaxConcurrency)
	srv := &http.Server{
		Addr: cfg.Listen, Handler: api.New(db, runner, cfg.APIToken).Handler(),
		ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	go func() {
		slog.Info("atomic-sync listening", "address", cfg.Listen, "version", buildinfo.Version, "commit", buildinfo.Commit)
		if e := srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			slog.Error("server failed", "error", e)
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		slog.Error("HTTP shutdown", "error", err)
		_ = srv.Close()
	}
	if err := runner.Shutdown(shutdown); err != nil {
		slog.Error("runner shutdown", "error", err)
	}
}
