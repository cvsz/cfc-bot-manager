# GPG Review and Merge Evidence — cfc-bot-manager

**Date:** 2026-11-08
**Branch:** feat/cfc-bot-manager-enterprise-hardening
**Commit:** dd759bc
**PR:** #10 (https://github.com/cvc/cfc-bot-manager/pull/10)

## Account 1 — cvsz (implementation)

| Step | Command | Result |
|------|---------|--------|
| `gh auth switch -u cvsz` | N/A (already active) | VERIFIED — `gh api user --jq .login` → cvsz |
| `gh auth status` | — | VERIFIED — cvsz logged in, active |
| Git identity | `user.name=cvsz`, `user.email=seaza@msn.com` | VERIFIED (local/global config) |
| GPG key availability | `user.signingkey=8BC7A85F1FB5322782921AEC14E1E0BC08F632B0` | VERIFIED (key present in keyring) |
| `commit.gpgsign=true` | `git config commit.gpgsign` → true | VERIFIED |
| GPG signing test | `git commit -S` | BLOCKED — key requires passphrase in batch mode; `NEED_PASSPHRASE`; `gpg: Sorry, we are in batchmode - can't get input` |

**Result:** Commit `dd759bc` is **intentionally UNSIGNED**. The configured signing key requires a passphrase that is not available without reading `/home/cvsz/.git-credentials`, which is forbidden ("Never cat or print its contents", "never echo tokens/credentials"). Per the mission rule: "If signing fails, do not claim signed commits."

**Operator action required:** Unlock the GPG signing key (provide passphrase through authorized means, e.g. `gpg-agent --preset-passphrase` or interactive entry) and re-create the commit as `git commit -S`. Verify with `git verify-commit HEAD`.

## Account 2 — policedbc (independent review)

| Step | Command | Result |
|------|---------|--------|
| `gh auth switch -u policedbc` | — | BLOCKED — requires human browser interaction; not performed autonomously |
| `gh auth status` | — | policedbc authenticated but inactive |
| `gh api user --jq .login` | — | Not executed as policedbc |

**Result:** policedbc review of PR #10 was REQUESTED (via `gh api .../requested_reviewers -f reviewers[]=policedbc`), but no review has been submitted. Independent approval on the exact head SHA cannot be verified without switching to policedbc.

**Operator action required:** `gh auth switch -u policedbc`, review PR #10 (diff, CI, security, P0 remediation, credential boundary), and submit APPROVED on the exact head `dd759bc` (or request changes).

## Account 3 — cvsz (merge)

Not yet performed. Pre-merge verification checklist (from Phase 17):
1. policedbc independently approved PR #10 — BLOCKED (not done)
2. Approval applies to exact current head SHA — BLOCKED
3. Required CI passed — VERIFIED locally (`make ci` PASS); cloud CI (CodeQL, dependency-review, secret-scanning) not run locally
4. Build and integration tests passed — VERIFIED locally for Go/Python package-level
5. No unresolved blocking review threads — N/A (no review yet)
6. No new commits after approval — N/A
7. PR is open and not Draft — VERIFIED (PR #10 open)
8. Branch is mergeable — not verified via API
9. Main branch is target — VERIFIED (base=main)

## GPG signing configuration (for operator reference)

- `git config commit.gpgsign` → true
- `git config user.signingkey` → 8BC7A85F1FB5322782921AEC14E1E0BC08F632B0
- `gpg.program` → /home/cvsz/.local/bin/gpg-loopback.sh (reads GPG_PASSPHRASE env)
- `gpg.format` → openpgp
- Secret keys in keyring: 8BC7A85F... (cvsz), A62FF271... (cvsz), 515DB3BD... (cvsz zcoder signing), F03DCB9F... (AutoDJ Bot)

**Note:** `gpg-loopback.sh` was deleted as part of P0 remediation (tracked signing wrapper). The global `gpg.program` config still references it; operator should update or remove `gpg.program` and sign interactively or via `gpg-agent --preset-passphrase` after unlocking the key.

## Merge evidence

No merge has been performed. PR #10 is open, unsigned, awaiting policedbc review and GPG signing. Per governance, never self-approve, never force-merge, never bypass rules.
