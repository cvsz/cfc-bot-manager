# Camfrog Protocol Compatibility — cfc-bot-manager

**Scope:** Observed Camfrog JSON-RPC and multipart form contract (from user-supplied HTML captures)
**Evidence boundary:** Undocumented/private interface; NOT an authenticated Camfrog session; NOT a public or stable upstream API
**Source audited:**
- `backend/internal/camfrog/rpc.go` (+ rpc_test.go)
- `docs/protocols/camfrog-observed-contract.md`
- `docs/product/plugin-compatibility-matrix.md`
- `docs/adr/0001-product-stack-and-adapter-boundary.md`
- `docs/security/threat-model.md`
**Date:** 2026-11-08

## Contract status: UNDOCUMENTED until verified

The observed contract is treated as undocumented until verified against authorized, supported installations. No runtime Camfrog integration exists. The Go `camfrog` package is a JSON-RPC serialization/validation boundary only.

## Observed RPC envelope (contract doc)

```typescript
interface CamfrogRpcRequest {
  id: number;
  method: string;
  params: unknown[];
}
```
- Endpoint: `POST /bot/ajax/`, content type `application/json`
- Observed methods: `status` ([]), `lang` ([language]), `switch_plugin` ([pluginId, enabled])

## Go implementation (rpc.go)

| Aspect | Implementation | Evidence state |
|--------|---------------|----------------|
| Envelope `{id, method, params}` | `RPCRequest` struct with json tags | VERIFIED (code) |
| Request ID > 0 | `NewRPCRequest` rejects id == 0 | VERIFIED (test) |
| Method allowlist | status, lang, switch_plugin; unknown rejected | VERIFIED (test) |
| `lang` locale allowlist | 10 observed UTF-8 locales | VERIFIED (test) |
| `switch_plugin` plugin ID allowlist | IDs 1-8, 11-16; 9 and 10 excluded | VERIFIED (test) |
| `switch_plugin` bool flag | second param must be bool | VERIFIED (test) |
| Bounded response read | `io.LimitReader`, MaxRPCResponseSize = 1 MiB | VERIFIED (test) |
| JSON validity | `json.Valid` before return | VERIFIED (test) |
| Schema-neutral response | preserves unknown fields; accepts objects/arrays | VERIFIED (test) |

## Network transport: NOT IMPLEMENTED

- No HTTP client to `/bot/ajax/`
- No session lifecycle, reconnection, or event delivery
- No multipart form transport
- No authentication, cookies, room keys, or session material handled
- No authorized integration environment available

## Compatibility claims: NONE

Per the engineering contract and threat model:
- No Cloud Bot or legacy standalone bot compatibility is claimed.
- Plugin ID 10's identity is NOT inferred from the partial heap-related capture.
- Plugin ID 9 (roulette) is retained only as a historical observation; the project does not implement wagering/betting/gambling.
- "OBSERVED" in the compatibility matrix means a configuration page appeared in a capture; it does NOT prove runtime execution or Cloud Bot support.

## Required future evidence (per plugin-compatibility-matrix.md promotion gate)

For every supported product/version:
- exact upstream product and version
- authorized environment and test identity (never committed)
- sanitized request/response fixture
- successful and negative test cases
- permission/role requirements
- timeout/session-expiry behavior
- independent test evidence reviewed in the PR

Until then, clients must show capabilities as unavailable/unverified and must not expose actions implying support.

## Verdict

Camfrog protocol support is **NOT IMPLEMENTED** and must not be claimed. The RPC serialization/validation boundary is implemented and tested but is not a network integration. No authorized integration testing was performed (no environment, no credentials, no target).
