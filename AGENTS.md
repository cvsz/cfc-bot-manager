# AGENTS.md — CFC Bot Manager Agent Contract

## Purpose
This repository implements **CFC Bot Manager** — a Camfrog Bot Management Platform with Web Control Panel, Windows Desktop Application (Tauri 2), shared Go backend, plugin architecture, and production deployment.

## Operating Rules
- Read `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`, and the closest `AGENTS.md` before editing.
- Keep code, configuration, filenames, commit messages, and technical documentation in English.
- Prefer the smallest reviewable change that satisfies the requested scope.
- Never weaken CI, security scanning, dependency review, branch protections, or release controls merely to make a check pass.
- Never commit credentials, tokens, private keys, production endpoints, personal data, or realistic secrets. Use documented placeholders.
- Do not invent project-specific owners, domains, deployment providers, package registries, cloud accounts, or credentials.
- Preserve template portability across languages and frameworks unless a file explicitly declares a narrower scope.
- Reuse existing workflows and documents instead of creating overlapping alternatives.
- Pin permissions for GitHub Actions to least privilege and prefer maintained first-party/verified actions.
- Treat external input, generated artifacts, pull requests from forks, and dependency metadata as untrusted.

## Template Placeholders
Use obvious placeholders such as `PROJECT_NAME`, `OWNER`, `example.com`, and `REPLACE_ME`. Any generated repository must be able to find and replace placeholders without exposing secrets.

## Review Skill
- Use [Scrutinize](.agents/skills/scrutinize/SKILL.md) when asked to review, audit, sanity-check, or give a second opinion on a plan, PR, diff, design, or code change, or when invoked with `/scrutinize` in a compatible agent.
- Question whether the change is necessary or can be smaller before tracing real code paths and verifying behavioral claims. Cite concrete file/line evidence and distinguish unverified claims from confirmed behavior.
- The skill guides agent behavior; it does not itself install a slash command or replace required tests, CI, or human review.

## Project Initialization
- For a repository created with **Use this template**, follow [docs/startup.md](docs/startup.md) and preview `scripts/bootstrap.py` before `--apply`; pass an actual GitHub user or org/team via `--codeowner` and verify write access.
- Do not present an initialized scaffold as a running or production-ready system; bootstrap changes project identity and ownership routing only.
- Replace Makefile and Dockerfile placeholders with actual project-specific commands and images before enabling an application delivery pipeline.
- Never overwrite existing application code, production secrets, DNS ownership or original license attribution during initialization.

## Change Workflow
1. Inspect the current exact branch/head and existing files.
2. Identify the smallest missing or inconsistent capability.
3. Add tests or validation first when practical.
4. Implement without widening scope.
5. Run the relevant validation and security checks.
6. Update documentation when behavior, setup, governance, or release procedures change.
7. Open a pull request; do not claim merge/release readiness without exact-head evidence.

## Verification
At minimum, verify Markdown/YAML syntax for touched files, workflow permissions/triggers, links and placeholders, absence of committed secrets, and consistency between README, templates, governance, security, and release documentation.

## Pull Requests and Releases
PRs must state scope, tests, security impact, compatibility/migration impact, documentation impact, deployment impact, and rollback. Releases require green required checks and explicit evidence; never infer production readiness from documentation alone.

## Security
Report vulnerabilities through SECURITY.md, not public issues. Security-related templates must redirect sensitive reports accordingly. Fail closed when a security-sensitive configuration is incomplete.

## Documentation Ownership
- `.github/`: GitHub automation, community health, ownership, issue/PR templates.
- `docs/`: versioned engineering, operations, release, repository-administration, and rollout guidance.
- `docs/adr/`: architecture decision records.
- `docs/ai/`: reusable prompts and playbooks.
- `docs/protocols/`: Camfrog protocol evidence (including Static RE).
- Root Markdown files: repository-wide policy and project lifecycle guidance.

## Nested AGENTS.md
Add a child AGENTS.md only when a subtree has durable rules that differ from this contract. The nearest AGENTS.md may add stricter local requirements but must not weaken repository-wide security rules.

## ZEAZ Reusable Execution Layer
- Use [ZEAZ-INTRODUCTION.md](ZEAZ-INTRODUCTION.md) as the cross-agent execution framework.
- Use [docs/ai/README.md](docs/ai/README.md) to select task-specific reusable prompts and playbooks.
- Task playbooks are additive and never weaken this repository contract or narrower subtree rules.
- Use the canonical evidence-state definitions and decision rule in [ZEAZ-INTRODUCTION.md](ZEAZ-INTRODUCTION.md); do not redefine them per harness or playbook.
- Keep implementation, verification, deployment, and production readiness as distinct states.

## Project-Specific Rules (CFC Bot Manager)

### Static RE Evidence Gates (Mandatory)
All Camfrog adapter work must respect these findings from static RE of `camfrog.exe`:
- **Outgoing messages = plain text only** — adapter must enforce; no HTML/CSS/rich text formatting
- **No rich text support** — no HTML/CSS formatting in adapter; `mshtml.dll`, `riched20.dll`, `msftedit.dll` not imported
- **Room message font editor = GUI preview only** — clearly labeled; adapter sends plain text
- **Plugin capabilities default to UNVERIFIED/DISABLED** until VERIFIED with authorized test evidence

### Architecture Boundaries
- **Backend**: Go 1.25, PostgreSQL, REST/WebSocket, JWT (Argon2id), MFA (TOTP), CSRF, RBAC
- **Frontend (planned)**: React + TypeScript, Tauri 2 for Windows Desktop
- **Camfrog Adapter**: JSON RPC + multipart transports, capability registry with Static RE evidence
- **Persistence**: PostgreSQL with UnitOfWork, 14 tables, migrations, triggers

### Evidence States
Use exactly: `VERIFIED`, `PARTIALLY VERIFIED`, `UNVERIFIED`, `BLOCKED`, `NOT APPLICABLE`
- Green CI alone is insufficient for `VERIFIED`
- Requires exact-head CI + independent review + deployment validation

### Static RE Evidence (Integrated)
From `camfrog-autodj/docs/camfrog-re/` — Static RE of `camfrog.exe`:
- **Outgoing messages = plain text only** (UIA send path)
- **Only 3 GDI font APIs** imported (`CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW`)
- **No rich-text libraries** (`mshtml.dll`, `riched20.dll`, `msftedit.dll` not imported)
- **HTML/CSS for incoming rendering only** (`font-family:`, `font-size:`, `color: #`, `font-weight: bold;`, `font-style: italic;`, `camfrog:room:openvideo:` URI)
- **IM font settings in Registry**: `IMFont`, `IMFontSize`, `FontName`, `Size`

### Adapter Interface (`internal/camfrog/adapter.go`)
- `CamfrogAdapter` with `Status()`, `Language()`, `SwitchPlugin()`, `GetPluginSettings()`, `SavePluginSettings()`, `RoomPause()`, `RoomSetLanguage()`
- `Transport` interface for JSON RPC + multipart
- `CapabilityRegistry` with `VERIFIED`/`PARTIALLY_VERIFIED`/`UNVERIFIED`/`UNSUPPORTED`/`DISABLED` states

### Testing Standards
- Unit tests with race detector: `go test -race ./...`
- Negative tests for tenant isolation (cross-org denial)
- Migration tests: up/down, savepoint, isolation levels
- Contract tests with sanitized fixtures (no cookies, sessions, room keys)

### Blocked Items (Admin Required)
- SEC-001: Rotate GPG signing key (Git history exposure)
- SEC-002: Set `INDEPENDENT_REVIEWER` variable
- SEC-003: Add `verify-independent-approval` to branch protection

## Documentation Ownership
- `.github/`: GitHub automation, community health, ownership, issue/PR templates.
- `docs/`: versioned engineering, operations, release, repository-administration, rollout guidance.
- `docs/adr/`: architecture decision records.
- `docs/ai/`: reusable prompts and playbooks.
- `docs/protocols/`: Camfrog protocol evidence (including Static RE).
- `docs/protocols/camfrog-re/`: Static RE evidence from Craft Anything decompiler.
- Root Markdown files: repository-wide policy and project lifecycle guidance.

## Nested AGENTS.md
Add a child AGENTS.md only when a subtree has durable rules that differ from this contract. The nearest AGENTS.md may add stricter local requirements but must not weaken repository-wide security rules.

## ZEAZ Reusable Execution Layer
- Use [ZEAZ-INTRODUCTION.md](ZEAZ-INTRODUCTION.md) as the cross-agent execution framework.
- Use [docs/ai/README.md](docs/ai/README.md) to select task-specific reusable prompts and playbooks.
- Task playbooks are additive and never weaken this repository contract or narrower subtree rules.
- Use the canonical evidence-state definitions and decision rule in [ZEAZ-INTRODUCTION.md](ZEAZ-INTRODUCTION.md); do not redefine them per harness or playbook.
- Keep implementation, verification, deployment, and production readiness as distinct states.

## Existing-Repository Rollout
When applying this template to an established repository, follow [docs/repository-rollout.md](docs/repository-rollout.md). Audit the target first and port only missing compatible controls. Do not overwrite repository-specific AGENTS rules, CI matrices, release workflows, infrastructure ownership, or operational evidence merely to match this template.

For repository administration, `scripts/github_admin.py` requires an authenticated GitHub identity with Administration permission. Treat successful read-back verification—not script presence—as evidence that the controls are effective.

## Git/GH Workflow for Commits, PRs, Reviews, and Merges

### Git Credentials
Use `/home/cvsz/.git-credentials` for git authentication:
```bash
git config --global credential.helper 'store --file /home/cvsz/.git-credentials'
```

### GPG-Signed Commits & PR Creation (as `cvsz`)
```bash
gh auth switch -u cvsz
# Configure GPG signing
git config --global commit.gpgsign true
git config --global user.signingkey <KEY_ID>

# Create branch, commit with GPG signature, push
git checkout -b <branch-name>
git commit -S -m "type: description"
git push origin <branch-name>

# Open PR
gh pr create --title "type: description" --body "..."
```

### Audit, Review & Approve (as `policedbc`)
```bash
gh auth switch -u policedbc
# Review PR
gh pr review <PR_NUMBER> --approve --body "Review comments..."

# Or request changes
gh pr review <PR_NUMBER> --request-changes --body "Changes needed..."
```

### Merge (as `cvsz`)
```bash
gh auth switch -u cvsz
# Merge with required checks passing
gh pr merge <PR_NUMBER> --merge --delete-branch

# Or squash merge
gh pr merge <PR_NUMBER> --squash --delete-branch
```

### GPG Key Setup (if not configured)
```bash
# Generate key if needed
gpg --full-generate-key
# Export public key for GitHub
gpg --armor --export <KEY_ID> | gh gpg-key add -
```

### Credential Security
- Never hardcode tokens in scripts or commit them
- Use `gh auth login` / `gh auth switch` for identity management
- Store GPG passphrase in OS keyring or CI secret store, never in shell history
- Use `gh auth status` to verify current identity before operations