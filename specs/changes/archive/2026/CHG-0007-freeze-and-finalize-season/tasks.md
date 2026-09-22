---
change_id: CHG-0007
artifact: tasks
status: proposed
updated: 2026-08-30
---

# Tasks

## T-001 — Add season state transitions and merit-acceptance policy

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN
- depends_on: none
- test_id: TEST-DOM-001, TEST-DOM-002, TEST-DOM-003
- test_type: unit
- requirements: REQ-SCC-004
- done_when: `Freeze`, `Finalize`, and `CanAcceptMerits` enforce the state machine and `ConfirmObjective` accepts `FREEZING` for in-progress contracts.
- red_evidence: `undefined: domain.Freeze` or `TestConfirmObjective_AcceptsFreezingWhenInProgress` fails with `ErrSeasonNotActive`.
- green_command: `go test -race ./tests/unit/missions/domain/... -run TestSeason` and `go test -race ./tests/unit/missions/domain/... -run TestConfirmObjective`.
- blue_refactor: Pure value-object transitions; state machine in one place.

### TEST-DOM-001 — Season state transitions

- Given a season `ACTIVE`, when `Freeze` is called, then it returns `FREEZING`.
- Given a season `FREEZING`, when `Finalize` is called, then it returns `FINALIZED`.
- Given a season `ACTIVE`, when `Finalize` is called, then it returns `ErrInvalidSeasonStateTransition`.
- Given a season `FINALIZED`, when `Freeze` is called, then it returns `ErrInvalidSeasonStateTransition`.

### TEST-DOM-002 — Season accepts merits only during eligible windows

- Given a `FREEZING` season and an `IN_PROGRESS` contract, when `CanAcceptMerits` is checked, then it returns true.
- Given a `FREEZING` season and an `ACCEPTED` contract, when `CanAcceptMerits` is checked, then it returns false.
- Given a `FINALIZED` season and any contract, when `CanAcceptMerits` is checked, then it returns false.

### TEST-DOM-003 — Confirm objective respects merit-acceptance policy

- Given a `FREEZING` season and an `IN_PROGRESS` contract, when a unit is confirmed, then a `MeritCredit` is recorded.
- Given a `FINALIZED` season, when a unit is confirmed, then it returns `ErrSeasonNotActive`.

## T-002 — Add FreezeSeasonHandler and FinalizeSeasonHandler

- status: completed
- development_layer: application
- layer_gate: Gate 3-APPLICATION
- depends_on: T-001
- test_id: TEST-APP-001, TEST-APP-002
- test_type: unit
- requirements: REQ-SCC-004
- done_when: Both handlers load, transition, persist, and surface `ErrSeasonNotFound` / `ErrInvalidSeasonStateTransition`.
- red_evidence: `undefined: application.FreezeSeasonHandler` when running `go test ./tests/unit/missions/application/... -run TestFreezeSeasonHandler`.
- green_command: `go test -race ./tests/unit/missions/application/... -run TestSeasonHandler`.
- blue_refactor: One command per handler; thin orchestration.

### TEST-APP-001 — Freeze command succeeds and fails correctly

- Given a `SeasonReader` returns an `ACTIVE` season, when `FreezeSeasonHandler.Handle` is called, then it calls `Save` with `FREEZING` and returns the season.
- Given a `SeasonReader` returns a `FINALIZED` season, when `FreezeSeasonHandler.Handle` is called, then it returns `ErrInvalidSeasonStateTransition` and does not call `Save`.

### TEST-APP-002 — Finalize command succeeds and fails correctly

- Given a `SeasonReader` returns a `FREEZING` season, when `FinalizeSeasonHandler.Handle` is called, then it calls `Save` with `FINALIZED`.
- Given a `SeasonReader` returns an `ACTIVE` season, when `FinalizeSeasonHandler.Handle` is called, then it returns `ErrInvalidSeasonStateTransition`.

## T-003 — Implement HTTP handlers for freeze and finalize

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-002
- test_id: TEST-HTTP-001, TEST-HTTP-002, TEST-HTTP-003
- test_type: integration
- requirements: REQ-SCC-004
- done_when: Both endpoints return `200`, `404`, or `422` with the correct public code.
- red_evidence: `404` on `POST /api/v1/missions/seasons/{id}/freeze` in integration tests.
- green_command: `go test -tags=integration -race ./tests/integration/composition/... -run TestCompositionServer_Freeze`.
- blue_refactor: HTTP layer maps only status and codes; business logic stays in application.

### TEST-HTTP-001 — Freeze a season through HTTP

- Given an `ACTIVE` season, when an operator `POST`s to `/api/v1/missions/seasons/{id}/freeze`, then the response is `200` with `State: FREEZING`.

### TEST-HTTP-002 — Finalize a season through HTTP

- Given a `FREEZING` season, when an operator `POST`s to `/api/v1/missions/seasons/{id}/finalize`, then the response is `200` with `State: FINALIZED`.

### TEST-HTTP-003 — Reject invalid state transitions through HTTP

- Given a `FINALIZED` season, when an operator `POST`s to `/api/v1/missions/seasons/{id}/freeze`, then the response is `422` with `code: SEASON_STATE_INVALID`.

## T-004 — Wire routes and update error mapping

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-003
- test_id: TEST-HTTP-001, TEST-HTTP-002, TEST-HTTP-003
- test_type: integration
- requirements: REQ-SCC-004
- done_when: `NewServer` registers both routes, maps `ErrSeasonNotFound` to `SEASON_NOT_FOUND`, `ErrInvalidSeasonStateTransition` to `SEASON_STATE_INVALID`, and integration tests pass.
- red_evidence: `http: pattern conflict` or missing `SEASON_STATE_INVALID` mapping.
- green_command: `go test -tags=integration -race ./tests/integration/composition/... -run TestCompositionServer`.
- blue_refactor: Reuse existing `SeasonStore` and `AuthMiddleware`.

## T-005 — Update API contract and canonical spec

- status: completed
- development_layer: verification
- depends_on: T-004
- test_id: all
- test_type: documentation
- requirements: REQ-SCC-004
- done_when: `docs/API-CONTRACT.md` describes freeze/finalize behavior and codes; canonical spec ready for consolidation.
- verification: review diff of `docs/API-CONTRACT.md` and planned canonical delta.

## T-006 — Verify and consolidate

- status: completed
- development_layer: verification
- depends_on: T-005
- test_id: all
- test_type: unit + integration
- verification_command: `gofmt -w . && go vet ./... && go test -race ./...`
- done_when: all gates pass, active change artifacts are updated with evidence, and the delta is ready for `common-sdd-verify-spec`.
- next_step: `common-sdd-verify-spec`, then canonical consolidation and archival.
