---
change_id: CHG-0012
title: Simplify mission publishing by removing MissionPool
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/active/2026/CHG-0012-simplify-remove-mission-pool/
archive_path: specs/changes/archive/2026/CHG-0012-simplify-remove-mission-pool/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CODE_QUALITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
context_scope: existing
context_snapshot_gate: not_applicable
context_snapshot_reason: The repository does not maintain a local AI context snapshot convention; durable context is provided by the global common rules.
verified_at: 2026-09-01
source_revision: uncommitted-working-tree-2026-09-01
created: 2026-09-01
updated: 2026-09-01
---

# Simplify Mission Publishing By Removing MissionPool

## Intent

Remove the `MissionPool` abstraction from the mission board. Offer publication will happen directly from a published `MissionDefinition`, reducing the public contract, domain model, persistence, and tests to the essential flow.

## Business Value

- Removes an unused endpoint and store (`POST /api/v1/missions/pools`, `PoolStore`) whose only implemented mode was `MANUAL`.
- Simplifies the operator flow: create/publish a definition, then publish an offer directly.
- Reduces maintenance surface and aligns the code with the actual scope of CHG-0011.
- Keeps the active-template exclusion and contract-definition preservation untouched.

## User Stories

- **US-MB-015**: As an operator, I want to publish a mission offer directly from a published definition, so I do not need to create a pool.

## Why L2

This change modifies a public endpoint (`POST /api/v1/missions/offers`), removes a domain entity and a persistence adapter, and affects unit and HTTP integration tests. It is a non-trivial API and persistence change but does not touch authentication, authorization, payments, or critical end-to-end trust boundaries, so it is classified as L2.

## Scope

- **Domain**: remove `MissionPool`, `PoolEntry`, `MissionPublishMode`, `ErrPoolInvalid`, and `ErrPoolNotFound`.
- **Application**: remove `CreateMissionPoolHandler` and the `PoolReader`/`PoolWriter` ports; rename `PublishMissionFromPoolHandler` to `PublishMissionHandler` and remove `poolID` handling.
- **Boundary**: remove `POST /api/v1/missions/pools` and its HTTP handler; change `POST /api/v1/missions/offers` payload (no `pool_id`).
- **Infrastructure**: remove `PoolStore` and pool-related DynamoDB persistence.
- **Composition**: update `NewServer` wiring.
- **Tests**: remove domain, application, HTTP, and integration tests for pools; adjust existing tests to the simplified publish path.
- **Documentation**: update `docs/API-CONTRACT.md` and `specs/capabilities/mission-board/spec.md`.

## Non-Goals

- Add scheduled or rotating publication.
- Change the active-mission limit or active-template exclusion.
- Modify contract fulfillment semantics.
- Migrate existing pool data (no pools are in production).

## Acceptance Summary

- `MissionPool` and its artifacts no longer exist in the codebase.
- `POST /api/v1/missions/offers` accepts only `definition_id` (and optional `definition_version`) plus `capacity`.
- `POST /api/v1/missions/pools` is gone.
- `PublishMissionHandler` publishes from a `PUBLISHED` definition bound to the active season.
- All tests pass, coverage for affected domain/application/infrastructure remains >= 90%.
- `gofmt`, `go vet`, `go test -race ./...` are clean.

## Verification Record

- **validation_status**: passed
- **reviewer**: agent
- **verified_at**: 2026-09-01
- **source_revision**: uncommitted-working-tree-2026-09-01
- **canonical_paths**:
  - `specs/capabilities/mission-board/spec.md`
- **intended_archive_path**: `specs/changes/archive/2026/CHG-0012-simplify-remove-mission-pool/`
- **commands**:
  - `tools/validate-sdd-change.sh`
  - `go test -race ./...`
- **coverage**:
  - domain 95.1%
  - application 94.6%
  - dynamodb infrastructure 90.2%
- **residual_risk**: None. Scheduled/rotating publication was a non-goal; it can be added later through a new governed change.
