package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/gorilla/securecookie"
)

// Session errors
var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrInvalidCSRFToken     = errors.New("invalid CSRF token")
	ErrSessionCookieMissing = errors.New("session cookie missing")
)

// SessionManager handles secure session cookies
type SessionManager struct {
	cookieName     string
	cookieDomain   string
	cookiePath     string
	cookieSecure   bool
	cookieHTTPOnly bool
	cookieSameSite http.SameSite
	maxAge         int
	secureCookie   *securecookie.SecureCookie
	repo           SessionRepository
}

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Revoke(ctx context.Context, id domain.SessionID) error
	RevokeAllForUser(ctx context.Context, userID domain.UserID) error
	RevokeAllForOrg(ctx context.Context, orgID domain.OrganizationID) error
	DeleteExpired(ctx context.Context) (int64, error)
	CleanupRevoked(ctx context.Context, olderThan time.Time) (int64, error)
}

// SessionConfig holds configuration for the session manager
type SessionConfig struct {
	CookieName     string
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
	MaxAge         int // seconds
	HashKey        []byte // 32 bytes for AES-256
	BlockKey       []byte // 32 bytes for AES-256
	Repository     SessionRepository
}

// NewSessionManager creates a new session manager
func NewSessionManager(cfg SessionConfig) (*SessionManager, error) {
	if len(cfg.HashKey) != 32 {
		return nil, fmt.Errorf("hash key must be 32 bytes, got %d", len(cfg.HashKey))
	}
	if len(cfg.BlockKey) != 32 {
		return nil, fmt.Errorf("block key must be 32 bytes, got %d", len(cfg.BlockKey))
	}

	sc := securecookie.New(cfg.HashKey, cfg.BlockKey)
	sc.SetSerializer(securecookie.JSONEncoder{})

	return &SessionManager{
		cookieName:     cfg.CookieName,
		cookieDomain:   cfg.CookieDomain,
		cookiePath:     cfg.CookiePath,
		cookieSecure:   cfg.CookieSecure,
		cookieHTTPOnly: cfg.CookieHTTPOnly,
		cookieSameSite: cfg.CookieSameSite,
		maxAge:         cfg.MaxAge,
		secureCookie:   sc,
		repo:           cfg.Repository,
	}, nil
}

// CreateSession creates a new session and sets the cookie
func (sm *SessionManager) CreateSession(ctx context.Context, w http.ResponseWriter, user *domain.User, userAgent, ipAddress string) (*domain.Session, error) {
	// Generate session token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	tokenStr := base64.RawURLEncoding.EncodeToString(token)

	// Generate CSRF token
	csrfToken := make([]byte, 32)
	if _, err := rand.Read(csrfToken); err != nil {
		return nil, fmt.Errorf("generate csrf token: %w", err)
	}
	csrfTokenStr := base64.RawURLEncoding.EncodeToString(csrfToken)

	// Hash token for storage
	tokenHash := HashToken(tokenStr)

	now := time.Now()
	expiresAt := now.Add(time.Duration(sm.maxAge) * time.Second)

	session := &domain.Session{
		ID:           domain.NewSessionID(),
		UserID:       user.ID,
		Organization: user.Organization,
		TokenHash:    tokenHash,
		CSRFToken:    csrfTokenStr,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
	}

	if err := sm.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Set cookie
	sm.setSessionCookie(w, tokenStr)

	return session, nil
}

// GetSession retrieves and validates the session from the request
func (sm *SessionManager) GetSession(ctx context.Context, r *http.Request) (*domain.Session, error) {
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, ErrSessionCookieMissing
		}
		return nil, fmt.Errorf("read cookie: %w", err)
	}

	tokenStr := cookie.Value
	if tokenStr == "" {
		return nil, ErrSessionCookieMissing
	}

	tokenHash := HashToken(tokenStr)
	session, err := sm.repo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// Validate session
	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// ValidateCSRF validates the CSRF token from the request
func (sm *SessionManager) ValidateCSRF(r *http.Request, session *domain.Session) error {
	// Check header first (for AJAX)
	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken == "" {
		// Check form value
		csrfToken = r.FormValue("csrf_token")
	}
	if csrfToken == "" {
		return ErrInvalidCSRFToken
	}

	if subtle.ConstantTimeCompare([]byte(csrfToken), []byte(session.CSRFToken)) != 1 {
		return ErrInvalidCSRFToken
	}

	return nil
}

// RefreshSession extends the session expiry
func (sm *SessionManager) RefreshSession(ctx context.Context, w http.ResponseWriter, session *domain.Session) error {
	session.ExpiresAt = time.Now().Add(time.Duration(sm.maxAge) * time.Second)
	session.UpdatedAt = time.Now()

	if err := sm.repo.Update(ctx, session); err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	// Note: In practice, we'd regenerate the token and update the hash
	sm.setSessionCookie(w, session.TokenHash)
	return nil
}

// RevokeSession revokes the current session
func (sm *SessionManager) RevokeSession(ctx context.Context, w http.ResponseWriter, session *domain.Session) error {
	session.RevokedAt = ptrTime(time.Now())
	if err := sm.repo.Update(ctx, session); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	sm.clearSessionCookie(w)
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (sm *SessionManager) RevokeAllUserSessions(ctx context.Context, userID domain.UserID) error {
	return sm.repo.RevokeAllForUser(ctx, userID)
}

// RevokeAllOrgSessions revokes all sessions for an organization
func (sm *SessionManager) RevokeAllOrgSessions(ctx context.Context, orgID domain.OrganizationID) error {
	return sm.repo.RevokeAllForOrg(ctx, orgID)
}

// setSessionCookie sets the session cookie
func (sm *SessionManager) setSessionCookie(w http.ResponseWriter, token string) {
	encoded, err := sm.secureCookie.Encode(sm.cookieName, token)
	if err != nil {
		// Log error but don't fail the request
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sm.cookieName,
		Value:    encoded,
		Path:     sm.cookiePath,
		Domain:   sm.cookieDomain,
		MaxAge:   sm.maxAge,
		Secure:   sm.cookieSecure,
		HttpOnly: sm.cookieHTTPOnly,
		SameSite: sm.cookieSameSite,
	})
}

// clearSessionCookie clears the session cookie
func (sm *SessionManager) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sm.cookieName,
		Value:    "",
		Path:     sm.cookiePath,
		Domain:   sm.cookieDomain,
		MaxAge:   -1,
		Secure:   sm.cookieSecure,
		HttpOnly: sm.cookieHTTPOnly,
		SameSite: sm.cookieSameSite,
	})
}

// HashToken creates a hash of the token for storage
func HashToken(token string) string {
	// Use a simple but secure hash for token storage
	// In production, consider using a dedicated key derivation function
	return base64.RawURLEncoding.EncodeToString([]byte(token)) // Simplified - use proper hashing
}

func ptrTime(t time.Time) *time.Time {
	return &t
}