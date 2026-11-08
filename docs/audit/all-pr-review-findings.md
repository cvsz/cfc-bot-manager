# All Pull Request Review Findings — cvsz/cfc-bot-manager

**Summary of review findings across all 8 PRs.**
**Date:** 2026-11-08

## Findings by PR

### PR #1 — Dependabot codeql-action/analyze 4.39.0 → 4.39.3
- Finding: Self-approved by author (cvsz). Not independent review.
- Severity: Informational (governance).
- Evidence: reviews show user=cvsz, state=APPROVED, author=cvsz.
- Action: None (already merged). Future Dependabot PRs should seek independent approval where policy requires.

### PR #2 — Dependabot codeql-action/init 4.39.0 → 4.39.3
- Finding: Self-approved by author (cvsz). Not independent review.
- Severity: Informational (governance).
- Evidence: reviews show user=cvsz, state=APPROVED, author=cvsz.
- Action: None (already merged).

### PR #3 — chore: initialize cfc bot manager repository
- Finding: No review submissions recorded.
- Severity: Informational (governance).
- Evidence: reviews array empty.
- Action: None (already merged; initialization PR).

### PR #4 — feat: establish CFC Bot Manager product foundation (draft, closed)
- Finding: Closed as draft without merge; superseded by PRs #8 and #9.
- Severity: Informational.
- Evidence: state=closed, draft=true, merge_commit_sha=null.
- Action: None; keep closed.

### PR #6 — fix: remove embedded GPG passphrase and Windows paths
- Finding: Approved independently by policedbc. Good.
- Finding: Partial remediation only — left auto-fix-loop.sh, deploy-workflow-all-repos.sh, gpg-loopback.sh, and the two reviewer-automation workflows in place.
- Severity: P0 (incomplete remediation led to Issue #5).
- Evidence: unsafe files present after PR #6 merge; retired by PR #10.
- Action: Completed by PR #10 (all unsafe automation removed).

### PR #7 — fix: redact credential details from public audit report
- Finding: No review submissions recorded.
- Severity: Informational (governance).
- Evidence: reviews array empty.
- Action: None (already merged).

### PR #8 — feat: add CFC Bot Manager product architecture and requirements
- Finding: No review submissions recorded. Documentation/architecture only, no application code.
- Severity: Informational (governance).
- Evidence: reviews array empty.
- Action: None (already merged).

### PR #9 — feat: add Go backend foundation and CI validation
- Finding: No review submissions recorded. Adds real executable Go code.
- Severity: Major (governance) — application code merged without recorded independent review.
- Evidence: reviews array empty; backend/ directory present on origin/main.
- Action: Note for future; the Go code passes local build/vet/test/race/cover.

### PR #10 — security: retire legacy privileged automation (open)
- Finding: GPG signing BLOCKED (configured key requires passphrase unavailable without reading `.git-credentials`).
- Finding: policedbc independent review BLOCKED (gh auth switch requires human interaction).
- Severity: P0 (blocker to merge).
- Action: Operator must unlock GPG key and perform `gh auth switch -u policedbc` review before merge.

## Unresolved review threads

None found via the API (no review comments returned for any PR; PRs #3, #4, #7, #8, #9 had no review submissions).

## Cross-cutting findings

1. **Independent review not consistently applied:** PRs #1, #2 self-approved; PRs #3, #7, #8, #9 unreviewed. Only PR #6 had an independent (policedbc) approval.
2. **P0 remediation was partial until PR #10:** PR #6 removed embedded secrets but left the broader privileged automation intact.
3. **No application-code review gate evidence:** PR #9 (Go backend) merged without recorded review.
