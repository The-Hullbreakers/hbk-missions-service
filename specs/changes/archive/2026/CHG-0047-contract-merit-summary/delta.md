---
change_id: CHG-0047
artifact: delta
status: implemented
updated: 2026-10-20
---

# Delta — Contract Merit Summary

## What Is Changing

### Mission Contract Fulfillment

- Contract projections (`GET /api/v1/missions/contracts`,
  `GET /api/v1/missions/contracts/{contract_id}`, review queue) embed the
  offer's frozen `objectives` rehydrated from `Terms`, so the derived
  `offer.merit_summary` reports the earnable total of the accepted promise.
- `merit_summary` uses the REQ-MB-015 formula over the frozen terms:
  `objective_merits = Σ (merit_per_unit + bonus) × unit_limit`,
  `total_merits = objective_merits + completion_reward`.

## MODIFIED

### REQ-MCF-006 — List player contracts

`GET /api/v1/missions/contracts` returns the contracts where the
authenticated player is in the roster. Filters: `season_id` and `status`
(`ALL` default, `ACTIVE` bucket for non-terminal, or an exact
`ContractState`). Each item's embedded `Offer` carries the contract's
frozen `objectives` and a `merit_summary` derived from `Terms`
(`objective_merits`, `completion_reward`, `total_merits`). Contracts with
no terms (legacy) report `objective_merits: 0` and
`total_merits = completion_reward`.

### REQ-MCF-022 — Read own contract detail (projection alignment)

The detail projection shares the same materialization path; its embedded
offer exposes the same frozen objectives and `merit_summary`.

## BDD Scenarios

### SCN-MCF-MERIT-SUMMARY — Contract shows the earnable total

Given a contract whose `Terms` freeze objectives `(merit_per_unit: 450,
unit_limit: 1)` and `(merit_per_unit: 50, unit_limit: 1)` with
`completion_reward: 0`, when the member lists `GET /api/v1/missions/
contracts`, then the embedded offer shows
`merit_summary: {objective_merits: 500, completion_reward: 0,
total_merits: 500}` and the two frozen objectives.

### SCN-MCF-MERIT-BONUS-LIMITS — Bonus and unit limits count

Given a term objective with `merit_per_unit: 10`, `bonus: 5`,
`unit_limit: 3` and `completion_reward: 200`, when the contract is read,
then `objective_merits: 45` and `total_merits: 245`.

### SCN-MCF-MERIT-LEGACY — Contracts without terms degrade safely

Given a stored contract with no `Terms`/`Phases` attributes, when it is
listed, then `offer.objectives` is empty and `total_merits` equals
`completion_reward`.

### SCN-MCF-MERIT-FROZEN — Republishing keeps the contract promise

Given a contract accepted at definition `v1`, when the operator publishes
`v2` with different merits, then the listed contract still reports the
`v1` frozen terms and totals (REQ-MCF-009).

## HTTP Boundary Changes

- `GET /api/v1/missions/contracts` and
  `GET /api/v1/missions/contracts/{contract_id}` — embedded `offer` gains
  populated `objectives` and a non-zero `merit_summary`. Additive only.

## Error Vocabulary

- Unchanged — no new codes.
