# Go Backend Review — cfc-bot-manager

**Scope:** `backend/` (Go module `github.com/cvsz/cfc-bot-manager/backend`, go 1.23, dependency-free)
**Source audited:**
- `backend/cmd/api/main.go`
- `backend/internal/config/config.go`
- `backend/internal/httpapi/server.go`
- `backend/internal/camfrog/rpc.go`
- Tests: `config_test.go`, `httpapi/server_test.go`, `camfrog/rpc_test.go`
**Date:** 2026-11-08

## Implementation summary

A single vertical slice of a Go HTTP backend:
- Configuration loader with strict validation (env, bind address, timeouts)
- HTTP server with security headers, GET-only routes, readiness probe, graceful shutdown
- Camfrog JSON-RPC request-envelope serialization/validation (allow-listed methods, locales, plugin IDs; bounded schema-neutral response reader)

No Camfrog network transport, no authentication, no persistence, no WebSocket, no bot lifecycle. This is a serialization/validation boundary, not a protocol integration.

## Verification results (local, VERIFIED)

| Check | Command | Result |
|-------|---------|--------|
| gofmt | `cd backend && gofmt -l .` | Clean (0 files) — fixed during this audit |
| go vet | `cd backend && go vet ./...` | PASS |
| go build | `cd backend && go build ./...` | PASS |
| go test | `cd backend && go test ./...` | PASS (camfrog, config, httpapi) |
| go test -race | `cd backend && go test -race ./...` | PASS |
| go test -cover | `cd backend && go test -cover ./...` | PASS (camfrog 84.4%, config 90.9%, httpapi 100%) |

`cmd/api` has no test files (0.0% coverage) — entry point only; behavior exercised through config/httpapi packages.

## Phase 4 required test coverage

| Required test | Covered? | Location / notes |
|---------------|----------|------------------|
| Malformed JSON | Yes | `TestReadJSONResponseBoundedAndSchemaNeutral` ("not-json") |
| Empty requests | Yes | `TestReadJSONResponseBoundedAndSchemaNeutral` (empty body); `TestMissingAndUnexpectedParameters` |
| Oversized responses | Yes | `TestReadJSONResponseBoundedAndSchemaNeutral` (> MaxRPCResponseSize) |
| Invalid methods | Yes | `TestRPCRequestRejectsUnknownMethodAndZeroID` ("delete_everything") |
| Invalid locales | Yes | `TestLanguageRequestRequiresObservedLocale` ("xx_XX.UTF-8") |
| Invalid plugin IDs | Yes | `TestSwitchPluginRejectsUnknownIDAndBadTypes` (IDs 9, 10) |
| Unknown plugin IDs | Yes | Same as above |
| Duplicate request IDs | Yes | `TestRequestIDPreservedThroughMarshal`, `TestDuplicateRequestIDRoundTrip` |
| Missing parameters | Yes | `TestMissingAndUnexpectedParameters` |
| Unexpected response formats | Yes | `TestReadJSONResponseBoundedAndSchemaNeutral` (schema-neutral `[1,2,3]`, `{"id":4}`) |
| Timeout | No | Not testable without server transport; config timeouts validated |
| Cancellation | No | Not testable without server transport; signal shutdown in main.go |
| Concurrent requests | Yes | `TestConcurrentMarshalAndRead` |
| Graceful shutdown | Partial | Implemented in main.go (signal.NotifyContext + Shutdown); no server-level test |

## Observations

- `rpc.go` `observedPluginIDs` excludes 9 (roulette) and 10 (unknown) — consistent with the plugin safety policy (ID 9 disabled pending safety evidence; ID 10 unverified).
- `ReadJSONResponse` is schema-neutral: it does not assume an undocumented response envelope. It preserves unknown fields and rejects invalid JSON / empty / oversized bodies.
- `config.Load()` rejects empty duration env values rather than silently falling back (defensive).
- `httpapi` never leaks internal errors to responses (readiness failure returns 503 with sanitized body).
- No authentication or authorization middleware — all endpoints are currently unauthenticated (NOT IMPLEMENTED, documented as future).

## Gaps (NOT IMPLEMENTED / PARTIALLY)

- Server-level timeout and cancellation tests (require an HTTP transport harness; not implemented)
- Graceful shutdown server-level test
- Authentication/authorization middleware
- Request size limits beyond `MaxHeaderBytes` (body not bounded at handler level)
- Rate limiting
- Structured log redaction (slog JSON handler used; no redaction hook)

## Verdict

The Go backend is correct, well-structured, and secure-by-design for the slice it implements. It does **not** claim Camfrog network integration, authentication, or persistence. Tests cover the RPC validation boundary thoroughly; server-level concurrency/timeout/shutdown tests remain as future work.
