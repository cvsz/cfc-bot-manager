# All Pull Request Matrix — cvsz/cfc-bot-manager

**Classification scheme:** MERGED_VERIFIED / OPEN_ACTION_REQUIRED / DRAFT_INCOMPLETE / CLOSED_SUPERSEDED / DUPLICATE / SECURITY_REVIEW_REQUIRED / BLOCKED / READY_FOR_MERGE

**Date:** 2026-11-08
**Verified against:** GitHub API (gh auth as cvsz, active account)

| # | State | Draft | Base | Head | Classification | Merge verified on main? | Tests/docs on merge |
|---|-------|-------|------|------|----------------|-------------------------|---------------------|
| 1 | merged | no | main@5e020ff | codeql-action/analyze-4.39.3 | MERGED_VERIFIED | Yes (25242fb is ancestor of origin/main 2ad7be0) | CI workflow updated (codeql v4) |
| 2 | merged | no | main@5e020ff | codeql-action/init-4.39.3 | MERGED_VERIFIED | Yes (68e3ee6 is ancestor of origin/main 2ad7be0) | CI workflow updated (codeql v4) |
| 3 | merged | no | main@5e020ff | chore/bootstrap-cfc-bot-manager | MERGED_VERIFIED | Yes (6c21ce1 is ancestor of origin/main 2ad7be0) | Repository bootstrapped (Final-Report.md, docs, placeholder Makefile/Dockerfile) |
| 4 | closed | YES | main@5e020ff | feat/camfrog-manager-product-foundation | CLOSED_SUPERSEDED | Not merged (null) | Draft; superseded by PRs #8 (architecture) and #9 (Go backend) |
| 6 | merged | no | main@5e020ff | fix/security-remove-embedded-secrets-and-windows-paths | MERGED_VERIFIED | Yes (26436c0 is ancestor of origin/main 2ad7be0) | Removed embedded GPG passphrase and Windows paths from two scripts |
| 7 | merged | no | main@26436c0 | fix/redact-exposed-secret-from-public-report | MERGED_VERIFIED | Yes (b2b4c64 is ancestor of origin/main 2ad7be0) | Redacted credential details from public audit report |
| 8 | merged | no | main@26436c0 | feat/cfc-bot-manager-product-foundation-v2 | MERGED_VERIFIED | Yes (8caf9e4 is ancestor of origin/main 2ad7be0) | Product architecture, requirements, ADR 0001, Camfrog contract, threat model, compatibility matrix (documentation only) |
| 9 | merged | no | main@26436c0 | feat/go-backend-foundation | MERGED_VERIFIED | Yes (7c609dd is ancestor of origin/main 2ad7be0) | Go backend + tests + CI (go test -race, vet, build); docs updated |
| 10 | closed | no | main@8caf9e4 | feat/cfc-bot-manager-enterprise-hardening@dc7109d | MERGED | Yes (8c863be is ancestor of origin/main 2ad7be0) | make ci PASS; 25/25 Python tests; Go vet/build/test/race/cover PASS; GitHub CI + CodeQL success |
| 11 | closed | no | main@8c863be | docs/update-post-merge-status@dd6b6b3 | MERGED_VERIFIED | Yes (ea67566 is ancestor of origin/main 2ad7be0) | Audit docs updated to reflect PR #10 merge and process deviation |
| 12 | closed | no | main@ea67566 | docs/final-report@4e9ea6f | MERGED_VERIFIED | Yes (13a5fd3 is ancestor of origin/main 2ad7be0) | Comprehensive final audit and remediation report |
| 13 | closed | no | main@13a5fd3 | feat/cfc-bot-manager-audit-alignment@45ebba3 | CLOSED_SUPERSEDED | Not merged (null) | Created by policedbc; content superseded by PR #14 (same head SHA, cvsz author) |
| 14 | closed | no | main@13a5fd3 | feat/cfc-bot-manager-audit-alignment@45ebba3 | MERGED_VERIFIED | Yes (9411d27 is ancestor of origin/main 2ad7be0) | Controlled cycle #1: cvsz GPG-signed → policedbc APPROVE on exact head → CI/CodeQL/dependency-review SUCCESS → cvsz merge → 9411d27 |
| 15 | closed | no | main@9411d27 | docs/update-post-merge-status@999c35e | MERGED_VERIFIED | Yes (2ad7be0 is origin/main HEAD) | Controlled cycle #2: cvsz GPG-signed → policedbc APPROVE on exact head → CI/CodeQL/dependency-review SUCCESS → cvsz merge → 2ad7be0 |

## Notes on merged PRs

- PR #6 removed embedded secrets/paths but left the broader unsafe automation (auto-fix-loop.sh, deploy-workflow-all-repos.sh, gpg-loopback.sh, the two reviewer-automation workflows) in place. PR #10 completes that retirement.
- PR #3 created Final-Report.md which documented pre-existing working-tree changes (scripts/auto-fix-loop.sh, scripts/deploy-workflow-all-repos.sh, scripts/bootstrap.py, scripts/sync_template.py, Final-Report.md). Those fixes are now part of origin/main.
- PR #9 added the only executable application code (Go backend). It was not reviewed by an independent reviewer (no review submissions recorded). The Go code passes build/vet/test/race/cover locally.
- PRs #1 and #2 were self-approved by cvsz (author == approver), which is not independent review under the repository's independent-review policy.
- PR #10 merged via GitHub merge commit 8c863be before policedbc APPROVED review on final head — process deviation documented in docs/audit/executive-summary.md, docs/audit/issue-5-remediation.md, docs/audit/gpg-review-merge-evidence.md.

## Duplicates

No duplicate PRs identified. PR #4 and PR #8 both concern "product foundation" but #4 was a draft closed without merge; #8 is a distinct, merged documentation/architecture PR. No unique implementation was lost from #4 (it was a draft superseded by #8/#9).

## Counts

- Merged: 12
- Closed without merge: 2 (PR #4 draft; PR #13 superseded)
- Draft: 1 (PR #4)
- Open: 0
