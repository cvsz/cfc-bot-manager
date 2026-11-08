# Camfrog Observed Interface Contract

## Evidence boundary

This document records behavior visible in user-supplied HTML captures and descriptions. It is not based on an independently authenticated Camfrog session, and it does not establish a public or stable upstream API. The adapter remains experimental until authorized integration tests confirm the behavior against a stated Camfrog product/version.

Never commit raw captures containing cookies, Set-Cookie headers, session tokens, room keys, passwords, or personal data. Store only sanitized fixtures.

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

## Separate settings transport

Plugin settings were submitted as normal `POST` forms using `multipart/form-data`. This is distinct from the JSON RPC endpoint. Preserve exact upstream field names in transport DTOs; map to domain settings only after validation. Unchecked HTML checkboxes may be omitted from form data and need explicit normalization.

## Room page fields observed

- `_1`: room name
- `_2`: room access key
- `_3`: bot language
- `_4`: pause checkbox, value `y`

Captured language identifiers include en_US, es_ES, fr_FR, nl_NL, it_IT, id_ID, tr_TR, th_TH, fil_PH, and pt_BR UTF-8 locale strings. Exact supported values must be verified against the target runtime.

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

## Adapter safety requirements

- Treat plugin settings as untrusted input and preserve raw multiline strings unless a verified grammar exists.
- Validate numeric/select values against captured allowed options, while allowing versioned capability metadata to override only with test evidence.
- Never execute upstream-provided regexes locally without a bounded, safe regex engine and explicit resource limits.
- Never render messages, nicknames, reasons, or greetings as HTML.
- Do not use optimistic UI state for RPC mutations. Wait for a validated success response and reconcile failures.
- Do not retry non-idempotent form submissions or actions without an established idempotency contract.
- Use per-session serialization if upstream sessions are not proven safe for concurrent requests.
- Keep raw transport logging disabled or strictly redacted.

## Required future evidence

For every supported product/version: sanitized request/response fixtures, authorized test account, success and failure cases, timeout behavior, session expiry behavior, validation behavior, and compatibility test results. Cloud Bot and legacy standalone bot must be represented as separate product targets.


## Safety exclusion

Plugin ID 9 ("Roulette") is retained here solely as a historical configuration-page observation. It must remain disabled in CFC Bot Manager; this project does not implement or facilitate wagering, betting, stakes, payouts, or gambling workflows.
