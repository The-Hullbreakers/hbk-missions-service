---
change_id: CHG-0007
title: Freeze and finalize a season
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
change_path: specs/changes/archive/2026/CHG-0007-freeze-and-finalize-season/
archive_path: specs/changes/archive/2026/CHG-0007-freeze-and-finalize-season/
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_UPDATE_DOCUMENTATION_WORKFLOW
context_scope: existing
verified_at: 2026-08-30
source_revision: HEAD
created: 2026-08-30
updated: 2026-08-30
---

# Freeze And Finalize A Season

## Intent

Allow an authorized operator to close the active competitive window: freeze the current season so only in-progress work can still earn merits, and then finalize it so the ledger is locked and standings become immutable.

## Business Value

- Close a season without losing eligible work that started in time.
- Prevent late scoring once the freeze window begins.
- Produce a final, immutable state that future leaderboards and achievements can trust.

## User Stories

- **US-SCC-007**: As an operator, I want to freeze the active season, to stop new merits except contracts already started.
- **US-SCC-008**: As an operator, I want to end a frozen season, to publish immutable standings and close the ledger.

## Why L2

Administrative state transitions across two new endpoints; modifies an existing shared domain rule in `ConfirmObjective` to accept `FREEZING` for `IN_PROGRESS` contracts; persists the new state in DynamoDB. No identity, security, or external integration changes beyond the existing `AuthMiddleware`.

## Scope

- Domain: `Season` state transitions and `CanAcceptMerits` policy in `internal/missions/domain/season.go`.
- Domain: update `ConfirmObjective` to use the new policy.
- Application: `FreezeSeasonHandler` and `FinalizeSeasonHandler` with `SeasonReader`/`SeasonWriter` ports.
- Infrastructure: reuse `SeasonStore.Get` and `Save`; no schema changes.
- Interface: `POST /api/v1/missions/seasons/{id}/freeze` and `POST /api/v1/missions/seasons/{id}/finalize`.
- Composition: wire routes and error mapping in `NewServer`.
- Tests: domain, application, interface, and integration tests.
- Documentation: update `docs/API-CONTRACT.md`; canonical spec updated on consolidation.

## Non-Goals

- Leaderboards, standings projection, or Achievements integration.
- `ARCHIVED` state transition.
- Wall-clock "started before freeze" timestamp; v1 treats "started in time" as `contract.State == ContractInProgress`.
- Unfreeze or reopen a finalized season.
- Role-based authorization beyond the existing middleware.

## Acceptance Summary

- `POST /api/v1/missions/seasons/{id}/freeze` on an `ACTIVE` season returns `200` with the season now `FREEZING`.
- `POST /api/v1/missions/seasons/{id}/freeze` on a non-`ACTIVE` season returns `422` with `SEASON_STATE_INVALID`.
- `POST /api/v1/missions/seasons/{id}/finalize` on a `FREEZING` season returns `200` with the season now `FINALIZED`.
- `POST /api/v1/missions/seasons/{id}/finalize` on a non-`FREEZING` season returns `422` with `SEASON_STATE_INVALID`.
- A non-existent season for either endpoint returns `404` with `SEASON_NOT_FOUND`.
- Confirming an objective during `FREEZING` succeeds only when the contract is `IN_PROGRESS`.
- Confirming an objective during `FINALIZED` returns `SEASON_NOT_ACTIVE`.
- The `Official` flag is never changed by freeze or finalize.

## Current Outcome

Gate 1 approved on 2026-08-30. Gate 2 approved same day.

- T-001 — Domain RED/GREEN/BLUE completed: `domain.Freeze`, `domain.Finalize`, `Season.CanAcceptMerits`, and `ConfirmObjective` updated. Unit tests green.
- T-002 — Application completed: `FreezeSeasonHandler` and `FinalizeSeasonHandler` with `SeasonReader`/`SeasonWriter` ports. Unit tests green.
- T-003/T-004 — Boundary and composition completed: HTTP handlers, error mapping, route wiring, `SeasonStore.Get` returns `ErrSeasonNotFound`. Integration test `TestCompositionServer_FreezeAndFinalize` green.
- T-005 — `docs/API-CONTRACT.md` updated with freeze/finalize notes and error vocabulary.
- T-006 — `gofmt -w .`, `go vet ./...`, and `go test -race ./...` all pass on 2026-08-30.

Next: `common-sdd-verify-spec`, then canonical consolidation and archival.
