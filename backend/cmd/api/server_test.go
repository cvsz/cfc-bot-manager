package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/config"
)

func TestNewServerAppliesConfiguration(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:   "127.0.0.1:0",
		ReadHeader: 5 * time.Second,
		Read:       15 * time.Second,
		Write:      15 * time.Second,
		Idle:       60 * time.Second,
		Shutdown:   10 * time.Second,
	}
	server := NewServer(cfg, "test", nil)

	if server.Addr != "127.0.0.1:0" {
		t.Fatalf("server addr = %q, want %q", server.Addr, "127.0.0.1:0")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("read header timeout = %v, want %v", server.ReadHeaderTimeout, 5*time.Second)
	}
	if server.ReadTimeout != 15*time.Second {
		t.Fatalf("read timeout = %v, want %v", server.ReadTimeout, 15*time.Second)
	}
	if server.WriteTimeout != 15*time.Second {
		t.Fatalf("write timeout = %v, want %v", server.WriteTimeout, 15*time.Second)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("idle timeout = %v, want %v", server.IdleTimeout, 60*time.Second)
	}
	if server.MaxHeaderBytes != 1<<20 {
		t.Fatalf("max header bytes = %d, want %d", server.MaxHeaderBytes, 1<<20)
	}
	if server.Handler == nil {
		t.Fatal("handler must not be nil")
	}
}

// readBody reads the full response body and closes it.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(body)
}

func TestServerStartsAndResponds(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:   "127.0.0.1:0", // ephemeral port
		ReadHeader: 5 * time.Second,
		Read:       15 * time.Second,
		Write:      15 * time.Second,
		Idle:       60 * time.Second,
		Shutdown:   10 * time.Second,
	}
	server := NewServer(cfg, "test", nil)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("create listener: %v", err)
	}
	server.Addr = ln.Addr().String()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ln)
	}()
	t.Cleanup(func() {
		_ = server.Close()
	})

	// Wait for the server to accept connections by polling /healthz.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + server.Addr + "/healthz")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if body := readBody(t, resp); body != `{"status":"ok"}`+"\n" {
					t.Fatalf("unexpected healthz body: %q", body)
				}
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not respond to /healthz within timeout")
}

func TestGracefulShutdownCompletes(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:   "127.0.0.1:0",
		ReadHeader: 5 * time.Second,
		Read:       15 * time.Second,
		Write:      15 * time.Second,
		Idle:       60 * time.Second,
		Shutdown:   5 * time.Second,
	}
	server := NewServer(cfg, "test", nil)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("create listener: %v", err)
	}
	server.Addr = ln.Addr().String()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ln)
	}()

	// Verify the server is accepting connections before shutting down.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + server.Addr + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}
	if err := <-serveErr; err != nil && err != http.ErrServerClosed {
		t.Fatalf("server returned unexpected error after shutdown: %v", err)
	}
}

func TestReadinessFailureReturns503(t *testing.T) {
	server := NewServer(config.Config{
		HTTPAddr:   "127.0.0.1:0",
		ReadHeader: 5 * time.Second,
		Read:       15 * time.Second,
		Write:      15 * time.Second,
		Idle:       60 * time.Second,
		Shutdown:   10 * time.Second,
	}, "test", func(context.Context) error {
		return errors.New("readiness check failed")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("create listener: %v", err)
	}
	server.Addr = ln.Addr().String()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ln)
	}()
	t.Cleanup(func() {
		_ = server.Close()
	})

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + server.Addr + "/readyz")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusServiceUnavailable {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not return 503 for /readyz within timeout")
}
