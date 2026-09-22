---
change_id: CHG-0013
artifact: tasks
status: archived
updated: 2026-09-02
---

# Tasks — Schedule Seasons With Automatic Transitions

## T-001 — Finalize BDD scenarios and design

- **Owner**: agent
- **Wave**: Gate 1 → Gate 2
- **Dependencies**: none
- **Outcome**: `delta.md` and `design.md` reviewed, no open design questions.
- **Verification**: All scenarios have Given/When/Then and map to tests.
- **Next**: T-002

## T-002 — Domain RED for scheduled season

- **Owner**: agent
- **Wave**: Domain RED
- **Dependencies**: T-001
- **Outcome**: `SeasonSeries`, `SeasonSchedule`, and `Season.ShouldFreeze/ShouldFinalize/NextScheduledSeason` fail the first RED tests.
- **Verification**: `go test ./tests/unit/missions/domain/...` shows expected RED.
- **Next**: T-003

## T-003 — Domain GREEN

- **Owner**: agent
- **Wave**: Domain GREEN
- **Dependencies**: T-002
- **Outcome**: Domain constructors and schedule transitions pass unit tests.
- **Verification**: `go test ./tests/unit/missions/domain/...` green.
- **Next**: T-004

## T-004 — Domain BLUE

- **Owner**: agent
- **Wave**: Domain BLUE
- **Dependencies**: T-003
- **Outcome**: Small value objects, no time framework in domain, `< 150` lines per file.
- **Verification**: `gofmt`, `go vet`, and line-count check pass.
- **Next**: T-005

## T-005 — Application RED for scheduled creation and scheduler

- **Owner**: agent
- **Wave**: Application RED
- **Dependencies**: T-004
- **Outcome**: `CreateScheduledSeasonHandler` and `SeasonScheduler` fail their unit tests against hand-written doubles.
- **Verification**: `go test ./tests/unit/missions/application/...` RED.
- **Next**: T-006

## T-006 — Application GREEN

- **Owner**: agent
- **Wave**: Application GREEN
- **Dependencies**: T-005
- **Outcome**: Handlers and scheduler pass unit tests.
- **Verification**: `go test ./tests/unit/missions/application/...` green.
- **Next**: T-007

## T-007 — Boundary and composition

- **Owner**: agent
- **Wave**: Boundary/Composition
- **Dependencies**: T-006
- **Outcome**: `POST /api/v1/seasons` with schedule, season responses, and `SeasonScheduler` goroutine wired in `NewServer`.
- **Verification**: `tests/integration/composition/...` and `tests/integration/http/...` green.
- **Next**: T-008

## T-008 — Infrastructure persistence

- **Owner**: agent
- **Wave**: Infrastructure
- **Dependencies**: T-007
- **Outcome**: `SeasonStore` persists and loads `StartsAt` and `Schedule`; manual seasons load unchanged.
- **Verification**: `tests/integration/infrastructure/dynamodb/...` green.
- **Next**: T-009

## T-009 — Mutation and critical E2E

- **Owner**: agent
- **Wave**: Verification
- **Dependencies**: T-008
- **Outcome**: Mutation and critical E2E evidence recorded; affected scope coverage ≥ 90%.
- **Verification**: `go test -race ./...`, mutation report, and critical E2E pass.
- **Next**: T-010

## T-010 — Documentation and rollout

- **Owner**: agent
- **Wave**: Converge
- **Dependencies**: T-009
- **Outcome**: `specs/capabilities/seasonal-category-competition/spec.md`, `docs/API-CONTRACT.md`, `docs/DATA-MODEL.md`, and `docs/DOMAIN-RULES.md` updated; `rollout.md` and `verification.md` complete.
- **Verification**: Docs match code and `tools/validate-sdd-change.sh` passes.
