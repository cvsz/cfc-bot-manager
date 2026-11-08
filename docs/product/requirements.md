# Product Requirements

## Goal

Provide one consistent management experience for Camfrog bot/room administration through a Windows desktop client and a web control panel, backed by one authenticated service and protocol adapter.

## Personas and access

The initial product must support distinct roles without trusting client-side role checks:
- Owner/administrator: manage account-level configuration, users, and bot/room assignments.
- Operator/moderator: operate only explicitly assigned bots/rooms and permitted features.
- Viewer: read permitted status and activity.
- Auditor: read audit/activity records without mutation.

Exact permission matrices and multi-tenant onboarding flows must be finalized in implementation ADRs. Default deny applies.

## Functional areas

- Authentication, session management, profile, optional MFA.
- Dashboard with bot/room health and recent activity.
- Bot and room inventory; status and lifecycle controls where supported upstream.
- Plugin inventory, capability discovery, settings forms, enable/disable, and import/export only where verified.
- Real-time status/event updates.
- User, role, and resource assignment management.
- Audit trail for privileged changes.
- System configuration, health, and operational settings.

## Client requirements

Web and Windows desktop clients use the same API contracts, authorization semantics, and core workflows. Desktop is packaged with Tauri 2. Tauri capabilities must be least-privilege and explicitly scoped. No backend secrets or Camfrog credentials are embedded in either client.

## Non-functional requirements

- Secure by default, least privilege, server-side authorization, tenant isolation.
- Explicit API and event schemas with compatibility/version strategy.
- Bounded request sizes, timeouts, cancellation, rate limiting, and safe failure semantics.
- Structured logs and correlation IDs with secret redaction.
- Observable background work and health/readiness signals.
- Deterministic builds, dependency pinning, CI validation, documented recovery/rollback.
- Accessibility and responsive layouts for web; keyboard support for desktop.

## Out of scope until evidence exists

- Claiming legacy standalone plugin compatibility with current Cloud Bot.
- Treating undocumented/private Camfrog pages as a supported public API.
- Guessing plugin ID 10's identity or behavior.
- Automatic moderation/ban actions without verified semantics, explicit authorization, and safe operational controls.
- Cloud deployment or public DNS changes as part of application implementation.

## Acceptance criteria

- Both clients authenticate against one backend and receive the same server-enforced permissions.
- Cross-tenant and cross-resource authorization tests prove isolation.
- All mutating API paths are covered by authorization, CSRF/origin checks where applicable, validation, audit, and failure-path tests.
- Camfrog adapter contract tests prove the behavior for each explicitly supported product/version.
- Plugin controls are enabled only when capability evidence is verified; unknown capability states are visibly disabled.
- No user-provided strings are interpreted as HTML.
- A fresh clone can install, test, build, and package using documented commands.
- Deployment, backup/restore, rollback, monitoring, and incident runbooks are verified before production claims.
