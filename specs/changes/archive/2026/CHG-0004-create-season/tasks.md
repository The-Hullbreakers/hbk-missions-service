---
change_id: CHG-0004
artifact: tasks
status: proposed
updated: 2026-08-30
---

# Tasks

## T-001 — Add `Official` flag to `Season`

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN
- depends_on: none
- test_id: TEST-DOM-001
- test_type: unit
- requirements: REQ-SCC-005
- done_when: `NewSeason` defaults to `Official: true`, `NewSeasonWithOfficial` exists, and unit tests pass.
- verification: `go test -race ./tests/unit/missions/domain/...` passes; `gofmt` and `go vet` clean.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run TestSeasonOfficial`
- Output: `undefined: NewSeasonWithOfficial`.
- Evidence: `Gate 3-DOMAIN` approved to continue T-001.

#### GREEN

- Production changes: `internal/missions/domain/season.go` adds `Official bool`, `NewSeason` defaults to `true`, and `NewSeasonWithOfficial` constructor.
- Command: `go test ./tests/unit/missions/domain/... -run TestSeason`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- `NewSeason` calls `NewSeasonWithOfficial` with default `true` to avoid breaking existing callers.
- `tests/unit/missions/domain/season_test.go` verifies default and explicit `Official` values.
- `LAYER-GATE-DOMAIN` passed for TEST-DOM-001.

### TEST-DOM-001 — Season can be official or unofficial

- Given a season is created with `Official: true`, then `season.Official` is `true`.
- Given a season is created with `Official: false`, then `season.Official` is `false`.
- Given `NewSeason` is called with only `id` and `state`, then `Official` defaults to `true`.

## T-002 — Implement `CreateSeasonHandler`

- status: completed
- development_layer: application
- layer_gate: Gate 3-APPLICATION
- depends_on: T-001
- test_id: TEST-APP-001
- test_type: unit
- requirements: REQ-SCC-005
- done_when: handler validates input, persists the season, and returns domain errors.
- verification: `go test -race ./tests/unit/missions/application/...` passes.

#### RED

- Command: `go test ./tests/unit/missions/application/... -run TestCreateSeasonHandler`
- Output: `undefined: CreateSeasonHandler`.
- Evidence: `Gate 3-APPLICATION` approved to continue T-002.

#### GREEN

- Production changes: `internal/missions/application/create_season.go` adds `CreateSeasonHandler`, `SeasonWriter` port, `ErrSeasonInvalid`, and `INVALID_SEASON` mapping.
- Command: `go test ./tests/unit/missions/application/... -run TestCreateSeasonHandler`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/application`.

#### BLUE

- Reuses existing `SeasonReader` from `confirm_objective.go` (not needed by handler).
- `tests/unit/missions/application/create_season_test.go` uses a recording mock for `SeasonWriter`.
- `LAYER-GATE-APPLICATION` passed for TEST-APP-001.

### TEST-APP-001 — Create season with official flag

- Given a valid request with `season_id` and `official_statistics: false`, when `CreateSeasonHandler` runs, then it persists the season with `Official: false` and returns it.
- Given an empty `season_id`, when `CreateSeasonHandler` runs, then it returns `ErrSeasonInvalid`.

## T-003 — Expose HTTP endpoint and persist `Official` flag

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY-RED
- depends_on: T-002
- test_id: TEST-HTTP-001, TEST-HTTP-002
- test_type: integration
- requirements: REQ-SCC-005
- done_when: HTTP endpoint accepts a season, persists it in DynamoDB, and returns `201` or `INVALID_SEASON`.
- verification: `go test -tags=integration -race ./tests/integration/...` passes; `gofmt` and `go vet` clean.

#### RED

- Command: `go test -tags=integration ./tests/integration/composition/... -run TestCompositionServer_CreateSeason`
- Output: `undefined: missionhttp.CreateSeasonHandler` / `undefined: NewCreateSeasonHandler`.
- Evidence: `Gate 3-BOUNDARY-RED` approved to continue T-003.

#### GREEN

- Production changes: `internal/missions/infrastructure/dynamodb/season_store.go` persists and loads `Official` with default `true`; `internal/missions/interface/http/create_season.go` decodes `official_statistics` (optional); `internal/missions/composition/server.go` wires `CreateSeasonHandler` and route `POST /api/v1/seasons`; `internal/missions/interface/http/errors.go` maps `ErrSeasonInvalid` to `INVALID_SEASON`.
- Command: `go test -tags=integration ./tests/integration/composition/... -run TestCompositionServer_CreateSeason`
- Output: `ok github.com/hbk/hbk-missions-service/tests/integration/composition`.

#### BLUE

- `getBool` helper added to `attribute_helpers.go` for future use; `SeasonStore` defaults missing `Official` to `true`.
- HTTP handler uses `*bool` for `official_statistics` to distinguish missing from `false` and defaults to `true`.
- `tests/integration/composition/create_season_test.go` covers success and invalid `season_id`.
- `LAYER-GATE-BOUNDARY-RED` passed for TEST-HTTP-001 and TEST-HTTP-002.

### TEST-HTTP-001 — Create a season through HTTP

- Given an authenticated operator POSTs `{ "season_id": "season-2026-q4", "official_statistics": false }` to `/api/v1/seasons`.
- When the request is valid.
- Then the season is persisted with `Official: false` and returned with `201`.

### TEST-HTTP-002 — Reject invalid season through HTTP

- Given an authenticated operator POSTs `{ "season_id": "" }` to `/api/v1/seasons`.
- When the request is processed.
- Then the response is `422 Unprocessable Entity` with `code: INVALID_SEASON`.

## T-004 — Verify and consolidate

- status: completed
- development_layer: verification
- depends_on: T-003
- test_id: all
- test_type: unit + integration
- verification_command: `gofmt -w . && go vet ./... && go test -race ./...`
- done_when: all gates pass and active change artifacts are updated with evidence.
- outcome: `go test -race ./...` passes; `gofmt` and `go vet` clean; no `tools/validate-sdd-change.sh` present.
- next_step: User review or proceed to `common-sdd-verify-spec`.
