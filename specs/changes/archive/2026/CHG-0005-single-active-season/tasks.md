---
change_id: CHG-0005
artifact: tasks
status: proposed
updated: 2026-08-30
---

# Tasks

## T-001 — Add active-season uniqueness rule in Domain

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN
- depends_on: none
- test_id: TEST-DOM-001
- test_type: unit
- requirements: REQ-SCC-005
- done_when: `ErrActiveSeasonAlreadyExists` exists and domain validation rejects when an active season is present.
- red_evidence: `undefined: domain.ValidateCanCreateSeason` and `undefined: domain.ErrActiveSeasonAlreadyExists` when running `go test ./tests/unit/missions/domain/... -run TestValidateCanCreateSeason`.
- green_command: `go test -race ./tests/unit/missions/domain/... -run TestValidateCanCreateSeason` passed.
- blue_refactor: Added a pure domain function `ValidateCanCreateSeason` and a sentinel error; kept `Season` constructor unchanged.

### TEST-DOM-001 — Domain rejects creation if an active season exists

- Given a list containing a `Season` with `State: ACTIVE`, when validating creation, then it returns `ErrActiveSeasonAlreadyExists`.
- Given no `ACTIVE` season, when validating creation, then it returns nil.
- Given a `FREEZING` season, when validating creation, then it returns nil.

## T-002 — Update `CreateSeasonHandler` to check active seasons

- status: completed
- development_layer: application
- layer_gate: Gate 3-APPLICATION
- depends_on: T-001
- test_id: TEST-APP-001
- test_type: unit
- requirements: REQ-SCC-005
- done_when: `CreateSeasonHandler` calls `ActiveSeasonsReader.ListActive()` and fails fast when an active season exists.
- red_evidence: `too many arguments in call to application.NewCreateSeasonHandler` when running unit tests.
- green_command: `go test -race ./tests/unit/missions/application/... -run TestCreateSeasonHandler` passed.
- blue_refactor: Introduced `ActiveSeasonsReader` port with a single `ListActive` operation; kept `CreateSeasonHandler` decoupled from infrastructure.

### TEST-APP-001 — Handler rejects when an active season exists

- Given an existing `ACTIVE` season, when `CreateSeasonHandler` is called, then it returns `ErrActiveSeasonAlreadyExists` and does not save.
- Given no `ACTIVE` season, when `CreateSeasonHandler` is called, then it saves and returns the new season.

## T-003 — Implement `ListActive` in `SeasonStore` and HTTP mapping

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-002
- test_id: TEST-HTTP-001, TEST-HTTP-002
- test_type: integration
- requirements: REQ-SCC-005
- done_when: `SeasonStore.ListActive()` filters active seasons, the endpoint maps the error, and integration tests pass.
- red_evidence: `not enough arguments in call to application.NewCreateSeasonHandler` and `DynamoDB ValidationException: Invalid FilterExpression: Attribute name is a reserved keyword; reserved keyword: State`.
- green_command: `go test -race ./tests/integration/composition/...` passed.
- blue_refactor: Added `#state` expression attribute name in `ListActive`; added `clearMissionsTable` integration test helper to isolate shared local DynamoDB state.

### TEST-HTTP-001 — Reject second active season through HTTP

- Given an `ACTIVE` season exists in DynamoDB, when an operator POSTs a new season, then the response is `422` with `code: ACTIVE_SEASON_EXISTS`.

### TEST-HTTP-002 — Create a season when none is active

- Given no `ACTIVE` season in DynamoDB, when an operator POSTs a new season, then the response is `201 Created`.

## T-004 — Verify and consolidate

- status: completed
- development_layer: verification
- depends_on: T-003
- test_id: all
- test_type: unit + integration
- verification_command: `gofmt -w . && go vet ./... && go test -race ./...`
- green_command: `gofmt -w . && go vet ./... && go test -race ./...` passed.
- done_when: all gates pass and active change artifacts are updated with evidence.
- next_step: common-sdd-verify-spec, then canonical consolidation and archival.
