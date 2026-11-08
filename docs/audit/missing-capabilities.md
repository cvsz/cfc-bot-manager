# Missing Capabilities — cfc-bot-manager

**Date:** 2026-11-08
**Rule:** Missing components are marked NOT IMPLEMENTED, not PASS. Blocked platform tests are marked BLOCKED, not PASS.

## Not implemented (application runtime)

| Capability | State | Justification / notes |
|------------|-------|----------------------|
| React + TypeScript web control panel | NOT IMPLEMENTED | No frontend code; intended by architecture |
| Tauri 2 Windows desktop | NOT IMPLEMENTED | No desktop code; Windows tests BLOCKED (Linux env) |
| Camfrog network transport/adapter | NOT IMPLEMENTED | RPC serialization boundary only; no authorized environment |
| Authentication / session management | NOT IMPLEMENTED | No authn/authz layer |
| PostgreSQL persistence | NOT IMPLEMENTED | No database; justified ("do not add a database before justified requirement") |
| Bot lifecycle (start/stop/restart/status) | NOT IMPLEMENTED | No bot manager runtime |
| Plugin runtime activation | NOT IMPLEMENTED | Allowlist exists; no activation path |
| WebSocket events | NOT IMPLEMENTED | Intended; not implemented |
| Audit logging | NOT IMPLEMENTED | Not implemented |
| Rate limiting / DoS resistance | NOT IMPLEMENTED | Not implemented |
| Encryption of sensitive data | NOT IMPLEMENTED | No data store |
| Backup / restore | NOT IMPLEMENTED | Not implemented |
| Deployment / rollback | NOT IMPLEMENTED | Not implemented |
| Observability (metrics/traces/alerts) | PARTIAL | slog structured logging only |
| Docker delivery | NOT IMPLEMENTED | Dockerfile is placeholder |
| MFA | NOT IMPLEMENTED | Intended; not implemented |

## Not verified (blocked)

| Capability | State | Reason |
|------------|-------|--------|
| GPG-signed commits | BLOCKED | Configured key requires passphrase unavailable without reading `.git-credentials` |
| policedbc independent review | BLOCKED | `gh auth switch -u policedbc` requires human interaction |
| Main branch protection effective state | BLOCKED | `github_admin.py --verify` requires admin identity |
| GitHub secret scanning / push protection / private vuln reporting | BLOCKED | Requires GitHub API read-back |
| CodeQL / dependency review effective results | BLOCKED | GitHub Actions cloud; not run locally |
| Windows/Tauri tests | BLOCKED | No Windows environment |
| Authorized Camfrog integration test | BLOCKED | No authorized environment/account/credentials |

## Verified (implemented)

| Capability | State |
|------------|-------|
| Go backend config loader + validation | VERIFIED |
| Go HTTP server (health/ready/meta, security headers, graceful shutdown) | VERIFIED |
| Go Camfrog RPC serialization/validation | VERIFIED |
| Python repository validation + tests | VERIFIED |
| P0 unsafe automation retirement | VERIFIED (code); regression tests pass |
| CI (`make ci`) | VERIFIED (local) |

## Deliberately NOT added (per mission/governance)

- Database (no justified product requirement yet)
- Microservice architecture (avoid unnecessary decomposition)
- Automation violating platform rules or user authorization
- Claimed Camfrog/legacy/Cloud Bot compatibility without evidence
- Fake adapters or mock-backed production behavior (engineering-contract.md:49)
