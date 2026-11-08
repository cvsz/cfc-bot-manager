# Go Backend Foundation

This module currently provides:
- validated HTTP server configuration with conservative loopback default;
- liveness, readiness, and version metadata endpoints;
- bounded HTTP timeouts and graceful shutdown;
- JSON structured logs;
- Camfrog RPC request/response contract helpers based only on observed fields.

## Run

```sh
cd backend
go run ./cmd/api
```

The default listener is `127.0.0.1:8080`. Override it explicitly with `HTTP_ADDR` when a different bind address is required.

## Endpoints

- `GET /healthz`: process liveness.
- `GET /readyz`: readiness; dependency checks will be injected as real dependencies are implemented.
- `GET /api/v1/meta`: build version metadata.

No authenticated application API or Camfrog network integration is exposed by this initial foundation. It must not be represented as a complete bot manager.

## Validate

```sh
go test ./...
go vet ./...
go build ./...
```

Camfrog RPC helpers serialize the observed `{id, method, params}` envelope and validate known parameter shapes. They do not prove that a captured plugin is supported by Cloud Bot or any particular legacy version.
