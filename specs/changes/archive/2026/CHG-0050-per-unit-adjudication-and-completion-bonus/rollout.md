---
change_id: CHG-0050
artifact: rollout
status: pending
updated: 2026-10-20
---

# Rollout — Per-Unit Adjudication And Completion Bonus

## Plan

1. Deploy the service change — `approve` accepts optional
   `rejected_units`; empty body keeps approve-all behavior, so current
   clients and any in-flight `PENDING_VALIDATION` contracts remain
   adjudicable.
2. Run the catalog seed (`pvp-missions.json`, `racing-missions.json`)
   per environment — publishes version `3` definitions and refreshed
   offers; existing offers/contracts keep their frozen terms.
3. Deploy the web-app companion (change `173-per-unit-review-verdicts`)
   so reviewers see per-unit verdict controls.

## Compatibility

- Backward compatible API: `approve` without a body = approve all.
- Legacy contracts (two-objective terms) adjudicate fine — rejecting a
  staged unit just works; their `unit_limit: 1` objectives make the
  tightened completion predicate equivalent.
- No data migration: `RejectedUnits` decodes as absent on old items.

## Rollback

- Revert the service deploy; `approve` still accepts the old no-body
  calls. Contracts already completed keep written credits (append-only
  ledger).
- Catalog rollback: publish corrected content as a new version
  (`"4"`); never mutate immutable versions.
