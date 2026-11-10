# CFC Bot Manager Architecture

## Status and scope

This document defines the intended product architecture. It is not evidence that an application runtime exists or that Camfrog upstream compatibility is verified.

## System context

CFC Bot Manager has two clients and one shared server-side application:
- Web Control Panel: React + TypeScript.
- Windows Desktop: Tauri 2 + React + TypeScript.
- Backend: Go, with PostgreSQL persistence and REST/WebSocket interfaces.

Both clients authenticate to the same backend and share generated/validated API contracts. Desktop-specific functionality must remain in Tauri commands and narrowly scoped capabilities; no privileged filesystem or process access is exposed to web content.

## Logical components

1. **Client applications**
   - Web and desktop use a shared TypeScript API client, contracts, and reusable UI primitives.
   - Route/page modules organize dashboard, bots, rooms, plugins, users, audit/activity, profile, and settings views.
   - Client state is a cache, never an authorization source.

2. **API boundary**
   - Versioned REST endpoints for identity, sessions, bot/room resources, plugin settings, capability discovery, and audit history.
   - WebSocket for authenticated status and event updates.
   - Request IDs, bounded payloads, schema validation, explicit errors, timeouts, and cancellation.

3. **Application/domain layer**
   - User, tenant, bot, room, plugin capability, settings, session, and audit domains.
   - Authorization decisions are made server-side for every resource and operation.
   - Domain services do not depend on Camfrog HTML, HTTP forms, or RPC envelope details.

4. **Camfrog adapter**
   - Implements a narrow interface for status, language, plugin enable/disable, settings read/write, and supported actions only where evidence exists.
   - Owns transport, session handling, wire DTOs, version/capability negotiation, response validation, and compatibility tests.
   - Must distinguish unsupported, unknown, unavailable, and transient failure states.
   - No automatic retries for non-idempotent actions unless idempotency is established.

5. **Persistence and background work**
   - PostgreSQL is the server-side source of truth.
   - Migrations are explicit and tested.
   - Background jobs are bounded, observable, idempotent where possible, and safe to retry.
   - SQLite may be used as a desktop cache only after cache invalidation and encryption requirements are specified.

6. **Security/operations**
   - Secure sessions, CSRF defense for cookie auth, strict CORS/origin policy, RBAC, tenant isolation, rate limits, audit trail, secret redaction, and structured telemetry.
   - Health/readiness endpoints do not disclose credentials or internal topology.
   - Configuration is validated at startup; secrets are injected through runtime configuration, not committed files.

## Proposed repository organization

Create only directories with implemented content; this is a target layout, not a mandate to create empty scaffolding.

```text
apps/
  web/                 React + TypeScript control panel
  desktop/             Tauri 2 + React desktop application
packages/
  contracts/           Shared API/event schemas and TS types
  client/              Typed REST/WebSocket client
  ui/                  Shared UI components
backend/
  cmd/api/
  internal/
    auth/              Sessions, MFA, RBAC, middleware
    domain/             Entities and domain contracts
    api/                HTTP handlers, middleware, WebSocket
    service/            Use cases
    repository/         Persistence interfaces
    persistence/        PostgreSQL adapters/migrations
    camfrog/            Adapter interface and protocol clients
    security/           CSRF, origins, redaction, limits
    observability/
camfrog/
  contracts/            Captured/sanitized protocol fixtures
  compatibility/        Version/capability evidence matrix
docs/
  adr/
  product/
  protocols/
  security/
tests/
  contract/
  integration/
  security/
  e2e/
deployments/
  docker/
```

## Protocol boundary and evidence

Observed pages show a JSON RPC POST to `/bot/ajax/` with `{id, method, params}`, as well as separate multipart form submissions. These are observed interfaces, not a published supported SDK. See [observed contract](protocols/camfrog-observed-contract.md).

**Critical Finding from Reverse Engineering:** Static analysis of `camfrog.exe` (SHA-256: `0b744602c004536dcf81e9965e8cb17358eb276b3876a1e70c50f72777c27e06`) confirms that **Camfrog Desktop Client does NOT support rich text in outgoing chat messages**. All messages sent via UI Automation are plain text only. The adapter must not attempt to send HTML/CSS/rich text formatting.

Evidence from static RE:
- Only 3 GDI font APIs imported: `CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW`
- No HTML/CSS parsing libraries imported (`mshtml.dll`, `riched20.dll`, `msftedit.dll` not imported)
- HTML/CSS strings found are used **only for rendering incoming messages** in embedded web view
- Chat messages sent via UI Automation are plain text only (verified in UIA adapter send path)
- Camfrog uses `camfrog:room:openvideo:` custom URI scheme for room links

The adapter must preserve wire compatibility in transport DTOs and translate into typed domain values. Unknown fields and semantics are not to be guessed. Contract fixtures must be sanitized and contain no cookies, session identifiers, or room keys.

**Outgoing messages are plain text only — the adapter must not attempt to send HTML/CSS/rich text formatting.**

## Authentication and authorization

- Establish an authenticated principal and tenant context on every request.
- Enforce resource-level access in backend use cases and repositories.
- Use secure, rotated sessions; MFA for privileged users; server-side RBAC.
- Protect cookie-authenticated mutations with CSRF tokens and strict Origin validation.
- Validate WebSocket authentication and re-check authorization on subscriptions/actions.
- Log actor, tenant, operation, target, result, and correlation ID; redact secrets.

## Failure semantics

A remote Camfrog operation can fail independently of local persistence. Model remote operations as explicit states and never report success optimistically. For plugin toggles, reconcile the displayed state only after a validated upstream acknowledgement; on failure retain the last confirmed state and surface the error.

## Key decisions

See [ADR 0001](adr/0001-product-stack-and-adapter-boundary.md) for the initial stack and adapter boundary. Amend via a new ADR when evidence or constraints change.
