# Plugin Safety — cfc-bot-manager

**Source audited:**
- `backend/internal/camfrog/rpc.go` (plugin ID inventory)
- `docs/product/plugin-compatibility-matrix.md`
- `docs/protocols/camfrog-observed-contract.md` (plugin evidence inventory + safety requirements)
- `docs/product/requirements.md` (out of scope)
**Date:** 2026-11-08

## Plugin inventory (observed, not verified)

| ID | Name | Evidence | Runtime behavior | Cloud Bot | Go allowlist |
|---:|------|----------|------------------|-----------|--------------|
| 1 | 8 Ball | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 2 | IM Control | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 3 | Messenger | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 4 | Greeter | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 5 | Math Trivia | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 6 | Bad Words | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 7 | Trivia | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 8 | Word Trivia | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 9 | Roulette | OBSERVED | UNVERIFIED | UNVERIFIED | EXCLUDED |
| 10 | Unknown | UNVERIFIED | UNVERIFIED | UNVERIFIED | EXCLUDED |
| 11 | Ghost | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 12 | Who Banned | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 13 | Talkback | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 14 | Dice | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 15 | DJ Manager | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |
| 16 | Auto Moderator | OBSERVED | UNVERIFIED | UNVERIFIED | allowed |

## Safety exclusions (enforced in code)

- **Plugin ID 9 (Roulette):** EXCLUDED from `observedPluginIDs` in `backend/internal/camfrog/rpc.go:25-28`. `rpc_test.go:37-39` asserts ID 9 is rejected with the comment "roulette plugin must remain disabled until non-wagering behavior is verified." The project does not implement wagering, betting, stakes, payouts, or gambling workflows. ID 9 is never enabled through automated changes.
- **Plugin ID 10 (Unknown):** EXCLUDED from `observedPluginIDs`. `rpc_test.go:40-42` asserts ID 10 is rejected ("unidentified plugin 10 must be rejected"). ID 10's identity is never inferred from the partial heap-related capture and never displayed to users.

## Plugin safety requirements (contract doc, docs/protocols/camfrog-observed-contract.md:88-97)

| Requirement | Status |
|-------------|--------|
| Plugin settings as untrusted input | Documented; not implemented (no settings transport) |
| Validate numeric/select against captured options | Documented; not implemented |
| Never execute upstream regexes locally without bounds | Documented; not implemented |
| Never render messages/nicknames/reasons/greetings as HTML | Documented; not implemented |
| No optimistic UI for RPC mutations | Documented; not implemented (no UI) |
| No retry of non-idempotent submissions without idempotency contract | Documented; not implemented |
| Per-session serialization until concurrency verified | Documented; not implemented |
| Raw transport logging disabled or redacted | Documented; not implemented |

## Unsupported plugins cannot be activated accidentally

The Go `NewRPCRequest` rejects any plugin ID not in `observedPluginIDs`, and IDs 9 and 10 are absent. There is no plugin activation UI or backend endpoint. Unknown/unverified capabilities are visibly disabled by absence from the allowlist.

## Legacy Bot / Cloud Bot

Separate compatibility assessments are required per the mission. No evidence supports treating captured legacy settings as Cloud Bot. Cloud Bot and legacy standalone bot must be represented as separate product targets (ADR 0001).

## Verdict

Plugin safety controls are correctly implemented at the RPC validation boundary (IDs 9 and 10 excluded; allowlist enforced). No plugin can be activated through the current backend. Runtime plugin behavior is UNVERIFIED for all IDs and must not be claimed.
