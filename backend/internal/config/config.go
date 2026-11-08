package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	HTTPAddr string
	ReadHeader time.Duration
	Read time.Duration
	Write time.Duration
	Idle time.Duration
	Shutdown time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment: "development",
		HTTPAddr: "127.0.0.1:8080",
		ReadHeader: 5 * time.Second,
		Read: 15 * time.Second,
		Write: 15 * time.Second,
		Idle: 60 * time.Second,
		Shutdown: 10 * time.Second,
	}
	if value, ok := os.LookupEnv("APP_ENV"); ok {
		cfg.Environment = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("HTTP_ADDR"); ok {
		cfg.HTTPAddr = strings.TrimSpace(value)
	}
	durations := []struct {
		name string
		dst *time.Duration
	}{
		{"HTTP_READ_HEADER_TIMEOUT", &cfg.ReadHeader},
		{"HTTP_READ_TIMEOUT", &cfg.Read},
		{"HTTP_WRITE_TIMEOUT", &cfg.Write},
		{"HTTP_IDLE_TIMEOUT", &cfg.Idle},
		{"HTTP_SHUTDOWN_TIMEOUT", &cfg.Shutdown},
	}
	for _, item := range durations {
		if value, ok := os.LookupEnv(item.name); ok {
			parsed, err := parseDuration(item.name, value)
			if err != nil {
				return Config{}, err
			}
			*item.dst = parsed
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	switch c.Environment {
	case "development", "test", "production":
	default:
		return fmt.Errorf("APP_ENV must be development, test, or production")
	}
	host, portText, err := net.SplitHostPort(c.HTTPAddr)
	if err != nil {
		return fmt.Errorf("HTTP_ADDR must be host:port: %w", err)
	}
	if host == "" {
		return fmt.Errorf("HTTP_ADDR host must be explicit; use 127.0.0.1 for local-only binding")
	}
	if strings.ContainsAny(host, " \t\r\n") {
		return fmt.Errorf("HTTP_ADDR host contains whitespace")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("HTTP_ADDR port must be between 1 and 65535")
	}
	timeouts := []struct {
		name string
		value time.Duration
	}{
		{"HTTP_READ_HEADER_TIMEOUT", c.ReadHeader},
		{"HTTP_READ_TIMEOUT", c.Read},
		{"HTTP_WRITE_TIMEOUT", c.Write},
		{"HTTP_IDLE_TIMEOUT", c.Idle},
		{"HTTP_SHUTDOWN_TIMEOUT", c.Shutdown},
	}
	for _, item := range timeouts {
		if item.value <= 0 || item.value > 10*time.Minute {
			return fmt.Errorf("%s must be greater than 0 and no more than 10m", item.name)
		}
	}
	return nil
}

func parseDuration(name, raw string) (time.Duration, error) {
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	if value <= 0 || value > 10*time.Minute {
		return 0, fmt.Errorf("%s must be greater than 0 and no more than 10m", name)
	}
	return value, nil
}
