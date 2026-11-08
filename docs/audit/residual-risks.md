# Residual Risks — cfc-bot-manager

**Date:** 2026-11-08
**Classification:** Risks remaining after the P0 remediation and this audit. None are claimed mitigated without evidence.

| ID | Severity | Risk | Evidence | State | Required action |
|----|----------|------|----------|-------|-----------------|
| R-001 | P0 | `/home/cvsz/.git-credentials` (outside repo, mode 600) may contain real GitHub PATs and GPG passphrases | Final-Report.md observation; file exists | BLOCKED (contents not inspected) | Operator: review, revoke/rotate affected credentials, protect file |
| R-002 | P1 | GPG signing key `8BC7A85F...` requires passphrase; commits cannot be signed autonomously | `git commit -S` failed: `NEED_PASSPHRASE` / batch mode | BLOCKED | Operator: unlock key via `gpg-agent --preset-passphrase` or interactive entry; re-sign commit; verify `git verify-commit HEAD` |
| R-003 | P1 | policedbc independent review cannot be completed | `gh auth switch -u policedbc` requires human browser interaction | BLOCKED | Operator: switch to policedbc, review PR #10 on exact head, submit APPROVED/request-changes |
| R-004 | P1 | Main branch protection not applied/verified | `github_admin.py --verify` not run; `INDEPENDENT_REVIEWER` var not configured | BLOCKED | Operator: set `INDEPENDENT_REVIEWER`, run `github_admin.py --repo cvsz/cfc-bot-manager --apply --verify` with admin identity |
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
