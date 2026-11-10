# CODEX.md — ZEAZ Codex Operating Instructions

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

**CFC Bot Manager** — Camfrog Bot Management Platform

### Stack
- **Go 1.25** backend with PostgreSQL
- **React + TypeScript** (planned) web control panel
- **Tauri 2** (planned) Windows Desktop
- **PostgreSQL** with versioned migrations
- **JWT auth** with Argon2id, MFA (TOTP), CSRF, RBAC

### Architecture
```
backend/
├── cmd/api/           # HTTP server
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

### Static RE Evidence (Integrated)
From `camfrog-autodj/docs/camfrog-re/` — Static RE of `camfrog.exe`:
- **Outgoing messages = plain text only** (no rich text support)
- **Only 3 GDI font APIs** imported (`CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW`)
- **No rich-text libraries** (`mshtml.dll`, `riched20.dll`, `msftedit.dll` not imported)
- **HTML/CSS for incoming only** — embedded web view rendering
- **Camfrog URI scheme**: `camfrog:room:openvideo:` for room links
- **Registry IM settings**: `IMFont`, `IMFontSize`, `FontName`, `Size`

### Current State
- ✅ Go 1.25, config, health endpoints
- ✅ Domain entities (14), repository interfaces (13)
- ✅ PostgreSQL (14 tables, FKs, triggers, migrations)
- ✅ Auth: Argon2id, cookie sessions + CSRF, TOTP MFA, RBAC
- ✅ Camfrog adapter: interface, JSON RPC + multipart transports, capability registry
- ✅ Room message font editor (GUI preview only)
- ✅ Tenant isolation tests (8 negative cases)
- ✅ Migration rollback tests (up/down, savepoint, isolation levels)
- ✅ All CI passing (Go race detector, Python tests, vet, build)

### Key Files
| Area | Files |
|------|-------|
| Auth | `internal/auth/*.go` (password, session, mfa, rbac, middleware) |
| Camfrog | `internal/camfrog/` (adapter, transport/, capability/) |
| Persistence | `internal/persistence/*.go` (13 repos + provider + tx) |
| Migrations | `migrations/000001_initial_schema.up/down.sql` |
| Room Style | `internal/httpapi/room_style_handler.go`, `internal/messagestyle/` |

### Evidence Gates
- Outgoing messages = plain text only (adapter enforces)
- No rich text in adapter (no HTML/CSS)
- Room font editor = GUI preview only
- Plugin capabilities = UNVERIFIED until VERIFIED

### Blocked (Admin Required)
- SEC-001: Rotate GPG signing key
- SEC-002: Set `INDEPENDENT_REVIEWER` variable
- SEC-003: Add `verify-independent-approval` to branch protection