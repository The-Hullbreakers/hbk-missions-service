---
change_id: CHG-0025
artifact: delta
status: proposed
updated: 2026-10-09
---

# Delta — Rename Definition Identifier To Definition Code

## What Is Changing

### Mission Board

- The mission-definition identifier is named `definition_code` end to end:
  create body, publish path, offer reference, list responses, and board
  exclusion vocabulary.

### Mission Contract Fulfillment

- Contracts preserve `definition_code` and `definition_version` from the
  offer; the frozen-terms and summary language uses the new name.

## MODIFIED

### REQ-MB-008 — Hide an active template

If a player participates in an `ACCEPTED`, `IN_PROGRESS`, or
`PENDING_VALIDATION` contract from a stable definition, the board excludes
every offer of the same `definition_code`, even from another version.
Acceptance revalidates the rule to prevent bypass and concurrency.

### REQ-MB-010 — List mission offers

`GET /api/v1/missions/offers` returns each offer carrying `definition_code`
(instead of `definition_id`) plus `definition_version`.

### REQ-MB-012 — Read catalog data only inside the caller scope

All definition lookups — public and internal during acceptance and
confirmation — resolve by `definition_code`; an unknown code resolves as not
found. (Post-CHG-0024 this is a single shared scope.)

### REQ-MCF — Preserve definition reference in the contract

A contract preserves the accepted `definition_code` and `definition_version`
from the offer, plus the frozen `ContractTerms` snapshot.

## HTTP Boundary Changes

- `POST /api/v1/missions/definitions` — body field `definition_id` renamed to
  `definition_code` (required, operator-chosen, no format constraint beyond
  non-empty).
- `POST /api/v1/missions/definitions/{definition_code}/publish` — path
  segment renamed.
- `POST /api/v1/missions/offers` — body field `definition_id` renamed to
  `definition_code`.
- All JSON responses exposing `DefinitionID` now expose `DefinitionCode`.

## BDD Scenarios

### SCN-DEF-CODE-CREATE — Create a definition with an operator code

Given an authenticated operator, when they `POST /api/v1/missions/definitions`
with `definition_code: "def-kills-100"`, `category`, `capacity`, and
objectives, then the definition is created in `DRAFT` and the response echoes
`DefinitionCode: "def-kills-100"`.

### SCN-DEF-CODE-REQUIRED — Code is mandatory

Given an authenticated operator, when they create a definition without
`definition_code` (or empty), then the response is `422` with the definition
validation error.

### SCN-DEF-CODE-PUBLISH — Publish by code in the path

Given a `DRAFT` definition `def-kills-100`, when the operator calls
`POST /api/v1/missions/definitions/def-kills-100/publish` with a version, then
the definition transitions to `PUBLISHED` under that code.

### SCN-DEF-CODE-OFFER — Offer references the code

Given a published definition `def-kills-100` at version `v1`, when an offer is
created with `definition_code: "def-kills-100"` and `definition_version:
"v1"`, then the offer and any resulting contract carry `DefinitionCode`.

### SCN-DEF-CODE-LEGACY-READ — Old items still resolve

Given a stored contract or offer item with legacy attribute `DefinitionID`,
when it is read, then the value maps to `DefinitionCode` (same fallback rule
as the existing legacy-version handling).

## Error Vocabulary

- `DEFINITION_NOT_FOUND` — unchanged; now triggers on unknown code.
- Definition validation error (`ErrDefinitionInvalid` mapping) — unchanged;
  empty `definition_code` fails identically to empty `definition_id` before.
