# Bugs and Findings — cfc-bot-manager

**Date:** 2026-11-08
**Classification:** Verified defects and findings from the audit. Each finding includes severity, file/line, evidence, impact, fix, and regression test where applicable.

## F-001 — P0: Legacy privileged automation (Issue #5)

- **Severity:** P0 (Security)
- **Files:**
  - `.github/workflows/auto-fix-open-prs.yml:33-35` — requires `secrets.GPG_PASSPHRASE`, `secrets.GPG_SIGNING_KEY`, `secrets.POLIEDBC_PAT`
  - `.github/workflows/automated-audit-review-approve.yml:34,93,103` — `secrets.POLIEDBC_PAT` as `GH_TOKEN`; `:98` `gh pr review --request-changes`
  - `.github/scripts/gpg-loopback.sh:5-6` — tracked signing wrapper reads `GPG_PASSPHRASE`
  - `scripts/auto-fix-loop.sh:168-169` — GPG loopback commit + `git push origin HEAD`
  - `scripts/deploy-workflow-all-repos.sh:23-25,74-75,88-94` — POLIEDBC_PAT from credential store; `gh secret set POLIEDBC_PAT`; `git push origin HEAD`
  - `docs/ai/playbooks/auto-audit-review-approve-policedbc.md` — obsolete automated-approval documentation
- **Evidence:** Grep matches across the tree (see Phase 1 investigation); Issue #5 text.
- **Impact:** Reviewer credential distribution (POLIEDBC_PAT), cross-repository mutation, direct-push automation, execution of repo-controlled source with reviewer credentials, tracked signing wrapper.
- **Fix:** All 7 files deleted (commit dd759bc, PR #10).
- **Regression test:** `tests/test_security_automation_retired.py` — 6 tests asserting retired files absent, no credential distribution in workflows, no automated `gh pr review`/`gh pr comment`, no script combining credential access with direct push.
- **State:** REMEDIATED (code); GPG signing + policedbc review + Main protection BLOCKED (operator actions).

## F-002 — P2: Go source not gofmt-clean

- **Severity:** P2 (Code quality)
- **Files:** `backend/cmd/api/main.go`, `backend/internal/camfrog/rpc.go`, `backend/internal/config/config.go`, `backend/internal/httpapi/server.go`, `backend/internal/httpapi/server_test.go`
- **Evidence:** `gofmt -l .` listed 5 files before fix.
- **Impact:** Inconsistent formatting; CI does not enforce gofmt.
- **Fix:** `gofmt -w .` applied (26 insertions, 26 deletions across 5 files).
- **Regression test:** `gofmt -l .` returns 0 files (verified).
- **State:** FIXED (VERIFIED).

## F-003 — P2: Go test gaps (Phase 4 required coverage)

- **Severity:** P2 (Test coverage)
- **Missing:** duplicate request IDs, missing/unexpected parameters, concurrent RPC, invalid environment/host/port, unknown HTTP paths, security headers.
- **Fix:** Added tests to `rpc_test.go`, `config_test.go`, `server_test.go`.
- **State:** FIXED (VERIFIED; go test -race PASS).

## F-004 — P1: Governance — independent review not consistently applied

- **Severity:** P1 (Governance)
- **Evidence:** PR reviews — PRs #1/#2 self-approved by cvsz (author == approver); PRs #3, #7, #8, #9 no review submissions; only PR #6 independently approved by policedbc. `INDEPENDENT_REVIEWER` variable not configured.
- **Impact:** Merge gate (`verify-independent-approval`) cannot enforce independent review.
- **Fix:** Preserved `request-independent-review.yml`; PR #10 requests policedbc review.
- **State:** BLOCKED — `gh auth switch -u policedbc` requires human interaction; `INDEPENDENT_REVIEWER` var not set.

## F-005 — P1: GPG signing blocked

- **Severity:** P1 (Release governance)
- **Evidence:** `git commit -S` failed: `NEED_PASSPHRASE` / `gpg: Sorry, we are in batchmode - can't get input`. Configured key `8BC7A85F...` requires passphrase.
- **Impact:** Commit dd759bc is UNSIGNED; cannot merge as signed.
- **Fix:** Operator must unlock GPG key (passphrase via authorized means) and re-sign.
- **State:** BLOCKED — passphrase not available without reading `.git-credentials` (forbidden).

## F-006 — P1: No authentication/authorization (NOT IMPLEMENTED)

- **Severity:** P1 (Security)
- **Evidence:** `backend/` has no authn/authz middleware; all endpoints unauthenticated.
- **Impact:** Endpoints exposed without authorization once bound beyond loopback.
- **Fix:** Implement authn/authz (future phase per implementation-plan.md).
- **State:** NOT IMPLEMENTED.

## F-007 — P1: No Camfrog network integration (NOT IMPLEMENTED)

- **Severity:** P1 (Functional)
- **Evidence:** `backend/internal/camfrog/rpc.go` is serialization/validation boundary only; no HTTP client, session, reconnection, event delivery.
- **Impact:** Protocol support cannot be claimed.
- **Fix:** Authorized integration evidence required before implementation.
- **State:** NOT IMPLEMENTED.

## Non-findings (reviewed, not defects)

- Go dependency-free module: no dependency vulnerabilities.
- Actions pinned to immutable SHAs: supply-chain compliant.
- No committed secrets: `git ls-files` secret scan clean.
- Plugin IDs 9 and 10 excluded from allowlist: correctly enforced in code.

## Open questions (not defects)

- Which Camfrog product/version is the supported target? (threat-model.md)
- What authentication/session lifecycle does the upstream adapter require?
- Can the upstream interface be used under applicable terms and authorization?
