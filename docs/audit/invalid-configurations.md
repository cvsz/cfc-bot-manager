# Invalid Configurations — cfc-bot-manager

**Date:** 2026-11-08
**Finding:** No invalid committed configurations were found in the current tree. Configurations that are intentionally placeholders (Makefile, Dockerfile) are documented as such, not invalid.

## Configuration items audited

| Item | File | State | Notes |
|------|------|-------|-------|
| Go module | `backend/go.mod` | Valid | `go 1.23`, dependency-free |
| CI workflow | `.github/workflows/ci.yml` | Valid | YAML valid; Go + Python validation |
| CodeQL | `.github/workflows/codeql.yml` | Valid | Pinned SHA v4 |
| Dependency review | `.github/workflows/dependency-review.yml` | Valid | Pinned SHA v5.0.0 |
| Independent review | `.github/workflows/request-independent-review.yml` | Valid | `pull_request_target` gate |
| Template sync | `.github/workflows/template-sync.yml` | Valid | `persist-credentials: true`; pushes feature branch + opens reviewable PR (not auto-merge) |
| Dependabot | `.github/dependabot.yml` | Valid | GitHub Actions + Docker |
| CODEOWNERS | `.github/CODEOWNERS` | Valid | @cvsz ownership |
| Makefile | `Makefile` | Placeholder | Template targets; not invalid |
| Dockerfile | `Dockerfile` | Placeholder | `python:3.12-alpine`; not for production yet |
| `.env.example` | `.env.example` | Valid | Non-secret names only |

## Removed configurations (P0 remediation)

The following configurations were removed as unsafe (Issue #5 P0):
- `.github/workflows/auto-fix-open-prs.yml` — required `secrets.GPG_PASSPHRASE`, `secrets.GPG_SIGNING_KEY`, `secrets.POLIEDBC_PAT`
- `.github/workflows/automated-audit-review-approve.yml` — used `secrets.POLIEDBC_PAT` as `GH_TOKEN` for `gh pr review --request-changes`
- `.github/scripts/gpg-loopback.sh` — tracked signing wrapper reading `GPG_PASSPHRASE`
- `.github/scripts/post-review-comment.sh` — orphaned helper
- `scripts/auto-fix-loop.sh` — cross-repo direct-push with GPG loopback
- `scripts/deploy-workflow-all-repos.sh` — cross-repo mutation setting `POLIEDBC_PAT` secret
- `docs/ai/playbooks/auto-audit-review-approve-policedbc.md` — obsolete automated-approval documentation

## template-sync.yml note

`.github/workflows/template-sync.yml` uses `persist-credentials: true` and `git push origin "$branch"` to push a feature branch and open a reviewable PR. It does NOT auto-merge ("No automatic merge." in body), does not use reviewer credentials, and does not do cross-repository mutation. It was NOT retired under Issue #5, but its `persist-credentials: true` is a least-privilege consideration for future hardening.

## Invalid values

None found. No invalid locale, plugin ID, port, timeout, or environment values are committed; the Go config validator rejects them at load time.
