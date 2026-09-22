---
change_id: CHG-0002
title: Confirm a bounty objective and credit seasonal merit
status: archived
risk: L3
canonical_action: create
change_action: create_delta
change_path: specs/changes/archive/2026/CHG-0002-confirm-bounty-objective-and-credit-merit/
archive_path: specs/changes/archive/2026/CHG-0002-confirm-bounty-objective-and-credit-merit/
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_UPDATE_DOCUMENTATION_WORKFLOW
context_scope: existing
created: 2026-08-27
updated: 2026-08-29
---

# Confirm A Bounty Objective And Credit Seasonal Merit

## Intent

Model the confirmation of a verifiable objective in an active contract and the credit of Season Merits in the corresponding category and track, using the concrete example of a PvP Bounty mission.

## Business Value

- Produces the measurable result of a mission: competitive merits.
- Connect the contract to the season without permanent XP or global ranking.
- Establishes the idempotent basis for subsequent trusted sources and reviews.

## User Story

- **US-MCF-001**: As a solo or team participant, I want to complete verifiable objectives, to obtain fair merits in my discipline.

## Why L3

The change affects competitive adjudication, merit ledger, concurrent confirmations, multi-tenant isolation and idempotence. A defect can duplicate scores, cross categories or reward the wrong participant.

## Scope

- Confirmation of target unit in an active contract.
- Validation of contract, window, objective, receiver, source and uniqueness.
- Merit credit to the correct tenant, season, category, mode and recipient.
- Rejection of duplicates and credits outside `ACTIVE`.
- Example `BOUNTY_HUNTING`: 100 merits per single target.
- Initial views of the ledger and standing as auditable projections.

## Non-Goals

- Generic objective rules engine.
- Synchronous integration with Achievements.
- Disputes, manual reviews or corrections (remain in CHG-0001).
- Procedural generation of contracts.

## Acceptance Summary

- A valid confirmation confirms exactly the published merits only once.
- The same unit, objective, contract and recipient does not generate a second credit.
- The credit belongs to the correct tenant, season, category and track.
- A `FREEZING` or `FINALIZED` season rejects new credits except eligible work started on time.
- The individual standing shows the total derived from the ledger.
- A team credits the team track, not each member.

## Current Outcome

Gate 1 approved on August 27, 2026. It is derived from `CHG-0001-initialize-missions-platform`, which already defined `MissionContract`, `MissionOffer`, acceptance and active limit. Gate 2 and implementation are pending.
