---
change_id: CHG-0002
artifact: design
status: proposed
updated: 2026-08-27
layer_scope:
  domain: affected
  application: affected
  boundary: affected
  infrastructure: affected
  interface: affected
  composition: affected
---

# Business-Centered Design

## Bounded Context

`MissionContract` has objectives, confirmations and results. `Competition` has season, category, track and merit ledger. `Identity` and `PartyManager` remain outside the domain.

## Aggregates

- `MissionContract`: acceptance, mode, frozen roster, phases, objectives and results.
- `Objective`: published definition of a scoring unit within a phase.
- `Confirmation`: attempt to credit a specific unit.
- `MeritCredit`: immutable credit in the ledger.
- `Season`: category window and policy.

## Taxonomy

`MissionCategory` determines the competitive bucket; `Track` is `INDIVIDUAL` or `TEAM`. For `SOLO` the receiver is the player; for `TEAM`, it is the team identifier.

## Objective And Confirmation

- Each `Objective` declares merit per unit, optional bonus and unit limit.
- A `Confirmation` identifies contract, target, receiver, unit and source.
- The confirmation is valid if: contract `IN_PROGRESS`, objective belongs to active phase, unit not repeated, season `ACTIVE`, recipient with right.
- A credit is generated with idempotent identity derived from `contract_id`, `objective_id`, `unit_id`, `recipient_id` and `track`.

## Merit Ledger

- Each credit is an immutable fact: tenant, season, category, track, recipient, amount, confirmation_id, timestamp.
- The recipient's total in a category is the sum of credits and adjustments.
- Standing is projection, not a second source of truth.

## Season State Guard

- `ACTIVE`: accept new credits.
- `FREEZING`: only accepts credits for work started on time.
- `FINALIZED`/`ARCHIVED`: reject all new credit.

## Development Sequence

1. Domain: `Objective`, `Confirmation`, `MeritCredit`, `Season`, `Track`, `MeritLedger`.
2. Application: use case `ConfirmObjectiveAndCreditMerit` with ports CQRS.
3. Boundary RED: endpoint `POST /api/v1/missions/contracts/{contract_id}/evidence`.
4. Infrastructure, Interface and Composition with DI owned by the module.
5. Security, concurrency, mutation, coverage, E2E, rollout and review.

## Rollback Principle

Rollback preserves existing ledger and disables acceptance of new confirmations through configuration. It never deletes credits or standing.
