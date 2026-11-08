# TODO / FIXME / HACK / XXX / Stub / Placeholder Inventory — cfc-bot-manager

**Date:** 2026-11-08
**Method:** `grep -rn -iE 'TODO|FIXME|HACK|XXX|NotImplemented|stub|placeholder|unreachable|dead code|unused|deprecated|invalid'` across code, config, docs (excluding `.git`, `__pycache__`, `.ruff_cache`, `Final-Report.md`)
**Classification:** Intentional / Implemented / Incomplete / Defective / Deprecated / Unsafe / Blocked (per Phase 13)

## Code-level TODOs/stubs

| File | Line | Text | Classification | Notes |
|------|------|------|----------------|-------|
| scripts/audit-repo.sh | 28 | `pass() { echo "PASS: $*" ... }` | Implemented | Helper function named `pass`; not a TODO |
| scripts/audit-repo.sh | 58,75,97,99,111,123 | `pass "..."` calls | Implemented | Audit script pass markers |
| scripts/rollout-independent-review.sh | 89 | --body text mentions INDEPENDENT_REVIEWER config | Implemented | Instructional message |

## Documentation TODOs (intentional, project not yet defined)

| File | Line | Text | Classification | Notes |
|------|------|------|----------------|-------|
| ABOUT.md | 8 | Technical ownership: TODO | Intentional | Maintainers not assigned |
| ABOUT.md | 12 | Problem and target users: TODO | Intentional | Product scope undefined |
| ABOUT.md | 13 | Scope and non-goals: TODO | Intentional | |
| ABOUT.md | 14 | Success metrics: TODO | Intentional | |
| ABOUT.md | 15 | Data classification and compliance: TODO | Intentional | |
| ABOUT.md | 16 | Support and service expectations: TODO | Intentional | |
| templates/project-about.md | 8-16 | Same TODOs (template) | Intentional | Template placeholders |

## Placeholders (template artifacts)

| File | Line | Text | Classification | Notes |
|------|------|------|----------------|-------|
| docs/startup.md | 64 | "replace the placeholder Dockerfile and Makefile application commands" | Intentional | Template guidance |
| docs/development.md | 8 | "Replace placeholder Makefile targets with real project commands" | Intentional | |
| docs/profiles.md | 12 | "replace placeholder Dockerfile/Makefile application targets" | Intentional | |
| AGENTS.md | 31 | "Replace Makefile and Dockerfile placeholders..." | Intentional | |
| Makefile | all | Placeholder targets (validate-template, test, lint, build, security, ci) | Implemented (for template) | Python/template-only; no app commands |
| Dockerfile | all | `FROM python:3.12-alpine` placeholder baseline | Intentional | Not for production delivery yet |
| .github/ISSUE_TEMPLATE/bug_report.yml | 13,19 | `placeholder: v1.2.3 or commit SHA`, `placeholder: OS, runtime...` | Intentional | Issue template placeholder examples |

## Protocol "placeholder" observations (legitimate, not code stubs)

| File | Line | Text | Classification | Notes |
|------|------|------|----------------|-------|
| docs/protocols/camfrog-observed-contract.md | 75 | Greeter: "%1% placeholders observed but semantics not fully verified" | Observed evidence | Capture artifact; not code |
| docs/protocols/camfrog-observed-contract.md | 80 | Roulette: "placeholder/game semantics unknown" | Observed evidence | ID 9 excluded |
| docs/protocols/camfrog-observed-contract.md | 82 | Who Banned: "placeholder and precedence semantics unknown" | Observed evidence | |
| docs/protocols/camfrog-observed-contract.md | 83 | Talkback: "placeholders; full grammar unknown" | Observed evidence | |

## Unsafe / defective (already remediated by PR #10)

| File | Line | Text | Classification | Disposition |
|------|------|------|----------------|-------------|
| scripts/auto-fix-loop.sh | 168 | `git -c gpg.program=...gpg-loopback.sh commit ...; git push origin HEAD` | Unsafe (direct-push + GPG) | DELETED |
| scripts/deploy-workflow-all-repos.sh | 23-25,74-75,88-94 | POLIEDBC_PAT from credential store; `gh secret set`; `git push origin HEAD` | Unsafe (cross-repo mutation + reviewer cred) | DELETED |
| .github/scripts/gpg-loopback.sh | 5-6 | Reads GPG_PASSPHRASE; passes to gpg | Unsafe (tracked signing wrapper) | DELETED |
| .github/workflows/automated-audit-review-approve.yml | 34,93,103 | `secrets.POLIEDBC_PAT`; `gh pr review --request-changes` | Unsafe (reviewer credential distribution) | DELETED |
| .github/workflows/auto-fix-open-prs.yml | 33-35 | `secrets.GPG_PASSPHRASE`, `GPG_SIGNING_KEY`, `POLIEDBC_PAT` | Unsafe (credential-requiring workflow) | DELETED |
| docs/ai/playbooks/auto-audit-review-approve-policedbc.md | all | Documents automated policedbc approval + PAT distribution | Obsolete automated-approval documentation | DELETED |

## Ruff findings (pre-existing, not introduced by this audit)

| File | Line | Rule | Classification | Notes |
|------|------|------|----------------|-------|
| scripts/github_admin.py | 31 | UP022 (capture_output) | Unsafe fix — NOT applied | ruff explicitly marks unsafe |
| tests/test_bootstrap.py | 3 | I001 import sorting | Code quality | Pre-existing |
| tests/test_independent_review.py | 3 | I001 import sorting | Code quality | Pre-existing |

## ShellCheck findings (pre-existing)

| File | Rule | Classification | Notes |
|------|------|--------------|-------|
| gh-switch.sh | SC2034 (unused vars) | Warning, not defect | Pre-existing |
| post-review-comment.sh | SC2034 | Warning | DELETED with the file |
| deploy-workflow-all-repos.sh | SC2034 | Warning | DELETED with the file |

## Do-not-remove items

Several documentation "TODO" and "placeholder" entries are intentional (product not yet defined; template guidance). They are NOT removed per Phase 13 ("Do not mechanically remove TODOs").

## Verdict

No code-level TODO/FIXME/HACK/XXX stubs remain in the implemented Go/Python code. All unsafe automation identified as defective has been deleted. Remaining items are intentional documentation placeholders or pre-existing minor lint warnings.
