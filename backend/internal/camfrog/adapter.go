package camfrog

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog/capability"
)

// CamfrogAdapter defines the interface for communicating with Camfrog upstream.
// It isolates protocol specifics and exposes only evidence-backed capabilities.
type CamfrogAdapter interface {
	// Status queries the bot status from upstream.
	Status(ctx context.Context) (*StatusResult, error)

	// Language sets the bot language.
	Language(ctx context.Context, locale string) error

	// SwitchPlugin enables or disables a plugin.
	SwitchPlugin(ctx context.Context, pluginID int, enabled bool) error

	// GetPluginSettings retrieves plugin settings from upstream.
	GetPluginSettings(ctx context.Context, pluginID int, roomID string) (map[string]string, error)

	// SavePluginSettings saves plugin settings to upstream via multipart form.
	SavePluginSettings(ctx context.Context, pluginID int, roomID string, settings map[string]string) error

	// RoomPause pauses or resumes a room.
	RoomPause(ctx context.Context, roomID string, pause bool) error

	// RoomSetLanguage sets the room language.
	RoomSetLanguage(ctx context.Context, roomID, locale string) error

	// Close closes the adapter and releases resources.
	Close() error
}

// StatusResult represents the bot status response from upstream.
type StatusResult struct {
	Uptime    string `json:"uptime"`
	Reason    string `json:"reason"`
	Status    string `json:"status"` // e.g., "initializing", "connected", "working", "paused", "error"
	Raw       json.RawMessage
}

// Re-export capability types for external use
type CapabilityStatus = capability.CapabilityStatus
type PluginCapability = capability.PluginCapability
type ConfigField = capability.ConfigField

const (
	CapabilityVerified         = capability.CapabilityVerified
	CapabilityPartiallyVerified = capability.CapabilityPartiallyVerified
	CapabilityUnverified       = capability.CapabilityUnverified
	CapabilityUnsupported      = capability.CapabilityUnsupported
	CapabilityDisabled         = capability.CapabilityDisabled
)

// Transport interface for underlying HTTP communication.
type Transport interface {
	// DoRequest performs an HTTP request with timeout and cancellation.
	DoRequest(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*Response, error)

	// SetSessionCookie sets the upstream session cookie.
	SetSessionCookie(cookie string)

	// GetSessionCookie returns the current session cookie.
	GetSessionCookie() string

	// Close closes the transport.
	Close() error
}

// Response represents an HTTP response from upstream.
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       io.ReadCloser
}

// SessionManager manages upstream Camfrog session state.
type SessionManager interface {
	// GetSession retrieves the current session.
	GetSession() *Session

	// SetSession updates the session.
	SetSession(session *Session)

	// IsAuthenticated returns true if the session is valid.
	IsAuthenticated() bool

	// RefreshSession attempts to refresh the session.
	RefreshSession(ctx context.Context) error
}

// Session represents an upstream Camfrog session.
type Session struct {
	Cookie      string
	CSRFToken   string
	ExpiresAt   time.Time
	UserAgent   string
	LastRefresh time.Time
}

// Config holds adapter configuration.
type Config struct {
	BaseURL           string
	Timeout           time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
	UserAgent         string
	InsecureSkipVerify bool // For testing only
}