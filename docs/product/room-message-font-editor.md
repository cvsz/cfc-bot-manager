# Room Message Font Editor

## Scope

The first font-editor slice provides a browser-based editor, live preview, global defaults, per-room overrides, and PostgreSQL persistence. It does not send room messages.

## Style contract

- `fontFamily`: Arial, Tahoma, Verdana, Georgia, Times New Roman, or Courier New.
- `fontSize`: integer from 10 through 28.
- `color`: six-digit hexadecimal color.
- `bold`, `italic`, `underline`: booleans.

The backend normalizes color to uppercase and rejects unknown fields, invalid values, malformed JSON, oversized bodies, and invalid room identifiers.

## Precedence

Room override > global default. A room without an override inherits the global default. Deleting a room override restores inheritance.

## API and access boundary

The API uses `ROOM_STYLE_API_TOKEN`, a bearer token of at least 32 characters compared in constant time. The editor requests the token and keeps it only in page memory.

This is a single-operator bootstrap mechanism, not multi-user identity, tenant isolation, or RBAC. Keep the service loopback-bound unless protected by a separately reviewed HTTPS/authentication gateway. Do not expose it as a public multi-tenant API.

## Camfrog compatibility

The preview is browser-only. Actual Camfrog message rendering remains disabled. The adapter must send plain text unless rich-text support is verified against an authorized product/version with sanitized contract fixtures and end-to-end evidence.

## Migration and rollback

Apply the versioned migration with `make db-migrate`. The migration runner uses a transaction and PostgreSQL advisory lock. Before any rollback, stop the service and take a verified database backup. Dropping the style tables is destructive and requires explicit operator approval.
