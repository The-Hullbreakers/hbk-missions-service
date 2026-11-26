---
change_id: CHG-0020
title: Derive caller identity and isolate tenant state
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/2026/CHG-0020-derive-caller-identity-and-tenant-isolation/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_FIX_BUG_WORKFLOW
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_AWS_LAMBDA_REST_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
archive_path: specs/changes/archive/2026/CHG-0020-derive-caller-identity-and-tenant-isolation/
verified_at: 2026-10-08
source_revision: working-tree (uncommitted, reviewed diff)
created: 2026-10-08
updated: 2026-10-08
---

# Derive Caller Identity and Isolate Tenant State

## Intent

Make the authenticated OIDC session the only source of caller identity
(`tenant_id`, `player_id`, roles, team membership) and scope all stored state to
the caller tenant, so no request can act on behalf of another player or tenant.

## Business Value

- Closes the cross-tenant and impersonation path reproduced in review: a member
  of tenant A could create a contract for another player in tenant B and credit
  it into tenant A's leaderboard.
- Restores the constitutional invariant "all state and standings are isolated
  by tenant" at every boundary and in storage.

## User Stories

- **US-SEC-IDENTITY**: As a platform owner, I want tenant and player identity taken only from the authenticated session, so no caller can act for another tenant or player.
- **US-SEC-EVIDENCE**: As a platform owner, I want evidence submissions checked against the caller's authority over the contract, so nobody credits another player's objectives.
- **US-SEC-TENANT-DATA**: As a platform owner, I want definitions and seasons stored per tenant, so one tenant cannot read or rewrite another tenant's catalog.

## Why L3

Tenant isolation and authorization are explicit L3 examples, this service's
constitution is classified L3, and the reproduced exploit crosses the public
trust boundary with leaderboard impact. Requires security gate, mutation gate,
critical E2E, coverage ≥90%, and independent review.

## Defects

- **BUG-0020-01 — Acceptance trusts body identity**: `internal/missions/interface/http/accept_mission.go` reads `tenant_id`/`player_id`/`team_id` from the request body. Reproduced: tenant-A caller created a contract for a tenant-B player.
- **BUG-0020-02 — Evidence trusts body identity and skips authority**: `internal/missions/interface/http/handler.go` (`POST .../evidence`) accepts `tenant_id`/`player_id` from the body and does not check roster membership or reviewer authority. Reproduced: credit forged for another player's contract.
- **BUG-0020-03 — Definition reads ignore tenant**: `internal/missions/infrastructure/dynamodb/definition_store.go` reads definitions without tenant identity, exposing another tenant's catalog.
- **BUG-0020-04 — Seasons stored globally**: season persistence and the single-`ACTIVE` invariant are global, so one tenant can create, freeze, or overwrite seasons for all tenants.

## Scope

- **Domain**: `CallerIdentity` value object (tenant, player, roles, team memberships); roster-membership rule for evidence authority.
- **Application**: accept, confirm-evidence, and season handlers consume `CallerIdentity`; handlers reject work the caller is not entitled to.
- **Infrastructure**: DynamoDB keys for definitions and seasons carry tenant; single-active-season invariant enforced per tenant.
- **Interface**: `AuthMiddleware` exposes `CallerIdentity`; handlers ignore body `tenant_id`/`player_id` (approved decision: derive silently, keep contract forgiving).
- **Composition**: wiring changes only.

## Non-Goals

- New roles, permissions, or reviewer workflows beyond roster membership plus existing `admin`/`moderator` authority.
- Identity provider changes; the OIDC session contract stays as-is.
- Backfill of historical miscredited ledgers (handled as operational follow-up in `rollout.md`).
- Strict `400` rejection when body identity conflicts with session identity (forgiving derive was approved; strict mode is a documented option).

## Sequencing

Runs first of the approved review-remediation set. CHG-0021, CHG-0022, and
CHG-0023 wait until this change is consolidated.

## Approvals

- Gate 1: approved 2026-10-08 (all four remediation changes; forgiving
  derive-and-ignore contract decision approved).
