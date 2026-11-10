package domain

import (
	"time"

	"github.com/google/uuid"
)

// ID types for type safety
type (
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	RoleID         uuid.UUID
	PermissionID   uuid.UUID
	SessionID      uuid.UUID
	RoomID         uuid.UUID
	BotID          uuid.UUID
	PluginID       uuid.UUID
	PluginConfigID uuid.UUID
	AuditLogID     uuid.UUID
	APIKeyID       uuid.UUID
)

// NewID generates a new UUID for the given type
func NewOrganizationID() OrganizationID { return OrganizationID(uuid.New()) }
func NewUserID() UserID                 { return UserID(uuid.New()) }
func NewRoleID() RoleID                 { return RoleID(uuid.New()) }
func NewPermissionID() PermissionID     { return PermissionID(uuid.New()) }
func NewSessionID() SessionID           { return SessionID(uuid.New()) }
func NewRoomID() RoomID                 { return RoomID(uuid.New()) }
func NewBotID() BotID                   { return BotID(uuid.New()) }
func NewPluginID() PluginID             { return PluginID(uuid.New()) }
func NewPluginConfigID() PluginConfigID { return PluginConfigID(uuid.New()) }
func NewAuditLogID() AuditLogID         { return AuditLogID(uuid.New()) }
func NewAPIKeyID() APIKeyID             { return APIKeyID(uuid.New()) }

// ParseID functions for input validation
func ParseOrganizationID(s string) (OrganizationID, error) { id, err := uuid.Parse(s); return OrganizationID(id), err }
func ParseUserID(s string) (UserID, error)                 { id, err := uuid.Parse(s); return UserID(id), err }
func ParseRoleID(s string) (RoleID, error)                 { id, err := uuid.Parse(s); return RoleID(id), err }
func ParsePermissionID(s string) (PermissionID, error)     { id, err := uuid.Parse(s); return PermissionID(id), err }
func ParseSessionID(s string) (SessionID, error)           { id, err := uuid.Parse(s); return SessionID(id), err }
func ParseRoomID(s string) (RoomID, error)                 { id, err := uuid.Parse(s); return RoomID(id), err }
func ParseBotID(s string) (BotID, error)                   { id, err := uuid.Parse(s); return BotID(id), err }
func ParsePluginID(s string) (PluginID, error)             { id, err := uuid.Parse(s); return PluginID(id), err }
func ParsePluginConfigID(s string) (PluginConfigID, error) { id, err := uuid.Parse(s); return PluginConfigID(id), err }
func ParseAuditLogID(s string) (AuditLogID, error)         { id, err := uuid.Parse(s); return AuditLogID(id), err }
func ParseAPIKeyID(s string) (APIKeyID, error)             { id, err := uuid.Parse(s); return APIKeyID(id), err }

// String() methods for logging/debugging
func (id OrganizationID) String() string { return uuid.UUID(id).String() }
func (id UserID) String() string         { return uuid.UUID(id).String() }
func (id RoleID) String() string         { return uuid.UUID(id).String() }
func (id PermissionID) String() string   { return uuid.UUID(id).String() }
func (id SessionID) String() string      { return uuid.UUID(id).String() }
func (id RoomID) String() string         { return uuid.UUID(id).String() }
func (id BotID) String() string          { return uuid.UUID(id).String() }
func (id PluginID) String() string       { return uuid.UUID(id).String() }
func (id PluginConfigID) String() string { return uuid.UUID(id).String() }
func (id AuditLogID) String() string     { return uuid.UUID(id).String() }
func (id APIKeyID) String() string       { return uuid.UUID(id).String() }

// Organization represents a tenant/organization
type Organization struct {
	ID          OrganizationID `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Slug        string         `json:"slug" db:"slug"`
	Description string         `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// User represents a system user
type User struct {
	ID           UserID    `json:"id" db:"id"`
	Organization OrganizationID `json:"organization_id" db:"organization_id"`
	Email        string    `json:"email" db:"email"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never serialize
	MFASecret    string    `json:"-" db:"mfa_secret"`    // Never serialize
	MFAEnabled   bool      `json:"mfa_enabled" db:"mfa_enabled"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Membership links users to organizations with roles
type Membership struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         UserID    `json:"user_id" db:"user_id"`
	OrganizationID OrganizationID `json:"organization_id" db:"organization_id"`
	RoleID         RoleID    `json:"role_id" db:"role_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Role defines a set of permissions
type Role struct {
	ID             RoleID `json:"id" db:"id"`
	OrganizationID OrganizationID `json:"organization_id" db:"organization_id"`
	Name           string `json:"name" db:"name"`
	Description    string `json:"description,omitempty" db:"description"`
	IsSystem       bool   `json:"is_system" db:"is_system"` // System roles cannot be deleted
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents a granular access right
type Permission struct {
	ID          PermissionID `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"` // e.g., "bots:read", "rooms:write", "plugins:configure"
	Description string       `json:"description,omitempty" db:"description"`
	Resource    string       `json:"resource" db:"resource"` // e.g., "bots", "rooms", "plugins", "users", "audit"
	Action      string       `json:"action" db:"action"`     // e.g., "read", "write", "delete", "configure"
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
}

// RolePermission links roles to permissions
type RolePermission struct {
	RoleID       RoleID       `json:"role_id" db:"role_id"`
	PermissionID PermissionID `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
}

// Session represents an authenticated user session
type Session struct {
	ID           SessionID `json:"id" db:"id"`
	UserID       UserID    `json:"user_id" db:"user_id"`
	Organization OrganizationID `json:"organization_id" db:"organization_id"`
	TokenHash    string    `json:"-" db:"token_hash"` // Hashed token, never serialize
	CSRFToken    string    `json:"-" db:"csrf_token"` // For double-submit CSRF
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	UserAgent    string    `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress    string    `json:"ip_address,omitempty" db:"ip_address"`
}

// Room represents a Camfrog room managed by a bot
type Room struct {
	ID             RoomID `json:"id" db:"id"`
	OrganizationID OrganizationID `json:"organization_id" db:"organization_id"`
	Name           string `json:"name" db:"name"`
	AccessKeyHash  string `json:"-" db:"access_key_hash"` // Hashed, never expose raw key
	BotID          *BotID `json:"bot_id,omitempty" db:"bot_id"`
	Language       string `json:"language" db:"language"` // e.g., "en_US.UTF-8"
	IsPaused       bool   `json:"is_paused" db:"is_paused"`
	Status         string `json:"status" db:"status"` // "connected", "disconnected", "error", "initializing"
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	Config         RoomConfig `json:"config" db:"config"` // JSONB
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// RoomConfig holds room-specific configuration
type RoomConfig struct {
	AutoJoin   bool   `json:"auto_join"`
	MaxUsers   int    `json:"max_users,omitempty"`
	WelcomeMsg string `json:"welcome_message,omitempty"`
}

// Bot represents a Camfrog bot instance
type Bot struct {
	ID             BotID `json:"id" db:"id"`
	OrganizationID OrganizationID `json:"organization_id" db:"organization_id"`
	Name           string `json:"name" db:"name"`
	Username       string `json:"username" db:"username"` // Camfrog bot username
	PasswordHash   string `json:"-" db:"password_hash"`   // Hashed, never expose
	Status         string `json:"status" db:"status"`     // "running", "stopped", "paused", "error", "starting"
	RoomID         *RoomID `json:"room_id,omitempty" db:"room_id"`
	LastConnectedAt *time.Time `json:"last_connected_at,omitempty" db:"last_connected_at"`
	LastError      string `json:"last_error,omitempty" db:"last_error"`
	Config         BotConfig `json:"config" db:"config"` // JSONB
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// BotConfig holds bot-specific configuration
type BotConfig struct {
	AutoReconnect bool   `json:"auto_reconnect"`
	LogLevel      string `json:"log_level"` // "debug", "info", "warn", "error"
	ProxyURL      string `json:"proxy_url,omitempty"`
}

// Plugin represents a Camfrog plugin definition
type Plugin struct {
	ID              PluginID `json:"id" db:"id"`
	PluginNumber    int      `json:"plugin_number" db:"plugin_number"` // Camfrog plugin ID (1-16, except 9,10)
	Name            string   `json:"name" db:"name"`
	Description     string   `json:"description,omitempty" db:"description"`
	Version         string   `json:"version" db:"version"`
	Compatibility   string   `json:"compatibility" db:"compatibility"` // "VERIFIED", "PARTIALLY_VERIFIED", "UNVERIFIED", "UNSUPPORTED", "DISABLED"
	CapabilityFlags []string `json:"capability_flags" db:"capability_flags"` // JSONB array
	ConfigSchema    string   `json:"config_schema" db:"config_schema"` // JSON Schema as string
	IsEnabled       bool     `json:"is_enabled" db:"is_enabled"` // Global enable/disable
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// PluginConfiguration holds plugin settings for a specific room/bot
type PluginConfiguration struct {
	ID              PluginConfigID  `json:"id" db:"id"`
	OrganizationID  OrganizationID  `json:"organization_id" db:"organization_id"`
	PluginID        PluginID        `json:"plugin_id" db:"plugin_id"`
	RoomID          *RoomID         `json:"room_id,omitempty" db:"room_id"`
	BotID           *BotID          `json:"bot_id,omitempty" db:"bot_id"`
	Settings        string          `json:"settings" db:"settings"` // JSON string (raw from Camfrog multipart form)
	IsEnabled       bool            `json:"is_enabled" db:"is_enabled"`
	Version         int             `json:"version" db:"version"` // For optimistic locking
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// AuditLog records privileged operations
type AuditLog struct {
	ID             AuditLogID `json:"id" db:"id"`
	OrganizationID OrganizationID `json:"organization_id" db:"organization_id"`
	ActorUserID    *UserID    `json:"actor_user_id,omitempty" db:"actor_user_id"`
	ActorType      string     `json:"actor_type" db:"actor_type"` // "user", "system", "api_key"
	Action         string     `json:"action" db:"action"`         // e.g., "bot.create", "plugin.enable", "user.delete"
	ResourceType   string     `json:"resource_type" db:"resource_type"` // "bot", "room", "plugin", "user", "role"
	ResourceID     string     `json:"resource_id" db:"resource_id"`
	OldValue       string     `json:"old_value,omitempty" db:"old_value"` // JSON
	NewValue       string     `json:"new_value,omitempty" db:"new_value"` // JSON
	IPAddress      string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      string     `json:"user_agent,omitempty" db:"user_agent"`
	CorrelationID  string     `json:"correlation_id" db:"correlation_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// APIKey represents a machine-to-machine credential
type APIKey struct {
	ID             APIKeyID `json:"id" db:"id"`
	OrganizationID OrganizationID `json:"organization_id" db:"organization_id"`
	Name           string `json:"name" db:"name"`
	KeyHash        string `json:"-" db:"key_hash"` // Hashed, never expose raw key
	Prefix         string `json:"prefix" db:"prefix"` // First 8 chars for identification
	Permissions    []string `json:"permissions" db:"permissions"` // JSONB array of permission names
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// BotConnection represents a bot's connection lifecycle event
type BotConnection struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BotID     BotID     `json:"bot_id" db:"bot_id"`
	RoomID    *RoomID   `json:"room_id,omitempty" db:"room_id"`
	EventType string    `json:"event_type" db:"event_type"` // "connected", "disconnected", "reconnecting", "error"
	Message   string    `json:"message,omitempty" db:"message"`
	Metadata  string    `json:"metadata,omitempty" db:"metadata"` // JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}