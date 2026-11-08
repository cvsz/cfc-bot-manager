# Security Assessment — cfc-bot-manager

**Scope:** Repository governance, CI/security workflows, Go backend, retired automation. No live penetration testing performed (no authorized staging target; no production targets attacked).
**Method:** Static code review, configuration inspection, local test execution. Evidence states per ZEAZ-INTRODUCTION.md.
**Date:** 2026-11-08

## P0 — Privileged automation (Issue #5)

Status: REMEDIATED (code), blocked on operator actions (GPG signing, policedbc review, Main protection).

All unsafe automation retired (see docs/audit/issue-5-remediation.md):
- `auto-fix-open-prs.yml`, `automated-audit-review-approve.yml` workflows deleted
- `gpg-loopback.sh`, `post-review-comment.sh` deleted
- `auto-fix-loop.sh`, `deploy-workflow-all-repos.sh` deleted
- `auto-audit-review-approve-policedbc.md` playbook deleted

Independent human review gate preserved: `request-independent-review.yml` (reviewer via Actions variable, never hard-coded, never auto-approves/merges).

## Go backend security (verified)

| Control | File | State |
|---------|------|-------|
| Loopback-only default bind | `backend/internal/config/config.go:69-78` | VERIFIED — host must be explicit; `127.0.0.1` default |
| HTTP timeouts | `backend/cmd/api/main.go:31-35` | VERIFIED — Read/Write/Idle/Header timeouts from config |
| Max header size | `backend/cmd/api/main.go:35` | VERIFIED — `MaxHeaderBytes: 1 << 20` |
| Bounded response reads | `backend/internal/camfrog/rpc.go:95-105` | VERIFIED — `MaxRPCResponseSize = 1 << 20`, `io.LimitReader` |
| JSON validity check | `backend/internal/camfrog/rpc.go:106` | VERIFIED — `json.Valid` before returning |
| Method allowlist | `backend/internal/camfrog/rpc.go:45-77` | VERIFIED — status/lang/switch_plugin only; unknown method rejected |
| Plugin ID allowlist (9 and 10 excluded) | `backend/internal/camfrog/rpc.go:25-28` | VERIFIED — IDs 1-8, 11-16 only; 9 (roulette) and 10 (unknown) not in inventory |
| Locale allowlist | `backend/internal/camfrog/rpc.go:19-23` | VERIFIED — 10 observed locales only |
| Request ID > 0 | `backend/internal/camfrog/rpc.go:39-41` | VERIFIED — zero ID rejected |
| Graceful shutdown | `backend/cmd/api/main.go:48-63` | VERIFIED — signal.NotifyContext + Shutdown with timeout |
| Security headers | `backend/internal/httpapi/server.go:47-55` | VERIFIED — nosniff, no-store, no-referrer, CSP default-src 'none' |
| No credential logging | `backend/internal/httpapi/server.go:29-41` | VERIFIED — readiness errors not leaked to response |
| GET-only routes | `backend/internal/httpapi/server.go:36-45` | VERIFIED — mutating methods return 405 |

## Security gaps (NOT IMPLEMENTED — not claimed)

- Authentication / session management (no authn/authz layer implemented)
- CSRF protection (no cookie-authenticated write endpoints yet)
- RBAC / tenant isolation (documented in requirements, not implemented)
- MFA (documented as future)
- Encryption of sensitive data at rest (no data store yet)
- Rate limiting / DoS resistance (no rate limiter implemented)
- Secret redaction in logs (structured logging exists; redaction not yet implemented)
- PostgreSQL persistence (no database)

## Supply-chain security

- Go module is dependency-free (`backend/go.mod`: `go 1.23`, no `require` blocks). No `go.sum` (no external dependencies).
- GitHub Actions pinned to immutable SHAs:
  - `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1`
  - `actions/dependency-review-action@a1d282b36b6f3519aa1f3fc636f609c47dddb294`
  - `actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`
  - `github/codeql-action/init@33d7b85da1a04171e211befe2c739303b9a4c066`
  - `github/codeql-action/analyze@33d7b85da1a04171e211befe2c739303b9a4c066`
- Dockerfile: `FROM python:3.12-alpine` (pinned digest not used; digest pinning recommended)
- No committed secrets (verified via `git ls-files` secret-name scan)

## Secret scanning / push protection

- CI checks for tracked secret-bearing file names (`.env`, `id_rsa`, `*.pem`, `*.key`).
- No `.env`, `id_rsa`, `*.pem`, `*.key` tracked in Git (VERIFIED).
- GitHub secret scanning / push protection effective state: BLOCKED (requires GitHub API read-back with admin/repo permissions; not run locally).

## CodeQL / Dependency Review

- `.github/workflows/codeql.yml` — CodeQL for `actions` language, pinned SHA (v4). Not run locally (GitHub Actions cloud).
- `.github/workflows/dependency-review.yml` — dependency-review-action v5.0.0. Not run locally.

## Branch protection

- `scripts/github_admin.py` defines the intended protection payload (required reviews, strict status checks, CODEOWNERS, conversation resolution, no force-push/deletion, enforce admins, read-only Actions, Actions cannot approve).
- Effective state: BLOCKED — `github_admin.py --verify` not executed (requires admin identity); `INDEPENDENT_REVIEWER` variable not configured.

## Credential exposure observations (private, no values published)

- `/home/cvsz/.git-credentials` exists (mode 600). Contents not inspected. May contain real PATs and GPG passphrases. Operator must review, revoke/rotate as needed.
- `scripts/gh-switch-accounts.conf` contains GPG public-key fingerprint `8BC7A85F1FB5322782921AEC14E1E0BC08F632B0` and account emails (not secret values).

## Penetration testing

No authorized penetration tests performed. Scope limited to local/staging synthetic testing only; no authorized staging environment or Camfrog target is available. Live user accounts, production hosts, and destructive operations are out of scope and were not touched.

## Residual risks

| ID | Severity | Risk | State |
|----|----------|------|-------|
| R-001 | P0 | `.git-credentials` contains real PATs/GPG passphrases outside repo | BLOCKED (operator must rotate) |
| R-002 | P1 | GPG signing key requires passphrase; cannot sign commits autonomously | BLOCKED |
| R-003 | P1 | No authentication/authorization layer; backend endpoints unauthenticated | NOT IMPLEMENTED |
| R-004 | P1 | Main branch protection not applied/verified | BLOCKED (admin identity) |
