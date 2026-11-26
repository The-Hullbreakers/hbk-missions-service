---
change_id: CHG-0042
title: Asynchronous team roster invitation flow
status: completed
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0042-team-roster-invitation-flow/
execution_mode: sequential
depends_on: 
  - CHG-0034
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
# Asynchronous team roster invitation flow

## Intent

Acceptance currently validates declared teammates synchronously — the spec's invitees-confirm then roster-freezes flow is a non-goal until now. This change adds the invitation lifecycle: invitees must confirm; the roster freezes when all confirm or the window lapses.

## Business Value

- Team contracts form by consent: teammates opt in rather than being committed silently.

## User Stories

- **US-MCF-INVITE**: As a teammate, I want to confirm a team mission invitation, so I only join contracts I agreed to.
- **US-MCF-FREEZE**: As a platform owner, I want the roster to freeze only after confirmations, so consent is enforced.

## Why L3

New async roster lifecycle: pending member states, expiry, and interplay with active-limit consumption and abandonment.

## Scope

- **Domain**: member confirmation state (PENDING/CONFIRMED/DECLINED/EXPIRED); roster freeze rule.
- **Application**: invite-at-accept produces pending members; confirm/decline commands; expiry sweep (scheduler?).
- **Infrastructure**: invitation state and `invited_at` persist inside the contract item; slots are reserved at invite time and released on decline/expiry (decision approved 2026-10-16).
- **Boundary**: confirm/decline endpoints; contract detail shows pending members.

## Non-Goals

- Notification delivery to invitees (transport concern).
- Replacing synchronous teammate validation (CHG-0034 remains the membership check).

## Approvals

- Gate 1: approved 2026-10-16 — design resolutions confirmed (reserve-at-invite, lazy 7-day expiry, start-without-declined) and four-task plan accepted.
