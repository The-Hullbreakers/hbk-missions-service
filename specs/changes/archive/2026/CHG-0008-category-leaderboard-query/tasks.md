---
change_id: CHG-0008
artifact: test-first-tasks
status: completed
updated: 2026-08-30
---

# Test-First Tasks

## T-001 — Modelar leaderboard, cursor y desempates

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: Gate 2
- test_id: TEST-SCC-LB-001
- test_type: unit
- test_file: `tests/unit/missions/domain/leaderboard_test.go`
- harness_files: `tests/unit/missions/domain/leaderboard_test.go`
- mapped_production_files:
  - `internal/missions/domain/leaderboard.go` — `Leaderboard`, `LeaderboardEntry`, `Rank`, compareStanding, encode/decode cursor.
  - `internal/missions/domain/standing.go` — `Standing` extended with `CompletedMissions` and `UniqueObjectives`; `ProjectStanding` calculates them.
  - `internal/missions/domain/merit_ledger.go` — `CreditsFor` added.
- requirements: REQ-SCC-007
- done_when: Descending merit order, deterministic tiebreakers, and stable cursor are green.
- outcome: `TEST-SCC-LB-001` to `TEST-SCC-LB-003` green. `BuildLeaderboard` orders by merits, tie-breaks (completed missions, unique objectives, recipient id), and paginates with an opaque base64 cursor. `ProjectStanding` uses `CreditsFor` to compute tie-break facts without a second source of truth.
- verification: `go test -race ./tests/unit/missions/domain/...` passes; `gofmt` and `go vet` clean. Files under 150 lines: `leaderboard.go` (124), `standing.go` (39), `merit_ledger.go` (37), `leaderboard_test.go` (88).
- next_step: T-002 Application query handler.

### TEST-SCC-LB-001 — Sort standings by merits

- Given standings of `player-1` (150) and `player-2` (100).
- When the leaderboard is sorted.
- Then `player-1` has `rank` 1 and `player-2` has `rank` 2.

### TEST-SCC-LB-002 — Apply deterministic tie-break

- Given `player-1` and `player-2` have 150 merits.
- When the tiebreaker is by `ConfirmationID` lexicographic.
- Then the order is stable and the rank does not repeat position 1.

### TEST-SCC-LB-003 — Cursor encodes last position

- Given a leaderboard of 25 entries.
- When the first page of 10 is requested.
- Then the cursor allows you to get exactly the next 10.

## T-002 — Ormissionar query de leaderboard

- status: completed
- development_layer: application
- layer_gate: Gate 3-APPLICATION, luego LAYER-GATE-APPLICATION
- depends_on: LAYER-GATE-DOMAIN
- test_id: TEST-APP-LB-001
- test_type: unit
- test_file: `tests/unit/missions/application/list_leaderboard_test.go`
- red_command: `go test ./tests/unit/missions/application/... -run TestListLeaderboard`
- expected_red: no existe `ListLeaderboardHandler`.
- mapped_production_files:
  - `internal/missions/application/list_leaderboard.go` — `ListLeaderboardHandler`, `LeaderboardReader`.
  - `internal/missions/domain/merit_credit.go` — `ParseTrack`, `ErrInvalidTrack`.
- requirements: REQ-SCC-007
- done_when: query valida season, track, tenant y aplica `min_merits`.
- outcome: `TEST-APP-LB-001` a `TEST-APP-LB-003` verdes. `ListLeaderboardHandler` valida `track`, verifica season y delega `BuildLeaderboard`. `ParseTrack` acepta `individual` y `team`.
- verification: `go test -race ./tests/unit/missions/application/...` pasa; `gofmt` y `go vet` limpios. Files under 150 lines: `list_leaderboard.go` (43), `list_leaderboard_test.go` (77).
- next_step: T-003 Infrastructure (GSI y leaderboard store).

### TEST-APP-LB-001 — Query returns standings

- Given credits to `tenant-1`, `season-1`, `BOUNTY_HUNTING`, `INDIVIDUAL`.
- When `ListLeaderboardHandler.Handle(...)` with `track=individual`.
- Then returns ordered standings with `next_cursor`.

### TEST-APP-LB-002 — Reject invalid track

- Given `track` invalid.
- When `Handle(...)` is invoked.
- Then returns error mappable to `INVALID_QUERY_PARAMETER`.

### TEST-APP-LB-003 — Season not found

- Given season inexistente.
- When `Handle(...)` se invoca.
- Then devuelve `SEASON_NOT_FOUND`.

## T-003 — Read standings from DynamoDB

- status: completed
- development_layer: infrastructure
- layer_gate: LAYER-GATE-INFRASTRUCTURE
- depends_on: LAYER-GATE-APPLICATION
- test_id: TEST-INFRA-LB-001
- test_type: integration/infrastructure
- test_file: `tests/integration/infrastructure/dynamodb/merit_ledger_store_list_test.go`
- mapped_production_files:
  - `internal/missions/infrastructure/dynamodb/dynamodb_client.go` — GSI2.
  - `internal/missions/infrastructure/dynamodb/merit_ledger_save.go` — saves `GSI2PK`/`GSI2SK`.
  - `internal/missions/infrastructure/dynamodb/merit_ledger_readers.go` — `ListStandings`, `TotalFor`.
  - `internal/missions/infrastructure/dynamodb/merit_ledger_keys.go` — key helpers.
- requirements: REQ-SCC-007
- done_when: the store groups credits by bucket and returns standings.
- outcome: `TEST-INFRA-LB-001` green. GSI2 allows per bucket query without `RecipientID`; `ListStandings` adds totals and tie-break facts. Store divided into files < 150 lines.
- verification: `go test -race ./tests/unit/missions/...` passes; `go test -tags=integration ./tests/integration/infrastructure/dynamodb/... -run TestMeritLedgerStore_ListStandingsByBucket` passes with local DynamoDB; `gofmt` and `go vet` clean. Files under 150 lines: `dynamodb_client.go` (100), `merit_ledger_save.go` (58), `merit_ledger_readers.go` (92), `merit_ledger_keys.go` (42).
- next_step: T-004 Boundary (HTTP endpoint).

## T-004 — Exponer endpoint HTTP

- status: completed
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY, then LAYER-GATE-BOUNDARY-RED
- depends_on: LAYER-GATE-INFRASTRUCTURE
- test_id: TEST-HTTP-LB-001
- test_type: integration/http
- test_files:
  - `tests/integration/http/leaderboard_helpers_test.go` — fakes and shared helpers.
  - `tests/integration/http/leaderboard_test.go` — happy path cases and domain errors.
  - `tests/integration/http/leaderboard_boundary_test.go` — auth and tenant boundary.
- mapped_production_files:
  - `internal/missions/interface/http/leaderboard.go` — `LeaderboardHandler`, `tenantFromContext`, `toLeaderboardResponse`.
  - `internal/missions/interface/http/errors.go` — mapping `ErrInvalidTrack`, `ErrInvalidCursor` to `INVALID_QUERY_PARAMETER`.
  - `internal/missions/composition/server.go` — wiring and routing.
- requirements: REQ-SCC-007
- done_when: the endpoint responds 200/400/404/403/401 and does not expose evidence.
- outcome: `TEST-HTTP-LB-001` to `TEST-HTTP-LB-006` green. The handler parses `track`, `limit`, `cursor`, `min_merits`; validates tenant and maps `INVALID_QUERY_PARAMETER`/`SEASON_NOT_FOUND`/`TENANT_BOUNDARY_DENIED`/`ErrInvalidCursor`. The DTO exposes `items` and `next_cursor`.
- verification: `go test -race ./...` passes; `gofmt` and `go vet` clean. Coverage > 90% on affected files (`LeaderboardHandler` 97.6%, `BuildLeaderboard` 96.9%, `ListStandings` 92.5%). Files under 150 lines: `leaderboard.go` (109), `leaderboard_helpers_test.go` (65), `leaderboard_test.go` (93), `leaderboard_boundary_test.go` (28), `server.go` (85), `errors.go` (53).
- next_step: T-005 Documentation.

### TEST-HTTP-LB-001 — Get individual leaderboard

- Given credits for individual `BOUNTY_HUNTING`.
-When `GET /api/v1/missions/seasons/season-1/leaderboards/BOUNTY_HUNTING?track=individual`.
- Then `200 OK` with ordered standings.

### TEST-HTTP-LB-002 — Missing track returns 400

- Given request without `track`.
- When the endpoint is invoked.
- Then `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

### TEST-HTTP-LB-003 — Tenant boundary

- Given consumer from another tenant.
- When you consult the leaderboard.
- Then `403 Forbidden` with `TENANT_BOUNDARY_DENIED`.

## T-005 — Converge documentation

- status: completed
- development_layer: documentation
- layer_gate: documentation review
- depends_on: LAYER-GATE-BOUNDARY-RED
- requirements: REQ-SCC-007
- done_when: `docs/API-CONTRACT.md` and delta reflect DTO and end filters.
- outcome: `docs/API-CONTRACT.md` updated with endpoint, filters, error codes and response DTO.
- verification: `go test -race ./...` passes; `gofmt` and `go vet` clean.
- next_step: Phase 5 — converge, verify and consolidate (review canonical spec and `specs/index.md`).
