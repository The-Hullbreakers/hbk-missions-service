---
change_id: CHG-0047
artifact: design
status: implemented
updated: 2026-10-20
---

# Design — Contract Merit Summary

## Domain — the fix

- `MissionContract` gains `MarshalJSON` (same alias pattern as
  `MissionOffer.MarshalJSON`): when `Offer.Objectives` is absent, the
  serialized offer is hydrated with `Terms.Objectives()` before encoding,
  so `offer.objectives` and the derived `offer.merit_summary` always
  reflect the frozen promise.
- When the offer already carries objectives (e.g. a fresh accept
  response), they take precedence untouched.
- `ContractTerms.Objectives()` already flattens phased terms; zero terms
  degrade to `objective_merits: 0`, `total_merits = completion_reward`.
- Domain placement beats a `mapToContract`-only fix: contract
  serialization flows through one upstream projection — store reads,
  seeds, in-memory doubles, and detail views all converge.

## Infrastructure (dynamodb)

- No change. Terms remain the single persisted source of truth; nothing
  new is written and no backfill is needed.

## Application / Boundary

- `ListPlayerContractsHandler`, `GetContractHandler`, review queue — no
  logic change; the contract already flows through.
- `MissionOffer.MarshalJSON` already emits `objectives` and
  `merit_summary`; the HTTP payload changes are automatic once the offer
  is hydrated.

## Affected projections

- `GET /api/v1/missions/contracts` (list), `GET .../contracts/{id}`
  (detail), and the review queue summary — all read via the same store
  mapping, so the fix is uniform.

## Risks

- Payload duplication: objectives appear under both `Offer.Objectives`
  and `Terms` — consistent with the published-offer shape; additive and
  harmless.
- Seeded contracts already persist `Terms`; no backfill needed.
