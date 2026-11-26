---
change_id: CHG-0032
title: List pending-review contracts for adjudicators
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0032-review-queue/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
context_scope: existing
created: 2026-10-12
updated: 2026-10-12
---

# Change — Review Queue for Pending Contracts

## Intent

CHG-0031 gated completion behind reviewer adjudication but shipped no way
for an `admin`/`moderator` to discover pending work — the security review
recorded this as a residual: reviewers must learn contract ids from
another channel. This change adds the read-only review queue: a
`GET /api/v1/missions/review-queue` endpoint returning every
`PENDING_VALIDATION` contract in the deployment scope.

## Why a New Use Case

`GET /api/v1/missions/contracts` is member-facing and roster-scoped —
`ListByPlayer` filters `IncludesPlayer(caller)`, so an adjudicator who is
not on the roster sees nothing. Reviewer discovery is a different
business intent with a different authorization boundary; it gets its own
handler and port rather than widening the member listing.

## Scope

- `GET /api/v1/missions/review-queue` behind `adminOrModerator`
  middleware, additionally enforced by the `isReviewer` application
  check.
- `PendingReviewLister` query port (one CQRS operation) and
  `ContractStore.ListPendingReview()` — Query `PK=CONTRACT` with a state
  `FilterExpression`; no new GSI, no persisted-shape change.
- Response items mirror the member contract summary shape plus the
  staged-credit count, so reviewers can size the adjudication work.

## Risk Classification — L2

- Read-only query on an existing trust boundary (`admin`/`moderator`).
- No mutation, no new GSI, no migration, no domain-rule change.
- Rollback: remove the route + port; nothing persisted depends on it.

## Tenant Note

Tenancy is a deployment/storage boundary in this service — no tenant
claim or persisted attribute exists. The queue returns all pending
contracts in the instance, consistent with `ListDefinitions` and every
other admin-scoped listing.
