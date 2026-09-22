---
change_id: CHG-0006
artifact: tasks
status: proposed
updated: 2026-08-30
---

# Tasks

## T-001 — Add `SeasonStatusFilter` in Domain

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN
- depends_on: none
- test_id: TEST-DOM-001
- test_type: unit
- requirements: REQ-SCC-006
- done_when: `SeasonStatusFilter` exists with `ACTIVE`, `INACTIVE`, `ALL` and `Match` works for all states.
- red_evidence: `undefined: domain.SeasonStatusFilter` when running `go test ./tests/unit/missions/domain/... -run TestSeasonStatusFilter`.
- green_command: `go test -race ./tests/unit/missions/domain/... -run TestSeasonStatusFilter` passed.
- blue_refactor: Pure value object; no side effects.

### TEST-DOM-001 — Filter matches the right seasons

- Given `ACTIVE`, when matching an `ACTIVE` season, then it returns true; a `FINALIZED` season returns false.
- Given `INACTIVE`, when matching a `FINALIZED` season, then it returns true; an `ACTIVE` season returns false.
- Given `ALL`, when matching any season, then it returns true.

## T-002 — Add `ListSeasonsHandler` and `SeasonReader` port

- status: completed
- development_layer: application
- layer_gate: Gate 3-APPLICATION
- depends_on: T-001
- test_id: TEST-APP-001
- test_type: unit
- requirements: REQ-SCC-006
- done_when: `ListSeasonsHandler` calls `SeasonReader.List` and returns the result.
- red_evidence: `undefined: application.ListSeasonsHandler` when running `go test ./tests/unit/missions/application/... -run TestListSeasonsHandler`.
- green_command: `go test -race ./tests/unit/missions/application/... -run TestListSeasonsHandler` passed.
- blue_refactor: Read-only query; one port operation.

### TEST-APP-001 — Handler returns filtered seasons

- Given a `SeasonReader` with an `ACTIVE` and a `FINALIZED` season, when `ListSeasonsHandler.Handle(ACTIVE)` is called, then it returns only the active one.
- Given `INACTIVE`, when called, then it returns only the finalized one.
- Given `ALL`, when called, then it returns both.

## T-003 — Implement `SeasonStore.List` in Infrastructure

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-002
- test_id: TEST-INFRA-001
- test_type: integration
- requirements: REQ-SCC-006
- done_when: `SeasonStore.List` returns the correct subset for each filter.
- red_evidence: `undefined: (*SeasonStore).List` when running `go test ./tests/integration/infrastructure/... -run TestSeasonStore_List`.
- green_command: `go test -race ./tests/integration/infrastructure/... -run TestSeasonStore_List` passed.
- blue_refactor: Query by `PK = SEASON`, then apply filter in Go to avoid composite condition issues.

### TEST-INFRA-001 — DynamoDB store filters by status

- Given saved `ACTIVE` and `ARCHIVED` seasons, when `List(ACTIVE)` is called, then it returns only the active one.
- Given `INACTIVE`, when called, then it returns only the archived one.
- Given `ALL`, when called, then it returns both.

## T-004 — Add HTTP `GET /api/v1/seasons` handler

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-002
- test_id: TEST-HTTP-001, TEST-HTTP-002
- test_type: integration
- requirements: REQ-SCC-006
- done_when: Route is registered, query param is parsed, and JSON list is returned.
- red_evidence: `404` on `GET /api/v1/seasons` in integration tests.
- green_command: `go test -race ./tests/integration/composition/... -run TestCompositionServer_ListSeasons` passed.
- blue_refactor: Query parsing stays in HTTP layer; application receives a typed filter.

### TEST-HTTP-001 — List active seasons through HTTP

- Given an `ACTIVE` season and a `FINALIZED` season, when `GET /api/v1/seasons` is called, then the response is `200` with only the active season.

### TEST-HTTP-002 — Reject unknown status values

- Given any state, when `GET /api/v1/seasons?status=closed` is called, then the response is `400` with `INVALID_QUERY_PARAMETER`.

## T-005 — Wire route in Composition

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-004
- test_id: TEST-HTTP-001, TEST-HTTP-002
- test_type: integration
- requirements: REQ-SCC-006
- done_when: `NewServer` registers `GET /api/v1/seasons` and uses `AuthMiddleware`.
- red_evidence: `http: pattern "GET /api/v1/seasons" conflicts` or route not found.
- green_command: `go test -race ./tests/integration/composition/... -run TestCompositionServer` passed.

## T-006 — Verify and consolidate

- status: completed
- development_layer: verification
- depends_on: T-005
- test_id: all
- test_type: unit + integration
- verification_command: `gofmt -w . && go vet ./... && go test -race ./...`
- green_command: `gofmt -w . && go vet ./... && go test -race ./...` passed.
- done_when: all gates pass, `docs/API-CONTRACT.md` is updated, and active change artifacts are current.
- next_step: common-sdd-verify-spec, then canonical consolidation and archival.
