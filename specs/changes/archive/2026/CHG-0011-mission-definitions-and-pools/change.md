---
change_id: CHG-0011
title: Add mission definitions and season pools
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0011-mission-definitions-and-pools/
archive_path: specs/changes/archive/2026/CHG-0011-mission-definitions-and-pools/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CODE_QUALITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
context_scope: existing
context_snapshot_gate: not_applicable
context_snapshot_reason: The repository does not maintain a local AI context snapshot convention; durable context is provided by the global common rules.
verified_at: 2026-08-30
source_revision: uncommitted-working-tree-2026-08-30
created: 2026-08-30
updated: 2026-08-30
---

# Add Mission Definitions And Season Pools

## Intent

Move mission publishing from an ad-hoc, raw-payload model to a reusable, template-driven model. `MissionDefinition` is a versioned template that can be used across seasons. `MissionPool` groups approved definitions for a tenant, category, and season. An operator manually publishes an `MissionOffer` from a pool, bound to the active season.

## Business Value

- Reusable mission templates reduce duplicated operator work across seasons.
- Season pools make competitive content explicit and auditable.
- Offer publication derives from approved definitions, eliminating arbitrary raw offers.
- Contracts preserve the exact definition version used at acceptance.
- No dead or unused code paths remain from the previous manual offer endpoint.

## User Stories

- **US-MB-011**: As an operator, I want to create a mission definition, so it can be reused in multiple season pools.
- **US-MB-012**: As an operator, I want to publish a definition version, so only immutable, approved templates enter pools.
- **US-MB-013**: As an operator, I want to create a mission pool for a tenant, category, and active season, so the board has a curated set of offers.
- **US-MB-014**: As an operator, I want to manually publish an offer from a pool, so members can see and accept a bound mission.
- **US-MCF-011**: As a member, I want my accepted contract to keep the definition version from the offer, so my rewards do not change after acceptance.

## Why L3

This change introduces new domain entities (`MissionDefinition`, `MissionPool`), alters the public contract of `POST /api/v1/missions/offers`, adds persistence, changes acceptance rules (active-template exclusion), and affects contract fulfillment semantics. It therefore requires security review, coverage, mutation testing, and critical E2E evidence.

## Scope

- **Domain**: add `MissionDefinition` and `MissionPool`; add `DefinitionID`/`DefinitionVersion` to `MissionOffer` and `MissionContract`; update `DecideAcceptance` to enforce active-template exclusion by stable definition identity; implement `OfferPublication` service to generate an offer from a pool/definition.
- **Application**: `CreateMissionDefinitionHandler`, `PublishMissionDefinitionHandler`, `CreateMissionPoolHandler`, `PublishMissionFromPoolHandler` (replaces the raw `PublishMissionHandler`).
- **Boundary**: `POST /api/v1/missions/definitions`, `POST /api/v1/missions/definitions/{definition_id}/publish`, `POST /api/v1/missions/pools`, `POST /api/v1/missions/offers` (new contract with `definition_id`/`pool_id`), `GET /api/v1/missions/pools`.
- **Infrastructure**: `DefinitionStore`, `PoolStore`; extend `OfferStore` and `ContractStore` for `DefinitionID`/`DefinitionVersion`; backward-compatible load for existing items without these fields.
- **Composition**: wire new stores, handlers, and routes in `NewServer`.
- **Tests**: unit tests for domain and application, integration tests for HTTP, composition, and DynamoDB.

## Non-Goals

- Scheduled or rotating pool publication; only manual publication in this change.
- Procedural generation of objectives or rewards.
- Full objective/merit rule engine.
- Cross-tenant or multi-category pools.
- Soft-delete or rollback of published definitions beyond `RETIRED` state.
- Migrating legacy offers to real definitions; they use a backward-compatible default.

## Acceptance Summary

- A definition can be created, published as an immutable version, and later retired.
- A pool can be created for a tenant, category, and active season with approved definitions.
- `POST /api/v1/missions/offers` accepts `definition_id` (and optional `pool_id`) and creates an offer bound to the active season.
- The generated `MissionOffer` carries the `definition_id` and `definition_version` of the source.
- `MissionContract` inherits and preserves `definition_id`/`definition_version`.
- A player with an active contract for a definition does not see or accept other offers with the same `definition_id`.
- Existing offers and contracts without `DefinitionID` load with a default value and still function.
- `go test -race ./...`, `gofmt`, and `go vet` pass.

## Current Outcome

- Gate 1 approved on 2026-08-30.
