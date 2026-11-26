---
change_id: CHG-0049
artifact: delta
status: active
updated: 2026-10-20
---

# Delta — Mission Description Format

## What Is Changing

### Mission Board

- `Description` on definitions, offers, and objectives is declared Markdown
  content: clients render paragraphs, emphasis, and lists; the service stores
  and returns it verbatim without sanitizing.
- `MissionDefinition` and `MissionOffer` carry `Rules`: an ordered list of
  short, self-contained requirement statements a client renders as bullets.
- `POST /api/v1/missions/definitions` accepts optional `rules` (list of
  non-empty strings); `GET /definitions`, `GET /offers`, `GET /offers/{id}`,
  `POST /offers`, and contract detail return `Rules`.
- Offers freeze `rules` from the published version, as title/objectives do.
- Seed catalogs write a short markdown `Description` and move the shared
  requirement sentences into `Rules` items.

## MODIFIED

### REQ-MB-009 — Publish an offer (payload)

`POST /api/v1/missions/offers` accepts the same request fields; the created
offer snapshots `title`, `description`, `completion_reward`, `objectives`,
**and `rules`** from the referenced published version.

### REQ-MB-015 — Expose mission content on the board and catalog

`GET /api/v1/missions/offers` and `GET /api/v1/missions/definitions` return
`Title`, `Description`, `CompletionReward`, **and `Rules`** alongside the
existing fields. `Description` is Markdown-formatted text; `Rules` is an
ordered list of requirement strings. Offers keep returning the frozen
`objectives` and `merit_summary` unchanged.

## BDD Scenarios

### SCN-DESC-MARKDOWN — Description round-trips as markdown

Given an authenticated operator creates a definition whose `description`
contains markdown (`## Requisitos`, list items, blank lines), when the
request is valid, then the definition, the published offer, and every read
endpoint return the string verbatim.

### SCN-RULES-STRUCTURED — Rules travel as a list

Given a definition created with `rules: ["Solo PU/LIVE", "Video continuo
sin cortes"]`, when an offer is published from that version, then
`GET /offers` returns `Rules: ["Solo PU/LIVE", "Video continuo sin cortes"]`
in order.

### SCN-RULES-FROZEN — Rules freeze per version

Given an accepted offer from `v1` with rules R1, when the operator
publishes `v2` with different rules, then the `v1` offer and the contract
still return R1.

### SCN-RULES-LEGACY — Legacy items degrade safely

Given a stored offer or definition written before this change (no `Rules`
attribute), when it is listed, then `Rules` decodes as an empty list and
`Description` is returned unchanged.

## HTTP Boundary Changes

- `POST /api/v1/missions/definitions` — request adds optional `rules`
  (`[]string`, snake_case `rules`). Additive.
- `GET /api/v1/missions/definitions`, `GET /api/v1/missions/offers`,
  `GET /api/v1/missions/offers/{offer_id}`, `POST /api/v1/missions/offers`,
  contract detail — responses add `Rules`. Additive.
- `Description` gains a declared Markdown format; the wire value stays a
  string. Contract-compatible.

## Error Vocabulary

- `INVALID_MISSION_DEFINITION` — extended: also returned when `rules`
  contains an empty or blank string item.
- All other codes unchanged.
