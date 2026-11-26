---
change_id: CHG-0021
title: Complete the solo contract journey with frozen terms and atomic scoring
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/2026/CHG-0021-solo-contract-journey-frozen-terms-atomic-scoring/
execution_mode: sequential
depends_on:
  - CHG-0020
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_FIX_BUG_WORKFLOW
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-08
updated: 2026-10-08
---

# Complete the Solo Contract Journey With Frozen Terms and Atomic Scoring

## Intent

A member can complete the full solo mission journey through the public API —
accept, start, confirm objectives, finish or abandon — while the contract
guarantees the objectives and rewards it promised at acceptance and credits
merits atomically so a failed write can never silently lose earnings.

## Business Value

- The primary player journey works end-to-end; contracts naturally finish and
  release active slots.
- What a player agreed to is what gets scored; no arbitrary merit amounts.
- Earned merits are never lost to partial writes, and retries are safe.

## User Stories

- **US-MCF-JOURNEY**: As a member, I want to start and finish (or abandon) my accepted contract through the API, so my active slots are released and the mission completes.
- **US-MCF-FROZEN**: As a member, I want my contract to keep the exact objectives and rewards published when I accepted it, so later edits cannot change my deal.
- **US-MCF-ATOMIC**: As a platform owner, I want confirmation and merit credit to commit together, so retries after a failure can never drop earned merits or double-credit.

## Why L3

Data-loss risk on merit crediting, idempotency on retries, a critical
end-to-end journey, and contract-integrity (frozen terms) all trigger L3 gates.

## Defects

- **BUG-0021-01 — Journey has no transitions**: acceptance creates `ACCEPTED` while confirmation requires `IN_PROGRESS`; no start/complete/abandon/expire is exposed (`composition/server.go:63`, `domain/mission_contract.go:67`). Tests bypass by starting contracts manually.
- **BUG-0021-02 — Frozen terms are not preserved**: definitions carry no objectives or reward rules; confirmation loads a global objective and only compares category. Reproduced: 999,999 merits credited via an unrelated same-category objective (`domain/mission_definition.go:23`, `application/confirm_objective.go:62`).
- **BUG-0021-03 — Merit loss on partial write**: contract saves `CONFIRMED` before the ledger credit; a ledger failure followed by retry returns duplicate while the standing stays zero (`application/confirm_objective.go:72`).
- **BUG-0021-04 — Bonus and unit limit dropped**: DynamoDB persistence discards `Bonus` and `UnitLimit` (`infrastructure/dynamodb/objective_store.go:30`).
- **BUG-0021-05 — Confirmation identity cannot parse generated ids**: contract ids contain colons while `ConfirmationID` expects exactly four colon-separated parts, so prior confirmations are unrecognized (`domain/confirmation_id.go:30`).
- **BUG-0021-06 — Re-acceptance overwrites history**: contract id `tenant:offer:player` is reused after abandoning, overwriting the previous attempt (`application/accept_mission.go:67`).

## Scope

- **Domain**: contract lifecycle transitions (`Start`, `Complete`, `Abandon`,
  `Expire`); `ContractTerms` snapshot (objective definitions + reward rules at
  acceptance); unique attempt identity; `ConfirmationID` with a safe delimiter
  and unambiguous parse.
- **Application**: lifecycle use cases; acceptance captures `ContractTerms`
  from the offer definition; confirmation executes against the snapshot and
  commits contract state + `MeritCredit` as one unit.
- **Boundary**: `POST /contracts/{id}/start|complete|abandon` (expire derived);
  contract responses expose snapshot terms and attempt identity.
- **Infrastructure**: persist `ContractTerms`, `Bonus`, `UnitLimit`; atomic
  confirm+credit via DynamoDB `TransactWriteItems`; unit-limit counter keyed by
  the corrected `ConfirmationID`.
- **Composition**: route wiring.

## Non-Goals

- Sequential phases and objective-unlock mechanics beyond one ordered phase set.
- Team participation semantics (CHG-0023).
- Cooldown/repetition policies and tenant-configurable limits (roadmap).
- Award tiers and Achievements delivery at finalization (roadmap).

## Sequencing

Second in the approved set; requires CHG-0020 consolidated first (identity is a
precondition for roster/ownership checks).

## Approvals

- Gate 1: approved 2026-10-08.
