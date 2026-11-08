# Implementation Plan

This plan prioritizes a verifiable vertical slice over creating a large empty monorepo.

## Phase 0 — Governance and product contract

- [x] Define product surfaces and intended architecture.
- [x] Preserve the inherited repository governance contract.
- [x] Record observed Camfrog transport and plugin evidence boundaries.
- [x] Mark plugin ID 10 as UNVERIFIED.
- [ ] Complete ADR review and scope sign-off.

## Phase 1 — Backend foundation

- [ ] Create Go module and pinned toolchain.
- [ ] Add configuration loader with strict validation and safe secret sourcing.
- [ ] Add health/readiness and structured logging with redaction.
- [ ] Define domain entities and repository interfaces.
- [ ] Add PostgreSQL schema/migrations and transaction boundaries.
- [ ] Implement authentication, session rotation, CSRF, MFA capability, and server-side RBAC.
- [ ] Add resource/tenant isolation tests before exposing CRUD endpoints.

## Phase 2 — Protocol adapter

- [ ] Define CamfrogAdapter interface and capability model.
- [ ] Separate JSON RPC and multipart settings transports.
- [ ] Implement request IDs, timeout, cancellation, response validation, and session handling.
- [ ] Add sanitized contract fixtures and fake-server tests.
- [ ] Verify protocol behavior only in an authorized environment.
- [ ] Keep unknown/unverified capabilities disabled.

## Phase 3 — Web control panel

- [ ] React + TypeScript app with typed API client.
- [ ] Authentication and route guards backed by server authorization.
- [ ] Dashboard, bot/room inventory, plugin capability/settings, activity, user/role, profile, and settings views.
- [ ] WebSocket status updates with reconnect/backoff and stale-state handling.
- [ ] Accessible forms, validation, loading/empty/error states.

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
