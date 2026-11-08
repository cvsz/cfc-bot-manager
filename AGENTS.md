# AGENTS.md — CFC Bot Manager Engineering Contract

## READ FIRST — Canonical ZeaZDev policy

Read https://github.com/cvsz/zeaz/blob/main/zeaz.md before every engineering task. This local contract adds product-specific requirements and must not weaken the canonical policy.

## Product scope

CFC Bot Manager is a cross-platform management product with:
- a Windows desktop application (Tauri 2 + React + TypeScript);
- a web control panel (React + TypeScript);
- one shared Go backend and Camfrog protocol adapter;
- PostgreSQL as the server-side source of truth;
- REST for request/response APIs and WebSocket for live status/events.

This describes the intended architecture, not evidence that runtime components already exist.

## Required workflow

Inspect → Understand → Plan → Validate → Apply → Test → Verify → Document. Read policy, current code, tests, workflows, and relevant docs before edits. Search for existing implementations before adding new ones. Keep PRs focused and reviewable.

## Protocol and compatibility constraints

- Isolate Camfrog-specific transport and payload behavior behind an adapter boundary.
- The observed POST /bot/ajax/ JSON RPC protocol and multipart form submissions are undocumented/private web interfaces, not a verified public SDK.
- Treat protocol support as experimental and compatibility-gated until verified using authorized test accounts/environments.
- Never scrape, bypass authentication, evade rate limits, or use captured credentials/session material.
- Never commit cookies, Set-Cookie headers, session tokens, room access keys, passwords, or private user data.
- Never claim a Cloud Bot or legacy plugin works without end-to-end evidence for that exact product/version.
- Preserve observed wire names and positional ordering in transport DTOs; map them into typed domain models at the adapter boundary.
- Unknown plugin #10 must remain UNVERIFIED. Do not infer its identity from the partial heap-related form capture.
- Preserve plugin settings as opaque/raw values when semantics or validation constraints are not evidenced.

## Security requirements

- Secure sessions, CSRF protection for cookie-authenticated writes, strict origin checks, secure cookie flags, RBAC, tenant/resource isolation, and audit logging.
- Validate all external inputs and remote responses against schemas.
- Render user/plugin-provided text as text, never trusted HTML.
- Use explicit request timeouts, cancellation, bounded payload sizes, and safe retry policies.
- Never forward Authorization across origins or redirects.
- State-changing actions must use appropriate HTTP methods and must not be implemented as unsafe GET links.
- Fail closed when authorization, protocol capability, or compatibility state is unknown.

## Implementation quality

- Go backend; React + TypeScript web; Tauri 2 desktop.
- Share contracts, API client, and reusable UI only where this reduces duplication without obscuring platform boundaries.
- Keep protocol transport separate from domain services and persistence.
- Prefer small dependencies and explicit contracts.
- Add tests for parsing, validation, authorization, protocol errors, and failure/rollback paths.
- Do not add fake adapters, mock-backed production behavior, placeholder endpoints, or controls that imply unsupported functionality.
- Keep secrets out of logs and test fixtures.

## Evidence states

Use VERIFIED (reproduced by test/authorized runtime check for stated version), OBSERVED (visible in supplied source/capture but not independently verified), INFERRED (design interpretation), UNVERIFIED (evidence absent/ambiguous), and UNSUPPORTED (known unavailable in stated environment). Every plugin capability and protocol behavior must carry an evidence state.

## Completion

Run relevant formatting, lint, type, unit/integration, build, security, and repository validation. Inspect final diff. State exactly which checks could not run and why. Do not claim deployment, compatibility, or production readiness based on documentation alone.
