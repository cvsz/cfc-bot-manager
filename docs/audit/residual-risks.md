# Residual Risks — cfc-bot-manager

**Date:** 2026-11-08
**Classification:** Risks remaining after the P0 remediation and this audit. None are claimed mitigated without evidence.

| ID | Severity | Risk | Evidence | State | Required action |
|----|----------|------|----------|-------|-----------------|
| R-001 | P0 | `/home/cvsz/.git-credentials` (outside repo, mode 600) may contain real GitHub PATs and GPG passphrases | Final-Report.md observation; file exists | BLOCKED (contents not inspected) | Operator: review, revoke/rotate affected credentials, protect file |
| R-002 | P1 | GPG signature verification on GitHub shows `null` for signed commits 58d0f76/dc7109d | Key 515DB3BD uploaded to GitHub; local `git verify-commit` passes | BLOCKED (GitHub-side propagation) | Operator: verify GitHub commit verification status |
| R-003 | P1 | policedbc independent APPROVED review not obtained on final head | PR #10 merged via commit 8c863be before policedbc approval | BLOCKED (cannot apply retroactively) | Process deviation documented; independent review gate must be enforced on future PRs |
| R-004 | P1 | Main branch protection not applied/verified | `github_admin.py --verify` not run; `INDEPENDENT_REVIEWER` var not configured; independent review gate workflow fails | BLOCKED | Operator: set `INDEPENDENT_REVIEWER`, run `github_admin.py --repo cvsz/cfc-bot-manager --apply --verify` with admin identity |
| R-005 | P1 | No authentication/authorization layer; backend endpoints unauthenticated | `backend/` has no auth middleware | NOT IMPLEMENTED | Implement authn/authz before exposing the API |
| R-006 | P1 | No Camfrog network integration; RPC helper is serialization only | `backend/internal/camfrog/rpc.go` is boundary only | NOT IMPLEMENTED | Authorized integration evidence before claiming protocol support |
| R-007 | P1 | No database/persistence | No PostgreSQL; justified as future capability | NOT IMPLEMENTED | Add only when justified by product requirement |
| R-008 | P2 | No frontend or Tauri desktop | No code present | NOT IMPLEMENTED | Implement per architecture (Phases 3-4 of plan) |
| R-009 | P2 | No rate limiting, audit logging, secret redaction in logs | Not implemented | NOT IMPLEMENTED | Add before production |
| R-010 | P2 | Effective GitHub security settings (secret scanning, push protection, private vuln reporting) not verified via API | Not run locally | BLOCKED | Operator read-back via GitHub API |
| R-011 | P3 | `ABOUT.md` has multiple TODO fields (ownership, product definition, metrics, data classification, support) | `ABOUT.md:8-16` | NOT IMPLEMENTED | Complete ABOUT.md |
| R-012 | P3 | Makefile/Dockerfile are placeholders | `Makefile`, `Dockerfile` | NOT IMPLEMENTED | Replace with real commands/images before delivery pipeline |

## Risk acceptance

No risk has been accepted as "mitigated" without evidence. P0 risks R-001–R-004 block the merge gate (MAIN remains NO-GO until Issue #5 acceptance conditions are met, including signed commit, policedbc approval, and verified Main protection).
