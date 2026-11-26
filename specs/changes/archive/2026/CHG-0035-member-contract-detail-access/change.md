---
change_id: CHG-0035
title: Allow roster members to read their contract detail
status: completed
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0035-member-contract-detail-access/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
context_scope: existing
created: 2026-10-14
updated: 2026-10-14
---
# Allow roster members to read their contract detail

## Intent

Contract detail is admin/moderator only. Roster members cannot read the terms, roster, or evidence references of their own contract. This change opens a member-scoped read with data minimization.

## Business Value

- Participants can inspect their own contract terms and progress without reviewer privileges.

## User Stories

- **US-MCF-DETAIL**: As a roster member, I want to read my contract detail, so I can verify terms and track evidence state.

## Why L2

Authorization scope change and data exposure surface — member view must not leak fields intended for reviewers.

## Scope

- **Application**: `GetContractHandler` gains caller-vs-roster authorization.
- **Boundary**: `GET /api/v1/missions/contracts/{contract_id}` admits roster members; admin/moderator retains full view.
- **Data minimization**: member projection omits reviewer-only fields if any (evidence media stays referenced, not embedded).

## Non-Goals

- Changing the admin/moderator detail contract.

## Approvals

- Gate 1: approved 2026-10-17 — D1 load-then-authorize: reviewer or active roster member, else `403 FORBIDDEN` (no anti-enumeration change); D2 member projection = same detail but `pending_credits` filtered to caller + `media_uri` omitted; D3 drop `adminOrModerator` route wrapper, auth in handler; D4 departed members lose read access.
