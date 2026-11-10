# Room Message Font Editor

## Scope

The first font-editor slice provides a browser-based editor, live preview, global defaults, per-room overrides, and PostgreSQL persistence. **It does not send room messages.**

## Critical: GUI Preview Only

**Static reverse engineering of `camfrog.exe` (SHA-256: `0b744602c004536dcf81e9965e8cb17358eb276b3876a1e70c50f72777c27e06`) confirms that Camfrog Desktop Client does NOT support rich text in outgoing chat messages.** All messages sent via UI Automation are plain text only.

The font editor is **GUI preview only** — it affects only the local preview display in the AutoDJ/editor GUI. Sent messages are **always plain text**; no bold, italic, underline, color, font size, or font family formatting is transmitted to Camfrog rooms.

### Evidence from Static Reverse Engineering

| Finding | Evidence | Confidence |
|---------|----------|------------|
| Only 3 GDI font APIs imported | `CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW` | High |
| No HTML/CSS/BBCode/RTF libraries imported | 0 imports of `mshtml.dll`, `riched20.dll`, `msftedit.dll` | High |
| HTML/CSS used for **incoming** rendering only | `font-family:`, `font-size:`, `color: #`, `font-weight: bold;`, `font-style: italic;` | High |
| Outgoing messages = plain text only | UIA send path analysis, no rich-text libraries imported | High |

The `MessageStyle` configuration is used **only** for the GUI preview panel. Sent messages are **always plain text**.

---

## Style contract

- `fontFamily`: Arial, Tahoma, Verdana, Georgia, Times New Roman, or Courier New.
- `fontSize`: integer from 10 through 28.
- `color`: six-digit hexadecimal color.
- `bold`, `italic`, `underline`: booleans.

The backend normalizes color to uppercase and rejects unknown fields, invalid values, malformed JSON, oversized bodies, and invalid room identifiers.

---

## Precedence

Room override > global default. A room without an override inherits the global default. Deleting a room override restores inheritance.

---

## API and access boundary

The API uses `ROOM_STYLE_API_TOKEN`, a bearer token of at least 32 characters compared in constant time. The editor requests the token and keeps it only in page memory.

This is a single-operator bootstrap mechanism, not multi-user identity, tenant isolation, or RBAC. Keep the service loopback-bound unless protected by a separately reviewed HTTPS/authentication gateway. Do not expose it as a public multi-tenant API.

---

## Camfrog compatibility

The preview is browser-only. **Actual Camfrog message rendering remains disabled.** The adapter **must send plain text**; rich-text support is NOT supported by Camfrog Desktop Client. Static RE confirms no rich-text libraries are imported and UIA send path is plain text only.

---

## Migration and rollback

Apply the versioned migration with `make db-migrate`. The migration runner uses a transaction and PostgreSQL advisory lock. Before any rollback, stop the service and take a verified database backup. Dropping the style tables is destructive and requires explicit operator approval.
