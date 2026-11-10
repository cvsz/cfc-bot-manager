# CFC Bot Manager - Adapter Implementation & Frontend-Backend Integration Plan

## Overview
Implement the Camfrog Adapter (JSON RPC + Multipart transports) and wire the React frontend to the Go backend API.

## Current State
- ✅ Backend: Go 1.25, PostgreSQL, JWT auth, tenant isolation, transactions
- ✅ Frontend: React 18 + TypeScript + Vite + Tailwind (structure complete)
- ✅ Camfrog Adapter interfaces defined (adapter.go, rpc.go, transports, capability registry)
- ⏳ Adapter implementation: Need concrete implementations
- ⏳ Frontend-backend integration: API wiring, auth flow, real data

---

## Phase 1: Camfrog Adapter Implementation

### 1.1 Concrete Adapter Implementation
Create `backend/internal/camfrog/adapter_impl.go`:
- Implement `CamfrogAdapter` interface with both transports
- JSON RPC for: Status, Language, SwitchPlugin
- Multipart for: GetPluginSettings, SavePluginSettings, RoomPause, RoomSetLanguage
- Session management with cookie/CSRF handling
- Proper error handling with evidence-backed constraints

### 1.2 Transport Layer
- Use existing `JSONRPCTransport` and `MultipartTransport`
- Add request ID generation (atomic counter)
- Response parsing with bounded reads
- Retry logic with exponential backoff

### 1.3 Capability Registry Integration
- Load from JSON config file (evidence-backed)
- Runtime updates with verification tracking
- Default registry with observed plugins (1-16, excluding 9,10)

### 1.4 Adapter Factory
Create `backend/internal/camfrog/factory.go`:
- `NewAdapter(cfg Config) (CamfrogAdapter, error)`
- Dependency injection for transports

---

## Phase 2: Backend API Endpoints

### 2.1 Camfrog HTTP Handlers
Add to `backend/internal/httpapi/`:
- `POST /api/v1/bots/{id}/status` - Get bot status via adapter
- `POST /api/v1/bots/{id}/language` - Set bot language
- `POST /api/v1/bots/{id}/plugins/{pluginId}/toggle` - Switch plugin
- `GET /api/v1/bots/{id}/plugins/{pluginId}/settings` - Get plugin settings
- `POST /api/v1/bots/{id}/plugins/{pluginId}/settings` - Save plugin settings
- `POST /api/v1/rooms/{id}/pause` - Pause/resume room
- `POST /api/v1/rooms/{id}/language` - Set room language

### 2.2 Bot Lifecycle Handlers
- `POST /api/v1/bots/{id}/start` - Start bot (spawn adapter goroutine)
- `POST /api/v1/bots/{id}/stop` - Stop bot (graceful shutdown)
- `POST /api/v1/bots/{id}/restart` - Restart bot

### 2.3 WebSocket for Real-time Updates
- `GET /api/v1/bots/{id}/ws` - Bot connection events, status changes
- Use existing `BotConnection` model

---

## Phase 3: Frontend-Backend Integration

### 3.1 API Client Updates
Update `frontend/src/api/client.ts`:
- Add all new endpoint methods
- Type-safe request/response interfaces
- Error handling with retry logic

### 3.2 Auth Flow
- Login page → JWT storage → auto-refresh
- MFA flow (TOTP setup/verify)
- Organization switching with context

### 3.3 Real Data Integration
Replace mock data in pages with real API calls:
- Dashboard: real stats from `/api/v1/stats`
- Rooms: CRUD with real adapter calls
- Bots: Start/stop/restart with status polling
- Plugins: Configuration with real multipart forms
- Audit Logs: Real data with filtering
- Settings: Profile, org, API keys, MFA

### 3.4 WebSocket Integration
- Bot status real-time updates
- Connection events streaming
- Toast notifications for events

---

## Phase 4: Testing & Validation

### 4.1 Adapter Tests
- Unit tests with mock transports
- Contract tests with sanitized fixtures
- Negative tests for disabled plugins (9, 10)
- Static RE evidence enforcement (plain text only)

### 4.2 Integration Tests
- Full bot lifecycle (start → status → stop)
- Plugin toggle with verification
- Room pause/language change

### 4.3 E2E Frontend Tests
- Login → Dashboard → Bot management flow
- Plugin configuration dialog
- Organization switching

---

## Phase 5: Production Hardening

### 5.1 Durable Objects for Bot State
- Per-bot Durable Object for state coordination
- WebSocket connections managed in DO
- Automatic cleanup on bot stop

### 5.2 K2 for Bot Connection Logs
- Stream connection events to K2
- Subscriptions for real-time monitoring

### 5.3 Observability
- Workers Logs + Traces enabled
- Structured JSON logging
- Metrics for bot uptime, plugin usage

---

## Implementation Order

1. **adapter_impl.go** - Core adapter with both transports
2. **factory.go** - Adapter factory
3. **HTTP handlers** - REST endpoints for bot/room/plugin ops
4. **WebSocket handler** - Real-time updates
5. **Frontend API client** - All new endpoints
6. **Frontend pages** - Real data integration
7. **Tests** - Unit, integration, E2E
8. **Durable Objects** - Production bot state
9. **Observability** - Logs, traces, metrics

---

## Static RE Evidence Gates (MANDATORY)
- ❌ **Outgoing messages = plain text only** - adapter must enforce
- ❌ **No rich text** - no HTML/CSS in adapter
- ❌ **Room font editor = GUI preview only** - clearly labeled
- ❌ **Plugin capabilities default UNVERIFIED/DISABLED** - until VERIFIED with evidence

---

## File References
- Adapter interface: `backend/internal/camfrog/adapter.go`
- RPC validation: `backend/internal/camfrog/rpc.go`
- JSON RPC transport: `backend/internal/camfrog/transport/jsonrpc.go`
- Multipart transport: `backend/internal/camfrog/transport/multipart.go`
- Capability registry: `backend/internal/camfrog/capability/registry.go`
- Frontend API: `frontend/src/api/client.ts`
- Frontend types: `frontend/src/types/api.ts`