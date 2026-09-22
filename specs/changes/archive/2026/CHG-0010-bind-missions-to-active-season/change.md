---
change_id: CHG-0010
title: Bind mission offers to the active season
status: consolidated
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0010-bind-missions-to-active-season/
archive_path: specs/changes/archive/2026/CHG-0010-bind-missions-to-active-season
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CODE_QUALITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
context_scope: existing
created: 2026-08-30
updated: 2026-08-30
---

# Bind Mission Offers To The Active Season

## Intent

Ensure every published mission offer is bound to the single active season, and that the resulting contract carries that season so confirmations cannot be credited against a different competitive window.

## Business Value

- Eliminates orphaned offers that have no competitive context.
- Prevents players or operators from accidentally crediting a contract to a season other than the one active when the offer was published.
- Makes pre-season/off-season analytics reliable by tying every offer, contract, and credit to a clearly identified season.

## User Stories

- **US-MB-010**: As an operator, I want published offers to be bound to the active season, so the board always reflects the current competitive window.
- **US-SCC-010**: As a competitor, I want every contract to resolve to a single season, so my merits are never split or misattributed.
- **US-MCF-010**: As a reviewer, I want confirmation requests to be rejected when the supplied season does not match the contract's season, so credits stay in the correct window.

## Why L2

The change adds a business rule to the `MissionOffer` and `MissionContract` domain models, modifies the public `POST /api/v1/missions/offers` contract, and touches persistence and composition. It does not introduce new trust boundaries, identity checks, or external integrations. The risk is comparable to `CHG-0005` (single active season) and `CHG-0004` (season creation with official flag).

## Scope

- Domain: add `SeasonID` to `MissionOffer` and `MissionContract`; validate season consistency during confirmation.
- Application: `PublishMissionHandler` queries the active season before creating an offer; `ConfirmObjectiveHandler` validates request season against contract season.
- Boundary: `POST /api/v1/missions/offers` returns the bound `season_id`; `POST /api/v1/missions/contracts/{contract_id}/evidence` rejects `SEASON_MISMATCH`.
- Infrastructure: `OfferStore` and `ContractStore` persist and load `SeasonID`.
- Composition: wire active-season lookup into `PublishMissionHandler`.

## Non-Goals

- Replace the `SeasonID` request field on `POST /api/v1/missions/contracts/{contract_id}/evidence`; it is kept for backward compatibility and validated against the contract.
- Add a new `PRESEASON` state; pre-season/off-season continue to be modeled with `Official: false`.
- Change freeze/finalize season rules or leaderboard calculation.
- Modify active-mission limits or offer capacity logic.

## Acceptance Summary

- A mission offer cannot be published while no season is `ACTIVE`.
- A published offer carries the `season_id` of the active season at creation time.
- A contract created from an offer inherits the offer's `season_id`.
- A confirmation is rejected when the request `season_id` does not match the contract's `season_id`.
- Existing DynamoDB items without `SeasonID` are treated as the active season on first read (backward-compatible default) or rejected with a clear error.

## Current Outcome

- Gate 1 approved on 2026-08-30.
- Gate 2 approved on 2026-08-30.
- Domain, Application, Infrastructure, Boundary, and Composition layers are GREEN.
- `go test -race ./...` passes.
- `gofmt -w .` and `go vet ./...` clean.
- Docs converged: `docs/API-CONTRACT.md`, `specs/capabilities/mission-board/spec.md`, `specs/capabilities/mission-contract-fulfillment/spec.md`, `specs/capabilities/seasonal-category-competition/spec.md`.
- Coverage tracking uses the existing repository coverage files; the affected domain/application/infrastructure packages are exercised by the unit and integration suites.
- Ready for final review before consolidation and archival.
