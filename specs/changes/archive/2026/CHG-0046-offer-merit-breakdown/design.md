---
change_id: CHG-0046
artifact: design
status: completed
updated: 2026-10-19
---

# Design — Offer Merit Breakdown

## Domain

- `MissionOffer` gains `Objectives []Objective` (defensive copies in/out),
  populated by `NewMissionOfferFromDefinition`. Direct constructors used by
  tests may leave it empty.
- `MeritSummary` value object: `ObjectiveMerits()`, `CompletionReward()`,
  `TotalMerits()` derived via
  `MissionOffer.MeritSummary()` — `Σ (MeritPerUnit + Bonus) × UnitLimit` +
  `CompletionReward`. Computation lives in Domain; serialization is a
  projection concern.

## Application

- `ListOffersHandler`/`GetOfferHandler`/`PublishMissionHandler` — no logic
  change; the offer already flows through.

## Boundary (interface/http)

- Offer payloads add `objectives` (existing `objectiveRuleResponse`-style
  shape) and `merit_summary`. Marshal through a small projection so the
  domain struct stays serialization-agnostic.

## Infrastructure (dynamodb)

- `offer_store.go` writes `Objectives` as a list attribute on publish.
- `offer_store_query.go` `mapOfferItem` decodes the list; a missing
  attribute yields `nil` (legacy tolerance, SCN-OFFER-MERIT-LEGACY).

## Seeds

- `seed_missions.py` `offer_item` stops excluding `Objectives` so seeded
  offers carry the frozen list.

## Risks

- Item size: objectives are ≤5 per seed rules and reuse the frozen
  description strings — acceptable within DynamoDB limits.
- Legacy items: decode tolerates absence; local board re-seeded.
