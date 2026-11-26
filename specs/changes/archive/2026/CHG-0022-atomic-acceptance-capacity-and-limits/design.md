---
change_id: CHG-0022
artifact: design
updated: 2026-10-08
---

# Design — Atomic Acceptance Capacity and Limits

## Canonical Layer Scope

Domain → Application → Infrastructure → Interface → Composition.

## Domain

- `DecideAcceptance` gains season-state input (`SeasonState` of the offer's
  bound season) and evaluates limit against the total active set.
- Effective-limit resolution stays in domain policy; no new aggregates.

## Application

- `AcceptMissionHandler` calls a single `AtomicAcceptanceWriter` port: writes
  contract + offer-capacity consumption + player-active counter under
  conditions that fail if capacity or limit would be exceeded.
- Season state is loaded within the same tenant scope (CHG-0020).

## Infrastructure

- `ContractStore`: `TransactWriteItems` — `Put` contract item,
  `Update` offer item decrementing `remaining_capacity` with
  `ConditionExpression remaining_capacity > 0`, `Update` player counter item
  with `ConditionExpression active_count < :limit`.
- Conditional-check failure maps to typed rejection reasons; no partial writes.

## Boundary — HTTP

- Rejection codes `OFFER_CAPACITY_EXHAUSTED`, `ACTIVE_LIMIT_REACHED`,
  `SEASON_NOT_ACTIVE` — `422`, stable vocabulary added to `errors.go`.

## Test Strategy

- Domain: decision matrix for limit/capacity/season-state partitions.
- Application: double that records the single write call; reason mapping.
- Infrastructure: real concurrent writes against DynamoDB local — prove
  exactly-once under contention (goroutines, `t.Parallel` style harness).
- HTTP: concurrent accept E2E for both capacity and limit.

## Risk And Rollback

- **Risk**: counter items must exist/be seeded — transaction uses `Update` with
  `if_not_exists` bootstrap for the player counter.
- **Rollback**: revert; counters become unused items (inert).
