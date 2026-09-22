# Tasks — Simplify Mission Publishing By Removing MissionPool

## Gate 1

- [x] T-001 — Identify canonical owner `mission-board` and classify risk as L2.
- [x] T-002 — Present Gate 1 plan and obtain approval.

## Gate 2

- [x] T-003 — Finalize `change.md`, `delta.md`, `design.md`, and `tasks.md`.
- [x] T-004 — Present written artifacts and obtain Gate 2 approval.

## Domain RED/GREEN/BLUE

- [x] T-005 — RED: remove `MissionPool` and `MissionPublishMode` from `internal/missions/domain/mission_pool.go` and delete the file.
- [x] T-006 — RED: remove `PoolEntry` from `internal/missions/domain/mission_pool.go` if not already removed.
- [x] T-007 — RED: remove `ErrPoolInvalid` and `ErrPoolNotFound` from `internal/missions/domain/errors.go`.
- [x] T-008 — GREEN: verify `NewMissionOfferFromDefinition` still works without pool; run `go test ./tests/unit/missions/domain`.
- [x] T-009 — BLUE: ensure all domain files remain <150 physical lines.

## Application RED/GREEN/BLUE

- [x] T-010 — RED: delete `internal/missions/application/create_mission_pool.go`.
- [x] T-011 — RED: rename `internal/missions/application/publish_mission.go` handler and remove `PoolReader` dependency.
- [x] T-012 — GREEN: adjust unit tests in `tests/unit/missions/application/publish_mission_test.go` and remove pool fixtures.
- [x] T-013 — BLUE: ensure application files and tests remain <150 physical lines.

## Boundary RED/GREEN/BLUE

- [x] T-014 — RED: delete `internal/missions/interface/http/create_mission_pool.go` and remove error mappings.
- [x] T-015 — RED: update `internal/missions/interface/http/publish_mission.go` (or equivalent) to remove `pool_id` from request.
- [x] T-016 — GREEN: adjust HTTP integration tests in `tests/integration/http`.
- [x] T-017 — BLUE: ensure HTTP handlers remain thin and <150 physical lines.

## Infrastructure RED/GREEN/BLUE

- [x] T-018 — RED: delete `internal/missions/infrastructure/dynamodb/pool_store.go`.
- [x] T-019 — RED: update `internal/missions/infrastructure/dynamodb/errors.go` if pool-specific errors are mapped.
- [x] T-020 — GREEN: adjust DynamoDB integration tests.
- [x] T-021 — BLUE: ensure infrastructure files remain <150 physical lines.

## Composition

- [x] T-022 — Update `internal/missions/composition/server.go` to remove pool wiring and route.
- [x] T-023 — Update `cmd/hbk-missions-service/main.go` if it explicitly initializes `PoolStore`.

## Documentation

- [x] T-024 — Update `docs/API-CONTRACT.md` to remove `POST /api/v1/missions/pools` and `pool_id` from `POST /api/v1/missions/offers`.
- [x] T-025 — Update `specs/capabilities/mission-board/spec.md` (after implementation, before consolidation).

## Verification

- [x] T-026 — Run `gofmt`, `go vet`, `go test -race ./...`.
- [x] T-027 — Run `tools/validate-sdd-change.sh`.
- [x] T-028 — Verify coverage >= 90% on affected packages.
