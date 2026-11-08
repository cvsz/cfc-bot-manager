# All Pull Request Dependency Map — cvsz/cfc-bot-manager

**Dependency changes across PRs and the dependency inventory.**
**Date:** 2026-11-08

## PR dependency changes

| # | Dependency change | Advisory | Compatibility | Regression tests |
|---|-------------------|----------|---------------|------------------|
| 1 | github/codeql-action/analyze 4.39.0 → 4.39.3 | None known | Patch bump (same major.minor) | CI workflow; no app tests affected |
| 2 | github/codeql-action/init 4.39.0 → 4.39.3 | None known | Patch bump (same major.minor) | CI workflow; no app tests affected |
| 3 | None (repository bootstrap) | N/A | N/A | bootstrap unit tests |
| 4 | None (draft, not merged) | N/A | N/A | N/A |
| 6 | None (removed embedded secrets/paths from scripts) | N/A | N/A | validate_repo, unittest |
| 7 | None (redacted report content) | N/A | N/A | validate_repo |
| 8 | None (documentation only) | N/A | N/A | validate_repo |
| 9 | Added `actions/setup-go` (b7ad1dad...) to CI; Go 1.23 module (dependency-free) | None | Go 1.23 toolchain | go test -race, go vet, go build |
| 10 | None (removed workflows/scripts) | N/A | N/A | make ci, unittest |

## Dependency inventory (current tree)

### Go backend (`backend/go.mod`)
```
module github.com/cvsz/cfc-bot-manager/backend

go 1.23
```
- No `require` blocks — dependency-free.
- No `go.sum` (no external dependencies).
- Local toolchain: go 1.27.0 (go.mod declares go 1.23; compatible).
- Security advisories: none applicable (no dependencies).
- Upgrade needed: none.

### GitHub Actions (pinned to immutable SHAs)
| Action | Ref | Pin type |
|--------|-----|----------|
| actions/checkout | 3d3c42e5aac5ba805825da76410c181273ba90b1 | commit SHA |
| actions/dependency-review-action | a1d282b36b6f3519aa1f3fc636f609c47dddb294 | commit SHA |
| actions/setup-go | b7ad1dad31e06c5925ef5d2fc7ad053ef454303e | commit SHA |
| github/codeql-action/init | 33d7b85da1a04171e211befe2c739303b9a4c066 | commit SHA |
| github/codeql-action/analyze | 33d7b85da1a04171e211befe2c739303b9a4c066 | commit SHA |

All baseline Actions pinned to immutable commit SHAs (supply-chain compliant). No floating tags.

### Docker
- `Dockerfile`: `FROM python:3.12-alpine` (tag-pinned, not digest-pinned). Recommend digest pinning for reproducible builds if used for delivery.

### Python tooling
- No `requirements.txt`, no `pyproject.toml`, no lockfile. Python stdlib only (`scripts/*.py`, `tests/*.py`). ruff 0.17.2 cache present (.ruff_cache) but ruff is not a declared project dependency.

## Dependency advisories

No dependency vulnerabilities identified (Go is dependency-free; Actions are pinned SHAs; no Python dependencies declared). Dependabot is configured for GitHub Actions and Docker (`.github/dependabot.yml`).

## Dependency upgrade recommendations

- None required for the current codebase.
- If the Dockerfile is brought into delivery, pin `python:3.12-alpine` to a digest.
- Consider declaring ruff as a dev dependency if Python linting is retained in CI.
