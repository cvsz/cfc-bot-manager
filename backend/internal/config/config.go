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
	Environment string
	HTTPAddr string
	DatabaseURL string
	RoomStyleToken string
	ReadHeader time.Duration
	Read time.Duration
	Write time.Duration
	Idle time.Duration
	Shutdown time.Duration
}

func Load() (Config, error) {
	cfg := Config{Environment:"development", HTTPAddr:"127.0.0.1:8080", ReadHeader:5*time.Second, Read:15*time.Second, Write:15*time.Second, Idle:60*time.Second, Shutdown:10*time.Second}
	if v, ok := os.LookupEnv("APP_ENV"); ok { cfg.Environment = strings.TrimSpace(v) }
	if v, ok := os.LookupEnv("HTTP_ADDR"); ok { cfg.HTTPAddr = strings.TrimSpace(v) }
	cfg.DatabaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	cfg.RoomStyleToken = os.Getenv("ROOM_STYLE_API_TOKEN")
	for _, item := range []struct{name string; dst *time.Duration}{
		{"HTTP_READ_HEADER_TIMEOUT",&cfg.ReadHeader},{"HTTP_READ_TIMEOUT",&cfg.Read},{"HTTP_WRITE_TIMEOUT",&cfg.Write},{"HTTP_IDLE_TIMEOUT",&cfg.Idle},{"HTTP_SHUTDOWN_TIMEOUT",&cfg.Shutdown},
	} {
		if v, ok := os.LookupEnv(item.name); ok {
			d, err := time.ParseDuration(strings.TrimSpace(v))
			if err != nil || d <= 0 || d > 10*time.Minute { return Config{}, fmt.Errorf("%s must be greater than 0 and no more than 10m", item.name) }
			*item.dst = d
		}
	}
	if err := cfg.Validate(); err != nil { return Config{}, err }
	return cfg, nil
}

func (c Config) Validate() error {
	switch c.Environment { case "development","test","production": default: return fmt.Errorf("APP_ENV must be development, test, or production") }
	host, portText, err := net.SplitHostPort(c.HTTPAddr)
	if err != nil || host == "" || strings.ContainsAny(host," \t\r\n") { return fmt.Errorf("HTTP_ADDR must contain an explicit host and port") }
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 { return fmt.Errorf("HTTP_ADDR port must be between 1 and 65535") }
	for name, value := range map[string]time.Duration{"HTTP_READ_HEADER_TIMEOUT":c.ReadHeader,"HTTP_READ_TIMEOUT":c.Read,"HTTP_WRITE_TIMEOUT":c.Write,"HTTP_IDLE_TIMEOUT":c.Idle,"HTTP_SHUTDOWN_TIMEOUT":c.Shutdown} {
		if value <= 0 || value > 10*time.Minute { return fmt.Errorf("%s must be greater than 0 and no more than 10m", name) }
	}
	if (c.DatabaseURL == "") != (c.RoomStyleToken == "") { return fmt.Errorf("DATABASE_URL and ROOM_STYLE_API_TOKEN must be configured together") }
	if c.DatabaseURL != "" {
		u, err := url.Parse(c.DatabaseURL)
		if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" { return fmt.Errorf("DATABASE_URL must be a valid postgres:// or postgresql:// URL") }
		if len(c.RoomStyleToken) < 32 { return fmt.Errorf("ROOM_STYLE_API_TOKEN must contain at least 32 characters") }
	}
	return nil
}
