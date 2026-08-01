package main

import (
	"errors"
	"net"
	"strings"
	"testing"
)

func TestRunReturnsListenFailure(t *testing.T) {
	want := errors.New("test listener unavailable")
	previous := listen
	listen = func(network, address string) (net.Listener, error) {
		if network != "tcp" || address != "127.0.0.1:18089" {
			t.Fatalf("unexpected listen request %q %q", network, address)
		}
		return nil, want
	}
	t.Cleanup(func() { listen = previous })

	t.Setenv("ATOMIC_LISTEN", "127.0.0.1:18089")
	t.Setenv("ATOMIC_DATA_DIR", t.TempDir())
	t.Setenv("ATOMIC_API_TOKEN", "0123456789abcdef0123456789abcdef")
	t.Setenv("ATOMIC_RCLONE_BIN", "rclone")

	err := run()
	if err == nil || !errors.Is(err, want) || !strings.Contains(err.Error(), "listen on") {
		t.Fatalf("occupied listener returned %v", err)
	}
}
