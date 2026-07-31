package config

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Listen          string
	DataDir         string
	APIToken        string
	RcloneBin       string
	MaxConcurrency  int
	RcloneTransfers int
	RcloneCheckers  int
	RcloneTPSLimit  int
	LogFormat       string
}

func Load() Config {
	data := env("ATOMIC_DATA_DIR", "/data")
	return Config{
		Listen: env("ATOMIC_LISTEN", "127.0.0.1:8080"), DataDir: data,
		APIToken: strings.TrimSpace(os.Getenv("ATOMIC_API_TOKEN")), RcloneBin: env("ATOMIC_RCLONE_BIN", "rclone"),
		MaxConcurrency:  intEnv("ATOMIC_MAX_CONCURRENCY", 2),
		RcloneTransfers: intEnv("ATOMIC_RCLONE_TRANSFERS", 2),
		RcloneCheckers:  intEnv("ATOMIC_RCLONE_CHECKERS", 2),
		RcloneTPSLimit:  intEnv("ATOMIC_RCLONE_TPS_LIMIT", 2),
		LogFormat:       env("ATOMIC_LOG_FORMAT", "json"),
	}
}

func (c Config) DBPath() string { return filepath.Join(c.DataDir, "atomic-sync.db") }

func (c Config) Validate() error {
	if c.APIToken != "" && len(c.APIToken) < 32 {
		return errors.New("ATOMIC_API_TOKEN must contain at least 32 characters when set")
	}
	if c.APIToken == "" {
		host, _, err := net.SplitHostPort(c.Listen)
		if err != nil {
			return errors.New("ATOMIC_LISTEN must be a host:port address")
		}
		if host != "localhost" {
			ip := net.ParseIP(host)
			if ip == nil || !ip.IsLoopback() {
				return errors.New("ATOMIC_API_TOKEN is required when ATOMIC_LISTEN is not loopback")
			}
		}
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return errors.New("ATOMIC_LOG_FORMAT must be json or text")
	}
	return nil
}
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value < 1 || value > 64 {
		return fallback
	}
	return value
}
