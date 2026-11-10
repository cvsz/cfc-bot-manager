package repository

import (
	"context"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
)

// Pagination and filtering common types

type Pagination struct {
	Limit  int
	Offset int
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type Sort struct {
	Field string
	Order SortOrder
}

type Filter map[string]any // Generic filter, implementation-specific

// ListResult wraps paginated results with total count
type ListResult[T any] struct {
	Items      []T
	TotalCount int64
	Pagination Pagination
}

// OrganizationRepository defines organization persistence operations
type OrganizationRepository interface {
	Create(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	List(ctx context.Context, pagination Pagination, sort Sort, filter Filter) (ListResult[domain.Organization], error)
	Update(ctx context.Context, org *domain.Organization) error
	Delete(ctx context.Context, id domain.OrganizationID) error
	ExistsBySlug(ctx context.Context, slug string, excludeID *domain.OrganizationID) (bool, error)
}

// UserRepository defines user persistence operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, orgID domain.OrganizationID, email string) (*domain.User, error)
	List(ctx context.Context, orgID domain.OrganizationID, pagination Pagination, sort Sort, filter Filter) (ListResult[domain.User], error)
	Update(ctx context.Context, user *domain.User) error
	UpdateLastLogin(ctx context.Context, id domain.UserID, t time.Time) error
	Delete(ctx context.Context, id domain.UserID) error
	ExistsByEmail(ctx context.Context, orgID domain.OrganizationID, email string, excludeID *domain.UserID) (bool, error)
}

// MembershipRepository defines user-organization-role membership operations
type MembershipRepository interface {
	Create(ctx context.Context, m *domain.Membership) error
	GetByUserAndOrg(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) (*domain.Membership, error)
	ListByUser(ctx context.Context, userID domain.UserID) ([]domain.Membership, error)
	ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination Pagination, sort Sort) (ListResult[domain.Membership], error)
	UpdateRole(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID, roleID domain.RoleID) error
	Delete(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) error
	CountByOrgAndRole(ctx context.Context, orgID domain.OrganizationID, roleID domain.RoleID) (int64, error)
}

// RoleRepository defines role persistence operations
type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id domain.RoleID) (*domain.Role, error)
	GetByName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Role, error)
	ListByOrg(ctx context.Context, orgID domain.OrganizationID, includeSystem bool) ([]domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id domain.RoleID) error
	AddPermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error
	RemovePermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error
	ListPermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error)
	IsSystemRole(ctx context.Context, id domain.RoleID) (bool, error)
}

// PermissionRepository defines permission persistence operations
type PermissionRepository interface {
	Create(ctx context.Context, perm *domain.Permission) error
	GetByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error)
	GetByName(ctx context.Context, name string) (*domain.Permission, error)
	List(ctx context.Context) ([]domain.Permission, error)
	ListByResource(ctx context.Context, resource string) ([]domain.Permission, error)
	Update(ctx context.Context, perm *domain.Permission) error
	Delete(ctx context.Context, id domain.PermissionID) error
}

// SessionRepository defines session persistence operations
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

// RoomRepository defines room persistence operations
type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, id domain.RoomID) (*domain.Room, error)
	GetByOrgAndName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Room, error)
	ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination Pagination, sort Sort, filter Filter) (ListResult[domain.Room], error)
	ListByBot(ctx context.Context, botID domain.BotID) ([]domain.Room, error)
	Update(ctx context.Context, room *domain.Room) error
	UpdateStatus(ctx context.Context, id domain.RoomID, status string, lastSeen *time.Time) error
	UpdateConfig(ctx context.Context, id domain.RoomID, config domain.RoomConfig) error
	AssignBot(ctx context.Context, roomID domain.RoomID, botID *domain.BotID) error
	Delete(ctx context.Context, id domain.RoomID) error
	ExistsByName(ctx context.Context, orgID domain.OrganizationID, name string, excludeID *domain.RoomID) (bool, error)
}

// BotRepository defines bot persistence operations
type BotRepository interface {
	Create(ctx context.Context, bot *domain.Bot) error
	GetByID(ctx context.Context, id domain.BotID) (*domain.Bot, error)
	GetByUsername(ctx context.Context, orgID domain.OrganizationID, username string) (*domain.Bot, error)
	ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination Pagination, sort Sort, filter Filter) (ListResult[domain.Bot], error)
	ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.Bot, error)
	Update(ctx context.Context, bot *domain.Bot) error
	UpdateStatus(ctx context.Context, id domain.BotID, status string, lastError string, lastConnected *time.Time) error
	UpdateConfig(ctx context.Context, id domain.BotID, config domain.BotConfig) error
	AssignRoom(ctx context.Context, botID domain.BotID, roomID *domain.RoomID) error
	Delete(ctx context.Context, id domain.BotID) error
	ExistsByUsername(ctx context.Context, orgID domain.OrganizationID, username string, excludeID *domain.BotID) (bool, error)
}

// PluginRepository defines plugin persistence operations
type PluginRepository interface {
	Create(ctx context.Context, plugin *domain.Plugin) error
	GetByID(ctx context.Context, id domain.PluginID) (*domain.Plugin, error)
	GetByNumber(ctx context.Context, number int) (*domain.Plugin, error)
	List(ctx context.Context, pagination Pagination, sort Sort, filter Filter) (ListResult[domain.Plugin], error)
	ListEnabled(ctx context.Context) ([]domain.Plugin, error)
	Update(ctx context.Context, plugin *domain.Plugin) error
	UpdateCompatibility(ctx context.Context, id domain.PluginID, compat string) error
	Delete(ctx context.Context, id domain.PluginID) error
	ExistsByNumber(ctx context.Context, number int, excludeID *domain.PluginID) (bool, error)
}

// PluginConfigRepository defines plugin configuration persistence operations
type PluginConfigRepository interface {
	Create(ctx context.Context, config *domain.PluginConfiguration) error
	GetByID(ctx context.Context, id domain.PluginConfigID) (*domain.PluginConfiguration, error)
	GetByPluginAndRoom(ctx context.Context, pluginID domain.PluginID, roomID domain.RoomID) (*domain.PluginConfiguration, error)
	GetByPluginAndBot(ctx context.Context, pluginID domain.PluginID, botID domain.BotID) (*domain.PluginConfiguration, error)
	ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.PluginConfiguration, error)
	ListByBot(ctx context.Context, botID domain.BotID) ([]domain.PluginConfiguration, error)
	Update(ctx context.Context, config *domain.PluginConfiguration) error
	UpdateSettings(ctx context.Context, id domain.PluginConfigID, settings string, version int) error
	ToggleEnabled(ctx context.Context, id domain.PluginConfigID, enabled bool) error
	Delete(ctx context.Context, id domain.PluginConfigID) error
	DeleteByRoom(ctx context.Context, roomID domain.RoomID) error
	DeleteByBot(ctx context.Context, botID domain.BotID) error
}

// AuditLogRepository defines audit log persistence operations
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination Pagination, sort Sort, filter Filter) (ListResult[domain.AuditLog], error)
	ListByResource(ctx context.Context, orgID domain.OrganizationID, resourceType, resourceID string, pagination Pagination) (ListResult[domain.AuditLog], error)
	ListByActor(ctx context.Context, orgID domain.OrganizationID, actorUserID domain.UserID, pagination Pagination) (ListResult[domain.AuditLog], error)
	DeleteOlderThan(ctx context.Context, orgID domain.OrganizationID, olderThan time.Time) (int64, error)
}

// APIKeyRepository defines API key persistence operations
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id domain.APIKeyID) (*domain.APIKey, error)
	GetByPrefix(ctx context.Context, orgID domain.OrganizationID, prefix string) (*domain.APIKey, error)
	ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination Pagination) (ListResult[domain.APIKey], error)
	Update(ctx context.Context, key *domain.APIKey) error
	UpdateLastUsed(ctx context.Context, id domain.APIKeyID, t time.Time) error
	Revoke(ctx context.Context, id domain.APIKeyID) error
	Delete(ctx context.Context, id domain.APIKeyID) error
}

// BotConnectionRepository defines bot connection event persistence
type BotConnectionRepository interface {
	Create(ctx context.Context, conn *domain.BotConnection) error
	ListByBot(ctx context.Context, botID domain.BotID, pagination Pagination) (ListResult[domain.BotConnection], error)
	ListByRoom(ctx context.Context, roomID domain.RoomID, pagination Pagination) (ListResult[domain.BotConnection], error)
	DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error)
}

// UnitOfWork defines transaction boundary for multiple repositories
type UnitOfWork interface {
	Begin(ctx context.Context) (UnitOfWork, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	// Repository accessors within transaction
	Organizations() OrganizationRepository
	Users() UserRepository
	Memberships() MembershipRepository
	Roles() RoleRepository
	Permissions() PermissionRepository
	Sessions() SessionRepository
	Rooms() RoomRepository
	Bots() BotRepository
	Plugins() PluginRepository
	PluginConfigs() PluginConfigRepository
	AuditLogs() AuditLogRepository
	APIKeys() APIKeyRepository
	BotConnections() BotConnectionRepository
}

// RepositoryProvider provides access to repositories (transactional or not)
type RepositoryProvider interface {
	Organizations() OrganizationRepository
	Users() UserRepository
	Memberships() MembershipRepository
	Roles() RoleRepository
	Permissions() PermissionRepository
	Sessions() SessionRepository
	Rooms() RoomRepository
	Bots() BotRepository
	Plugins() PluginRepository
	PluginConfigs() PluginConfigRepository
	AuditLogs() AuditLogRepository
	APIKeys() APIKeyRepository
	BotConnections() BotConnectionRepository

	WithTransaction(ctx context.Context, fn func(UnitOfWork) error) error
}