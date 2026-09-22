---
change_id: CHG-0013
artifact: delta
status: proposed
updated: 2026-09-02
---

# Delta — Schedule Seasons With Automatic Transitions

## What Is Changing

### Seasonal Category Competition

- `Season` gains `StartsAt` and an optional `SeasonSchedule`.
- `SeasonSchedule` contains `ActiveWeeks`, `OffWeeks`, and a `SeasonSeries` (`Prefix`, `Iteration`).
- `SeasonID` is derived as `{prefix}-{iteration}` for scheduled seasons.
- `Season.ShouldFreeze(now)` and `Season.ShouldFinalize(now)` evaluate the schedule against a clock.
- `Season.NextScheduledSeason()` produces the next `ACTIVE` season in the series.
- A `SeasonScheduler` application service runs in the background, freezes, finalizes, and creates the next season at the right instants.
- `POST /api/v1/seasons` accepts a `schedule` block.
- `GET /api/v1/seasons` returns `starts_at` and `schedule` when present.

## BDD Scenarios

### SCN-SCC-SCHEDULE-CREATE-001 — Create a scheduled season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/seasons` with `schedule.series.prefix: "season-2026"`, `schedule.series.iteration: 1`, `schedule.active_weeks: 4`, `schedule.off_weeks: 1`, and `starts_at: "2026-09-02T00:00:00Z"`, then the season is persisted as `ACTIVE` with `SeasonID: "season-2026-1"` and `201 Created`.

### SCN-SCC-SCHEDULE-CREATE-REJECT-001 — Reject scheduled season with zero active weeks

Given any caller, when they call `POST /api/v1/seasons` with `schedule.active_weeks: 0`, then the response is `422 Unprocessable Entity` with `INVALID_SEASON` and no season is persisted.

### SCN-SCC-SCHEDULE-AUTO-FREEZE-001 — Freeze when active window ends

Given a scheduled season `season-2026-1` with `starts_at` at `t0` and `active_weeks: 4`, when the scheduler tick occurs at `t0 + 4 weeks`, then the season transitions to `FREEZING`.

### SCN-SCC-SCHEDULE-AUTO-FINALIZE-001 — Finalize and create next season

Given a `FREEZING` scheduled season `season-2026-1` with `starts_at` at `t0`, `active_weeks: 4`, and `off_weeks: 1`, when the scheduler tick occurs at `t0 + 5 weeks`, then the season transitions to `FINALIZED` and a new `ACTIVE` season `season-2026-2` is created with `starts_at` at `t0 + 5 weeks`.

### SCN-SCC-SCHEDULE-IN-PROGRESS-GRACE-001 — Keep crediting in-progress contracts during off-season

Given a `FREEZING` scheduled season and a contract in `IN_PROGRESS`, when a confirmation arrives during the off-season, then the credit is recorded.

### SCN-SCC-SCHEDULE-MANUAL-UNAFFECTED-001 — Manual season is not auto-transitioned

Given a manual season without a `schedule`, when the scheduler runs, then the season state does not change.

## Error Vocabulary

- `INVALID_SEASON`
- `ACTIVE_SEASON_ALREADY_EXISTS`
- `SEASON_NOT_FOUND`
