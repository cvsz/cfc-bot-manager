# Security Threat Model — Initial

## Assets

Camfrog session material, room access keys, user identities, tenant data, bot configuration, moderation settings, audit records, desktop session state, and signing/release credentials.

## Trust boundaries

- Browser/desktop client ↔ Go API.
- Go API ↔ PostgreSQL.
- Go API ↔ Camfrog upstream web interface.
- WebSocket clients ↔ event subscriptions.
- Build/release workflows ↔ source/dependencies/artifacts.

## Principal risks and controls

| Risk | Required control |
|---|---|
| Cross-tenant/resource data access | Server-derived tenant context, repository scoping, deny-by-default authorization, adversarial tests |
| Session theft/fixation | Secure cookies or OS-backed storage, rotation, expiry, revocation, no tokens in URLs/logs |
| CSRF and origin abuse | CSRF token, strict Origin/Host validation, SameSite/HttpOnly/Secure cookies |
| XSS through plugin/user text | Contextual escaping; never inject captured text as HTML |
| SSRF/redirect credential leak | Fixed configured upstream origin, redirect rejection or strict same-origin policy, no cross-origin auth forwarding |
| RPC/form payload manipulation | Typed validation, allow-listed methods/fields, bounded payloads, authorization |
| Session concurrency/races | Per-session serialization until concurrency behavior is verified |
| Polling exhaustion | Single-flight polling, cancellation, bounded intervals and backoff |
| WebSocket privilege drift | Authenticate handshake, authorize subscriptions, revalidate privileged actions |
| Desktop privilege escalation | Minimal Tauri capabilities, scoped filesystem access, CSP, no arbitrary shell/process bridge |
| Secret leakage | Secret redaction, no raw protocol dumps, secret scanning, sanitized fixtures |
| Supply-chain compromise | Pinned actions/dependencies, dependency review, SBOM, reproducible builds |

## Open questions

- Which Camfrog product/version is the supported target?
- What authentication and session lifecycle does the upstream adapter require?
- Can the upstream interface be used under applicable terms and authorization?
- Which operations are idempotent and safe to retry?
- What exact tenant/business model and retention obligations apply?

## Release gate

This is an initial design threat model, not a completed security assessment. A production threat model requires review against implemented code, data flows, deployment topology, and actual upstream behavior.
