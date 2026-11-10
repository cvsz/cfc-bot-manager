# Plugin Compatibility Matrix

## Evidence terminology

- **OBSERVED**: configuration fields/page captured in user-supplied HTML; runtime behavior is not thereby verified.
- **STATIC_RE**: evidence from static reverse engineering of `camfrog.exe` (SHA-256: `0b744602c004536dcf81e9965e8cb17358eb276b3876a1e70c50f72777c27e06`) via Craft Anything native decompiler (clean-room lifter).
- **VERIFIED**: authorized runtime test confirms the described capability for a specific product/version.
- **UNVERIFIED**: missing, partial, or ambiguous evidence.
- **UNSUPPORTED**: confirmed unavailable in the stated product/version.

## Inventory

| ID | Name | Page/settings capture | Static RE evidence | Runtime behavior | Cloud Bot compatibility |
|---:|---|---|---|---|---|
| 1 | 8 Ball | OBSERVED | Font picker, bot name, messages | UNVERIFIED | UNVERIFIED |
| 2 | IM Control | OBSERVED | Owners textarea, font picker | UNVERIFIED | UNVERIFIED |
| 3 | Messenger | OBSERVED | Messages, display interval, font picker | UNVERIFIED | UNVERIFIED |
| 4 | Greeter | OBSERVED | Hidden JSON greeting array (18 categories), role/gender checkboxes, role selector, greeting textarea, font picker | UNVERIFIED | UNVERIFIED |
| 5 | Math Trivia | OBSERVED | Interval/timeout selects, notify/display checkboxes, operand count/class, addition/subtraction/multiplication toggles | UNVERIFIED | UNVERIFIED |
| 6 | Bad Words | OBSERVED | Rules textarea, violation threshold, punishment reason, regex checkbox, kick/punish checkbox, font picker | UNVERIFIED | UNVERIFIED |
| 7 | Trivia | OBSERVED | Interval/timeout, notify/display checkboxes, QUIZ import, `/bot/download/7` export | UNVERIFIED | UNVERIFIED |
| 8 | Word Trivia | OBSERVED | Interval/timeout, notify/display checkboxes, QUIZ import, `/bot/download/8` export | UNVERIFIED | UNVERIFIED |
| 9 | Roulette | OBSERVED | Messages, interval hours, notify/timer checkboxes, font picker | UNVERIFIED | UNVERIFIED |
| 10 | Unknown | UNVERIFIED | Partial heap-related fields only; no page/name/complete save flow | UNVERIFIED | UNVERIFIED |
| 11 | Ghost | OBSERVED | Level-selection, question timeout, notify/display checkboxes, word import, `/bot/download/11` export | UNVERIFIED | UNVERIFIED |
| 12 | Who Banned | OBSERVED | Eight event-message templates, warn/private notification checkboxes, retention days 1–7, font picker | UNVERIFIED | UNVERIFIED |
| 13 | Talkback | OBSERVED | Rules textarea with pipe-separated trigger/response and placeholders | UNVERIFIED | UNVERIFIED |
| 14 | Dice | OBSERVED | Dice type values 2–12, count 1–6 | UNVERIFIED | UNVERIFIED |
| 15 | DJ Manager | OBSERVED | `t1` default 30 with unknown unit/range; moderator-only/notify checkboxes | UNVERIFIED | UNVERIFIED |
| 16 | Auto Moderator | OBSERVED | Nickname textarea, report threshold select (3–10, 100, 1000), font picker | UNVERIFIED | UNVERIFIED |

## Static RE Evidence Summary (camfrog.exe)

| Finding | Evidence | Confidence |
|---------|----------|------------|
| Only 3 GDI font APIs imported | `CreateFontIndirectW`, `EnumFontFamiliesExW`, `ChooseFontW` | High |
| Font names stored as string literals | `main_font`, `gray_font`, `white_font`, `link_font`, `IMFont`, `IMFontSize`, `FontName`, `Size` | High |
| IM-specific font settings in Registry | `IMFont`, `IMFontSize`, `FontName`, `Size` at `\\Settings\\General\\`, `\\ProfileInfo\\` | High |
| HTML/CSS used for **incoming** message rendering only | `font-family:`, `font-size:`, `color: #`, `font-weight: bold;`, `font-style: italic;`, `camfrog:room:openvideo:` URI scheme | High |
| No HTML/CSS/BBCode/RTF libraries imported | 0 imports of `mshtml.dll`, `riched20.dll`, `msftedit.dll` | High |
| Outgoing messages = plain text only | UIA send path analysis, no rich-text libraries imported | High |

## Interpretation

- **Rich text in outgoing messages is NOT SUPPORTED** (Static RE: UNSUPPORTED). The room message font editor is GUI preview only.
- Do not infer that all captured pages belong to Cloud Bot. Captured legacy settings and the Cloud Bot product are distinct until product/version-specific tests prove otherwise.
- Do not create a guessed plugin #10 name from the incomplete heap-related capture. Preserve its raw evidence separately and keep it out of user-facing supported plugin lists.
- Plugin ID 9 (Roulette) is listed only because its configuration page appeared in a supplied capture. It must remain disabled; this project does not implement or facilitate wagering, betting, stakes, payouts, or gambling workflows.

## Promotion gate

A capability may be marked VERIFIED only with:
- exact upstream product and version;
- authorized environment and test identity (never committed);
- sanitized request and response fixture;
- successful and negative test cases;
- permission/role requirements;
- timeout/session-expiry behavior;
- independent test evidence reviewed in the PR.

Until then, clients must show the capability as unavailable/unverified and must not expose an action that implies it is supported.


## Safety exclusion

Plugin ID 9 (Roulette) is listed only because its configuration page appeared in a supplied capture. It is not an enabled or supported product capability. The product must not expose wagering, betting, stakes, payouts, or gambling workflows. Keep ID 9 disabled unless independent evidence establishes a strictly non-wagering use case and the scope is reviewed for age-appropriate safety.

## Static RE Evidence References

See `docs/protocols/camfrog-observed-contract.md` for full Static RE evidence details including:
- PE Import analysis (only 3 GDI font APIs)
- Font string literals (`main_font`, `gray_font`, `white_font`, `link_font`, `IMFont`, `IMFontSize`, `FontName`, `Size`)
- HTML/CSS rendering for incoming messages only (`font-family:`, `font-size:`, `color: #`, `font-weight: bold;`, `font-style: italic;`, `camfrog:room:openvideo:` URI scheme)
- No rich-text libraries imported (`mshtml.dll`, `riched20.dll`, `msftedit.dll` not imported)
- Decompiled functions: `sub_140437e60` (font initialization), `sub_1404adc70` (IM font settings), `sub_140446470` (HTML/CSS style building)
- Only 3 GDI font APIs imported; no `mshtml.dll`, `riched20.dll`, `msftedit.dll`
- UIA send path = plain text only
