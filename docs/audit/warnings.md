# Warnings — cfc-bot-manager

**Date:** 2026-11-08
**Scope:** Non-fatal warnings from lint/static analysis that are not defects.

## ShellCheck SC2034 (unused variables)

These are pre-existing warnings, not security defects. Some were deleted with the P0-remediated files.

| File | Line | Variable | State |
|------|------|----------|-------|
| scripts/gh-switch.sh | 49 | ORIG_URL | Warning — DELETED? No, gh-switch.sh remains |
| scripts/gh-switch.sh | 60 | ORIG_HOST | Warning — gh-switch.sh remains |
| scripts/gh-switch.sh | 95 | owner, repo | Warning — gh-switch.sh remains |
| scripts/gh-switch.sh | 112 | cur_mode | Warning — gh-switch.sh remains |
| scripts/deploy-workflow-all-repos.sh | 20 | ACCOUNTS_FILE | Warning — DELETED (P0) |
| .github/scripts/post-review-comment.sh | 15 | AUTO_APPROVE_SAFE | Warning — DELETED (P0) |

`gh-switch.sh` remains in the tree (it is a local-only config switcher, not automation). Its SC2034 warnings are cosmetic.

## Ruff findings

| File | Rule | Classification | Action |
|------|------|--------------|--------|
| scripts/github_admin.py | UP022 | Unsafe fix — NOT applied | Left as-is |
| tests/test_bootstrap.py | I001 | Import sorting | Pre-existing |
| tests/test_independent_review.py | I001 | Import sorting | Pre-existing |
| tests/test_security_automation_retired.py | — | — | Clean (ruff --fix applied) |

## Go

No warnings. `gofmt -l` clean, `go vet` clean.

## Summary

All warnings are cosmetic (unused variables, import ordering) or explicitly unsafe (ruff UP022, not applied). No warning indicates a security defect or incorrect behavior.
