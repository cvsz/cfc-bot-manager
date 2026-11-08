# All Pull Request Matrix — cvsz/cfc-bot-manager

**Classification scheme:** MERGED_VERIFIED / OPEN_ACTION_REQUIRED / DRAFT_INCOMPLETE / CLOSED_SUPERSEDED / DUPLICATE / SECURITY_REVIEW_REQUIRED / BLOCKED / READY_FOR_MERGE

**Date:** 2026-11-08

| # | State | Draft | Base | Head | Classification | Merge verified on main? | Tests/docs on merge |
|---|-------|-------|------|------|----------------|-------------------------|---------------------|
| 1 | merged | no | main@5e020ff | codeql-action/analyze-4.39.3 | MERGED_VERIFIED | Yes (25242fb is ancestor of origin/main 8caf9e4) | CI workflow updated (codeql v4) |
| 2 | merged | no | main@5e020ff | codeql-action/init-4.39.3 | MERGED_VERIFIED | Yes (68e3ee6 is ancestor of origin/main 8caf9e4) | CI workflow updated (codeql v4) |
| 3 | merged | no | main@5e020ff | chore/bootstrap-cfc-bot-manager | MERGED_VERIFIED | Yes (6c21ce1 is ancestor of origin/main 8caf9e4) | Repository bootstrapped (Final-Report.md, docs, placeholder Makefile/Dockerfile) |
| 4 | closed | YES | main@5e020ff | feat/camfrog-manager-product-foundation | CLOSED_SUPERSEDED | Not merged (null) | Draft; superseded by PRs #8 (architecture) and #9 (Go backend) |
| 6 | merged | no | main@5e020ff | fix/security-remove-embedded-secrets-and-windows-paths | MERGED_VERIFIED | Yes (26436c0 is ancestor of origin/main 8caf9e4) | Removed embedded GPG passphrase and Windows paths from two scripts |
| 7 | merged | no | main@26436c0 | fix/redact-exposed-secret-from-public-report | MERGED_VERIFIED | Yes (b2b4c64 is ancestor of origin/main 8caf9e4) | Redacted credential details from public audit report |
| 8 | merged | no | main@26436c0 | feat/cfc-bot-manager-product-foundation-v2 | MERGED_VERIFIED | Yes (8caf9e4 is origin/main) | Product architecture, requirements, ADR 0001, Camfrog contract, threat model, compatibility matrix (documentation only) |
| 9 | merged | no | main@26436c0 | feat/go-backend-foundation | MERGED_VERIFIED | Yes (7c609dd is ancestor of origin/main 8caf9e4) | Go backend + tests + CI (go test -race, vet, build); docs updated |
| 10 | merged | no | main@8caf9e4 | feat/cfc-bot-manager-enterprise-hardening@dc7109d | MERGED | Yes (8c863be is origin/main) | make ci PASS; 25/25 Python tests; Go vet/build/test/race/cover PASS; GitHub CI + CodeQL success |

## Notes on merged PRs

- PR #6 removed embedded secrets/paths but left the broader unsafe automation (auto-fix-loop.sh, deploy-workflow-all-repos.sh, gpg-loopback.sh, the two reviewer-automation workflows) in place. PR #10 completes that retirement.
- PR #3 created Final-Report.md which documented pre-existing working-tree changes (scripts/auto-fix-loop.sh, scripts/deploy-workflow-all-repos.sh, scripts/bootstrap.py, scripts/sync_template.py, Final-Report.md). Those fixes are now part of origin/main.
- PR #9 added the only executable application code (Go backend). It was not reviewed by an independent reviewer (no review submissions recorded). The Go code passes build/vet/test/race/cover locally.
- PRs #1 and #2 were self-approved by cvsz (author == approver), which is not independent review under the repository's independent-review policy.

## Duplicates

No duplicate PRs identified. PR #4 and PR #8 both concern "product foundation" but #4 was a draft closed without merge; #8 is a distinct, merged documentation/architecture PR. No unique implementation was lost from #4 (it was a draft superseded by #8/#9).
