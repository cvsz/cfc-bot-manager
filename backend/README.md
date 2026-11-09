# Go Backend

The backend provides:
- validated HTTP configuration with a loopback default;
- liveness, readiness, and version metadata endpoints;
- bounded HTTP timeouts, graceful shutdown, and JSON structured logs;
- an authenticated room-message-style API backed by PostgreSQL when configured;
- an embedded browser editor at `/room-style-editor/`;
- Camfrog RPC contract helpers based only on observed fields.

## Enable room-style persistence

Configure both `DATABASE_URL` and `ROOM_STYLE_API_TOKEN`. Generate the token with:

```sh
openssl rand -hex 32
```

Apply the schema and start the API:

```sh
make db-migrate
cd backend && go run ./cmd/api
```

The default listener is `127.0.0.1:8080`. Open `http://127.0.0.1:8080/room-style-editor/`. The editor asks for the token at runtime and does not store it in browser storage. Use HTTPS outside localhost.

The bearer token is a single-operator bootstrap mechanism, not the planned identity, tenant isolation, or RBAC system. Do not expose this API to multiple users or the public internet. Keep it loopback-bound or behind a separately reviewed HTTPS/authentication gateway.

## Style API

- `GET /api/v1/message-style/default`
- `PUT /api/v1/message-style/default`
- `GET /api/v1/rooms/{roomID}/message-style`
- `PUT /api/v1/rooms/{roomID}/message-style`
- `DELETE /api/v1/rooms/{roomID}/message-style`

All endpoints require `Authorization: Bearer <ROOM_STYLE_API_TOKEN>`. A room without an override inherits the global default. Deleting an override restores inheritance.

Style fields: `fontFamily`, `fontSize`, `color`, `bold`, `italic`, and `underline`. Supported fonts are Arial, Tahoma, Verdana, Georgia, Times New Roman, and Courier New. Font size is 10–28 px; colors are six-digit hexadecimal values.

## Camfrog formatting boundary

The editor only previews and persists preferences. It does not send messages and does not claim that Camfrog supports rich-text formatting. The message transport remains plain text until formatting is verified against an authorized Camfrog product/version.

## Validation

```sh
go test -race ./...
go vet ./...
go build ./...
```
