-- Initial schema for CFC Bot Manager
-- Version: 1
-- Description: Core tables for organizations, users, roles, permissions, rooms, bots, plugins, audit logs

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Organizations (Tenants)
-- ============================================================
CREATE TABLE organizations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_slug ON organizations(slug);

-- ============================================================
-- Users
-- ============================================================
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email               VARCHAR(255) NOT NULL,
    display_name        VARCHAR(255) NOT NULL,
    password_hash       VARCHAR(255) NOT NULL, -- Argon2id
    mfa_secret          VARCHAR(255),          -- Encrypted TOTP secret
    mfa_enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, email)
);

CREATE INDEX idx_users_org_email ON users(organization_id, email);
CREATE INDEX idx_users_org_active ON users(organization_id, is_active);

-- ============================================================
-- Roles
-- ============================================================
CREATE TABLE roles (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    is_system           BOOLEAN NOT NULL DEFAULT FALSE, -- System roles cannot be deleted
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, name)
);

CREATE INDEX idx_roles_org ON roles(organization_id);

-- ============================================================
-- Permissions
-- ============================================================
CREATE TABLE permissions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                VARCHAR(100) NOT NULL UNIQUE, -- e.g., "bots:read", "rooms:write"
    description         TEXT,
    resource            VARCHAR(50) NOT NULL,       -- e.g., "bots", "rooms", "plugins"
    action              VARCHAR(50) NOT NULL,       -- e.g., "read", "write", "delete", "configure"
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_resource ON permissions(resource);

-- ============================================================
-- Role-Permission Mapping (Many-to-Many)
-- ============================================================
CREATE TABLE role_permissions (
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id   UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================================
-- User-Organization-Role Membership
-- ============================================================
CREATE TABLE memberships (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id             UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, organization_id)
);

CREATE INDEX idx_memberships_user ON memberships(user_id);
CREATE INDEX idx_memberships_org ON memberships(organization_id);
CREATE INDEX idx_memberships_role ON memberships(role_id);

-- ============================================================
-- Sessions
-- ============================================================
CREATE TABLE sessions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token_hash          VARCHAR(255) NOT NULL UNIQUE, -- Hashed session token
    csrf_token          VARCHAR(255) NOT NULL,        -- For double-submit CSRF
    expires_at          TIMESTAMPTZ NOT NULL,
    revoked_at          TIMESTAMPTZ,
    user_agent          TEXT,
    ip_address          INET,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_org ON sessions(organization_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at) WHERE revoked_at IS NULL;

-- ============================================================
-- Rooms
-- ============================================================
CREATE TABLE rooms (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    access_key_hash     VARCHAR(255) NOT NULL, -- Hashed room access key
    bot_id              UUID REFERENCES bots(id) ON DELETE SET NULL, -- Forward ref, will add FK later
    language            VARCHAR(50) NOT NULL DEFAULT 'en_US.UTF-8',
    is_paused           BOOLEAN NOT NULL DEFAULT FALSE,
    status              VARCHAR(50) NOT NULL DEFAULT 'disconnected', -- connected, disconnected, error, initializing
    last_seen_at        TIMESTAMPTZ,
    config              JSONB NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, name)
);

CREATE INDEX idx_rooms_org ON rooms(organization_id);
CREATE INDEX idx_rooms_bot ON rooms(bot_id);
CREATE INDEX idx_rooms_status ON rooms(status);

-- ============================================================
-- Bots
-- ============================================================
CREATE TABLE bots (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                    VARCHAR(255) NOT NULL,
    username                VARCHAR(255) NOT NULL, -- Camfrog bot username
    password_hash           VARCHAR(255) NOT NULL, -- Hashed Camfrog password
    status                  VARCHAR(50) NOT NULL DEFAULT 'stopped', -- running, stopped, paused, error, starting
    room_id                 UUID REFERENCES rooms(id) ON DELETE SET NULL,
    last_connected_at       TIMESTAMPTZ,
    last_error              TEXT,
    config                  JSONB NOT NULL DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, username)
);

CREATE INDEX idx_bots_org ON bots(organization_id);
CREATE INDEX idx_bots_room ON bots(room_id);
CREATE INDEX idx_bots_status ON bots(status);

-- Add the forward-reference FK for rooms.bot_id
ALTER TABLE rooms ADD CONSTRAINT fk_rooms_bot_id FOREIGN KEY (bot_id) REFERENCES bots(id) ON DELETE SET NULL;

-- ============================================================
-- Plugins
-- ============================================================
CREATE TABLE plugins (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plugin_number           INTEGER NOT NULL UNIQUE, -- Camfrog plugin ID (1-16, except 9,10)
    name                    VARCHAR(100) NOT NULL,
    description             TEXT,
    version                 VARCHAR(50) NOT NULL,
    compatibility           VARCHAR(30) NOT NULL DEFAULT 'UNVERIFIED', -- VERIFIED, PARTIALLY_VERIFIED, UNVERIFIED, UNSUPPORTED, DISABLED
    capability_flags        JSONB NOT NULL DEFAULT '[]', -- Array of capability strings
    config_schema           JSONB NOT NULL DEFAULT '{}', -- JSON Schema for configuration
    is_enabled              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plugins_number ON plugins(plugin_number);
CREATE INDEX idx_plugins_compat ON plugins(compatibility);
CREATE INDEX idx_plugins_enabled ON plugins(is_enabled);

-- ============================================================
-- Plugin Configurations (per room/bot)
-- ============================================================
CREATE TABLE plugin_configurations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plugin_id           UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    room_id             UUID REFERENCES rooms(id) ON DELETE CASCADE,
    bot_id              UUID REFERENCES bots(id) ON DELETE CASCADE,
    settings            JSONB NOT NULL DEFAULT '{}', -- Raw settings from Camfrog multipart form
    is_enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    version             INTEGER NOT NULL DEFAULT 1, -- Optimistic locking
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_plugin_config_scope CHECK (
        (room_id IS NOT NULL AND bot_id IS NULL) OR
        (room_id IS NULL AND bot_id IS NOT NULL)
    )
);

CREATE INDEX idx_plugin_configs_plugin ON plugin_configurations(plugin_id);
CREATE INDEX idx_plugin_configs_room ON plugin_configurations(room_id);
CREATE INDEX idx_plugin_configs_bot ON plugin_configurations(bot_id);
CREATE UNIQUE INDEX uq_plugin_config_room ON plugin_configurations(plugin_id, room_id) WHERE room_id IS NOT NULL;
CREATE UNIQUE INDEX uq_plugin_config_bot ON plugin_configurations(plugin_id, bot_id) WHERE bot_id IS NOT NULL;

-- ============================================================
-- Audit Logs
-- ============================================================
CREATE TABLE audit_logs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_type          VARCHAR(30) NOT NULL, -- user, system, api_key
    action              VARCHAR(100) NOT NULL, -- e.g., bot.create, plugin.enable, user.delete
    resource_type       VARCHAR(50) NOT NULL, -- bot, room, plugin, user, role, organization
    resource_id         VARCHAR(255) NOT NULL,
    old_value           JSONB,
    new_value           JSONB,
    ip_address          INET,
    user_agent          TEXT,
    correlation_id      VARCHAR(255) NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_org_created ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_user_id);
CREATE INDEX idx_audit_logs_correlation ON audit_logs(correlation_id);

-- ============================================================
-- API Keys
-- ============================================================
CREATE TABLE api_keys (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    key_hash            VARCHAR(255) NOT NULL UNIQUE, -- Hashed API key
    prefix              VARCHAR(20) NOT NULL,         -- First chars for identification (e.g., "cfc_abcd12")
    permissions         JSONB NOT NULL DEFAULT '[]',  -- Array of permission names
    expires_at          TIMESTAMPTZ,
    last_used_at        TIMESTAMPTZ,
    revoked_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_org ON api_keys(organization_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(prefix);
CREATE INDEX idx_api_keys_revoked ON api_keys(revoked_at) WHERE revoked_at IS NULL;

-- ============================================================
-- Bot Connections (Connection lifecycle events)
-- ============================================================
CREATE TABLE bot_connections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bot_id          UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    room_id         UUID REFERENCES rooms(id) ON DELETE SET NULL,
    event_type      VARCHAR(50) NOT NULL, -- connected, disconnected, reconnecting, error
    message         TEXT,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bot_connections_bot ON bot_connections(bot_id);
CREATE INDEX idx_bot_connections_room ON bot_connections(room_id);
CREATE INDEX idx_bot_connections_created ON bot_connections(created_at DESC);

-- ============================================================
-- Updated_at Triggers
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_memberships_updated_at BEFORE UPDATE ON memberships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bots_updated_at BEFORE UPDATE ON bots FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plugin_configurations_updated_at BEFORE UPDATE ON plugin_configurations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Default System Permissions
-- ============================================================
INSERT INTO permissions (name, description, resource, action) VALUES
    -- Organization permissions
    ('organizations:read', 'View organization details', 'organizations', 'read'),
    ('organizations:write', 'Modify organization settings', 'organizations', 'write'),
    ('organizations:delete', 'Delete organization', 'organizations', 'delete'),
    -- User permissions
    ('users:read', 'View users in organization', 'users', 'read'),
    ('users:write', 'Create/modify users', 'users', 'write'),
    ('users:delete', 'Delete users', 'users', 'delete'),
    ('users:manage_roles', 'Assign/remove roles from users', 'users', 'manage_roles'),
    -- Role permissions
    ('roles:read', 'View roles', 'roles', 'read'),
    ('roles:write', 'Create/modify roles', 'roles', 'write'),
    ('roles:delete', 'Delete roles', 'roles', 'delete'),
    -- Room permissions
    ('rooms:read', 'View rooms', 'rooms', 'read'),
    ('rooms:write', 'Create/modify rooms', 'rooms', 'write'),
    ('rooms:delete', 'Delete rooms', 'rooms', 'delete'),
    ('rooms:configure', 'Configure room settings', 'rooms', 'configure'),
    -- Bot permissions
    ('bots:read', 'View bots', 'bots', 'read'),
    ('bots:write', 'Create/modify bots', 'bots', 'write'),
    ('bots:delete', 'Delete bots', 'bots', 'delete'),
    ('bots:control', 'Start/stop/pause bots', 'bots', 'control'),
    -- Plugin permissions
    ('plugins:read', 'View plugins and configurations', 'plugins', 'read'),
    ('plugins:write', 'Create/modify plugin configurations', 'plugins', 'write'),
    ('plugins:configure', 'Enable/disable plugins, change settings', 'plugins', 'configure'),
    -- Audit permissions
    ('audit:read', 'View audit logs', 'audit', 'read'),
    -- API Key permissions
    ('api_keys:read', 'View API keys', 'api_keys', 'read'),
    ('api_keys:write', 'Create/revoke API keys', 'api_keys', 'write')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- Default System Roles (created per organization via application logic)
-- ============================================================
-- Note: System roles are created per organization on first setup
-- Owner: all permissions
-- Admin: most permissions except organizations:delete, users:delete (self)
-- Operator: bots:control, rooms:configure, plugins:configure, read permissions
-- Viewer: read permissions only
-- Auditor: audit:read only