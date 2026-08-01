package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("ATOMIC_LISTEN", "127.0.0.1:9090")
	t.Setenv("ATOMIC_DATA_DIR", "/tmp/atomic-state")
	token := strings.Repeat("x", 32)
	t.Setenv("ATOMIC_API_TOKEN", "  "+token+"  ")
	t.Setenv("ATOMIC_RCLONE_BIN", "/usr/bin/rclone")
	t.Setenv("ATOMIC_MAX_CONCURRENCY", "7")
	t.Setenv("ATOMIC_RCLONE_TRANSFERS", "3")
	t.Setenv("ATOMIC_RCLONE_CHECKERS", "5")
	t.Setenv("ATOMIC_RCLONE_TPS_LIMIT", "2")
	t.Setenv("ATOMIC_LOG_FORMAT", "text")
	config := Load()
	if config.Listen != "127.0.0.1:9090" || config.APIToken != token || config.MaxConcurrency != 7 || config.RcloneTransfers != 3 || config.RcloneCheckers != 5 || config.RcloneTPSLimit != 2 || config.LogFormat != "text" {
		t.Fatalf("unexpected config: %#v", config)
	}
	if config.DBPath() != "/tmp/atomic-state/atomic-sync.db" {
		t.Fatalf("unexpected DB path: %s", config.DBPath())
	}
}

func TestValidateSecurityBoundary(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{name: "loopback without token", config: Config{Listen: "127.0.0.1:8080", LogFormat: "json"}, valid: true},
		{name: "ipv6 loopback without token", config: Config{Listen: "[::1]:8080", LogFormat: "text"}, valid: true},
		{name: "public without token", config: Config{Listen: ":8080", LogFormat: "json"}},
		{name: "short token", config: Config{Listen: ":8080", APIToken: "too-short", LogFormat: "json"}},
		{name: "public with token", config: Config{Listen: ":8080", APIToken: strings.Repeat("x", 32), LogFormat: "json"}, valid: true},
		{name: "bad log format", config: Config{Listen: "127.0.0.1:8080", LogFormat: "yaml"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.config.Validate(); (err == nil) != test.valid {
				t.Fatalf("Validate() error=%v, valid=%v", err, test.valid)
			}
		})
	}
}

func TestInvalidConcurrencyFallsBack(t *testing.T) {
	t.Setenv("ATOMIC_MAX_CONCURRENCY", "999")
	if got := Load().MaxConcurrency; got != 2 {
		t.Fatalf("got %d, want fallback 2", got)
	}
}

func TestTPSLimitCanBeDisabled(t *testing.T) {
	t.Setenv("ATOMIC_RCLONE_TPS_LIMIT", "0")
	if got := Load().RcloneTPSLimit; got != 0 {
		t.Fatalf("got %d, want disabled TPS limit", got)
	}
}
