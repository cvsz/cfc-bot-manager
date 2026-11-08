# Production Readiness — cfc-bot-manager

**Date:** 2026-11-08  
**Current HEAD:** 2ad7be06df6b75de3a5ebbddbb1bc0f9ee7066bb (main)  
**Verdict:** NOT PRODUCTION READY. The repository is a governance/tooling baseline with one Go backend vertical slice; P0 Issue #5 code remediation is complete but operator-level acceptance conditions are not met. The Camfrog Bot Manager application (frontend, desktop, database, Camfrog integration, authn/authz, deployment) is NOT IMPLEMENTED.

Per ZEAZ-INTRODUCTION.md: production readiness is an evidence-backed assessment across governance, security, reliability, data integrity, CI/CD, reproducibility, deployment, observability, backup/restore, DR, rollback, performance, capacity, documentation, incident response, ownership, and compliance. Green CI alone is insufficient.

## Readiness gate assessment

### Repository governance

| Gate | State | Evidence / reason |
|------|-------|-------------------|
| Repository governance docs (AGENTS.md, SECURITY.md, GOVERNANCE.md) | VERIFIED | Baseline docs present and current |
| Protected default branch | BLOCKED | `gh api branches/main/protection` returns HTTP 404 (not applied). Cannot apply via `github_admin.py --apply --verify` because `INDEPENDENT_REVIEWER` variable cannot be set via API (PUT returns 404). |
| PR review before merge | PARTIALLY VERIFIED | `request-independent-review.yml` present; `INDEPENDENT_REVIEWER` var not configured; PRs #1/#2 self-approved; #9 unreviewed; #10 merged before policedbc APPROVED on final head |
| Required status checks | PARTIALLY VERIFIED | CI workflow present; `verify-independent-approval` not enforceable (var missing); CodeQL and dependency-review workflows present |
| CODEOWNERS review | PARTIALLY VERIFIED | `.github/CODEOWNERS` present (@cvsz); not verified effective via API |
| Conversation resolution | NOT VERIFIED | Not checked via API |
| Blocked force pushes/deletion | NOT VERIFIED | Not checked via API (protection not applied) |
| Dependabot alerts/updates | VERIFIED (config present) | `.github/dependabot.yml`; Dependabot PRs #1/#2 merged |
| Private vulnerability reporting | NOT VERIFIED | SECURITY.md documents path; GitHub setting not checked via API |
| Secret scanning/push protection | PARTIALLY VERIFIED | GitHub API: secret_scanning enabled, secret_scanning_push_protection enabled; CI secret-name check present |
| Least-privilege Actions | VERIFIED (config) | `contents: read` on workflows; `default_workflow_permissions: read`; `can_approve_pull_request_reviews: false` |
| Actions cannot approve | VERIFIED (config) | `github_admin.py` payload specifies false; API read-back confirms `can_approve_pull_request_reviews: false` |
| CodeQL | VERIFIED (config present) | `.github/workflows/codeql.yml`; pinned SHA v4; GitHub CI success |
| Dependency review | VERIFIED (config present) | `.github/workflows/dependency-review.yml`; pinned SHA v5.0.0; GitHub CI success |

### Application runtime (NOT IMPLEMENTED)

| Gate | State | Reason |
|------|-------|--------|
| Authentication / authorization | NOT IMPLEMENTED | No auth layer; all backend endpoints unauthenticated |
| Session management | NOT IMPLEMENTED | No session layer |
| PostgreSQL persistence | NOT IMPLEMENTED | No database; justified as future capability |
| Camfrog network integration | NOT IMPLEMENTED | No transport; RPC serialization boundary only |
| Frontend (React + TypeScript) | NOT IMPLEMENTED | No code present |
| Tauri 2 desktop | NOT IMPLEMENTED | No code present; Windows tests BLOCKED (Linux env) |
| Encryption of sensitive data | NOT IMPLEMENTED | No data store |
| Rate limiting / DoS resistance | NOT IMPLEMENTED | Not implemented |
| Audit logging | NOT IMPLEMENTED | Not implemented |
| Backup / restore | NOT IMPLEMENTED | Not implemented |
| Deployment / rollback | NOT IMPLEMENTED | Not implemented |
| Observability / alerting | PARTIAL | slog structured logging only; no metrics/traces/alerts |
| Windows packaging | BLOCKED | No desktop; Linux environment |

## P0 Issue #5 acceptance status

| Acceptance criterion | State |
|---------------------|-------|
| Unsafe automation absent from Main tree | VERIFIED |
| Regression tests prevent reintroduction | VERIFIED |
| Exact signed PR head passes CI/CodeQL/dependency review/secret scanning | PARTIALLY VERIFIED |
| Fresh independent policedbc review on exact head | PARTIALLY VERIFIED (PR #10 not approved on final head) |
| Effective Main protection read back and verified | BLOCKED (API 404; INDEPENDENT_REVIEWER not settable via API) |
| Credential disposition / old-token rejection recorded privately | BLOCKED (not performed) |

## Conclusion

Production readiness is **NOT IMPLEMENTED / BLOCKED**, not PASS. Every applicable readiness gate must be VERIFIED before a production claim; currently most are NOT IMPLEMENTED or BLOCKED.

MAIN remains NO-GO until P0 acceptance conditions are met:
1. Main branch protection applied and verified
2. `INDEPENDENT_REVIEWER` variable configured
3. GPG signing operational with GitHub verification
4. policedbc independent review on PR #10 final head (process deviation cannot be retroactively fixed)
5. Credential rotation/revocation performed by operator

## Required before production-readiness claim

1. Apply and verify Main branch protection (`github_admin.py --apply --verify` with admin identity, after `INDEPENDENT_REVIEWER` is set)
2. Configure `INDEPENDENT_REVIEWER` repository variable via GitHub web UI
3. Resolve GPG signing blocker (unlock key via authorized means) and sign commits
4. Complete policedbc independent review of PR #10 (process deviation documented; future PRs must enforce)
5. Implement authentication, authorization (RBAC/tenant isolation), session management
6. Implement PostgreSQL persistence with migrations
7. Implement Camfrog adapter with authorized integration evidence (plugin 9 disabled, plugin 10 unverified)
8. Implement frontend and Tauri desktop
9. Run CodeQL, dependency review, secret scanning in CI and verify effective state via API
10. Add deployment, backup/restore, rollback, observability, incident response runbooks
