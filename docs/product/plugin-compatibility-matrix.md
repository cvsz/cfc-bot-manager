# Plugin Compatibility Matrix

## Evidence terminology

- OBSERVED: configuration fields/page captured in user-supplied HTML; runtime behavior is not thereby verified.
- VERIFIED: authorized runtime test confirms the described capability for a specific product/version.
- UNVERIFIED: missing, partial, or ambiguous evidence.
- UNSUPPORTED: confirmed unavailable in the stated product/version.

## Inventory

| ID | Name | Page/settings capture | Runtime behavior | Cloud Bot compatibility |
|---:|---|---|---|---|
| 1 | 8 Ball | OBSERVED | UNVERIFIED | UNVERIFIED |
| 2 | IM Control | OBSERVED | UNVERIFIED | UNVERIFIED |
| 3 | Messenger | OBSERVED | UNVERIFIED | UNVERIFIED |
| 4 | Greeter | OBSERVED | UNVERIFIED | UNVERIFIED |
| 5 | Math Trivia | OBSERVED | UNVERIFIED | UNVERIFIED |
| 6 | Bad Words | OBSERVED | UNVERIFIED | UNVERIFIED |
| 7 | Trivia | OBSERVED | UNVERIFIED | UNVERIFIED |
| 8 | Word Trivia | OBSERVED | UNVERIFIED | UNVERIFIED |
| 9 | Roulette | OBSERVED | UNVERIFIED | UNVERIFIED |
| 10 | Unknown | UNVERIFIED | UNVERIFIED | UNVERIFIED |
| 11 | Ghost | OBSERVED | UNVERIFIED | UNVERIFIED |
| 12 | Who Banned | OBSERVED | UNVERIFIED | UNVERIFIED |
| 13 | Talkback | OBSERVED | UNVERIFIED | UNVERIFIED |
| 14 | Dice | OBSERVED | UNVERIFIED | UNVERIFIED |
| 15 | DJ Manager | OBSERVED | UNVERIFIED | UNVERIFIED |
| 16 | Auto Moderator | OBSERVED | UNVERIFIED | UNVERIFIED |

## Interpretation

Do not infer that all captured pages belong to Cloud Bot. Captured legacy settings and the Cloud Bot product are distinct until product/version-specific tests prove otherwise.

Do not create a guessed plugin #10 name from the incomplete heap-related capture. Preserve its raw evidence separately and keep it out of user-facing supported plugin lists.

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

Plugin ID 9 is listed only because its configuration page appeared in a supplied capture. It is not an enabled or supported product capability. The product must not expose wagering, betting, stakes, payouts, or gambling workflows. Keep ID 9 disabled unless independent evidence establishes a strictly non-wagering use case and the scope is reviewed for age-appropriate safety.
