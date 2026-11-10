// API types matching backend domain models
export type UUID = string;

export interface Organization {
  id: UUID;
  name: string;
  slug: string;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: UUID;
  organization_id: UUID;
  email: string;
  display_name: string | null;
  password_hash: string;
  is_active: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: UUID;
  organization_id: UUID;
  name: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: UUID;
  name: string;
  resource: string;
  action: string;
  description: string | null;
  created_at: string;
}

export interface Session {
  id: UUID;
  user_id: UUID;
  organization_id: UUID;
  token_hash: string;
  csrf_token: string;
  expires_at: string;
  revoked_at: string | null;
  created_at: string;
}

export interface Room {
  id: UUID;
  organization_id: UUID;
  bot_id: UUID | null;
  name: string;
  access_key_hash: string;
  language: string;
  status: RoomStatus;
  config: RoomConfig;
  last_seen_at: string | null;
  created_at: string;
  updated_at: string;
}

export type RoomStatus = 'connected' | 'disconnected' | 'reconnecting' | 'error';

export interface RoomConfig {
  auto_reconnect: boolean;
  log_level: 'debug' | 'info' | 'warn' | 'error';
  proxy_url: string | null;
}

export interface Bot {
  id: UUID;
  organization_id: UUID;
  room_id: UUID | null;
  name: string;
  username: string;
  password_hash: string;
  status: BotStatus;
  config: BotConfig;
  last_error: string | null;
  last_connected_at: string | null;
  created_at: string;
  updated_at: string;
}

export type BotStatus = 'stopped' | 'starting' | 'running' | 'stopping' | 'error';

export interface BotStatusPayload {
  uptime: string;
  reason: string;
  status: string;
  raw?: unknown;
}

export interface BotConfig {
  auto_reconnect: boolean;
  log_level: 'debug' | 'info' | 'warn' | 'error';
  proxy_url: string | null;
}

export interface Plugin {
  id: UUID;
  plugin_number: number;
  name: string;
  description: string | null;
  version: string;
  compatibility: PluginCompatibility;
  capability_flags: string[];
  config_schema: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export type PluginCompatibility = 
  | 'VERIFIED' 
  | 'PARTIALLY_VERIFIED' 
  | 'UNVERIFIED' 
  | 'UNSUPPORTED' 
  | 'DISABLED';

export interface PluginConfiguration {
  id: UUID;
  organization_id: UUID;
  plugin_id: UUID;
  room_id: UUID | null;
  bot_id: UUID | null;
  settings: string; // JSON string
  is_enabled: boolean;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface AuditLog {
  id: UUID;
  organization_id: UUID;
  actor_user_id: UUID | null;
  actor_type: 'user' | 'system' | 'api_key';
  action: string;
  resource_type: string;
  resource_id: string;
  old_value: string | null;
  new_value: string | null;
  ip_address: string | null;
  user_agent: string | null;
  correlation_id: string | null;
  created_at: string;
}

export interface APIKey {
  id: UUID;
  organization_id: UUID;
  name: string;
  prefix: string;
  key_hash: string;
  permissions: string[];
  last_used_at: string | null;
  revoked_at: string | null;
  expires_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface BotConnection {
  id: UUID;
  bot_id: UUID;
  room_id: UUID | null;
  event_type: string;
  metadata: Record<string, unknown> | null;
  created_at: string;
}

// API Response types
export interface ListResponse<T> {
  items: T[];
  total_count: number;
  pagination: {
    limit: number;
    offset: number;
  };
}

export interface ErrorResponse {
  error: string;
  message: string;
  details?: Record<string, string[]>;
}

// Pagination
export interface PaginationParams {
  limit?: number;
  offset?: number;
  sort_field?: string;
  sort_order?: 'asc' | 'desc';
  filter?: Record<string, unknown>;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
  totp_code?: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
  requires_mfa: boolean;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface MFASetupResponse {
  secret: string;
  qr_code_url: string;
  backup_codes: string[];
}

export interface MFAVerifyRequest {
  code: string;
}

export interface MFAVerifyResponse {
  verified: boolean;
}