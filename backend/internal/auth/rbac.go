package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
)

// RBAC errors
var (
	ErrPermissionDenied    = errors.New("permission denied")
	ErrResourceNotFound    = errors.New("resource not found")
	ErrTenantMismatch      = errors.New("tenant mismatch: resource belongs to different organization")
	ErrInsufficientRole    = errors.New("insufficient role for operation")
	ErrUnauthenticated     = errors.New("unauthenticated")
)

// Permission represents a granular permission
type Permission string

const (
	// Organization permissions
	PermOrgRead       Permission = "organizations:read"
	PermOrgWrite      Permission = "organizations:write"
	PermOrgDelete     Permission = "organizations:delete"

	// User permissions
	PermUserRead       Permission = "users:read"
	PermUserWrite      Permission = "users:write"
	PermUserDelete     Permission = "users:delete"
	PermUserManageRoles Permission = "users:manage_roles"

	// Role permissions
	PermRoleRead   Permission = "roles:read"
	PermRoleWrite  Permission = "roles:write"
	PermRoleDelete Permission = "roles:delete"

	// Room permissions
	PermRoomRead      Permission = "rooms:read"
	PermRoomWrite     Permission = "rooms:write"
	PermRoomDelete    Permission = "rooms:delete"
	PermRoomConfigure Permission = "rooms:configure"

	// Bot permissions
	PermBotRead      Permission = "bots:read"
	PermBotWrite     Permission = "bots:write"
	PermBotDelete    Permission = "bots:delete"
	PermBotControl   Permission = "bots:control"

	// Plugin permissions
	PermPluginRead       Permission = "plugins:read"
	PermPluginWrite      Permission = "plugins:write"
	PermPluginConfigure  Permission = "plugins:configure"

	// Audit permissions
	PermAuditRead Permission = "audit:read"

	// API Key permissions
	PermAPIKeyRead  Permission = "api_keys:read"
	PermAPIKeyWrite Permission = "api_keys:write"
)

// AllPermissions returns all defined permissions
func AllPermissions() []Permission {
	return []Permission{
		PermOrgRead, PermOrgWrite, PermOrgDelete,
		PermUserRead, PermUserWrite, PermUserDelete, PermUserManageRoles,
		PermRoleRead, PermRoleWrite, PermRoleDelete,
		PermRoomRead, PermRoomWrite, PermRoomDelete, PermRoomConfigure,
		PermBotRead, PermBotWrite, PermBotDelete, PermBotControl,
		PermPluginRead, PermPluginWrite, PermPluginConfigure,
		PermAuditRead,
		PermAPIKeyRead, PermAPIKeyWrite,
	}
}

// ResourceType represents the type of resource for authorization
type ResourceType string

const (
	ResourceOrganization ResourceType = "organizations"
	ResourceUser         ResourceType = "users"
	ResourceRole         ResourceType = "roles"
	ResourceRoom         ResourceType = "rooms"
	ResourceBot          ResourceType = "bots"
	ResourcePlugin       ResourceType = "plugins"
	ResourceAudit        ResourceType = "audit"
	ResourceAPIKey       ResourceType = "api_keys"
)

// Action represents an action on a resource
type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
	ActionManage Action = "manage"
	ActionControl Action = "control"
	ActionConfigure Action = "configure"
)

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	UserID       domain.UserID
	Organization domain.OrganizationID
	ResourceType ResourceType
	ResourceID   string
	Action       Action
	Permissions  []Permission // User's permissions
}

// RBACManager handles role-based access control
type RBACManager struct {
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
	permissionRepo repository.PermissionRepository
	membershipRepo repository.MembershipRepository
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	membershipRepo repository.MembershipRepository,
) *RBACManager {
	return &RBACManager{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		membershipRepo:  membershipRepo,
	}
}

// GetUserPermissions retrieves all permissions for a user in an organization
func (rm *RBACManager) GetUserPermissions(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) ([]Permission, error) {
	// Get user's membership in the organization
	membership, err := rm.membershipRepo.GetByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	// Get role's permissions
	perms, err := rm.roleRepo.ListPermissions(ctx, membership.RoleID)
	if err != nil {
		return nil, err
	}

	// Convert to Permission type
	permissions := make([]Permission, len(perms))
	for i, p := range perms {
		permissions[i] = Permission(p.Name)
	}

	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (rm *RBACManager) CheckPermission(ctx context.Context, check PermissionCheck) error {
	// Get user's permissions if not provided
	permissions := check.Permissions
	if permissions == nil {
		var err error
		permissions, err = rm.GetUserPermissions(ctx, check.UserID, check.Organization)
		if err != nil {
			return err
		}
	}

	// Build required permission string
	requiredPerm := Permission(fmt.Sprintf("%s:%s", check.ResourceType, check.Action))

	// Check if user has the required permission
	for _, p := range permissions {
		if p == requiredPerm || p == Permission(fmt.Sprintf("%s:*", check.ResourceType)) || p == "*" {
			return nil
		}
	}

	return ErrPermissionDenied
}

// CheckResourceAccess checks if a user can access a specific resource
func (rm *RBACManager) CheckResourceAccess(ctx context.Context, check PermissionCheck, resourceOrgID domain.OrganizationID) error {
	// Check tenant isolation
	if check.Organization != resourceOrgID {
		return ErrTenantMismatch
	}

	// Check permission
	return rm.CheckPermission(ctx, check)
}

// RequirePermission creates a middleware that requires a specific permission
func (rm *RBACManager) RequirePermission(resourceType ResourceType, action Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session from context
			session := GetSessionFromContext(r.Context())
			if session == nil {
				http.Error(w, ErrUnauthenticated.Error(), http.StatusUnauthorized)
				return
			}

			// Get user permissions
			permissions, err := rm.GetUserPermissions(r.Context(), session.UserID, session.Organization)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			requiredPerm := Permission(fmt.Sprintf("%s:%s", resourceType, action))
			hasPermission := false
			for _, p := range permissions {
				if p == requiredPerm || p == Permission(fmt.Sprintf("%s:*", resourceType)) || p == "*" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				http.Error(w, ErrPermissionDenied.Error(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireResourceAccess creates a middleware that requires access to a specific resource
func (rm *RBACManager) RequireResourceAccess(resourceType ResourceType, action Action, getResourceOrgID func(*http.Request) (domain.OrganizationID, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSessionFromContext(r.Context())
			if session == nil {
				http.Error(w, ErrUnauthenticated.Error(), http.StatusUnauthorized)
				return
			}

			resourceOrgID, err := getResourceOrgID(r)
			if err != nil {
				http.Error(w, "resource not found", http.StatusNotFound)
				return
			}

			check := PermissionCheck{
				UserID:       session.UserID,
				Organization: session.Organization,
				ResourceType: resourceType,
				Action:       action,
			}

			if err := rm.CheckResourceAccess(r.Context(), check, resourceOrgID); err != nil {
				if errors.Is(err, ErrTenantMismatch) {
					http.Error(w, "resource not found", http.StatusNotFound)
					return
				}
				http.Error(w, ErrPermissionDenied.Error(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// System roles
const (
	RoleOwner      = "owner"
	RoleAdmin      = "admin"
	RoleOperator   = "operator"
	RoleViewer     = "viewer"
	RoleAuditor    = "auditor"
)

// DefaultRolePermissions returns the default permissions for each system role
func DefaultRolePermissions() map[string][]Permission {
	return map[string][]Permission{
		RoleOwner: {
			PermOrgRead, PermOrgWrite, PermOrgDelete,
			PermUserRead, PermUserWrite, PermUserDelete, PermUserManageRoles,
			PermRoleRead, PermRoleWrite, PermRoleDelete,
			PermRoomRead, PermRoomWrite, PermRoomDelete, PermRoomConfigure,
			PermBotRead, PermBotWrite, PermBotDelete, PermBotControl,
			PermPluginRead, PermPluginWrite, PermPluginConfigure,
			PermAuditRead,
			PermAPIKeyRead, PermAPIKeyWrite,
		},
		RoleAdmin: {
			PermOrgRead, PermOrgWrite,
			PermUserRead, PermUserWrite, PermUserManageRoles,
			PermRoleRead, PermRoleWrite,
			PermRoomRead, PermRoomWrite, PermRoomConfigure,
			PermBotRead, PermBotWrite, PermBotControl,
			PermPluginRead, PermPluginWrite, PermPluginConfigure,
			PermAuditRead,
			PermAPIKeyRead, PermAPIKeyWrite,
		},
		RoleOperator: {
			PermOrgRead,
			PermUserRead,
			PermRoomRead, PermRoomConfigure,
			PermBotRead, PermBotControl,
			PermPluginRead, PermPluginConfigure,
		},
		RoleViewer: {
			PermOrgRead,
			PermUserRead,
			PermRoomRead,
			PermBotRead,
			PermPluginRead,
		},
		RoleAuditor: {
			PermOrgRead,
			PermUserRead,
			PermRoomRead,
			PermBotRead,
			PermPluginRead,
			PermAuditRead,
		},
	}
}