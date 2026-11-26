---
change_id: CHG-0047
title: Show contract merit summary from frozen terms
status: completed
risk: L2
canonical_action: modify
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
execution_mode: sequential
---

# Show contract merit summary from frozen terms

## Defect analysis

- **BUG-0047 — Listed contracts report zero earnable merits.**
  `GET /api/v1/missions/contracts` embeds each contract's offer with
  `objectives: null` and `merit_summary` all zeros, even when the frozen
  `Terms` carry objectives paying merits (e.g. `merit_per_unit` 450 + 50).
- **Expected:** the embedded offer projection exposes the frozen objectives
  and a `merit_summary` computed from the contract's frozen `Terms`:
  `objective_merits = Σ (merit_per_unit + bonus) × unit_limit`,
  `total_merits = objective_merits + completion_reward`.
- **Actual:** `MeritSummary{0,0,0}` and `Objectives: null` on every
  contract read.
- **Root cause:** `mapToContract` (`internal/missions/infrastructure/
  dynamodb/contract_store_mapping.go`) rebuilds the embedded `MissionOffer`
  without `Objectives`. Contract items never duplicated objectives onto
  offer attributes — the frozen snapshot lives in `Terms`/`Phases` — so
  `MissionOffer.MeritSummary()` (REQ-MB-015) sums an empty list.

## Regression requirement

- **REG-0047 — Contract reads derive offer merit data from frozen terms.**
  Every contract projection (list, detail, review queue) rehydrates the
  embedded offer's objectives from `Terms` so `merit_summary` reflects the
  promise captured at acceptance (REQ-MCF-009), never live offer data.

## Non-goals

- No new response field: the change fixes the existing `offer.objectives`
  and `offer.merit_summary` projection (additive, backward compatible).
- No crediting-rule change; no duplicate `Objectives` attribute written to
  contract items — terms stay the single source of truth.
- No seed or migration backfill: stored `Terms` are already complete.

## Rollback

- Reverting restores the zero summary; stored items are unchanged and
  remain valid (no schema write added).
