package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors for common domain violations
var (
	// Not found errors
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrRoleNotFound         = errors.New("role not found")
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrSessionNotFound      = errors.New("session not found")
	ErrRoomNotFound         = errors.New("room not found")
	ErrBotNotFound          = errors.New("bot not found")
	ErrPluginNotFound       = errors.New("plugin not found")
	ErrPluginConfigNotFound = errors.New("plugin configuration not found")
	ErrAuditLogNotFound     = errors.New("audit log not found")
	ErrAPIKeyNotFound       = errors.New("API key not found")
	ErrMembershipNotFound   = errors.New("membership not found")

	// Conflict/duplicate errors
	ErrOrganizationSlugExists = errors.New("organization slug already exists")
	ErrUserEmailExists        = errors.New("user email already exists in organization")
	ErrRoleNameExists         = errors.New("role name already exists in organization")
	ErrRoomNameExists         = errors.New("room name already exists in organization")
	ErrBotUsernameExists      = errors.New("bot username already exists in organization")
	ErrPluginNumberExists     = errors.New("plugin number already exists")
	ErrAPIKeyPrefixExists     = errors.New("API key prefix already exists")

	// Authorization errors
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden: insufficient permissions")
	ErrTenantMismatch     = errors.New("tenant mismatch: resource belongs to different organization")
	ErrInsufficientRole   = errors.New("insufficient role for operation")
	ErrMFARequired        = errors.New("multi-factor authentication required")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRevoked     = errors.New("session revoked")
	ErrCSRFInvalid        = errors.New("invalid CSRF token")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Validation errors
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidOrganization  = errors.New("invalid organization")
	ErrInvalidUser          = errors.New("invalid user")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidSession       = errors.New("invalid session")
	ErrInvalidRoom          = errors.New("invalid room")
	ErrInvalidBot           = errors.New("invalid bot")
	ErrInvalidPlugin        = errors.New("invalid plugin")
	ErrInvalidPluginConfig  = errors.New("invalid plugin configuration")
	ErrInvalidAPIKey        = errors.New("invalid API key")
	ErrInvalidPagination    = errors.New("invalid pagination parameters")
	ErrInvalidSort          = errors.New("invalid sort parameters")
	ErrInvalidFilter        = errors.New("invalid filter parameters")

	// State errors
	ErrBotAlreadyRunning   = errors.New("bot is already running")
	ErrBotNotRunning       = errors.New("bot is not running")
	ErrRoomAlreadyHasBot   = errors.New("room already has a bot assigned")
	ErrBotAlreadyHasRoom   = errors.New("bot already has a room assigned")
	ErrPluginNotVerified   = errors.New("plugin is not verified for use")
	ErrPluginDisabled      = errors.New("plugin is disabled")
	ErrSystemRoleImmutable = errors.New("system roles cannot be modified or deleted")
	ErrLastAdminRole       = errors.New("cannot remove last admin role from organization")

	// External/upstream errors
	ErrCamfrogUnavailable   = errors.New("Camfrog service unavailable")
	ErrCamfrogAuthFailed    = errors.New("Camfrog authentication failed")
	ErrCamfrogTimeout       = errors.New("Camfrog request timed out")
	ErrCamfrogInvalidResponse = errors.New("Camfrog returned invalid response")
	ErrCamfrogRateLimited   = errors.New("Camfrog rate limit exceeded")

	// Persistence errors
	ErrOptimisticLock = errors.New("optimistic lock conflict: resource was modified")
	ErrTransaction    = errors.New("transaction failed")
)

// Error types with context

type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

func (e NotFoundError) Is(target error) bool {
	var nf NotFoundError
	return errors.As(target, &nf)
}

type ConflictError struct {
	Resource string
	Field    string
	Value    string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("%s already exists with %s=%s", e.Resource, e.Field, e.Value)
}

func (e ConflictError) Is(target error) bool {
	var ce ConflictError
	return errors.As(target, &ce)
}

type ForbiddenError struct {
	Action   string
	Resource string
	Reason   string
}

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s %s: %s", e.Action, e.Resource, e.Reason)
}

func (e ForbiddenError) Is(target error) bool {
	var fe ForbiddenError
	return errors.As(target, &fe)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

func (e ValidationError) Is(target error) bool {
	var ve ValidationError
	return errors.As(target, &ve)
}

type TenantMismatchError struct {
	ExpectedOrgID string
	ActualOrgID   string
	Resource      string
	ResourceID    string
}

func (e TenantMismatchError) Error() string {
	return fmt.Sprintf("tenant mismatch: %s %s belongs to organization %s, not %s", e.Resource, e.ResourceID, e.ActualOrgID, e.ExpectedOrgID)
}

func (e TenantMismatchError) Is(target error) bool {
	var tme TenantMismatchError
	return errors.As(target, &tme)
}

type OptimisticLockError struct {
	Resource string
	ID       string
}

func (e OptimisticLockError) Error() string {
	return fmt.Sprintf("optimistic lock conflict: %s %s was modified by another request", e.Resource, e.ID)
}

func (e OptimisticLockError) Is(target error) bool {
	var ole OptimisticLockError
	return errors.As(target, &ole)
}

// Helper constructors

func NewNotFoundError(resource, id string) error {
	return NotFoundError{Resource: resource, ID: id}
}

func NewConflictError(resource, field, value string) error {
	return ConflictError{Resource: resource, Field: field, Value: value}
}

func NewForbiddenError(action, resource, reason string) error {
	return ForbiddenError{Action: action, Resource: resource, Reason: reason}
}

func NewValidationError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}

func NewTenantMismatchError(expectedOrgID, actualOrgID, resource, resourceID string) error {
	return TenantMismatchError{
		ExpectedOrgID: expectedOrgID,
		ActualOrgID:   actualOrgID,
		Resource:      resource,
		ResourceID:    resourceID,
	}
}

func NewOptimisticLockError(resource, id string) error {
	return OptimisticLockError{Resource: resource, ID: id}
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	var nf NotFoundError
	return errors.As(err, &nf)
}

// IsConflict checks if error is a conflict error
func IsConflict(err error) bool {
	var ce ConflictError
	return errors.As(err, &ce)
}

// IsForbidden checks if error is a forbidden error
func IsForbidden(err error) bool {
	var fe ForbiddenError
	return errors.As(err, &fe)
}

// IsValidation checks if error is a validation error
func IsValidation(err error) bool {
	var ve ValidationError
	return errors.As(err, &ve)
}

// IsTenantMismatch checks if error is a tenant mismatch error
func IsTenantMismatch(err error) bool {
	var tme TenantMismatchError
	return errors.As(err, &tme)
}

// IsOptimisticLock checks if error is an optimistic lock error
func IsOptimisticLock(err error) bool {
	var ole OptimisticLockError
	return errors.As(err, &ole)
}