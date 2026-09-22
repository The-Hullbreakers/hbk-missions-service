---
change_id: CHG-0013
artifact: design
status: proposed
updated: 2026-09-02
---

# Design — Schedule Seasons With Automatic Transitions

## Overview

A scheduled season is bound to a `Clock` and a `SeasonSchedule`. The scheduler polls, freezes active seasons when the active window ends, finalizes freezing seasons at the end of the off-season, and immediately creates the next season in the series. Manual seasons remain unchanged and are never touched by the scheduler.

## Domain Model

- `SeasonSeries`: value object with `Prefix` and `Iteration`. It produces `SeasonID` (`{prefix}-{iteration}`) and `Next()`.
- `SeasonSchedule`: value object with `ActiveWeeks`, `OffWeeks`, and `Series`.
- `Season`: adds `StartsAt time.Time` and `Schedule *SeasonSchedule`. Manual seasons have `Schedule == nil`.
- `Clock`: domain/application port with `Now() time.Time`.
- Domain functions:
  - `NewScheduledSeason(seriesPrefix, iteration, startsAt, activeWeeks, offWeeks, official) (Season, error)`
  - `NewScheduledSeasonFromSchedule(schedule, startsAt, official) (Season, error)`
  - `Season.ShouldFreeze(now)`
  - `Season.ShouldFinalize(now)`
  - `Season.NextScheduledSeason()`
  - `Season.NextStartsAt()`

## Application Layer

- `CreateScheduledSeasonHandler`: validates no active season exists, builds a `Season` from a `CreateScheduledSeasonCommand`, and persists it.
- `SeasonScheduler`: owns the background loop.
  - Loads all `ACTIVE` and `FREEZING` seasons via `SeasonReader`.
  - For each `ACTIVE` season, calls `ShouldFreeze(now)`. If true, calls `FreezeSeasonHandler`.
  - For each `FREEZING` season, calls `ShouldFinalize(now)`. If true, calls `FinalizeSeasonHandler` and then creates the next scheduled season.
  - Repeats until a pass makes no transitions, so a late scheduler can catch up.
- Existing `FreezeSeasonHandler` and `FinalizeSeasonHandler` are reused; their public HTTP routes remain manual overrides.

## Boundary

- `POST /api/v1/seasons` accepts an optional `schedule` object:
  ```json
  {
    "tenant_id": "tenant-1",
    "official_statistics": true,
    "starts_at": "2026-09-02T00:00:00Z",
    "schedule": {
      "active_weeks": 4,
      "off_weeks": 1,
      "series": {
        "prefix": "season-2026",
        "iteration": 1
      }
    }
  }
  ```
- If `schedule` is absent, the request behaves exactly as today (manual season).
- Responses include `starts_at` and `schedule` when present.

## Infrastructure

- `SeasonStore` persists `StartsAt`, `ScheduleActiveWeeks`, `ScheduleOffWeeks`, `ScheduleSeriesPrefix`, and `ScheduleSeriesIteration`.
- Missing attributes are loaded as a manual season with `Schedule == nil` and `StartsAt` zero for backward compatibility.
- `SeasonStore.List` already supports `ALL`, so the scheduler can load every season and filter by state.

## Concurrency

- The scheduler is a single goroutine per process. Multiple running instances could race on freeze/finalize. Conditional writes or idempotent transitions are required.
- DynamoDB `PutItem` on `Season` overwrites the full row. The scheduler should call `Freeze`/`Finalize` through their existing handlers, which read the current row and write the new state. If two instances run concurrently, the last write wins, but the transition is still valid because `Freeze`/`Finalize` are state checks.
- The scheduler creates the next season only after the previous is `FINALIZED`, and `CreateScheduledSeasonHandler` enforces "no active season". Two concurrent instances may both see none and both try to create; DynamoDB `PutItem` on the same `SeasonID` is idempotent, and the active-season check will fail on the second if it sees the first. A small race window remains; for v1, daily season transitions do not require distributed locking.

## Backward Compatibility

- Manual seasons (`Schedule == nil`) are ignored by the scheduler.
- Existing `POST /api/v1/seasons` clients that omit `schedule` are unaffected.
- Existing `GET /api/v1/seasons` clients receive new fields they can ignore.
- DynamoDB items without schedule attributes load as manual seasons with `StartsAt` zero.

## Rollback

- Disable the scheduler via configuration (`scheduler_tick_minutes: 0`).
- Operators can still freeze/finalize manually.
- No data is erased; seasons are append-only.
