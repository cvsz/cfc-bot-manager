# Dependency Upgrade Report — cfc-bot-manager

**Date:** 2026-11-08
**Policy:** Upgrade in small groups; never blindly upgrade all; never downgrade to conceal failures; never remove integrity checks; run all relevant tests after each group.

## Dependency inventory

### Go backend (`backend/go.mod`)
```
module github.com/cvsz/cfc-bot-manager/backend

go 1.23
```
- Dependency-free: no `require` blocks, no `go.sum`.
- Local toolchain: go 1.27.0; go.mod declares go 1.23 (compatible).
- Upgrades needed: none. No transitive dependencies to audit.

### GitHub Actions (immutable SHA pins — supply-chain compliant)
| Action | Ref | Status |
|--------|-----|--------|
| actions/checkout | 3d3c42e5aac5ba805825da76410c181273ba90b1 | Current (pinned SHA) |
| actions/dependency-review-action | a1d282b36b6f3519aa1f3fc636f609c47dddb294 | Current (pinned SHA) |
| actions/setup-go | b7ad1dad31e06c5925ef5d2fc7ad053ef454303e | Current (pinned SHA, from PR #9) |
| github/codeql-action/init | 33d7b85da1a04171e211befe2c739303b9a4c066 | Current (pinned SHA, v4) |
| github/codeql-action/analyze | 33d7b85da1a04171e211befe2c739303b9a4c066 | Current (pinned SHA, v4) |

Dependabot is configured for GitHub Actions and Docker (`.github/dependabot.yml`). Dependabot PRs #1 and #2 were merged (codeql-action bumps 4.39.0 → 4.39.3).

### Docker
- `Dockerfile`: `FROM python:3.12-alpine` (tag-pinned). No application to containerize yet. Recommend digest pinning when used for delivery.

### Python tooling
- stdlib only; no `requirements.txt`/lockfile. ruff 0.17.2 cache present but ruff is not a declared project dependency.

## Security advisories reviewed

No dependency vulnerabilities identified:
- Go: no dependencies.
- Actions: pinned to immutable SHAs (no floating tags to drift into vulnerable versions).
- No Python dependencies declared.

## Upgrades performed

None required. No dependency groups were upgraded because no dependencies exist and no advisories apply.

## Upgrade governance followed

- No blind mass upgrade (no dependencies to upgrade).
- No downgrade to conceal failures.
- No integrity checks removed (SHA pins retained).
- Tests run after review: `make ci` PASS (includes Go test -race, vet, build; Python unittest; validate_repo).

## Recommendations

- When the Go backend gains dependencies, add `go.sum`, run `go mod tidy`, and review `govulncheck` output.
- When the Dockerfile is used for delivery, pin the base image to a digest.
- Consider declaring ruff as a dev dependency if Python linting is retained in CI.
