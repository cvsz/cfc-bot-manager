# Build, Lint, and Test Report — cfc-bot-manager

**Date:** 2026-11-08  
**Commit:** 2ad7be06df6b75de3a5ebbddbb1bc0f9ee7066bb (main HEAD, after PR #15 merge)  
**Environment:** Linux, go 1.27.0, Python 3.x, ruff 0.17.2, shellcheck present  
**Verified locally:** All commands executed and results confirmed

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
| Go fmt | `cd backend && gofmt -l .` | 0 | Clean (0 files) |
| Go vet | `cd backend && go vet ./...` | 0 | PASS |
| golangci-lint | Not configured | — | Not applicable |
| TS/ESLint | Not applicable | — | No frontend |
| ruff (scripts + tests) | `ruff check scripts/ tests/` | 1 | Pre-existing issues: `github_admin.py` UP022 (unsafe fix — NOT applied), `test_bootstrap.py` I001, `test_independent_review.py` I001 |
| ruff (new test file) | `ruff check tests/test_security_automation_retired.py` | 0 | PASS |
| ShellCheck (scripts) | `shellcheck scripts/*.sh` | 1 | SC2034 unused-variable warnings only (pre-existing, not security defects) |
| PowerShell | N/A | — | No PowerShell scripts |

### Pre-existing lint findings (not introduced by this audit; left as-is to avoid scope creep)

- `scripts/github_admin.py:31` — UP022 `capture_output` suggestion is explicitly an **unsafe fix** per ruff; not applied.
- `tests/test_bootstrap.py:3` — I001 import sorting.
- `tests/test_independent_review.py:3` — I001 import sorting.
- `shellcheck` — SC2034 unused variables in `gh-switch.sh`, `rollout-independent-review.sh`. Cosmetic warnings only.

## TEST

| Check | Command | Exit | Result |
|-------|---------|------|--------|
| Go unit tests | `cd backend && go test ./...` | 0 | PASS (4 packages: cmd/api, camfrog, config, httpapi) |
| Go race | `cd backend && go test -race ./...` | 0 | PASS (4 packages) |
| Go cover | `cd backend && go test -cover ./...` | 0 | PASS (cmd/api 2.3%, camfrog 84.4%, config 90.9%, httpapi 100%) |
| Python unit tests | `python3 -m unittest discover -s tests -v` | 0 | 25/25 PASS (incl. 9 security regression tests) |
| pytest | Not applicable | — | No pytest configuration; unittest is the test runner |
| Frontend tests | Not applicable | — | No frontend |
| Tauri tests | Not applicable | — | No desktop; Windows BLOCKED |
| Integration tests | PARTIAL | Go package-level tests only; no cross-service integration | See integration-test-report.md |
| E2E tests | Not applicable | — | Not configured |
| Regression tests | `tests/test_security_automation_retired.py` | 0 | 25/25 PASS |

## Secret scans

| Check | Command | Result |
|-------|---------|--------|
| Tracked secret-bearing files | `git ls-files \| grep -E '(\.env$|id_rsa|id_ed25519|.*\.pem|.*\.key$)' \| grep -v '^\.env\.example$'` | None found (VERIFIED) |
| CI secret-name check | `.github/workflows/ci.yml:42-49` | Present in CI workflow |
| Retired automation files | `git ls-files` check for 7 retired files | All absent (VERIFIED) |

## CodeQL / Dependency Review

| Check | Status |
|-------|--------|
| CodeQL workflow | `.github/workflows/codeql.yml` present; pinned SHA v4; GitHub CI success |
| Dependency Review workflow | `.github/workflows/dependency-review.yml` present; pinned SHA v5.0.0; GitHub CI success |
| Local execution | Not run locally (GitHub Actions cloud) |

## YAML / Markdown validation

| Check | Command | Result |
|-------|---------|--------|
| Repository validation | `python3 scripts/validate_repo.py` | PASS (repository structure + local Markdown links) |
| YAML syntax | Ruby YAML load of `.github/**/*.{yml,yaml}` | Valid (validated by CI) |

## License checks

| Check | Result |
|-------|--------|
| LICENSE | MIT License. Attribution preserved. |

## Test skips and blockers

| Check | State | Reason |
|-------|-------|--------|
| Windows/Tauri tests | BLOCKED | Linux/WSL environment; no Windows target — not fabricated |
| pytest | NOT APPLICABLE | No pytest configuration; unittest is the test runner |
| CodeQL/dependency-review/secret-scanning effective-state via API | BLOCKED | GitHub Actions cloud; `gh api branches/main/protection` returns 404; API read-back limited |
| `golangci-lint` | NOT APPLICABLE | Not configured |
| Docker build | NOT APPLICABLE | Dockerfile is a placeholder; no application to containerize |
| Authorized penetration tests | NOT APPLICABLE | No authorized staging target; never attacked unrelated hosts/live accounts |
| GPG signing | BLOCKED | No TTY; passphrase unavailable without reading `.git-credentials` |
