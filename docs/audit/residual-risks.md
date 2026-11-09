# Residual Risks — cfc-bot-manager

**Date:** 2026-11-08  
**Current HEAD:** 2ad7be06df6b75de3a5ebbddbb1bc0f9ee7066bb (main)  
**Classification:** Risks remaining after the P0 remediation and this audit. None are claimed mitigated without evidence.

| ID | Severity | Risk | Evidence | State | Required action |
|----|----------|------|----------|-------|-----------------|
| R-001 | P0 | `/home/cvsz/.git-credentials` (outside repo, mode 600) may contain real GitHub PATs and GPG passphrases | File exists; contents NOT inspected per policy | BLOCKED (contents not inspected) | Operator: review file, revoke/rotate affected credentials, protect file (chmod 600) |
| R-002 | P1 | GPG signature verification on GitHub shows `null` for signed commits 58d0f76/dc7109d/45ebba3/999c35e | Key 515DB3BD uploaded to GitHub; local `git verify-commit` passes; GitHub API shows `null` | BLOCKED (GitHub-side propagation) | Operator: verify GitHub commit verification status |
| R-003 | P1 | policedbc independent APPROVED review not obtained on PR #10 final head | PR #10 merged via commit 8c863be before policedbc approval | BLOCKED (cannot apply retroactively) | Process deviation documented; independent review gate must be enforced on future PRs |
| R-004 | P1 | Main branch protection not applied/verified | `gh api branches/main/protection` returns HTTP 404; `INDEPENDENT_REVIEWER` var not settable via API (404); `github_admin.py --verify` not executed | BLOCKED | Operator: set `INDEPENDENT_REVIEWER` via GitHub web UI, run `github_admin.py --repo cvsz/cfc-bot-manager --apply --verify` with admin identity |
| R-005 | P1 | No authentication/authorization layer; backend endpoints unauthenticated | `backend/` has no auth middleware | NOT IMPLEMENTED | Implement authn/authz before exposing the API beyond loopback |
| R-006 | P1 | No Camfrog network integration; RPC helper is serialization only | `backend/internal/camfrog/rpc.go` is boundary only | NOT IMPLEMENTED | Authorized integration evidence before claiming protocol support |
| R-007 | P1 | No database/persistence | No PostgreSQL; justified as future capability | NOT IMPLEMENTED | Add only when justified by product requirement |
| R-008 | P2 | No frontend or Tauri desktop | No code present | NOT IMPLEMENTED | Implement per architecture (docs/product/implementation-plan.md) |
| R-009 | P2 | No rate limiting, audit logging, secret redaction in logs | Not implemented | NOT IMPLEMENTED | Add before production |
| R-010 | P2 | Effective GitHub security settings (secret scanning, push protection, private vuln reporting) not fully verified via API | Partial API read-back: secret_scanning enabled, push protection enabled; private vuln reporting not checked | PARTIALLY VERIFIED | Operator read-back via GitHub API |
| R-011 | P3 | `ABOUT.md` has multiple TODO fields (ownership, product definition, metrics, data classification, support) | `ABOUT.md:8-16` | NOT IMPLEMENTED | Complete ABOUT.md |
| R-012 | P3 | Makefile/Dockerfile are placeholders | `Makefile`, `Dockerfile` | NOT IMPLEMENTED | Replace with real commands/images before delivery pipeline |

## Risk acceptance

No risk has been accepted as "mitigated" without evidence. P0 risks R-001–R-004 block the merge gate (MAIN remains NO-GO until Issue #5 acceptance conditions are met, including signed commit, policedbc approval on final head, and verified Main protection).

## GPG signing blocker detail

GPG signing is configured (`commit.gpgsign=true`, `user.signingkey=515DB3BD...`) but cannot be exercised:
- `git commit -S` fails: `gpg: cannot open '/dev/tty': No such device or address`
- No TTY available in this environment
- Passphrase unavailable without reading `.git-credentials` (forbidden per policy)
- Locally signed commits (58d0f76, dc7109d, 45ebba3, 999c35e) verified via `git verify-commit` as "Good signature"
- GitHub API shows `verified: null` for signed commits (not claimed as GitHub-verified)

## Branch protection blocker detail

`gh api branches/main/protection` returns HTTP 404 — branch protection is NOT applied on GitHub.
`scripts/github_admin.py --apply --verify` fails because `INDEPENDENT_REVIEWER` repository variable cannot be set via the GitHub API (PUT returns HTTP 404).
This requires operator action via the GitHub web UI to set the variable, then run `github_admin.py --apply --verify`.

## Credential exposure note

Historical exposure: a GPG passphrase previously appeared in repository history and an audit report (treated as compromised). Rotation recommended by operator — not performed or claimed in this session. No credential values were read, printed, echoed, or committed during this audit.
