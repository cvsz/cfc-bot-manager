package transportapi

import (
	"context"
	"io"
)

// Transport interface for underlying HTTP communication.
type Transport interface {
	// DoRequest performs an HTTP request with timeout and cancellation.
	DoRequest(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*Response, error)

	// SetSessionCookie sets the upstream session cookie.
	SetSessionCookie(cookie string)

	// GetSessionCookie returns the current session cookie.
	GetSessionCookie() string

	// SetSession sets the full session.
	SetSession(session *Session)

	// GetSession returns the current session.
	GetSession() *Session

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
	ExpiresAt   int64 // Unix timestamp
	UserAgent   string
	LastRefresh int64 // Unix timestamp
}

// Config holds adapter configuration.
type Config struct {
	BaseURL            string
	Timeout            int64 // nanoseconds
	MaxRetries         int
	RetryDelay         int64 // nanoseconds
	UserAgent          string
	InsecureSkipVerify bool // For testing only
}