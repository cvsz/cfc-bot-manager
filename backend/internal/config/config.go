package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment    string
	HTTPAddr       string
	DatabaseURL    string
	RoomStyleToken string
	ReadHeader     time.Duration
	Read           time.Duration
	Write          time.Duration
	Idle            time.Duration
	Shutdown        time.Duration
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
	cfg.DatabaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	cfg.RoomStyleToken = os.Getenv("ROOM_STYLE_API_TOKEN")
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
			parsed, err := time.ParseDuration(strings.TrimSpace(value))
			if err != nil || parsed <= 0 || parsed > 10*time.Minute {
				return Config{}, fmt.Errorf("%s must be greater than 0 and no more than 10m", item.name)
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
	if err != nil || host == "" || strings.ContainsAny(host, " \t\r\n") {
		return fmt.Errorf("HTTP_ADDR must contain an explicit host and port")
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
	if (c.DatabaseURL == "") != (c.RoomStyleToken == "") {
		return fmt.Errorf("DATABASE_URL and ROOM_STYLE_API_TOKEN must be configured together")
	}
	if c.DatabaseURL != "" {
		u, err := url.Parse(c.DatabaseURL)
		if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" {
			return fmt.Errorf("DATABASE_URL must be a valid postgres:// or postgresql:// URL")
		}
		if len(c.RoomStyleToken) < 32 || len(c.RoomStyleToken) > 256 || strings.ContainsAny(c.RoomStyleToken, " \t\r\n") {
			return fmt.Errorf("ROOM_STYLE_API_TOKEN must contain 32 to 256 non-whitespace characters")
		}
	}
	return nil
}
