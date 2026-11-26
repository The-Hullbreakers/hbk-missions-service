---
change_id: CHG-0049
artifact: design
status: active
updated: 2026-10-20
---

# Design — Mission Description Format

## Domain

- `MissionDefinition` gains `Rules []string`; `NewMissionDefinition` accepts
  `rules` and rejects any blank item with `ErrDefinitionInvalid`. Nil is
  valid (no rules).
- `MissionOffer` gains `Rules []string`; `NewMissionOfferFromDefinition`
  freezes a defensive copy alongside `Title`/`Description`.
- `Objective` keeps `Description` unchanged; markdown is a rendering
  contract, not a domain invariant — the domain stores text verbatim.

## Application

- `CreateMissionDefinitionHandler` signature gains `rules []string` passed
  through to `NewMissionDefinition`.

## Boundary (interface/http)

- `createMissionDefinitionRequest` adds `rules` (`[]string`).
- Response projections echo `Rules`; `Description` serialization unchanged
  (markdown lives inside the existing string field).

## Infrastructure (dynamodb)

- `definition_store.go`, `offer_store.go`, `contract_item.go` write `Rules`
  as an `L` of `S` attributes (omit when empty).
- `definition_store_queries.go`, `offer_store_query.go`,
  `contract_store_mapping.go` decode `Rules`; a missing attribute yields
  `nil` (SCN-RULES-LEGACY).

## Seeds

- `seed_missions.py`: catalogs gain a top-level `rules: [...]` list and a
  short `description` markdown template; the script stops concatenating the
  rules prose into `Description`, writes `Rules` on definitions and offers.
- `seeds/pvp-missions.json`, `seeds/racing-missions.json`: regenerate with
  markdown `Description` + `Rules` items (the current wall-of-text sentences
  become list items).

## Risks

- Contract items already store the offer snapshot; adding `Rules` is
  additive and legacy reads tolerate absence.
- Item size shrinks (rules deduplicated from objective descriptions is out
  of scope; objective descriptions keep their current text).
