# Architecture Review — cfc-bot-manager

**Date:** 2026-11-08
**Source audited:** `docs/architecture.md`, `docs/adr/0001-product-stack-and-adapter-boundary.md`, `docs/product/engineering-contract.md`, `docs/product/implementation-plan.md`, `docs/product/requirements.md`, `backend/`

## Intended architecture (documented)

- Go shared backend (REST + Camfrog adapter)
- React + TypeScript web control panel
- Tauri 2 Windows desktop (shares API contracts/client/UI)
- Shared TypeScript API contracts/client/UI packages
- PostgreSQL as server-side source of truth; SQLite optional desktop cache only
- REST for commands/queries; authenticated WebSocket for live events
- `CamfrogAdapter` interface isolating protocol specifics

## Actual implementation

| Component | Implemented? | Evidence |
|-----------|-------------|----------|
| Go backend | PARTIAL | `backend/` vertical slice (config, HTTP API, RPC serialization); no auth, no persistence, no Camfrog transport |
| React + TypeScript web panel | NOT IMPLEMENTED | No frontend code |
| Tauri 2 desktop | NOT IMPLEMENTED | No desktop code |
| Camfrog adapter (network) | NOT IMPLEMENTED | RPC serialization boundary only |
| PostgreSQL persistence | NOT IMPLEMENTED | No database |
| Plugin compatibility layer | NOT IMPLEMENTED | Allowlist in rpc.go only (no runtime activation) |
| RPC boundary | PARTIAL | Serialization/validation; no network |
| Event/session management | NOT IMPLEMENTED | No session layer |

## Boundary and ownership audit

- **Module ownership:** `backend/` is the single backend module; no clear separation of domain services vs. transport vs. persistence yet (only one slice exists).
- **Dependency inversion:** Not yet applicable (no multiple layers to invert).
- **API contracts:** REST routes (/healthz, /readyz, /api/v1/meta) defined; no OpenAPI/schema.
- **State ownership:** No state persisted; in-memory only.
- **Error handling:** Go errors returned; HTTP errors sanitized (no leak); config validation strict.
- **Data flow:** Config → server → handler → response; Camfrog RPC is unidirectional serialization.
- **Concurrency:** `main.go` uses `select` on signal + serveErr; `rpc.go` helpers are pure functions (safe for concurrent use); no shared mutable state.
- **Lifecycle:** `main.go` graceful shutdown via `signal.NotifyContext` + `server.Shutdown` with timeout.
- **Security boundaries:** Loopback bind by default; MaxHeaderBytes; bounded response reads; security headers; no authn/authz yet.
- **Windows compatibility:** Not applicable (no desktop code); Linux environment.

## Architectural concerns

1. **No microservice created** — correct; single Go backend, monolithic within its scope.
2. **Camfrog adapter boundary is sound in principle** (interface isolates protocol) but not implemented beyond serialization.
3. **No trust boundary between client and backend** yet — no authentication, so no server-side authorization.
4. **No persistence layer** — consistent with "do not add a database before there is a justified product requirement."

## Verdict

Architecture is documented and consistent (one backend, isolated adapter boundary, no microservices). The implemented slice (Go backend) respects the boundaries. The majority of the architecture is NOT IMPLEMENTED and must not be claimed.
