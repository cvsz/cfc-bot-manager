# Compatibility Matrix — cfc-bot-manager

**Date:** 2026-11-08
**Basis:** `docs/product/plugin-compatibility-matrix.md`, `docs/protocols/camfrog-observed-contract.md`, `backend/internal/camfrog/rpc.go`

## Evidence terminology (per plugin-compatibility-matrix.md)

- **OBSERVED:** configuration fields/page captured in user-supplied HTML; runtime behavior not verified.
- **VERIFIED:** authorized runtime test confirms the capability for a specific product/version.
- **UNVERIFIED:** missing, partial, or ambiguous evidence.
- **UNSUPPORTED:** confirmed unavailable in the stated product/version.

## Plugin inventory

| ID | Name | Page/settings capture | Runtime behavior | Cloud Bot | Go allowlist (rpc.go) |
|---:|------|-----------------------|------------------|-----------|----------------------|
| 1 | 8 Ball | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 2 | IM Control | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 3 | Messenger | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 4 | Greeter | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 5 | Math Trivia | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 6 | Bad Words | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 7 | Trivia | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 8 | Word Trivia | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 9 | Roulette | OBSERVED | UNVERIFIED | UNVERIFIED | **EXCLUDED** (disabled pending safety) |
| 10 | Unknown | UNVERIFIED | UNVERIFIED | UNVERIFIED | **EXCLUDED** (unverified) |
| 11 | Ghost | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 12 | Who Banned | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 13 | Talkback | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 14 | Dice | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 15 | DJ Manager | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 16 | Auto Moderator | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |

## Safety exclusions (enforced)

- **Plugin ID 9:** Disabled. Never enabled through automated changes. Project does not implement wagering/betting/gambling. Requires independent evidence of strictly non-wagering use case + safety review before any enablement.
- **Plugin ID 10:** Unverified. Identity never inferred from partial heap-related capture. Never displayed to users.

## Component compatibility

| Component | Compatibility | Evidence |
|-----------|---------------|----------|
| Go backend ↔ config | VERIFIED | config.Load validated by tests |
| Go backend ↔ HTTP API | VERIFIED | httptest tests pass |
| Go backend ↔ Camfrog RPC (serialization) | VERIFIED (boundary only) | rpc_test.go; no network |
| Go backend ↔ Camfrog network | NOT IMPLEMENTED | No transport |
| Frontend ↔ backend | NOT IMPLEMENTED | No frontend |
| Desktop ↔ backend | NOT IMPLEMENTED | No desktop |
| Backend ↔ PostgreSQL | NOT IMPLEMENTED | No database |
| Legacy Bot | Separate assessment required | No evidence |
| Cloud Bot | Separate assessment required | No evidence |

## Promotion gate (to mark VERIFIED)

Per docs/product/plugin-compatibility-matrix.md:37-46, a capability may be marked VERIFIED only with exact product/version, authorized environment, sanitized fixture, success+negative tests, permission requirements, timeout/session-expiry behavior, and independent PR review evidence. None are currently VERIFIED.
