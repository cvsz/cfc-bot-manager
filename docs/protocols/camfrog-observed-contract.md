# Camfrog Observed Interface Contract

## Evidence boundary

This document records behavior visible in:
- User-supplied HTML captures and descriptions (legacy)
- **Static reverse engineering of Camfrog Desktop Client `camfrog.exe`** (SHA-256: `0b744602c004536dcf81e9965e8cb17358eb276b3876a1e70c50f72777c27e06`) via Craft Anything native decompiler (clean-room lifter)

**Analysis Date:** 2026-11-09  
**Tools:** Craft Anything native toolset (clean-room lifter decompiler)  
**Status:** Static PE inspection only — Hopper license not available

This is not based on an independently authenticated Camfrog session, and it does not establish a public or stable upstream API. The adapter remains experimental until authorized integration tests confirm the behavior against a stated Camfrog product/version.

Never commit raw captures containing cookies, Set-Cookie headers, session tokens, room keys, passwords, or personal data. Store only sanitized fixtures.

---

## Critical Finding: No Rich Text Support in Outgoing Messages

**Camfrog Desktop Client does NOT support rich text in outgoing chat messages.** All messages sent via UI Automation are plain text only.

### Evidence from Static Reverse Engineering

| Finding | Evidence | Confidence |
|---------|----------|------------|
| Only 3 GDI font APIs imported | `CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW` | High |
| Font names stored as string literals | `main_font`, `gray_font`, `white_font`, `link_font`, etc. | High |
| IM-specific font settings | `IMFont`, `IMFontSize`, `FontName`, `Size` | High |
| HTML/CSS used for **incoming** message rendering | `font-family:`, `font-size:`, `color: #`, `font-weight: bold;`, `font-style: italic;` | High |
| No HTML/CSS/BBCode/RTF libraries imported | Import table analysis (0 imports of `mshtml.dll`, `riched20.dll`, `msftedit.dll`) | High |
| Chat messages sent via UIA = plain text | UIA adapter send path analysis | High |

**Conclusion:** Camfrog renders incoming messages with HTML/CSS in an embedded web view for display purposes, but **outgoing messages sent via UI Automation are plain text only**. The CFC Bot Manager cannot send rich-text messages (bold, italic, underline, color, font size, font family) to Camfrog rooms.

---

## Shared RPC endpoint

Observed request:
- Method: `POST`
- Path: `/bot/ajax/`
- Content type: `application/json`
- Envelope:

```typescript
interface CamfrogRpcRequest {
  id: number;
  method: string;
  params: unknown[];
}
```

Observed methods:
- `status` with `[]`
- `lang` with `[language]`
- `switch_plugin` with `[pluginId, enabled]`

Status results expose `uptime`, `reason`, and `status`. Captured status values include initialization, connection, working/paused, and account/room failure states. Treat upstream values as an open string union and preserve unknown values; do not reject newly introduced status text merely because it is unknown.

The source page polls status approximately every five seconds. A new client must prevent overlapping polls, apply request timeouts/cancellation, and validate response shape.

---

## Separate settings transport

Plugin settings were submitted as normal `POST` forms using `multipart/form-data`. This is distinct from the JSON RPC endpoint. Preserve exact upstream field names in transport DTOs; map to domain settings only after validation. Unchecked HTML checkboxes may be omitted from form data and need explicit normalization.

---

## Room page fields observed

- `_1`: room name
- `_2`: room access key
- `_3`: bot language
- `_4`: pause checkbox, value `y`

Captured language identifiers include en_US, es_ES, fr_FR, nl_NL, it_IT, id_ID, tr_TR, th_TH, fil_PH, and pt_BR UTF-8 locale strings. Exact supported values must be verified against the target runtime.

---

## HTML/CSS Rendering (Incoming Messages Only)

Camfrog uses HTML/CSS templates for **rendering incoming messages** in its embedded web view. These are **not** used for sending messages.

### Username Link Templates

```html
%5%<a class='username' style='color: %1%; font-weight: %2%; text-decoration: underline;' href='camfrog:room:openvideo:%3%'>%4%</a>
```

```html
%4%<span class='username' style='color: %1%; font-weight: %2%;'>%3%</span>
```

These templates render **incoming** username links in chat with:
- Custom color (`color: %1%`)
- Font weight (`font-weight: %2%`)
- Text decoration (`text-decoration: underline`)
- Camfrog room protocol URI (`camfrog:room:openvideo:%3%`)

### General HTML/CSS Style Strings (Incoming Message Rendering Only)

| String | Purpose |
|--------|---------|
| `<P style=" color: #808080; font-family: Segoe UI; font-size: 10pt; text-align: center;">` | Centered system message style |
| `font-weight: bold;` | Bold text style |
| `font-style: italic;` | Italic text style |
| `font-size: ` | Font size CSS property |
| `font-family: ` | Font family CSS property |
| `" style='"` | HTML style attribute opening |
| `"text-decoration: "` | Text decoration CSS property |
| `" underline"` | Underline text decoration value |
| `" line-through"` | Line-through text decoration value |
| `"font-weight: bold; "` | Bold font weight CSS |
| `"font-style: italic; "` | Italic font style CSS |
| `"font-size: "` | Font size CSS property |
| `" color: #"` | Font color CSS property |
| `" font-family: "` | Font family CSS property |

### Link Style Template (Incoming)

```html
<a style='color: #0000ff; font-weight: bold; text-decoration: underline; cursor: pointer;' href='%1%'>%2%</a>
```

Blue hyperlink style for clickable links in chat.

### Camfrog Protocol URI Scheme

Camfrog uses a custom URI scheme for room links:
- `camfrog:room:openvideo:%3%` — Opens a video chat room

---

## Plugin evidence inventory

| ID | Captured plugin name | Evidence |
|---:|---|---|
| 1 | 8 Ball | OBSERVED |
| 2 | IM Control | OBSERVED |
| 3 | Messenger | OBSERVED |
| 4 | Greeter | OBSERVED |
| 5 | Math Trivia | OBSERVED |
| 6 | Bad Words | OBSERVED |
| 7 | Trivia | OBSERVED |
| 8 | Word Trivia | OBSERVED |
| 9 | Roulette | OBSERVED |
| 10 | Unknown | UNVERIFIED — partial heap-related fields only; no page/name/complete save flow |
| 11 | Ghost | OBSERVED |
| 12 | Who banned | OBSERVED |
| 13 | Talkback | OBSERVED |
| 14 | Dice | OBSERVED |
| 15 | DJ Manager | OBSERVED |
| 16 | Auto Moderator | OBSERVED |

“OBSERVED” means a user-supplied HTML capture exposed a configuration page; it does not prove runtime execution, semantics, or Cloud Bot support. For ID 10, never display a guessed plugin name. Distinguish observed configuration fields from verified behavior.

---

## Selected field-level observations

- **1 — 8 Ball:** `_1` bot name, `_2` messages, `font_picker` JSON.
- **2 — IM Control:** `_1` owners textarea, `font_picker`.
- **3 — Messenger:** `_1` messages, `_2` display interval seconds (default 60), `font_picker`.
- **4 — Greeter:** hidden JSON greeting array (18 categories), role/gender checkboxes, role selector, greeting textarea, `font_picker`; `%1%` placeholders observed but semantics not fully verified.
- **5 — Math Trivia:** interval and timeout selects; notify/display checkboxes; operand count/class; addition/subtraction/multiplication toggles; no division option in the capture.
- **6 — Bad Words:** rules textarea, violation threshold, punishment reason, regex checkbox, kick/punish checkbox, `font_picker`. Rules grammar and action semantics unknown.
- **7 — Trivia:** interval/timeout, notify/display checkboxes, QUIZ import and `/bot/download/7` export.
- **8 — Word Trivia:** interval/timeout, notify/display checkboxes, QUIZ import and `/bot/download/8` export.
- **9 — Roulette:** messages, interval hours, notify and timer checkboxes, `font_picker`; placeholder/game semantics unknown.
- **11 — Ghost:** level-selection and question timeout, notify/display checkboxes, word import and `/bot/download/11` export.
- **12 — Who Banned:** eight event-message templates, warn/private notification checkboxes, retention days 1–7, `font_picker`; placeholder and precedence semantics unknown.
- **13 — Talkback:** rules textarea with one example of pipe-separated trigger/response and placeholders; full grammar unknown.
- **14 — Dice:** dice type values 2–12 and count 1–6; meaning of “type” not confirmed.
- **15 — DJ Manager:** `t1` default 30 with unknown unit/range; moderator-only and notify checkboxes; queue/turn semantics unknown.
- **16 — Auto Moderator:** nickname textarea, report threshold select (3–10, 100, 1000), `font_picker`; nickname grammar and action semantics unknown.

---

## Adapter safety requirements

- Treat plugin settings as untrusted input and preserve raw multiline strings unless a verified grammar exists.
- Validate numeric/select values against captured allowed options, while allowing versioned capability metadata to override only with test evidence.
- Never execute upstream-provided regexes locally without a bounded, safe regex engine and explicit resource limits.
- Never render messages, nicknames, reasons, or greetings as HTML.
- Do not use optimistic UI state for RPC mutations. Wait for a validated success response and reconcile failures.
- Do not retry non-idempotent form submissions or actions without an established idempotency contract.
- Use per-session serialization if upstream sessions are not proven safe for concurrent requests.
- Keep raw transport logging disabled or strictly redacted.
- **Outgoing messages are plain text only — do not attempt to send HTML/CSS/rich text formatting.**

---

## Required future evidence

For every supported product/version: sanitized request/response fixtures, authorized test account, success and failure cases, timeout behavior, session expiry behavior, validation behavior, and compatibility test results. Cloud Bot and legacy standalone bot must be represented as separate product targets.

---

## Safety exclusion

Plugin ID 9 ("Roulette") is retained here solely as a historical configuration-page observation. It must remain disabled in CFC Bot Manager; this project does not implement or facilitate wagering, betting, stakes, payouts, or gambling workflows.