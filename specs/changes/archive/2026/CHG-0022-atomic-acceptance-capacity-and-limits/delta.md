---
change_id: CHG-0022
artifact: delta
status: consolidated
updated: 2026-10-08
---

# Delta — Atomic Acceptance Capacity and Limits

## What Is Changing

### Mission Board

- Acceptance validates capacity, active-template exclusion, effective limit
  (default + category override), and bound-season state, then commits the
  contract and the capacity/limit counters in one conditional transaction.
- The effective limit applies to the player's total active contracts across
  all categories (category overrides raise/lower the limit for that category's
  contracts but the check is against the whole active set).
- Offers whose bound season is not `ACTIVE` cannot be accepted.

## MODIFIED

### REQ-MB-003 — Limit active contracts

The limit check evaluates the player's total `ACCEPTED`, `IN_PROGRESS`, and
`PENDING_VALIDATION` contracts regardless of category, using the effective
limit resolved from tenant default and category override.

### REQ-MB-004 — Consistent acceptance

The eligibility decision and the contract write are one atomic operation.
Concurrent acceptances serialize: the first commits, the rest receive the
explainable rejection (`OFFER_CAPACITY_EXHAUSTED`, `ACTIVE_LIMIT_REACHED`, or
`SEASON_NOT_ACTIVE`).

## ADDED

### REQ-MB-014 — Accept only within a live season

Acceptance is rejected when the offer's bound season is not `ACTIVE`
(`FREEZING`, `FINALIZED`, `ARCHIVED`), with code `SEASON_NOT_ACTIVE` and no
contract or capacity consumption.

## BDD Scenarios

### SCN-MB-ATOMIC-CAPACITY — Protect the last offer slot

Given an offer with one remaining slot, when two acceptances execute
concurrently, then exactly one contract exists and the other request receives
`OFFER_CAPACITY_EXHAUSTED`.

### SCN-MB-ATOMIC-LIMIT — Protect the last personal slot

Given `player-1` is one below their effective limit, when they submit two
concurrent acceptances, then exactly one succeeds and the other receives
`ACTIVE_LIMIT_REACHED`.

### SCN-MB-LIMIT-TOTAL — Count the total across categories

Given a default limit of 3 and `player-1` holding one active contract in each
of three categories, when they accept a fourth offer in any category, then the
acceptance is rejected with `ACTIVE_LIMIT_REACHED`.

### SCN-MB-SEASON-CLOSED — Reject acceptance on finalized season

Given an offer bound to a `FINALIZED` season, when a member accepts it, then
the response is `422` with `SEASON_NOT_ACTIVE` and capacity is unchanged.

## Error Vocabulary

- `OFFER_CAPACITY_EXHAUSTED` — `422`
- `ACTIVE_LIMIT_REACHED` — `422`
- `SEASON_NOT_ACTIVE` — `422` (existing, reused)
- `TEMPLATE_ALREADY_ACTIVE` — `422` (existing REQ-MB-008, unchanged)
