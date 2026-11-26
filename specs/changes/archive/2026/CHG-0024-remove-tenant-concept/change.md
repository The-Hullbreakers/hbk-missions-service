---
change_id: CHG-0024
title: Remove tenant concept
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/archive/2026/CHG-0024-remove-tenant-concept/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_AWS_LAMBDA_REST_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-08
updated: 2026-10-08
---

# Remove Tenant Concept

## Intent

Remove `tenant_id` and every tenant-scoping rule from the service because the
HBK Identity server does not emit a `tenant_id` claim and the platform operates
a single mission scope. Player identity, roles, and team membership continue to
come exclusively from the authenticated OIDC session.

## Business Value

- Unblocks authentication: today every authenticated request fails
  `NewCallerIdentity` because the session carries no `tenant_id` claim.
- Removes an isolation model the product does not operate; there is exactly one
  player population, one catalog, and one competitive space.

## User Stories

- **US-REM-TENANT-AUTH**: As a platform owner, I want caller identity built only from claims the identity server actually emits (`player_id`, roles, `team_ids`), so authenticated requests succeed.
- **US-REM-TENANT-DATA**: As a platform owner, I want all missions, seasons, contracts, and standings stored in a single scope, so there is no dead per-tenant partitioning.
- **US-REM-TENANT-CONTRACT**: As an API consumer, I want a public contract with no `tenant_id` fields or tenant filters, so the surface matches what the identity server supports.

## Why L3

The change modifies identity/authorization semantics and the public contract
across all three capabilities, rewrites storage keys, and removes a
trust-boundary mechanism introduced by CHG-0020. Requires security gate,
mutation gate, critical E2E, coverage >= 90%, and independent review.

## Defects

- **BUG-0024-01 — Identity claim the issuer never emits**: `CallerIdentityFromContext` requires `claims["tenant_id"]`; HBK Identity does not emit it, so every authenticated call resolves an invalid caller and fails. Reproduced by any authenticated request through `AuthMiddleware`.

## Scope

- **Domain**: `CallerIdentity` drops `tenantID`; `NewSystemIdentity` becomes parameterless scope for the scheduler; tenant removed from `MeritCredit`, `Standing`, `Season`, contract/definition/offer fields, and the `tenant:offer:player:attempt` contract-id format.
- **Application**: all handlers stop resolving or forwarding tenant; season uniqueness and `ACTIVE_SEASON_EXISTS` become global; leaderboard, contract, offer, and definition queries operate in one scope.
- **Infrastructure**: DynamoDB `PK`/`SK`/`GSI1` keys drop the tenant segment for seasons, definitions, offers, contracts, active-set markers, and the merit ledger.
- **Interface**: `CallerIdentityFromContext` no longer reads `tenant_id`; `tenantFromContext` removed; body `tenant_id` ignored as today (forgiving derive decision from CHG-0020 preserved for remaining identity fields).
- **Composition**: wiring only.
- **Docs**: `AUTHENTICATION-AND-MULTITENANCY.md`, `API-GUIDE-FRONTEND.md`, `API-CONTRACT.md` lose tenant claims, filters, and scenarios.

## Non-Goals

- Multi-tenant replacement (sharding, org id, workspace id) — none; scope is singular.
- Identity provider changes; the session contract is whatever HBK Identity emits minus `tenant_id`.
- Migration/backfill of existing tenant-prefixed DynamoDB items — operational follow-up recorded in `rollout.md`.
- Weakening roster/authority checks: `admin`/`moderator` and roster membership still gate evidence and lifecycle transitions.

## Sequencing

Independent of CHG-0014, CHG-0017, and CHG-0023, but CHG-0023 tasks that touch
team identity must rebase onto the tenant-free `CallerIdentity` if this lands
first.

## Approvals

- Gate 1: approved 2026-10-08.
