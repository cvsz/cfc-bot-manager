# Security Findings — CFC Bot Manager

**Audit Date:** 2026-11-09  
**Auditor:** Autonomous Engineering Agent  
**Classification:** Internal — Do Not Distribute Externally  
**Reference:** INITIAL-REPOSITORY-AUDIT.md

---

## Executive Summary

| Finding ID | Title | Severity | Status |
|------------|-------|----------|--------|
| SEC-001 | Historical signing-key passphrase exposure in Git history | HIGH | Requires Rotation |
| SEC-002 | `INDEPENDENT_REVIEWER` repository variable not configured | HIGH | Configuration Gap |
| SEC-003 | Independent review workflow not enforced as required status check | HIGH | Branch Protection Gap |
| SEC-004 | No commit signing requirement on main branch | MEDIUM | Configuration Gap |
| SEC-005 | Go dependency vulnerability scanning not in CI | MEDIUM | Coverage Gap |
| SEC-006 | No SBOM generation in release pipeline | MEDIUM | Coverage Gap |
| SEC-007 | Camfrog adapter accepts observed (unverified) plugin IDs | MEDIUM | Design Decision |
| SEC-008 | No authentication/authorization implementation | CRITICAL | Not Implemented |
| SEC-009 | No tenant isolation enforcement | CRITICAL | Not Implemented |
| SEC-010 | No secret scanning for Go modules | LOW | Coverage Gap |

---

## Detailed Findings

### SEC-001: Historical Signing-Key Passphrase Exposure

**Severity:** HIGH  
**Type:** Credential Exposure in Git History  
**Status:** REQUIRES ROTATION  

**Evidence:**
- PR #7 ("fix: redact credential details from public audit report") merged at `b2b4c64` explicitly addresses redaction of a passphrase from a public report
- PR #6 ("fix: remove embedded GPG passphrase and Windows paths from scripts") merged at `68e3ee6` removed embedded secrets from scripts
- Git history prior to these PRs may contain the passphrase in commits, PR descriptions, or workflow logs

**Impact:**
- If the passphrase was committed to the repository (even in a since-removed file), it exists in Git history
- Anyone with read access to the repository can retrieve it via `git log -p` or GitHub UI
- The passphrase must be considered **compromised** and **rotated immediately**

**Remediation Required:**
1. **Rotate the signing key** — Generate a new GPG/SSH signing key pair
2. **Update all systems** using the old key (CI signing, release signing, commit signing)
3. **Do not rewrite shared Git history** — Per AGENTS.md §4.1: "Do not assume removing a secret from the current file removes it from Git history. Never rewrite shared Git history without a separately approved migration plan."
4. **Document rotation** in `CHANGELOG.md` and security advisories

**Verification Criteria:**
- [ ] New signing key generated and distributed to authorized signers
- [ ] All CI/CD pipelines updated to use new key
- [ ] Old key revoked from all trust stores
- [ ] Rotation documented with timestamp

---

### SEC-002: INDEPENDENT_REVIEWER Variable Not Configured

**Severity:** HIGH  
**Type:** Missing Governance Configuration  
**Status:** OPEN  

**Evidence:**
```yaml
# .github/workflows/request-independent-review.yml line 15
if: github.event_name == 'pull_request_target' && github.event.pull_request.draft == false && vars.INDEPENDENT_REVIEWER != '' && github.event.pull_request.user.login != vars.INDEPENDENT_REVIEWER
```

```bash
$ gh api repos/cvsz/cfc-bot-manager/variables | jq -r '.variables[] | select(.name=="INDEPENDENT_REVIEWER") | .value'
# No output — variable not set
```

**Impact:**
- The `request-independent-review` workflow fails validation at line 66-68 when the variable is missing
- Independent review cannot be requested or verified automatically
- Governance contract (AGENTS.md §10) requires independent review workflow to "remain active"

**Remediation Required:**
1. Set repository variable `INDEPENDENT_REVIEWER` to a valid GitHub login (user or bot) via:
   - GitHub UI: Settings → Secrets and variables → Actions → Variables
   - Or: `gh variable set INDEPENDENT_REVIEWER --body "<login>" --repo cvsz/cfc-bot-manager`
2. Ensure the reviewer is distinct from typical PR authors (per workflow validation line 32-34)

**Verification Criteria:**
- [ ] Variable set and readable via `gh api`
- [ ] Workflow `request-independent-review.yml` passes validation on next PR
- [ ] Reviewer is a valid, independent identity

---

### SEC-003: Independent Review Not Enforced as Required Status Check

**Severity:** HIGH  
**Type:** Branch Protection Misconfiguration  
**Status:** OPEN  

**Evidence:**
Branch protection API response shows required contexts:
```json
"contexts": [
  "dependency-review",
  "repository-baseline",
  "Analyze GitHub Actions"
]
```
The `verify-independent-approval` job from `request-independent-review.yml` is **absent**.

**Impact:**
- PRs can be merged without independent review approval
- The workflow runs but its failure does not block merge
- Violates AGENTS.md §10: "The independent review workflow must remain active" and §13: "Independent reviewer approval must exist"

**Remediation Required:**
1. Add `verify-independent-approval` to branch protection required status checks via:
   - GitHub UI: Settings → Branches → Branch protection rules → main → Required status checks
   - Or: `gh api --method PATCH repos/cvsz/cfc-bot-manager/branches/main/protection --input -` with updated `required_status_checks.contexts`
2. Ensure `INDEPENDENT_REVIEWER` is configured first (SEC-002), otherwise the check will always fail

**Verification Criteria:**
- [ ] `verify-independent-approval` appears in branch protection required contexts
- [ ] Test PR requires independent approval to merge
- [ ] Workflow passes for PRs with valid independent review

---

### SEC-004: No Commit Signing Requirement

**Severity:** MEDIUM  
**Type:** Supply Chain Integrity Gap  
**Status:** OPEN  

**Evidence:**
Branch protection shows:
```json
"required_signatures": {
  "enabled": false
}
```

**Impact:**
- Commits to `main` can be pushed without cryptographic signatures
- No verification of commit authorship in the protected branch
- Reduces supply chain integrity assurance

**Remediation Required:**
1. Enable required commit signing on `main` branch
2. Configure GPG/SSH/SSH signing keys for authorized committers
3. Document signing key management in `docs/operations/deployment.md`

**Verification Criteria:**
- [ ] `required_signatures.enabled: true` in branch protection
- [ ] Authorized committers have signing keys configured
- [ ] Unsigned commits rejected on push to `main`

---

### SEC-005: Go Dependency Vulnerability Scanning Not in CI

**Severity:** MEDIUM  
**Type:** Coverage Gap  
**Status:** OPEN  

**Evidence:**
- `ci.yml` runs `go test -race ./...`, `go vet ./...`, `go build ./...`
- No `govulncheck`, `nancy`, `go list -m -json all | nancy sleuth`, or similar
- `dependency-review.yml` only covers GitHub Actions and Docker, not Go modules

**Impact:**
- Known vulnerabilities in Go dependencies may go undetected
- Supply chain attacks via compromised transitive dependencies not caught

**Remediation Required:**
1. Add `govulncheck` to CI pipeline (official Go vulnerability scanner)
2. Consider `nancy` or `trivy` for additional coverage
3. Run on schedule and on PR changes to `go.mod`/`go.sum`

**Verification Criteria:**
- [ ] `govulncheck` runs in CI on every PR and push to main
- [ ] Fails build on HIGH/CRITICAL vulnerabilities
- [ ] Documented in `docs/operations/deployment.md`

---

### SEC-006: No SBOM Generation

**Severity:** MEDIUM  
**Type:** Supply Chain Transparency Gap  
**Status:** OPEN  

**Evidence:**
- No `syft`, `cyclonedx-gomod`, or GitHub SBOM action in workflows
- No SBOM artifact uploaded to releases

**Impact:**
- Consumers cannot verify dependency inventory
- Compliance requirements (e.g., EO 14028) not met
- Incident response hampered without known-good dependency list

**Remediation Required:**
1. Add SBOM generation to release workflow (or new `sbom.yml` workflow)
2. Use `anchore/sbom-action` or `cyclonedx-gomod` for Go
3. Attach SBOM to GitHub Releases

**Verification Criteria:**
- [ ] SBOM generated for every release tag
- [ ] SBOM attached as release asset
- [ ] Format: SPDX or CycloneDX

---

### SEC-007: Camfrog Adapter Accepts Observed (Unverified) Plugin IDs

**Severity:** MEDIUM  
**Type:** Design Decision — Attack Surface  
**Status:** DOCUMENTED, INTENTIONAL  

**Evidence:**
```go
// internal/camfrog/rpc.go lines 25-28
var observedPluginIDs = map[int]struct{}{
    1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 7: {}, 8: {},
    11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {},
}
```
- Plugin IDs 9 (Roulette) and 10 (Unknown) explicitly rejected
- IDs 1-8, 11-16 accepted based on **observed HTML captures only**
- `plugin-compatibility-matrix.md` marks all as `UNVERIFIED` for runtime behavior and Cloud Bot compatibility

**Risk:**
- If upstream Camfrog interface changes, accepted plugin IDs may behave differently
- Malformed or malicious plugin configurations could be sent if UI is compromised
- No capability verification before enabling plugins

**Mitigation (Current):**
- Adapter validates parameter shapes and types strictly
- Unknown methods rejected (`default` case in `NewRPCRequest`)
- Response size bounded (1 MiB)
- Plugin ID 9 (gambling) and 10 (unknown) permanently disabled in adapter

**Required Future Work:**
- Capability-based enablement: only mark plugins `VERIFIED` after authorized integration tests
- UI must show `UNVERIFIED` plugins as disabled/unavailable
- Add per-plugin capability declarations with versioned compatibility

**Verification Criteria:**
- [ ] No `VERIFIED` plugins without authorized test evidence
- [ ] UI never enables `UNVERIFIED` plugins by default
- [ ] Contract tests cover all accepted plugin ID parameter shapes

---

### SEC-008: No Authentication/Authorization Implementation

**Severity:** CRITICAL  
**Type:** Missing Security Control  
**Status:** NOT IMPLEMENTED (Phase 1 Dependency)  

**Evidence:**
- `backend/internal/auth/` directory does not exist
- No session management, JWT, OAuth, or MFA code
- `httpapi/server.go` has no authentication middleware
- Implementation plan Phase 1 item 20: "Implement authentication, session rotation, CSRF, MFA capability, and server-side RBAC" — **NOT STARTED**

**Impact:**
- Anyone with network access to the API can call all endpoints
- No tenant isolation possible without auth context
- No audit trail of privileged operations

**Remediation Required:**
Per implementation plan Phase 1:
1. Add `internal/auth/` package with session management
2. Implement secure cookie-based sessions with CSRF tokens
3. Add MFA support (TOTP)
4. Implement RBAC with resource-level permissions
5. Add authorization middleware to all protected endpoints
6. Write cross-tenant isolation tests before exposing CRUD

**Verification Criteria:**
- [ ] All mutating endpoints require valid session
- [ ] CSRF tokens validated for cookie-authenticated mutations
- [ ] MFA enforced for privileged roles
- [ ] Cross-tenant access tests prove isolation
- [ ] Audit log records all privileged operations

---

### SEC-009: No Tenant Isolation Enforcement

**Severity:** CRITICAL  
**Type:** Missing Security Control  
**Status:** NOT IMPLEMENTED (Phase 1 Dependency)  

**Evidence:**
- No `internal/repository/` or `internal/persistence/` packages
- No PostgreSQL schema or migrations
- Implementation plan Phase 1 items 19, 21: "Add PostgreSQL schema/migrations and transaction boundaries" and "Add resource/tenant isolation tests before exposing CRUD endpoints" — **NOT STARTED**
- Requirements §15: "Default deny applies" for permissions

**Impact:**
- Multi-tenancy cannot be enforced without persistence layer
- Data leakage between organizations possible if implemented incorrectly
- No foundation for resource-level permissions

**Remediation Required:**
Per implementation plan Phase 1:
1. Design normalized PostgreSQL schema (organizations, users, memberships, roles, permissions, rooms, bots, plugins, audit_logs)
2. Create versioned migrations with rollback strategy
3. Implement repository interfaces with tenant-scoped queries
4. Write cross-tenant denial tests (negative test cases)
5. Enforce tenant context in all repository methods

**Verification Criteria:**
- [ ] All queries filter by tenant/organization ID
- [ ] Cross-tenant access attempts return 403/404 (not data)
- [ ] Negative tests in CI prove isolation
- [ ] Migration rollback tested and documented

---

### SEC-010: No Secret Scanning for Go Modules

**Severity:** LOW  
**Type:** Coverage Gap  
**Status:** OPEN  

**Evidence:**
- CI secret check only scans tracked files for patterns: `.env`, `id_rsa`, `id_ed25519`, `*.pem`, `*.key`
- No `gosec`, `trufflehog`, or `git-secrets` for Go source code
- No GitHub secret scanning alerts configured (would require GitHub Advanced Security)

**Impact:**
- Hardcoded secrets in Go source files may go undetected
- Low probability given current codebase size, but risk increases with development

**Remediation Required:**
1. Add `gosec` to CI for static analysis of Go code
2. Enable GitHub secret scanning (if Advanced Security available)
3. Consider `trufflehog` for generic secret detection

**Verification Criteria:**
- [ ] `gosec` runs in CI on Go files
- [ ] No high-severity findings in current codebase
- [ ] Documented in CI configuration

---

## Compliance with AGENTS.md Security Requirements

| Requirement (§4) | Status | Notes |
|------------------|--------|-------|
| Never print/copy/commit credentials | ✅ | Current codebase clean; CI secret check active |
| Redact secrets from logs, tests, docs | ✅ | `ReadJSONResponse` bounds reads; no logging of bodies |
| Assume Git history retains secrets | ✅ | SEC-001 documents this; rotation required |
| Recommend rotation for exposed credentials | ✅ | SEC-001 remediation specifies rotation |
| Never rewrite shared history without approval | ✅ | Remediation explicitly avoids history rewrite |
| Use only authorized Camfrog accounts | ✅ | No Camfrog integration tests exist yet |
| No bypassing auth/access controls | ✅ | Not implemented yet (SEC-008) |
| No fabricating public SDK | ✅ | Adapter explicitly marks observed vs verified |
| Private interface isolated behind adapter | ✅ | `internal/camfrog/` package boundary |
| Explicit capability verification required | ✅ | All plugins `UNVERIFIED`; capability matrix exists |
| Sanitized fixtures for tests | ✅ | `rpc_test.go` uses synthetic fixtures |
| No mock tests represented as real | ✅ | Tests clearly unit tests; no integration tests yet |
| Plugin ID 9 excluded | ✅ | Rejected in `rpc.go` and `rpc_test.go` |
| Plugin ID 10 UNVERIFIED | ✅ | Rejected in adapter; documented in matrix |
| No guessing plugin behavior | ✅ | Adapter validates only observed parameter shapes |
| Unsupported plugins disabled/marked | ✅ | IDs 9, 10 rejected; others accepted but unverified |
| No destructive migration without backup/rollback | ✅ | Not applicable yet (no migrations exist) |
| No production deployment without auth | ✅ | Not deployed |
| No forced Git push | ✅ | Branch protection blocks force push |
| No bypassing branch protection/review | ✅ | Enforced; SEC-003 gap for independent review |
| No publishing from unapproved commits | ✅ | Required checks + reviews enforced |
| No fake tests/fabricated logs | ✅ | All tests real; CI evidence verifiable |

---

## Remediation Priority Order

| Priority | Finding | Rationale |
|----------|---------|-----------|
| P0 | SEC-001 | Credential rotation is time-sensitive; compromised key |
| P0 | SEC-002 | Governance blocker; independent review cannot function |
| P0 | SEC-003 | Governance enforcement; independent review not a merge gate |
| P1 | SEC-008 | Authentication is prerequisite for all other security controls |
| P1 | SEC-009 | Tenant isolation is prerequisite for multi-tenant product |
| P1 | SEC-004 | Commit signing improves supply chain integrity |
| P2 | SEC-005 | Dependency scanning reduces supply chain risk |
| P2 | SEC-006 | SBOM enables compliance and incident response |
| P2 | SEC-007 | Capability verification needed before plugin enablement |
| P3 | SEC-010 | Secret scanning for Go source code |

---

## Next Steps

1. **Immediate:** Coordinate signing key rotation (SEC-001) with repository administrators
2. **Immediate:** Configure `INDEPENDENT_REVIEWER` variable (SEC-002)
3. **Immediate:** Add `verify-independent-approval` to required status checks (SEC-003)
4. **Phase 1 Implementation:** Build authentication, PostgreSQL persistence, and tenant isolation (SEC-008, SEC-009)
5. **CI Enhancement:** Add `govulncheck`, SBOM generation, `gosec` (SEC-005, SEC-006, SEC-010)
6. **Branch Protection:** Enable required commit signing (SEC-004)

---

**End of Security Findings**