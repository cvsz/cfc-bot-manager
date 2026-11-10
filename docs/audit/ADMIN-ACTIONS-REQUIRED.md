# Admin Actions Required — CFC Bot Manager

**Date:** 2026-11-09  
**Context:** Three critical security/governance actions require GitHub repository administrator permissions. This environment cannot perform them. An authorized operator with `admin` permission on `cvsz/cfc-bot-manager` must execute these.

---

## Required Actions (Priority Order)

### 1. Rotate Compromised Signing Key (SEC-001, P0)

**Background:** A signing-key passphrase was exposed in historical repository content (pre-PR #7). Per AGENTS.md §4.1, the key must be rotated. Shared Git history must not be rewritten.

**Actions Required:**
```bash
# 1. Generate new GPG signing key (run locally by authorized signer)
gpg --full-generate-key
# Select: RSA/RSA, 4096 bits, no expiry, name/email for project

# 2. Export public key for GitHub
gpg --armor --export <KEY_ID> > cfc-bot-manager-gpg.pub

# 3. Add to GitHub: Settings → SSH and GPG keys → New GPG key
# Paste the armored public key

# 4. Configure git to use new key (for each authorized committer)
git config --global user.signingkey <KEY_ID>
git config --global commit.gpgsign true

# 5. Update CI/CD signing (if using automated signing)
# Add GPG_PRIVATE_KEY as repository secret (Settings → Secrets → Actions)
# Update workflows that sign commits/releases

# 6. Revoke old key
gpg --gen-revoke <OLD_KEY_ID>
# Publish revocation to keyservers if previously published
```

**Verification:**
- New signed commits show "Verified" badge on GitHub
- Old key shows as revoked
- CHANGELOG.md updated with rotation timestamp

---

### 2. Set INDEPENDENT_REVIEWER Variable (SEC-002, P0)

**Background:** The `request-independent-review.yml` workflow requires this variable to function. It is currently unset.

**Actions Required:**
```bash
# Option A: GitHub CLI (requires auth with admin:repo_hook scope)
gh variable set INDEPENDENT_REVIEWER --body "<github-login>" --repo cvsz/cfc-bot-manager

# Option B: GitHub UI
# Settings → Secrets and variables → Actions → Variables → New repository variable
# Name: INDEPENDENT_REVIEWER
# Value: <github-login> (must be valid GitHub user/login, not the typical PR author)

# Option C: API (with PAT having repo scope)
curl -X POST -H "Authorization: Bearer <PAT>" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/cvsz/cfc-bot-manager/actions/variables \
  -d '{"name":"INDEPENDENT_REVIEWER","value":"<github-login>"}'
```

**Constraints from workflow (line 28-34):**
- Value must match `^[A-Za-z0-9-]+$`
- Must not be the same as typical PR authors (case-insensitive)

**Verification:**
```bash
gh api repos/cvsz/cfc-bot-manager/variables --jq '.variables[] | select(.name=="INDEPENDENT_REVIEWER")'
# Should return the configured login
```

---

### 3. Add Independent Review to Required Status Checks (SEC-003, P0)

**Background:** Branch protection on `main` requires three checks but not `verify-independent-approval`. After SEC-002, this must be added.

**Actions Required:**
```bash
# 1. Fetch current protection
gh api repos/cvsz/cfc-bot-manager/branches/main/protection > current-protection.json

# 2. Update required_status_checks.contexts to include "verify-independent-approval"
# Using jq to modify:
jq '.required_status_checks.contexts += ["verify-independent-approval"]' current-protection.json > updated-protection.json

# 3. Apply update (requires admin permission)
gh api --method PATCH repos/cvsz/cfc-bot-manager/branches/main/protection \
  --input updated-protection.json

# Alternative: GitHub UI
# Settings → Branches → Branch protection rules → main → Edit
# Add "verify-independent-approval" to "Require status checks to pass before merging"
```

**Verification:**
```bash
gh api repos/cvsz/cfc-bot-manager/branches/main/protection --jq '.required_status_checks.contexts'
# Should include: "dependency-review", "repository-baseline", "Analyze GitHub Actions", "verify-independent-approval"
```

**Test:** Open a test PR — it should request review from INDEPENDENT_REVIEWER and block merge until they approve the exact head SHA.

---

### 4. Enable Required Commit Signing (SEC-004, P1) — After SEC-001

**Background:** `required_signatures.enabled: false` on main branch.

**Actions Required:**
```bash
# GitHub UI: Settings → Branches → Branch protection rules → main → Edit
# Check "Require signed commits"

# Or API:
gh api --method PATCH repos/cvsz/cfc-bot-manager/branches/main/protection \
  -f required_signatures[enabled]=true
```

**Verification:**
```bash
gh api repos/cvsz/cfc-bot-manager/branches/main/protection --jq '.required_signatures.enabled'
# Should return true
```

---

## Summary for Operator

| Action | Command | Prerequisites |
|--------|---------|---------------|
| 1. Rotate signing key | Manual (GPG + GitHub UI) | Authorized signers available |
| 2. Set INDEPENDENT_REVIEWER | `gh variable set INDEPENDENT_REVIEWER --body "<login>" --repo cvsz/cfc-bot-manager` | Valid independent GitHub login |
| 3. Update branch protection | `gh api --method PATCH ...` with updated contexts | Action 2 complete |
| 4. Enable commit signing | `gh api --method PATCH ... -f required_signatures[enabled]=true` | Action 1 complete |

**Note:** Actions 2 and 3 can be done by any repository admin. Actions 1 and 4 require coordination with committers.

---

## Next Code Implementation (Unblocked)

While admin actions are pending, the following code work can proceed in parallel:

1. **WP-BE-001:** Domain entities & repository interfaces (`backend/internal/domain/`, `backend/internal/repository/`)
2. **WP-CA-001:** Complete CamfrogAdapter interface (`backend/internal/camfrog/` refactor)
3. **WP-BE-002:** PostgreSQL schema & migrations (after WP-BE-001)

These require no admin permissions and are the critical path for backend completion.