package camfrog

import (
	"errors"
	"time"
)

// NewAdapter creates a new Camfrog adapter with the given configuration.
// This is the public factory function for creating adapters.
func NewAdapter(cfg Config) (CamfrogAdapter, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("base URL is required")
	}

	// Default configuration values
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 1 * time.Second
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "CFC Bot Manager/1.0"
	}

	// Delegate to internal implementation
	return newAdapterImpl(cfg)
}