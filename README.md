# CFC Bot Manager

CFC Bot Manager is a cross-platform management product for Camfrog bot instances, designed around a shared backend and protocol adapter.

**Project status:** product foundation and requirements are being established. The repository currently contains governance/bootstrap foundations; application runtime, protocol compatibility, and production readiness must be demonstrated by implementation and tests.

## Product surfaces

- **Web Control Panel:** React + TypeScript.
- **Windows Desktop:** Tauri 2 + React + TypeScript.
- **Shared backend:** Go API, authorization, persistence, jobs, and real-time event delivery.
- **Protocol boundary:** a versioned Camfrog adapter that isolates observed upstream web-interface behavior from the application domain.

## Architecture principles

1. Web and desktop clients use the same backend contracts and authorization rules.
2. Camfrog-specific details remain behind an adapter; business logic never depends on raw upstream HTML or transport payloads.
3. PostgreSQL is the server-side source of truth. A local SQLite cache is optional and must never become a second authority.
4. REST handles commands and queries; WebSocket carries authenticated live updates.
5. Compatibility is capability-based and evidence-backed. Unknown features are disabled rather than guessed.
6. Security controls apply server-side to every API, not only in the UI.

See [product requirements](docs/product/requirements.md), [architecture](docs/architecture.md), and [protocol evidence](docs/protocols/camfrog-observed-contract.md).

## Plugin compatibility

Captured legacy plugin pages are not proof of Cloud Bot compatibility. The current evidence inventory covers plugin IDs 1–9 and 11–16; ID 10 has only a partial, unidentified capture and remains **UNVERIFIED**. Each capability must be verified against the actual supported Camfrog product/version before it is enabled.

## Development status

This repository currently contains project governance and planning artifacts. Do not treat it as a functioning bot manager until the runtime implementation, CI checks, authorized integration tests, and deployment evidence exist.

Start with [development guide](docs/development.md) and [implementation plan](docs/product/implementation-plan.md).

## Security

Do not commit Camfrog cookies, session tokens, room access keys, passwords, or private user data. Report security issues through [SECURITY.md](SECURITY.md).
