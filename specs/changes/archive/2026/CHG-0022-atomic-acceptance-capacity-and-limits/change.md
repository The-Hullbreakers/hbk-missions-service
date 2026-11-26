---
change_id: CHG-0022
title: Enforce acceptance capacity, player limits, and season eligibility atomically
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/active/2026/CHG-0022-atomic-acceptance-capacity-and-limits/
execution_mode: sequential
depends_on:
  - CHG-0021
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

# Enforce Acceptance Capacity, Player Limits, and Season Eligibility Atomically

## Intent

Offer capacity and per-player active-contract limits are enforced inside the
same atomic write that creates the contract, and acceptance is refused for
offers whose season is no longer `ACTIVE` — so no race or stale state can
oversell capacity or exceed a player's limit.

## Business Value

- Capacity-one offers can never be sold twice.
- A player's concurrent acceptances can never exceed their configured limit.
- The effective limit applies to the player's total active contracts, not
  per-category slices.
- Contracts are only created against a live season.

## User Stories

- **US-MB-ATOMIC**: As an operator, I want capacity and player-limit checks committed atomically with the contract, so concurrent acceptances cannot oversell an offer or exceed a player's limit.
- **US-MB-SEASON**: As an operator, I want acceptance rejected when the offer's season is finalized, so no new contracts attach to a closed competitive window.

## Why L3

Concurrency/idempotency on a competitive resource with leaderboard impact;
reproduced races violate published invariants (`REQ-MB-003`, `REQ-MB-004`).

## Defects

- **BUG-0022-01 — Capacity race**: two players accepted a capacity-one offer; check and insert are not atomic (`application/accept_mission.go:46`, `infrastructure/dynamodb/contract_store.go:89`).
- **BUG-0022-02 — Limit race**: two concurrent acceptances by one player both passed the limit check and exceeded it.
- **BUG-0022-03 — Limit counted per category**: a player exceeded the tenant default limit by accepting missions spread across categories; the effective limit was not applied to the total.
- **BUG-0022-04 — Finalized-season acceptance**: offers from finalized seasons still accepted contracts.

## Scope

- **Domain**: `DecideAcceptance` extended with season-state eligibility.
- **Application**: acceptance performs check+write through one conditional
  persistence operation (capacity counter + limit counter + contract insert).
- **Infrastructure**: DynamoDB `TransactWriteItems`/`ConditionExpression`
  enforcing offer capacity and player limit counters in one transaction.
- **Boundary**: stable rejection codes for each failure reason.

## Non-Goals

- Tenant-configurable limit administration API (roadmap; default + existing
  category overrides remain).
- Team roster capacity checks (CHG-0023).
- Retry/backoff UX on contention.

## Sequencing

Third in the approved set; builds on CHG-0021 attempt identity and CHG-0020
identity derivation.

## Approvals

- Gate 1: approved 2026-10-08.
