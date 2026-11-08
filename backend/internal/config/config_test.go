package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "")
	// Empty duration values are invalid rather than silently falling back.
	if _, err := Load(); err == nil {
		t.Fatal("expected empty duration to be rejected")
	}
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "5s")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" || cfg.ReadHeader != 5*time.Second {
		t.Fatalf("unexpected configuration: %+v", cfg)
	}
}

func TestValidateRejectsInvalidAddressAndTimeout(t *testing.T) {
	cfg := Config{
		Environment: "test", HTTPAddr: "127.0.0.1:0",
		ReadHeader: time.Second, Read: time.Second, Write: time.Second,
		Idle: time.Second, Shutdown: time.Second,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid port to fail")
	}
	cfg.HTTPAddr = ":8080"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unspecified host to fail")
	}
	cfg.HTTPAddr = "127.0.0.1:8080"
	cfg.Write = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected zero timeout to fail")
	}
}

func TestLoadRejectsMalformedDuration(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("expected malformed duration to fail")
	}
}
