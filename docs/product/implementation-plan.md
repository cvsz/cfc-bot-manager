# Implementation Plan

This plan prioritizes a verifiable vertical slice over creating a large empty monorepo.

## Phase 0 — Governance and product contract

- [x] Define product surfaces and intended architecture.
- [x] Preserve the inherited repository governance contract.
- [x] Record observed Camfrog transport and plugin evidence boundaries.
- [x] Mark plugin ID 10 as UNVERIFIED.
- [x] **Integrate Static RE evidence from camfrog.exe reverse engineering** (SHA-256: `0b744602c004536dcf81e9965e8cb17358eb276b3876a1e70c50f72777c27e06`)
- [ ] Complete ADR review and scope sign-off.

## Phase 1 — Backend foundation

- [x] Create Go module and pinned toolchain (Go 1.25).
- [x] Add configuration loader with strict validation and safe secret sourcing.
- [x] Add health/readiness and structured logging with redaction.
- [x] Define domain entities and repository interfaces.
- [x] Add PostgreSQL schema/migrations and transaction boundaries.
- [x] Implement authentication, session rotation, CSRF, MFA capability, and server-side RBAC.
- [x] Add resource/tenant isolation tests before exposing CRUD endpoints.
- [ ] Transaction boundaries & rollback strategy (P1-08).

## Phase 2 — Protocol adapter

- [x] Define CamfrogAdapter interface and capability model.
- [x] Separate JSON RPC and multipart settings transports.
- [x] Implement request IDs, timeout, cancellation, response validation, and session handling.
- [x] Add sanitized contract fixtures and fake-server tests (P2-04).
- [ ] Verify protocol behavior only in an authorized environment (P2-05).
- [x] Keep unknown/unverified capabilities disabled (P2-06).
- [ ] **Integrate Static RE evidence into adapter**: outgoing messages = plain text only; no rich text support; HTML/CSS for incoming rendering only; camfrog:room:openvideo: URI scheme for room links; Registry-based IM font settings (IMFont, IMFontSize, FontName, Size).

## Phase 3 — Web control panel

- [ ] React + TypeScript app with typed API client.
- [ ] Authentication and route guards backed by server authorization.
- [ ] Dashboard, bot/room inventory, plugin capability/settings, activity, user/role, profile, and settings views.
- [ ] WebSocket status updates with reconnect/backoff and stale-state handling.
- [ ] Accessible forms, validation, loading/empty/error states.
- [x] **Room message font editor (GUI preview only)** — clearly labeled as preview; adapter sends plain text only.

## Phase 4 — Windows desktop

- [ ] Tauri 2 shell with minimal capabilities.
- [ ] Reuse contracts, API client, and shared UI.
- [ ] Secure token/session handling using OS-backed storage where appropriate.
- [ ] Explicit permissions for filesystem/dialog/process access.
- [ ] Signed/reproducible packaging and update strategy.

## Phase 5 — Operations and release

- [ ] Docker Compose for local development and test environment.
- [ ] CI for Go, TypeScript, Rust/Tauri, security, contract, and integration checks.
- [ ] SBOM and dependency review.
- [ ] Backup/restore and migration rollback strategy.
- [ ] Observability, alerting, deployment, and rollback runbooks.
- [ ] Independent review and exact-head verification before release.

## Dependency order

Backend contracts and authorization precede client feature work. Protocol transport is isolated and tested before any UI control claims a feature is available. Do not build plugin forms from assumptions where field semantics are incomplete.

**Static RE evidence gates:**
- Outgoing messages = plain text only (adapter must enforce)
- No rich text support in adapter (no HTML/CSS formatting)
- Room message font editor = GUI preview only (clearly labeled)
- Plugin capabilities default to UNVERIFIED/DISABLED until VERIFIED
