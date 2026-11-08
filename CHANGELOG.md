# Changelog

All notable changes to this template are documented here.

The format follows Keep a Changelog conventions; generated projects should adopt an explicit versioning policy appropriate to their product.

## [Unreleased]

### Added

- Safe project identity bootstrap with explicit dry-run/apply and idempotence tests.
- Generated README/ABOUT templates and startup/profile documentation.
- CI bootstrap test coverage and fail-closed Makefile placeholders.
- ZEAZ cross-agent engineering execution framework.
- Claude Code and OpenCode adapters.
- Reusable AI guides, playbooks, prompts, skills, component manifests, and plugin-source manifests.
- ECC OSS CLI install configuration with explicit separation from hosted ECC App evidence.
- Repository structure and local Markdown-link validator.
- GitHub administration automation with dry-run, explicit apply, and read-back verification.
- Repository rollout guidance for applying the baseline safely to existing repositories.
- `docs/audit/` audit evidence bundle (executive summary, P0 Issue #5 remediation, security assessment, PR inventory/matrix/disposition/findings/dependency map, Go backend review, Camfrog compatibility, plugin safety, frontend/desktop/build-lint-test/integration/penetration/dependency-upgrade/GPG-merge-evidence/production-readiness/residual-risks, architecture review, bugs and findings, TODO inventory, invalid configurations, warnings, missing capabilities, compatibility matrix).

### Changed

- Reorganized reusable AI documentation into `guides/`, `playbooks/`, and `prompts/`.
- Pinned baseline first-party GitHub Actions to immutable commit SHAs.
- Expanded CODEOWNERS coverage for repository policy, AI, ECC, skills, components, and plugin manifests.
- Clarified that release-note configuration is not an artifact-publishing workflow.
- Documented branch/security administration as an explicit evidence gate rather than a documentation-only checklist.

### Fixed

- Corrected generated README link validation so template links are resolved from their generated root location.
- Removed stale wording that could imply green CI or configuration files alone establish production readiness.
- `gofmt` applied to backend Go files (`cmd/api/main.go`, `internal/camfrog/rpc.go`, `internal/config/config.go`, `internal/httpapi/server.go`, `internal/httpapi/server_test.go`).
- Added Go test coverage for duplicate request IDs, concurrent RPC, missing/unexpected parameters, invalid environment/host/port, unknown HTTP paths, and security headers.
- Added Python regression tests (`tests/test_security_automation_retired.py`) preventing reintroduction of retired privileged automation.

### Security

- Added fail-closed repository-administration verification.
- Added protected-branch controls for required reviews/checks, conversation resolution, force-push prevention, and deletion prevention.
- Added documented verification for Dependabot, private vulnerability reporting, secret scanning/push protection, and least-privilege Actions permissions where supported.
- **P0 (Issue #5):** Retired legacy privileged automation — deleted `.github/workflows/auto-fix-open-prs.yml`, `.github/workflows/automated-audit-review-approve.yml`, `.github/scripts/gpg-loopback.sh`, `.github/scripts/post-review-comment.sh`, `scripts/auto-fix-loop.sh`, `scripts/deploy-workflow-all-repos.sh`, and `docs/ai/playbooks/auto-audit-review-approve-policedbc.md`. Preserved `request-independent-review.yml` as the deliberate independent human-review gate. Added regression tests. GPG signing of the remediation commit is BLOCKED (configured key requires a passphrase unavailable without reading `.git-credentials`); policedbc independent review and Main protection verification are BLOCKED (require operator action).
