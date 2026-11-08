# ADR 0001: Product Stack and Camfrog Adapter Boundary

- Status: Proposed
- Date: 2026-11-08

## Context

The product requires both a Windows desktop app and a web control panel, with one consistent backend and one isolated Camfrog integration boundary. Supplied HTML captures show an undocumented/private interface with both JSON RPC and multipart form submissions. Public SDK/API support has not been established.

## Decision

Propose:
- Go for the backend.
- React + TypeScript for web.
- Tauri 2 + React + TypeScript for Windows desktop.
- Shared TypeScript API contracts/client/UI packages where useful.
- REST for commands/queries and authenticated WebSocket for live events.
- PostgreSQL as server-side source of truth.
- SQLite only as an optional desktop cache, never a competing source of truth.
- A CamfrogAdapter interface that isolates protocol specifics and exposes only evidence-backed capabilities.

## Consequences

- One backend centralizes authorization and audit logic across clients.
- Tauri requires a narrowly scoped capability model and explicit OS integration permissions.
- Shared contracts reduce drift, but cross-client packages must not leak desktop privileges into web.
- The undocumented upstream interface may change without notice; compatibility gates and fixtures are mandatory.
- Cloud Bot and legacy standalone bot compatibility must be treated as separate targets.
- Features with incomplete evidence remain disabled or explicitly marked unverified.

## Rejected alternatives

- Duplicating the backend in desktop and web: inconsistent authorization and protocol behavior.
- Embedding Camfrog HTTP calls directly in each client: credential exposure and duplicated fragile protocol logic.
- Treating captured HTML as a stable public SDK: unsupported by the available evidence.
