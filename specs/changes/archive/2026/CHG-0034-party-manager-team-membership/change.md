---
change_id: CHG-0034
title: Integrate Party Manager team membership
status: consolidated
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0034-party-manager-team-membership/
execution_mode: sequential
depends_on: 
  - CHG-0023
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-14
updated: 2026-10-14
---
# Integrate Party Manager team membership

## Intent

Team acceptance currently rejects every declared teammate because composition wires DenyAllReader. This change integrates the authoritative team roster (Party Manager) so real teammates can be validated — failing closed when the source is unavailable.

## Business Value

- TEAM/FLEX-as-team missions become playable end to end for the first time.
- Phantom teammates remain impossible: membership is verified against the authoritative source, not trusted from the request.

## User Stories

- **US-MCF-MEMBERSHIP**: As a platform owner, I want teammate membership verified against Party Manager, so team contracts contain only real teammates.

## Why L3

Authorization-adjacent trust boundary and a new external dependency: wrong fail-open behavior lets any caller consume another player's active-contract slot.

## Scope

- **Application**: no signature change — `TeamMembershipReader` port already exists.
- **Infrastructure**: real `TeamMembershipReader` adapter against Party Manager.
- **Composition**: wire the real adapter; keep `DenyAllReader` available for tests.
- **Failure mode**: Party Manager unreachable → acceptance fails closed (`TEAMMATE_NOT_MEMBER` or a dedicated unavailability code).

## Non-Goals

- Invitation/confirm handshake (CHG-0042).
- Caching or replica reads of team rosters beyond one lookup per acceptance.

## Approvals

- Gate 1: approved 2026-10-15 — user confirmed create_delta and canonical modify of `mission-contract-fulfillment`.
- Gate 2: approved 2026-10-15 — RED start authorized.
- Final review: approved 2026-10-15 — consolidated and archived.
