---
change_id: CHG-0013
artifact: verification
status: passed
updated: 2026-09-02
---

# Verification — Schedule Seasons With Automatic Transitions

## Validation Plan

| Gate | Command / Evidence | Expected Result |
| --- | --- | --- |
| Domain RED | `go test ./tests/unit/missions/domain/... -run TestScheduledSeason` | Fails with undefined schedule behavior |
| Domain GREEN | `go test ./tests/unit/missions/domain/...` | Pass |
| Application RED | `go test ./tests/unit/missions/application/... -run TestCreateScheduledSeason\|TestSeasonScheduler` | Fails with missing handler/scheduler |
| Application GREEN | `go test ./tests/unit/missions/application/...` | Pass |
| Boundary/Composition | `go test ./tests/integration/composition/... ./tests/integration/http/...` | Pass |
| Infrastructure | `go test ./tests/integration/infrastructure/dynamodb/...` | Pass |
| Race | `go test -race ./...` | Pass |
| Coverage | `tools/validate-sdd-change.sh` | Domain/application/infrastructure ≥ 90% |
| Mutation | Infection / go-mutesting on affected packages | No unexplained surviving mutants |
| Critical E2E | `tests/integration/e2e/...` | Full schedule lifecycle end-to-end |

## Domain Evidence

- `NewScheduledSeason` rejects `active_weeks <= 0` and `off_weeks < 0`.
- `Season.ShouldFreeze` is true exactly at `starts_at + active_weeks`.
- `Season.ShouldFinalize` is true exactly at `starts_at + active_weeks + off_weeks`.
- `Season.NextScheduledSeason` derives `SeasonID` with iteration `+ 1` and starts at `NextStartsAt()`.

## Application Evidence

- `CreateScheduledSeasonHandler` rejects creation when an active season exists.
- `SeasonScheduler` freezes, finalizes, and creates the next season in a single tick.
- `SeasonScheduler` with a manual `Clock` is deterministic and catches up when far behind.
- Manual seasons are skipped by the scheduler.

## Boundary Evidence

- `POST /api/v1/seasons` returns `201` with schedule and `422` for invalid durations.
- `GET /api/v1/seasons` includes `starts_at` and `schedule` when present.

## Infrastructure Evidence

- `SeasonStore` round-trip for scheduled and manual seasons.
- Existing items without schedule attributes load as manual seasons.

## Final Approval

- All gates pass before `common-sdd-verify-spec` and consolidation.
