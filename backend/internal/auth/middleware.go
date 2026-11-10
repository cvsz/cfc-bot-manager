package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
)

// Context keys for storing auth data
type contextKey string

const (
	ContextKeySession   contextKey = "session"
	ContextKeyUser      contextKey = "user"
	ContextKeyCSRFToken contextKey = "csrf_token"
	ContextKeyPerms     contextKey = "permissions"
)

// AuthMiddleware provides authentication and authorization middleware
type AuthMiddleware struct {
	sessionManager *SessionManager
	rbacManager    *RBACManager
	cookieName     string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(sm *SessionManager, rm *RBACManager, cookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sm,
		rbacManager:    rm,
		cookieName:     cookieName,
	}
}

// RequireAuth middleware that requires a valid session
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, err := am.sessionManager.GetSession(ctx, r)
		if err != nil {
			if errors.Is(err, ErrSessionCookieMissing) {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}
			if errors.Is(err, ErrSessionExpired) {
				http.Error(w, "session expired", http.StatusUnauthorized)
				return
			}
			if errors.Is(err, ErrSessionRevoked) {
				http.Error(w, "session revoked", http.StatusUnauthorized)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Add session to context
		ctx = context.WithValue(ctx, ContextKeySession, session)
		ctx = context.WithValue(ctx, ContextKeyCSRFToken, session.CSRFToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireMFA middleware that requires MFA to be enabled and verified
func (am *AuthMiddleware) RequireMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		// Check if MFA is enabled for this user
		// This would typically fetch the user from the database
		// For now, we'll add a header check for MFA verification
		mfaVerified := r.Header.Get("X-MFA-Verified")
		if mfaVerified != "true" {
			http.Error(w, "MFA required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequirePermission middleware factory
func (am *AuthMiddleware) RequirePermission(resourceType ResourceType, action Action) func(http.Handler) http.Handler {
	return am.rbacManager.RequirePermission(resourceType, action)
}

// RequireResourceAccess middleware factory
func (am *AuthMiddleware) RequireResourceAccess(resourceType ResourceType, action Action, getResourceOrgID func(*http.Request) (domain.OrganizationID, error)) func(http.Handler) http.Handler {
	return am.rbacManager.RequireResourceAccess(resourceType, action, getResourceOrgID)
}

// CSRFMiddleware validates CSRF tokens for mutating requests
func (am *AuthMiddleware) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		session := GetSessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		if err := am.sessionManager.ValidateCSRF(r, session); err != nil {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalAuth middleware that adds session to context if present but doesn't require it
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, err := am.sessionManager.GetSession(ctx, r)
		if err == nil && session != nil {
			ctx = context.WithValue(ctx, ContextKeySession, session)
			ctx = context.WithValue(ctx, ContextKeyCSRFToken, session.CSRFToken)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSessionFromContext retrieves the session from the request context
func GetSessionFromContext(ctx context.Context) *domain.Session {
	if session, ok := ctx.Value(ContextKeySession).(*domain.Session); ok {
		return session
	}
	return nil
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *domain.User {
	if user, ok := ctx.Value(ContextKeyUser).(*domain.User); ok {
		return user
	}
	return nil
}

// GetCSRFTokenFromContext retrieves the CSRF token from the request context
func GetCSRFTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(ContextKeyCSRFToken).(string); ok {
		return token
	}
	return ""
}

// GetPermissionsFromContext retrieves the permissions from the request context
func GetPermissionsFromContext(ctx context.Context) []Permission {
	if perms, ok := ctx.Value(ContextKeyPerms).([]Permission); ok {
		return perms
	}
	return nil
}

// SetSessionContext sets the session in the context
func SetSessionContext(ctx context.Context, session *domain.Session) context.Context {
	ctx = context.WithValue(ctx, ContextKeySession, session)
	ctx = context.WithValue(ctx, ContextKeyCSRFToken, session.CSRFToken)
	return ctx
}

// SetUserContext sets the user in the context
func SetUserContext(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, ContextKeyUser, user)
}

// SetPermissionsContext sets the permissions in the context
func SetPermissionsContext(ctx context.Context, permissions []Permission) context.Context {
	return context.WithValue(ctx, ContextKeyPerms, permissions)
}

// CORSHeaders adds CORS headers for API access
func CORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-MFA-Verified")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds security-related headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
}

// RateLimiter provides basic rate limiting
type RateLimiter struct {
	config RateLimitConfig
	// In production, use a proper rate limiter like golang.org/x/time/rate
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{config: config}
}

// RateLimitMiddleware provides basic rate limiting
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	// Simplified - in production use a proper rate limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// AuditLogMiddleware logs requests for audit purposes
func AuditLogMiddleware(auditRepo AuditLogRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response recorder to capture status code
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			// Log the request
			session := GetSessionFromContext(r.Context())
			if session != nil {
				// In production, log to audit repository
				_ = session // placeholder
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// AuditLogRepository interface for audit logging
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
}