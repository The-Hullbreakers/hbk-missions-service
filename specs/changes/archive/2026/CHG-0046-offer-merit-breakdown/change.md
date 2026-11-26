---
change_id: CHG-0046
title: Expose offer objectives and merit breakdown
status: completed
risk: L2
canonical_action: modify
canonical_paths:
  - specs/capabilities/mission-board/spec.md
execution_mode: sequential
---

# Expose offer objectives and merit breakdown

## Defect analysis

- **BUG-0046 — Offers hide earnable merits.** `GET /api/v1/missions/offers`
  returns `CompletionReward` only; every seeded mission sets it to `0`, so
  clients display "0 merits" although objectives pay merits on confirmation.
- **Expected:** the offer exposes structured objectives
  (`objective_id`, `name`, `description`, `category`, `merit_per_unit`,
  `bonus`, `unit_limit`) and a merit summary
  (`objective_merits`, `completion_reward`, `total_merits`).
- **Actual:** no `objectives` on the offer payload or item; no derived total.
- **Root cause:** `MissionOffer` never froze objectives (unlike
  `Title`/`Description`/`CompletionReward` from CHG-0026), so neither the
  DynamoDB item, the seed writer, nor the HTTP payload can surface them.

## Regression requirement

- **REG-0046 — Offer carries frozen objectives and the maximum earnable
  merit total.** Objective freeze follows the REQ-MB-005 snapshot rule;
  `total_merits = Σ (merit_per_unit + bonus) × unit_limit +
  completion_reward`.

## Non-goals

- No change to crediting rules (REQ-MCF-COMPLETION-REWARD untouched).
- No breaking change to existing offer fields; the delta is additive.
- Legacy offers without stored `Objectives` serialize an empty list and a
  `total_merits` of `completion_reward`; the local board is re-seeded.

## Rollback

- The delta is additive; reverting removes fields and keeps stored items
  valid (extra DynamoDB attributes are ignored on decode).
