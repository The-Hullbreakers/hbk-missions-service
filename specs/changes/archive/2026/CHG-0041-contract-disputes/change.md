---
change_id: CHG-0041
title: Contract dispute lifecycle
status: completed
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0041-contract-disputes/
execution_mode: sequential
depends_on: []
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
# Contract dispute lifecycle

## Intent

`DISPUTES_PENDING` exists in the error vocabulary but no dispute model exists. This change adds the dispute entity, its lifecycle, and the guards that block leave/abandon/adjudication while disputes are pending.

## Business Value

- Contested results get a review path; members cannot dodge moderation by leaving mid-dispute.

## User Stories

- **US-MCF-DISPUTE**: As a roster member, I want to dispute a rejection, so contested decisions get reviewed.
- **US-MCF-DISPUTE-REVIEWER**: As a moderator, I want disputes to block leave/abandon, so adjudication completes before exit.

## Why L3

New domain state machine interacting with contract lifecycle (leave/abandon/review) — adjudication integrity is trust-critical.

## Scope

- **Domain**: `Dispute` (id, contract, raiser, reason, state OPEN/RESOLVED); guards in leave/abandon.
- **Application**: raise/resolve dispute commands; dispute list for reviewers.
- **Infrastructure**: dispute persistence on the contract item or separate partition.
- **Boundary**: POST dispute, GET disputes (reviewer), resolve endpoint.

## Non-Goals

- Dispute escalation workflows; automatic re-adjudication.

## Approvals

- Gate 1: approved 2026-10-17 — D1 `Dispute{DisputeID, ContractID, ObjectiveID, RaisedBy, Reason, State: OPEN|RESOLVED}` **targets a specific rejected credit by objective** (decision B); D2 `Leave`/`Abandon` → `ErrDisputesPending` (422 `DISPUTES_PENDING`) while any OPEN dispute; `ErrDisputeStateInvalid` (422 `DISPUTE_STATE_INVALID`) on re-resolve; D3 `SK` children `DISPUTE#{id}` under contract `PK`; D4 `POST .../disputes`, `GET .../disputes` (reviewer), `POST .../disputes/{dispute_id}/resolve` (moderator); raisable by roster member only while contract is DISQUALIFIED.
