# All Pull Request Inventory — cvsz/cfc-bot-manager

**Method:** `gh api --paginate repos/cvsz/cfc-bot-manager/pulls -X GET -f state=all -f per_page=100` (complete pagination; single page returned all 8 PRs)
**PR #5:** Does not exist as a pull request — #5 is Issue #5 ("[P0] Retire legacy privileged automation...").
**Total discovered:** 8 pull requests
**Date:** 2026-11-08

## Complete inventory (all states)

| # | URL | State | Draft | Title | Author | Base (SHA) | Head (SHA) | Merge SHA |
|---|-----|-------|-------|-------|--------|------------|------------|-----------|
| 1 | https://github.com/cvsz/cfc-bot-manager/pull/1 | merged | no | Bump github/codeql-action/analyze from 4.39.0 to 4.39.3 | dependabot[bot] | main@5e020ff | dependabot/github_actions/github/codeql-action/analyze-4.39.3@5af2666 | 25242fb |
| 2 | https://github.com/cvsz/cfc-bot-manager/pull/2 | merged | no | Bump github/codeql-action/init from 4.39.0 to 4.39.3 | dependabot[bot] | main@5e020ff | dependabot/github_actions/github/codeql-action/init-4.39.3@c72083d | 68e3ee6 |
| 3 | https://github.com/cvsz/cfc-bot-manager/pull/3 | merged | no | chore: initialize cfc bot manager repository | cvsz | main@5e020ff | chore/bootstrap-cfc-bot-manager@6c21ce1 | 6c21ce1 |
| 4 | https://github.com/cvsz/cfc-bot-manager/pull/4 | closed | YES | feat: establish CFC Bot Manager product foundation | cvsz | main@5e020ff | feat/camfrog-manager-product-foundation@6cad5c5 | null |
| 6 | https://github.com/cvsz/cfc-bot-manager/pull/6 | merged | no | fix: remove embedded GPG passphrase and Windows paths from scripts | cvsz | main@5e020ff | fix/security-remove-embedded-secrets-and-windows-paths@4bd5b31 | 26436c0 |
| 7 | https://github.com/cvsz/cfc-bot-manager/pull/7 | merged | no | fix: redact credential details from public audit report | cvsz | main@26436c0 | fix/redact-exposed-secret-from-public-report@875b974 | b2b4c64 |
| 8 | https://github.com/cvsz/cfc-bot-manager/pull/8 | merged | no | feat: add CFC Bot Manager product architecture and requirements | cvsz | main@26436c0 | feat/cfc-bot-manager-product-foundation-v2@1d76f0c | 8caf9e4 |
| 9 | https://github.com/cvsz/cfc-bot-manager/pull/9 | merged | no | feat: add Go backend foundation and CI validation | cvsz | main@26436c0 | feat/go-backend-foundation@1fc607a | 7c609dd |
| 10 | https://github.com/cvsz/cfc-bot-manager/pull/10 | open | no | security: retire legacy privileged automation (P0 Issue #5) and harden validation | cvsz | main@8caf9e4 | feat/cfc-bot-manager-enterprise-hardening@dd759bc | null |

## Counts

- Open: 1 (PR #10)
- Draft: 1 (PR #4, closed as draft; not merged)
- Merged: 7 (PRs #1, #2, #3, #6, #7, #8, #9)
- Closed without merge: 2 (PR #4 closed draft; none others)
- Dependabot: 2 (PRs #1, #2)

## Review submissions (from GitHub API)

| # | Reviewers | States |
|---|-----------|--------|
| 1 | cvsz | APPROVED (commit 5af2666) |
| 2 | cvsz | APPROVED (commit c72083d) |
| 3 | — | no reviews |
| 4 | — | no reviews (draft, closed) |
| 6 | policedbc | APPROVED (commit 4bd5b31) |
| 7 | — | no reviews |
| 8 | — | no reviews |
| 9 | — | no reviews |
| 10 | policedbc | requested (pending) |

Note: PRs #1 and #2 were self-approved by cvsz (author == approver). PR #6 was approved by policedbc (independent). Per repository policy, independent review is required; self-approval by the author is not independent.
