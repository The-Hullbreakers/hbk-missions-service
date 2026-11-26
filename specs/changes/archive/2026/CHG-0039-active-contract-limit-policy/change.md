---
change_id: CHG-0039
title: Manage the active-contract limit at runtime
status: completed
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0039-active-contract-limit-policy/
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
# Manage the active-contract limit at runtime

## Intent

The active-contract limit is hardcoded (5, with nil overrides) at composition. The API contract exposes `PUT /api/v1/missions/policies/active-contract-limit`. This change makes the limit a persisted, admin-managed policy read at acceptance time.

## Business Value

- Operations can tune active limits without redeploying.

## User Stories

- **US-MCF-POLICY**: As an operator, I want to update the active-contract limit at runtime, so balance tuning needs no deploy.

## Why L2

Admin write affecting acceptance behavior globally; policy read path changes for every acceptance.

## Scope

- **Domain**: policy value object stays; add per-category override map persistence decision.
- **Application**: `UpdateActiveLimitHandler` (admin) + `ActiveLimitPolicyReader` port consumed by acceptance.
- **Infrastructure**: persisted policy document.
- **Boundary**: `PUT /api/v1/missions/policies/active-contract-limit` (admin only).

## Non-Goals

- Per-player limits; audit log of policy changes (nice-to-have).

## Approvals

- Gate 1: approved 2026-10-17 — D1 `ActiveLimitPolicyReader`/`Writer` ports + `POLICY#active-contract-limit` DynamoDB document, handler keeps injected policy from the reader; D2 PUT body `{default, overrides?}` all-positive else `422 INVALID_POLICY`; D3 `adminOrModerator` only; D4 missing doc falls back to `NewActiveLimitPolicy(5, nil)`.
