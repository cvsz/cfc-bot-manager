package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	// Verify all sentinel errors are non-nil and have messages
	sentinelErrors := []error{
		ErrOrganizationNotFound,
		ErrUserNotFound,
		ErrRoleNotFound,
		ErrPermissionNotFound,
		ErrSessionNotFound,
		ErrRoomNotFound,
		ErrBotNotFound,
		ErrPluginNotFound,
		ErrPluginConfigNotFound,
		ErrAuditLogNotFound,
		ErrAPIKeyNotFound,
		ErrMembershipNotFound,
		ErrOrganizationSlugExists,
		ErrUserEmailExists,
		ErrRoleNameExists,
		ErrRoomNameExists,
		ErrBotUsernameExists,
		ErrPluginNumberExists,
		ErrAPIKeyPrefixExists,
		ErrUnauthorized,
		ErrForbidden,
		ErrTenantMismatch,
		ErrInsufficientRole,
		ErrMFARequired,
		ErrSessionExpired,
		ErrSessionRevoked,
		ErrCSRFInvalid,
		ErrInvalidCredentials,
		ErrInvalidInput,
		ErrInvalidOrganization,
		ErrInvalidUser,
		ErrInvalidRole,
		ErrInvalidSession,
		ErrInvalidRoom,
		ErrInvalidBot,
		ErrInvalidPlugin,
		ErrInvalidPluginConfig,
		ErrInvalidAPIKey,
		ErrInvalidPagination,
		ErrInvalidSort,
		ErrInvalidFilter,
		ErrBotAlreadyRunning,
		ErrBotNotRunning,
		ErrRoomAlreadyHasBot,
		ErrBotAlreadyHasRoom,
		ErrPluginNotVerified,
		ErrPluginDisabled,
		ErrSystemRoleImmutable,
		ErrLastAdminRole,
		ErrCamfrogUnavailable,
		ErrCamfrogAuthFailed,
		ErrCamfrogTimeout,
		ErrCamfrogInvalidResponse,
		ErrCamfrogRateLimited,
		ErrOptimisticLock,
		ErrTransaction,
	}

	for _, err := range sentinelErrors {
		require.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	t.Parallel()

	err := NewNotFoundError("User", "user-123")
	require.Error(t, err)
	assert.Equal(t, "User not found: user-123", err.Error())

	// Test Is method - custom NotFoundError.Is matches NotFoundError type
	assert.True(t, IsNotFound(err))
	// errors.Is with sentinel ErrUserNotFound returns false because different instances
	// The custom error's Is method matches the NotFoundError type, not specific sentinels
	assert.False(t, errors.Is(err, ErrUserNotFound))
	assert.False(t, IsConflict(err))
}

func TestConflictError(t *testing.T) {
	t.Parallel()

	err := NewConflictError("Organization", "slug", "my-org")
	require.Error(t, err)
	assert.Equal(t, "Organization already exists with slug=my-org", err.Error())

	// Test Is method
	assert.True(t, IsConflict(err))
	assert.False(t, IsNotFound(err))
}

func TestForbiddenError(t *testing.T) {
	t.Parallel()

	err := NewForbiddenError("delete", "bot", "not owner")
	require.Error(t, err)
	assert.Equal(t, "forbidden: delete bot: not owner", err.Error())

	// Test Is method
	assert.True(t, IsForbidden(err))
	assert.False(t, IsNotFound(err))
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	err := NewValidationError("email", "invalid format")
	require.Error(t, err)
	assert.Equal(t, "validation failed for email: invalid format", err.Error())

	// Test Is method
	assert.True(t, IsValidation(err))
	assert.False(t, IsNotFound(err))
}

func TestTenantMismatchError(t *testing.T) {
	t.Parallel()

	err := NewTenantMismatchError("org-1", "org-2", "Room", "room-123")
	require.Error(t, err)
	assert.Equal(t, "tenant mismatch: Room room-123 belongs to organization org-2, not org-1", err.Error())

	// Test Is method
	assert.True(t, IsTenantMismatch(err))
	assert.False(t, IsNotFound(err))
}

func TestOptimisticLockError(t *testing.T) {
	t.Parallel()

	err := NewOptimisticLockError("PluginConfiguration", "config-456")
	require.Error(t, err)
	assert.Equal(t, "optimistic lock conflict: PluginConfiguration config-456 was modified by another request", err.Error())

	// Test Is method
	assert.True(t, IsOptimisticLock(err))
	assert.False(t, IsNotFound(err))
}

func TestErrorIsChain(t *testing.T) {
	t.Parallel()

	// Test that custom errors can wrap sentinel errors
	baseErr := ErrUserNotFound
	wrappedErr := NewNotFoundError("User", "user-123")

	// Custom error Is should match NotFoundError type
	assert.True(t, IsNotFound(wrappedErr))

	// But not the specific sentinel (different instances)
	assert.False(t, errors.Is(wrappedErr, baseErr))

	// Test ForbiddenError wrapping
	forbiddenErr := NewForbiddenError("update", "room", "tenant mismatch")
	assert.True(t, IsForbidden(forbiddenErr))
	// errors.Is with sentinel ErrForbidden returns false because different instances
	assert.False(t, errors.Is(forbiddenErr, ErrForbidden))
}