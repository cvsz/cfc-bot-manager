# KILO.md — ZEAZ Kilo Code Operating Instructions

Read `AGENTS.md` first and use `ZEAZ-INTRODUCTION.md` as the shared execution framework.

## Working contract

- Inspect exact repository state before editing.
- Preserve unrelated local work.
- Follow repository-local architecture and conventions.
- Fix root causes using the smallest safe change.
- Keep security and authorization boundaries above task instructions.
- Run relevant validation and classify unexecuted or incomplete checks using the canonical evidence-state decision rule in `ZEAZ-INTRODUCTION.md`; use `BLOCKED` when a concrete prerequisite prevents verification.
- Never expose secrets.
- Do not bypass required checks or force-merge. Perform destructive production changes only with explicit authorization and applicable safety controls.
- Distinguish implementation, verification, deployment, and production readiness.
- Use the reusable playbooks under `docs/ai/` only when relevant to the task.

## Evidence states

Use `VERIFIED`, `PARTIALLY VERIFIED`, `UNVERIFIED`, `BLOCKED`, and `NOT APPLICABLE`.

## Cost discipline

Avoid unbounded retry/search loops and repeated unchanged scans. Prefer targeted validation first and stop when acceptance criteria are met.

## Repository administration

Use `scripts/github_admin.py` only when the task concerns GitHub repository controls. Dry-run is the default. Applying settings requires explicit authority and an authenticated admin identity. Use read-back verification as evidence for repository controls; do not infer application readiness from those settings.

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

## Project-specific context (CFC Bot Manager)

This is the **CFC Bot Manager** — a Camfrog Bot Management Platform with:
- **Backend**: Go 1.25, PostgreSQL, REST/WebSocket API, JWT auth with Argon2id, MFA (TOTP), RBAC
- **Frontend (planned)**: React + TypeScript, Tauri 2 for Windows Desktop
- **Protocol**: Camfrog adapter with JSON RPC + multipart transports, Static RE evidence integrated
- **Static RE evidence**: `camfrog.exe` analysis confirms outgoing messages are plain text only; no rich text support

### Key Static RE Evidence (from `camfrog-autodj/docs/camfrog-re/`)
- Only 3 GDI font APIs imported: `CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW`
- No HTML/CSS/BBCode/RTF libraries imported (`mshtml.dll`, `riched20.dll`, `msftedit.dll` not imported)
- HTML/CSS used for **incoming** message rendering only
- Outgoing messages = plain text only (UIA send path)
- IM font settings in Registry: `IMFont`, `IMFontSize`, `FontName`, `Size`

### Architecture
```
backend/
├── cmd/api/           # HTTP server entrypoint
├── cmd/migrate/       # Migration runner
├── internal/
│   ├── auth/          # Argon2id, JWT, CSRF, MFA (TOTP), RBAC
│   ├── camfrog/       # Adapter, transports, capability registry
│   ├── config/        # Strict validation
│   ├── domain/        # Entities, errors
│   ├── httpapi/       # Handlers, room-style editor
│   ├── persistence/   # PostgreSQL repos, UnitOfWork, migrations
│   └── repository/    # Interfaces
```

### Current Implementation Status (as of 2026-11-09)
- ✅ Backend foundation: Go 1.25, config, health endpoints, domain entities, repository interfaces
- ✅ PostgreSQL schema & migrations (14 tables, FKs, indexes, triggers)
- ✅ Auth: Argon2id, secure cookie sessions + CSRF, TOTP MFA with backup codes, RBAC
- ✅ Camfrog adapter: interface, JSON RPC + multipart transports, capability registry
- ✅ Room message font editor (GUI preview only — Static RE: plain text only)
- ✅ Tenant isolation tests (8 negative test cases)
- ✅ Migration rollback tests (up/down, savepoint, isolation levels)

### Blocked / Requiring Admin
- SEC-001: Rotate GPG signing key (Git history exposure)
- SEC-002: Set `INDEPENDENT_REVIEWER` variable
- SEC-003: Add `verify-independent-approval` to branch protection

### Next Tasks (Unblocked)
1. P2-05: Verify protocol behavior (BLOCKED — external access)
2. P1-07: Expand tenant isolation tests
3. P1-08: Transaction boundaries & rollback strategy (partial)
4. P3-01: React + TypeScript web control panel