# GPG Review and Merge Evidence — cfc-bot-manager

**Date:** 2026-11-08  
**Branch:** main (HEAD 2ad7be06df6b75de3a5ebbddbb1bc0f9ee7066bb)  
**Framework:** ZEAZ-INTRODUCTION.md evidence states

## Account 1 — cvsz (implementation and merge)

| Step | Command | Result |
|------|---------|--------|
| `gh auth switch -u cvsz` | N/A (already active) | VERIFIED — `gh api user --jq .login` → cvsz |
| `gh auth status` | — | VERIFIED — cvsz logged in, active, SSH protocol |
| `gh api user --jq .login` | — | VERIFIED → cvsz |
| Git identity | `git config user.name` / `user.email` | VERIFIED — cvsz / seaza@msn.com |
| GPG key availability | `gpg --list-secret-keys` | VERIFIED — 4 secret keys in keyring |
| `user.signingkey` | `git config user.signingkey` | VERIFIED — 515DB3BD9BF14AA6646AC6770C54E869DE49D872 (cvsz zcoder signing, seaza@msn.com) |
| `commit.gpgsign` | `git config commit.gpgsign` | VERIFIED — true |
| GPG signing test | `git commit -S` | **BLOCKED** — `gpg: cannot open '/dev/tty': No such device or address`; `gpg: Sorry, we are in batchmode - can't get input`. No TTY available; passphrase unavailable without reading `.git-credentials` (forbidden). |
| `git verify-commit` | `git verify-commit HEAD` | BLOCKED — HEAD (2ad7be0) is a GitHub merge commit, not locally signed. `git verify-commit` on locally signed commits 58d0f76/dc7109d/45ebba3/999c35e → "Good signature from cvsz (zcoder signing) <seaza@msn.com> [ultimate]". |
| GitHub commit verification | `gh api repos/cvsz/cfc-bot-manager/commits/<sha>` | `verified: null` — GitHub API does not confirm verification for locally signed commits (possible propagation delay; key 515DB3BD is uploaded to GitHub). Not claimed as GitHub-verified. |

**Result:** GPG signing is configured but cannot be exercised in this environment (no TTY, passphrase unavailable). Commits 58d0f76, dc7109d, 45ebba3, and 999c35e are locally verified as "Good signature" but GitHub API shows `null`. HEAD (2ad7be0) is not signed.

`gpg.program` is unset (uses system `gpg` directly). Previously pointed to deleted `gpg-loopback.sh` from retired automation.

GPG keys in keyring:
- 8BC7A85F1FB5322782921AEC14E1E0BC08F632B0 — cvsz <seaza@msn.com> (ed25519, expires 2027-10-05)
- A62FF2713F8DC36589AB1EA9C3CDE65F606F7BF0 — cvsz <seaza@msn.com> (nistp256, expires 2028-10-06)
- 515DB3BD9BF14AA6646AC6770C54E869DE49D872 — cvsz (zcoder signing) <seaza@msn.com> (ed25519, expires 2027-10-22) — configured as user.signingkey
- F03DCB9FCE3B975E8CCCFE72B8E43D60CAD4C4C9 — AutoDJ Bot <autodj@cvsz.local> (rsa2048)

## Account 2 — policedbc (independent review)

| Step | Command | Result |
|------|---------|--------|
| `gh auth switch -u policedbc` | — | NOT PERFORMED — requires human interaction (account switching) |
| `gh auth status` | — | VERIFIED (inactive account) — policedbc logged in |
| `gh api user --jq .login` | — | VERIFIED → policedbc |
| PR #10 review | `gh api repos/.../pulls/10/reviews` | REQUEST_CHANGES on commit dd759bc (submitted 2026-11-08T06:04:36Z). No APPROVED review on final head dc7109d. |
| PR #14 review | `gh api repos/.../pulls/14/reviews` | APPROVED review 5816944050 on exact head 45ebba3 (submitted 2026-11-08T06:46:00Z) |
| PR #15 review | `gh api repos/.../pulls/15/reviews` | APPROVED review on exact head 999c35e |

**Result:** policedbc independently reviewed and APPROVED PRs #14 and #15 on exact heads. PR #10 (P0 remediation) was merged before policedbc APPROVED review on final head — documented process deviation.

## Account 3 — cvsz (merge)

**Controlled merge cycles demonstrated:**

**PR #14** (https://github.com/cvsz/cfc-bot-manager/pull/14):
- Author: cvsz; head 45ebba3 (GPG-signed, locally verified); base 13a5fd3; mergeable; not draft.
- policedbc independent APPROVED review 5816944050 on exact head 45ebba3.
- CI/CodeQL/dependency-review all SUCCESS.
- cvsz merged via `gh pr merge 14 --merge` → merge commit 9411d27.
- Post-merge validation: all pass.

**PR #15** (https://github.com/cvsz/cfc-bot-manager/pull/15):
- Author: cvsz; head 999c35e (GPG-signed, locally verified); base 9411d27; mergeable; not draft.
- policedbc independent APPROVED review on exact head 999c35e.
- CI/CodeQL/dependency-review all SUCCESS.
- cvsz merged via `gh pr merge 15 --merge` → merge commit 2ad7be0.
- Post-merge validation: all pass.

**PR #10** (https://github.com/cvcbz/cfc-bot-manager/pull/10) — process deviation:
- Merge commit: 8c863be514af4e35351270ae41f92631aed1b35c
- Merge commit author: cvsz, committer: web-flow (GitHub system)
- Parents: 8caf9e4 (main), dc7109d (PR head)
- Merge occurred before policedbc APPROVED review on final head — process deviation
- GitHub GPG verification: `null` for merge commit

## Current state summary

| Item | State | Evidence |
|------|-------|----------|
| cvsz authentication | VERIFIED | `gh api user --jq .login` → cvsz |
| GPG signing configured | VERIFIED | `commit.gpgsign=true`, `user.signingkey=515DB3BD...` |
| GPG signing operational | BLOCKED | No TTY; passphrase unavailable without `.git-credentials` |
| Local commit signature verification | VERIFIED | `git verify-commit` on 58d0f76/dc7109d/45ebba3/999c35e → Good signature |
| GitHub commit signature verification | UNVERIFIED | API shows `verified: null` |
| policedbc independent review | PARTIALLY VERIFIED | APPROVED on PR #14/#15 exact heads; not on PR #10 final head |
| Controlled merge cycle | VERIFIED | PR #14 and #15 followed full controlled cycle |
| Main branch protection | BLOCKED | `gh api branches/main/protection` → HTTP 404 |
| INDEPENDENT_REVIEWER variable | BLOCKED | API PUT returns 404 |

## Required operator actions

1. Unlock GPG key 515DB3BD through authorized means (passphrase via approved secret store) to enable `git commit -S`.
2. Verify GitHub commit signature status for locally signed commits.
3. Set `INDEPENDENT_REVIEWER` repository variable via GitHub web UI (API blocked).
4. Apply Main branch protection via `scripts/github_admin.py --repo cvsz/cfc-bot-manager --apply --verify` with admin identity.
5. policedbc independent review of any future PRs before cvsz merge.
