---
change_id: CHG-0046
artifact: delta
status: completed
updated: 2026-10-19
---

# Delta — Offer Merit Breakdown

## What Is Changing

### Mission Board

- `MissionOffer` freezes the published definition's objectives alongside
  `title`, `description`, and `completion_reward`.
- `GET /api/v1/missions/offers`, `GET /api/v1/missions/offers/{offer_id}`,
  and `POST /api/v1/missions/offers` responses include `objectives` and a
  `merit_summary` breakdown.
- `merit_summary` derives `objective_merits` (Σ `(merit_per_unit + bonus) ×
  unit_limit` over all objectives), echoes `completion_reward`, and reports
  `total_merits` as their sum — the maximum a contract can earn.

## MODIFIED

### REQ-MB-015 — Expose mission content on the board and catalog

`GET /api/v1/missions/offers` and `GET /api/v1/missions/definitions` return
`Title`, `Description`, and `CompletionReward` alongside the existing
fields. Offers additionally return the frozen `objectives` (each with
`objective_id`, `name`, `description`, `category`, `merit_per_unit`,
`bonus`, `unit_limit`) and `merit_summary` (`objective_merits`,
`completion_reward`, `total_merits`). Publishing a new version may change
`title`/`description`; objectives, `completion_reward`, and derived totals
stay frozen per version so live offers and contracts never change.

### REQ-MB-009 — Publish an offer (payload)

`POST /api/v1/missions/offers` accepts the same request fields; the created
offer snapshots `title`, `description`, `completion_reward`, **and
`objectives`** from the referenced published version.

## BDD Scenarios

### SCN-OFFER-MERIT-BREAKDOWN — Offer shows the earnable total

Given a published definition `pvp-live-air-10 v1` with objectives of
`(merit_per_unit: 270, unit_limit: 1)` and `(merit_per_unit: 30,
unit_limit: 1)` and `completion_reward: 0`, when a member lists
`GET /api/v1/missions/offers`, then the offer carries both structured
objectives and `merit_summary: {objective_merits: 300,
completion_reward: 0, total_merits: 300}`.

### SCN-OFFER-MERIT-BONUS-AND-LIMITS — Bonus and unit limits count

Given an objective with `merit_per_unit: 10`, `bonus: 5`, `unit_limit: 3`
and `completion_reward: 200`, when the offer is read, then
`objective_merits: 45` and `total_merits: 245`.

### SCN-OFFER-MERIT-LEGACY — Legacy offers degrade safely

Given a stored offer written before this change (no `Objectives`
attribute), when it is listed, then `objectives: []` and `total_merits`
equals `completion_reward`.

### SCN-OFFER-MERIT-FROZEN — Republishing keeps the snapshot

Given an accepted offer from `v1`, when the operator publishes `v2` with
changed objectives, then the `v1` offer still returns `v1` objectives and
totals.

## HTTP Boundary Changes

- `GET /api/v1/missions/offers`, `GET /api/v1/missions/offers/{offer_id}`,
  `POST /api/v1/missions/offers` — responses add `objectives` and
  `merit_summary`. Additive only.

## Error Vocabulary

- `DEFINITION_INVALID` — unchanged.
- `OFFER_NOT_AVAILABLE` — unchanged.
