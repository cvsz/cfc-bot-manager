# Build, Lint, and Test Report — cfc-bot-manager

**Date:** 2026-11-08
**Commit:** dd759bc (feat/cfc-bot-manager-enterprise-hardening), branch vs main@8caf9e4
**Environment:** Linux, go 1.27.0, Python 3.x, ruff 0.17.2, shellcheck present

## BUILD

| Check | Command | Exit | Result |
|-------|---------|------|--------|
| Go build | `cd backend && go build ./...` | 0 | PASS |
| Python compile | `python3 -m py_compile scripts/*.py tests/*.py` | 0 | PASS |
| Docker build | Not attempted | — | Dockerfile is a placeholder (`python:3.12-alpine`); no application to containerize |
| Package validation | N/A | — | No packages to publish |

## LINT

| Check | Command | Exit | Result |
|-------|---------|------|--------|
| Go fmt | `cd backend && gofmt -l .` | 0 | Clean (0 files) — fixed during audit |
| Go vet | `cd backend && go vet ./...` | 0 | PASS |
| golangci-lint | Not configured | — | Not applicable |
| TS/ESLint | Not applicable | — | No frontend |
| ruff (scripts + tests) | `ruff check scripts/ tests/` | 1 | Pre-existing issues: `github_admin.py` UP022 (unsafe fix — NOT applied), `test_bootstrap.py` I001, `test_independent_review.py` I001 |
| ruff (new test file) | `ruff check tests/test_security_automation_retired.py` | 0 | PASS |
| ShellCheck (scripts) | `shellcheck scripts/*.sh .github/scripts/*.sh` | 1 | SC2034 unused-variable warnings only (pre-existing, not security defects) |
| PowerShell | N/A | — | No PowerShell scripts |

### Pre-existing lint findings (not introduced by this audit; left as-is to avoid scope creep)

- `scripts/github_admin.py:31` — UP022 `capture_output` suggestion is explicitly an **unsafe fix** per ruff; not applied.
- `tests/test_bootstrap.py:3` — I001 import sorting.
- `tests/test_independent_review.py:3` — I001 import sorting.
- `shellcheck` — SC2034 unused variables in `deploy-workflow-all-repos.sh` (deleted), `gh-switch.sh`, `post-review-comment.sh` (deleted). Remaining scripts (`audit-repo.sh`, `audit-all-repos.sh`, `bootstrap`? no shell), `gh-switch.sh`, `rollout-independent-review.sh` — re-check needed after deletions.

## TEST

| Check | Command | Exit | Result |
|-------|---------|------|--------|
| Go unit tests | `cd backend && go test ./...` | 0 | PASS (camfrog, config, httpapi) |
| Go race | `cd backend && go test -race ./...` | 0 | PASS |
| Go cover | `cd backend && go test -cover ./...` | 0 | PASS (camfrog 84.4%, config 90.9%, httpapi 100%) |
| Python unit tests | `python3 -m unittest discover -s tests -v` | 0 | 22/22 PASS (incl. 6 new regression tests) |
| pytest | Not applicable | — | No Python test framework configured; unittest used |
| Frontend tests | Not applicable | — | No frontend |
| Tauri tests | Not applicable | — | No desktop |
| Integration tests | PARTIAL | Go package-level tests only; no cross-service integration | See integration-test-report.md |
| E2E tests | Not applicable | — | Not configured |
| Regression tests | `tests/test_security_automation_retired.py` | 0 | 6/6 PASS |

## Secret scans

- `git ls-files | grep -E '(\.env$|id_rsa|id_ed25519|.*\.pem|.*\.key$)' | grep -v '^\.env\.example$'` — no tracked secret-bearing files (VERIFIED)
- CI secret-name check present in `.github/workflows/ci.yml:42-49`

## CodeQL / Dependency Review

- Not run locally (GitHub Actions cloud). Workflows present and pinned to immutable SHAs.

## YAML / Markdown validation

- `python3 scripts/validate_repo.py` — PASS (repository structure + local Markdown links)
- `.github/workflows/*.yml` YAML valid (validated by CI ruby YAML load)

## License checks

- LICENSE: MIT License. Attribution preserved (template source retains original copyright per README).

## Test skips and blockers

- Windows/Tauri tests: BLOCKED (Linux/WSL environment; no Windows target) — not fabricated.
- pytest: NOT APPLICABLE (no pytest configuration; unittest is the test runner).
- CodeQL/dependency-review/secret-scanning effective-state: BLOCKED (GitHub Actions cloud; requires API read-back).
- `golangci-lint`: NOT APPLICABLE (not configured).
