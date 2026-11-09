# All Pull Request Inventory — cvsz/cfc-bot-manager

**Method:** `gh api --paginate repos/cvsz/cfc-bot-manager/pulls -X GET -f state=all -f per_page=100` (complete pagination; single page returned all 14 PRs)
**PR #5:** Does not exist as a pull request — #5 is Issue #5 ("[P0] Retire legacy privileged automation...").
**Total discovered:** 14 pull requests
**Date:** 2026-11-08
**Verified against:** GitHub API (gh auth as cvsz, active account)

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
| 10 | https://github.com/cvsz/cfc-bot-manager/pull/10 | closed | no | security: retire legacy privileged automation (P0 Issue #5) and harden validation | cvsz | main@8caf9e4 | feat/cfc-bot-manager-enterprise-hardening@dc7109d | 8c863be |
| 11 | https://github.com/cvsz/cfc-bot-manager/pull/11 | closed | no | docs: update audit docs to reflect PR #10 merge and process deviation | cvsz | main@8c863be | docs/update-post-merge-status@dd6b6b3 | ea67566 |
| 12 | https://github.com/cvsz/cfc-bot-manager/pull/12 | closed | no | docs: add comprehensive final audit and remediation report | cvsz | main@ea67566 | docs/final-report@4e9ea6f | 13a5fd3 |
| 13 | https://github.com/cvsz/cfc-bot-manager/pull/13 | closed | no | docs: align audit reports with current main HEAD and branch protection | policedbc | main@13a5fd3 | feat/cfc-bot-manager-audit-alignment@45ebba3 | null |
| 14 | https://github.com/cvsz/cfc-bot-manager/pull/14 | closed | no | docs: align audit reports with current main HEAD and branch protection | cvsz | main@13a5fd3 | feat/cfc-bot-manager-audit-alignment@45ebba3 | 9411d27 |
| 15 | https://github.com/cvsz/cfc-bot-manager/pull/15 | closed | no | docs: reflect post-merge main SHA 9411d27 (PR #14) in audit reports | cvsz | main@9411d27 | docs/update-post-merge-status@999c35e | 2ad7be0 |

## Counts

- Open: 0
- Draft: 1 (PR #4, closed as draft; not merged)
- Merged: 12 (PRs #1, #2, #3, #6, #7, #8, #9, #10, #11, #12, #14, #15)
- Closed without merge: 2 (PR #4 closed draft; PR #13 closed without merge_sha)
- Dependabot: 2 (PRs #1, #2)

## Review submissions (from GitHub API)

| # | Reviewers | States |
|---|-----------|--------|
| 1 | cvsz | APPROVED (self-approval, author == approver) |
| 2 | cvsz | APPROVED (self-approval, author == approver) |
| 3 | — | no reviews |
| 4 | — | no reviews (draft, closed) |
| 6 | policedbc | APPROVED (independent) |
| 7 | — | no reviews |
| 8 | — | no reviews |
| 9 | — | no reviews |
| 10 | policedbc | REQUEST_CHANGES on dd759bc (prior head); no APPROVED on final head dc7109d |
| 11 | — | no reviews |
| 12 | — | no reviews |
| 13 | — | no reviews (policedbc author; closed without merge) |
| 14 | policedbc | APPROVED review on exact head 45ebba3 |
| 15 | policedbc | APPROVED review on exact head 999c35e |

## Notes

- No PR #5 exists; #5 is Issue #5.
- PR #13 was created by policedbc and closed without merge; its content was superseded by PR #14 (cvsz author, same head SHA 45ebba3).
- PRs #14 and #15 followed the controlled merge cycle: cvsz GPG-signed implementation commit → policedbc independent APPROVED review on exact head → CI/CodeQL/dependency-review SUCCESS → cvsz merge.
- PR #10 (P0 remediation) was merged via GitHub merge commit 8c863be before policedbc submitted an APPROVED review on the final head — documented process deviation.
- PRs #1 and #2 were self-approved by cvsz (not independent).
- PR #9 (Go backend) merged without recorded independent review.
- 100% of discovered PRs have a documented disposition (see all-pr-disposition.md).
