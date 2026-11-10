-- Rollback for initial schema
-- Version: 1

-- Drop triggers first
DROP TRIGGER IF EXISTS update_plugin_configurations_updated_at ON plugin_configurations;
DROP TRIGGER IF EXISTS update_plugins_updated_at ON plugins;
DROP TRIGGER IF EXISTS update_bots_updated_at ON bots;
DROP TRIGGER IF EXISTS update_rooms_updated_at ON rooms;
DROP TRIGGER IF EXISTS update_memberships_updated_at ON memberships;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS bot_connections;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS plugin_configurations;
DROP TABLE IF EXISTS plugins;
DROP TABLE IF EXISTS bots;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

-- Drop extensions (optional, may be used by other schemas)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";