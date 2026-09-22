---
capability_id: CAP-SEASONAL-CATEGORY-COMPETITION
title: Seasonal Category Competition
status: current
last_consolidated_change: CHG-0010-bind-missions-to-active-season
updated: 2026-08-30
---

# Seasonal Category Competition

## Intent

Track, freeze, and publish competitive standings per season, category, and track, using the ledger of seasonal merits as the single source of truth.

## Scope

- Record every merit credit under exactly one tenant, season, category, track, and recipient.
- Separate individual and team leaderboards per category.
- Freeze and finalize seasons so no new credit can be added after the window closes.
- Derive standings from the ledger; do not maintain a second source of truth.

## Non-Goals

- Permanent player XP or level.
- Cross-season global ranking.
- Dispute resolution or manual adjustments (see CHG-0001).

## Key Concepts

- `MeritCredit`: immutable record of tenant, season, category, track, recipient, amount, and confirmation identity.
- `MeritLedger`: append-only collection of credits.
- `Standing`: projection of ledger totals for a bucket.
- `Season`: a competitive window with `SeasonState`, plus an `Official` flag that decides whether the season counts for official statistics. A season may have a `StartsAt` timestamp and a `SeasonSchedule`.
- `SeasonSchedule`: value object defining `active_weeks`, optional `off_weeks`, and a `SeasonSeries` (`prefix` + `iteration`) that derives the `SeasonID` (`{prefix}-{iteration}`).
- `NewActiveSeason`: validating constructor for an `ACTIVE` manual season.
- `NewScheduledSeason`: validating constructor for an `ACTIVE` scheduled season bound to a `StartsAt` and a `SeasonSchedule`.
- `SeasonState`: `ACTIVE`, `FREEZING`, `FINALIZED`, `ARCHIVED`.
- `Track`: `INDIVIDUAL` or `TEAM`.

## Requirements

### REQ-SCC-001 — Record Seasonal Merits

Each credit belongs to exactly one tenant, season, category, track, and recipient. No permanent user XP exists.

#### SCN-SCC-CATEGORY-CREDIT — Credit the right bucket

Given a Bounty Hunting objective is confirmed, when it is credited, then only the Bounty standing for the matching season and track increases.

### REQ-SCC-002 — Separate standings

Each category has independent individual and team leaderboards. Team points do not auto-replicate to each member's individual standing.

#### SCN-SCC-SEPARATE-TRACKS — Avoid team duplication

Given a team contract generates merits, when standings update, then the team standing increases and individual member standings do not.

### REQ-SCC-003 — Project auditable leaderboards

The leaderboard shows rank, merits, and public tie-break facts derived from the ledger. Corrections append an adjustment with a reason and never rewrite the original credit.

#### SCN-SCC-CORRECTION — Correct without hiding history

Given a credit was awarded in error, when a reviewer applies a correction, then the total changes through an auditable adjustment.

### REQ-SCC-004 — Freeze and finalize

When a season moves from `ACTIVE` to `FREEZING`, new merits stop except for eligible work started in time. After disputes resolve, a `FINALIZED` immutable standing is published. Only `Official` seasons produce official standings; unofficial seasons may still be finalized for display but do not generate permanent achievements.

#### SCN-SCC-OFFICIAL-FREEZE — Official seasons produce standings

Given an `Official` season is `FREEZING`, when all eligible credits are recorded, then the final standing counts toward official history.

#### SCN-SCC-UNOFFICIAL-FREEZE — Unofficial seasons do not produce official standings

Given an `Official: false` season is `FREEZING`, when all eligible credits are recorded, then the final standing is available for display but is not sent to Achievements as an official reward.

#### SCN-SCC-FREEZE-ACTIVE — Freeze the active season
Given a season `season-2026-q4` is `ACTIVE`, when an operator calls `POST /api/v1/missions/seasons/season-2026-q4/freeze`, then it is persisted as `FREEZING` and returned with `200 OK`.

#### SCN-SCC-FREEZE-REJECT — Reject freeze from non-active
Given a season is not `ACTIVE`, when `POST /api/v1/missions/seasons/{id}/freeze` is called, then it returns `422` with `SEASON_STATE_INVALID`.

#### SCN-SCC-FINALIZE-FREEZING — Finalize a freezing season
Given a season `season-2026-q4` is `FREEZING`, when an operator calls `POST /api/v1/missions/seasons/season-2026-q4/finalize`, then it is persisted as `FINALIZED` and returned with `200 OK`.

#### SCN-SCC-FINALIZE-REJECT — Reject finalize from non-freezing
Given a season is not `FREEZING`, when `POST /api/v1/missions/seasons/{id}/finalize` is called, then it returns `422` with `SEASON_STATE_INVALID`.

#### SCN-SCC-FREEZE-ACCEPT-IN-PROGRESS — Allow started work during freeze
Given a `FREEZING` season and a contract that is already `IN_PROGRESS`, when a unit is confirmed, then a `MeritCredit` is recorded.

#### SCN-SCC-FINALIZED-REJECT-CONFIRM — Reject confirmations after finalization
Given a `FINALIZED` season, when any unit is confirmed, then it returns `SEASON_NOT_ACTIVE` and no credit is recorded.

### REQ-SCC-005 — Create a season with an official-statistics flag

An authorized operator can create a season and decide whether it counts for official standings. The default is official. The flag is set at creation and never changes.

#### SCN-SCC-CREATE-OFFICIAL — Create an official season

Given no season `season-2026-q4` exists, when an operator creates it with `official_statistics: true`, then the season is persisted as `ACTIVE` and `Official: true` and returned with `201 Created`.

#### SCN-SCC-CREATE-UNOFFICIAL — Create a practice season

Given no season `practice-2026-q4` exists, when an operator creates it with `official_statistics: false`, then the season is persisted as `ACTIVE` and `Official: false` and returned with `201 Created`.

#### SCN-SCC-CREATE-DEFAULT — Create a season with default official flag

Given no season `season-default` exists, when an operator creates it without `official_statistics`, then the season is persisted as `ACTIVE` and `Official: true` and returned with `201 Created`.

#### SCN-SCC-CREATE-REJECT-EMPTY — Reject an empty season id

Given the operator sends an empty `season_id`, when the create request is processed, then it returns `INVALID_SEASON` with `422 Unprocessable Entity` and no season is persisted.

#### SCN-SCC-CREATE-REJECT-ALREADY-ACTIVE — Reject creation when a season is already active

Given a season already exists with `State: ACTIVE`, when an operator attempts to create a new season, then the request is rejected with `ACTIVE_SEASON_EXISTS` and no season is persisted.

### REQ-SCC-006 — List seasons by lifecycle status

Authorized consumers can list seasons and filter them by the business buckets `active`, `inactive`, and `all`. The default filter is `active` so the mission board can immediately display the current competitive window.

#### SCN-SCC-CONSUME-ACTIVE — Active season consumed by mission publishing

Given a season `season-2026-q4` is `ACTIVE`, when the mission board publishes an offer, then the offer is bound to `season-2026-q4` and no second season can be active at the same time.

#### SCN-SCC-LIST-ACTIVE-DEFAULT — List active seasons by default

Given seasons `season-2026-q4` is `ACTIVE` and `season-2026-q3` is `FINALIZED`, when a consumer calls `GET /api/v1/seasons` without `status`, then the response contains only `season-2026-q4` with `200 OK`.

#### SCN-SCC-LIST-ACTIVE — List active seasons explicitly

Given the same seasons, when a consumer calls `GET /api/v1/seasons?status=active`, then the response contains only `season-2026-q4`.

#### SCN-SCC-LIST-INACTIVE — List non-active seasons

Given the same seasons, when a consumer calls `GET /api/v1/seasons?status=inactive`, then the response contains only `season-2026-q3`.

#### SCN-SCC-LIST-ALL — List every season

Given the same seasons, when a consumer calls `GET /api/v1/seasons?status=all`, then the response contains both seasons.

#### SCN-SCC-LIST-EMPTY — Return empty when no season matches

Given no seasons exist, when a consumer calls `GET /api/v1/seasons?status=active`, then the response is `200 OK` with an empty array.

#### SCN-SCC-LIST-INVALID-STATUS — Reject unknown status values

Given any seasons, when a consumer calls `GET /api/v1/seasons?status=closed`, then the response is `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

### REQ-SCC-007 — Query public leaderboards by season, category and track

Authorized consumers can query the public leaderboard for a given season, category, and track. The response shows rank, total merits, and public tie-break facts only; it never exposes evidence, contracts, or another tenant's standings.

#### SCN-SCC-LEADERBOARD-001 — List individual leaderboard for a category

Given season `season-2026-q4` has `BOUNTY_HUNTING` `INDIVIDUAL` credits for `player-1` (150) and `player-2` (100), when a consumer calls `GET /api/v1/missions/seasons/season-2026-q4/leaderboards/BOUNTY_HUNTING?track=individual`, then `player-1` is returned at rank 1 and `player-2` at rank 2 with `200 OK`.

#### SCN-SCC-LEADERBOARD-002 — List team leaderboard for a category

Given season `season-2026-q4` has `BOUNTY_HUNTING` `TEAM` credits for `team-alpha` (300) and `team-beta` (200), when a consumer calls `GET /api/v1/missions/seasons/season-2026-q4/leaderboards/BOUNTY_HUNTING?track=team`, then `team-alpha` is returned at rank 1 and `team-beta` at rank 2 with `200 OK`.

#### SCN-SCC-LEADERBOARD-003 — Reject missing track

Given a consumer calls the endpoint without `track`, when the request is processed, then it returns `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

#### SCN-SCC-LEADERBOARD-004 — Reject invalid track

Given a consumer calls with `track=squad`, when the request is processed, then it returns `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

#### SCN-SCC-LEADERBOARD-005 — Paginate with limit and cursor

Given a leaderboard with more than 20 participants, when it is called with `limit=10&cursor=<page1>`, then it returns the next 10 standings and a new cursor.

#### SCN-SCC-LEADERBOARD-006 — Filter by minimum merits

Given participants with 0, 50, and 150 merits, when it is called with `min_merits=50`, then the participant with 0 merits is excluded.

#### SCN-SCC-LEADERBOARD-007 — Tenant boundary denied

Given the consumer belongs to `tenant-a`, when they request the leaderboard of `tenant-b`, then it returns `403 Forbidden` with `TENANT_BOUNDARY_DENIED`.

#### SCN-SCC-LEADERBOARD-008 — Season not found

Given season `season-unknown` does not exist, when its leaderboard is requested, then it returns `404 Not Found` with `SEASON_NOT_FOUND`.

#### SCN-SCC-LEADERBOARD-009 — Empty leaderboard

Given a season has no credits for the requested category and track, when it is requested, then it returns `200 OK` with an empty array and empty `next_cursor`.

### REQ-SCC-008 — Apply published tie-breaks

When participants are tied on total merits, the leaderboard resolves their order by published public facts: completed missions, then unique objectives, then recipient id. These criteria are frozen before the season begins.

#### SCN-SCC-TIE — Resolve a deterministic tie

Given two participants have the same merits, when the leaderboard is finalized, then the published tie-break facts are applied in order.

### REQ-SCC-009 — Create a scheduled season with active and off weeks

An authorized operator can create a season with a fixed `active_weeks` duration, optional `off_weeks`, and a `SeasonSeries` (`prefix` + `iteration`). The `SeasonID` is derived as `{prefix}-{iteration}`, and the season starts at the provided `starts_at` timestamp. Manual seasons continue to use an explicit `season_id`.

#### SCN-SCC-SCHEDULE-CREATE-001 — Create a scheduled season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/seasons` with `schedule.series.prefix: "season-2026"`, `schedule.series.iteration: 1`, `schedule.active_weeks: 4`, `schedule.off_weeks: 1`, and `starts_at: "2026-09-02T00:00:00Z"`, then the season is persisted as `ACTIVE` with `SeasonID: "season-2026-1"` and `201 Created`.

#### SCN-SCC-SCHEDULE-CREATE-REJECT-001 — Reject scheduled season with zero active weeks

Given any caller, when they call `POST /api/v1/seasons` with `schedule.active_weeks: 0`, then the response is `422 Unprocessable Entity` with `INVALID_SEASON` and no season is persisted.

### REQ-SCC-010 — Auto-freeze, finalize, and start the next season

A `SeasonScheduler` evaluates seasons against a `Clock`. It freezes an `ACTIVE` scheduled season when `active_weeks` have elapsed, finalizes a `FREEZING` scheduled season when `off_weeks` have elapsed, and immediately creates the next `ACTIVE` season in the series. Manual seasons are ignored by the scheduler.

#### SCN-SCC-SCHEDULE-AUTO-FREEZE-001 — Freeze when active window ends

Given a scheduled season `season-2026-1` with `starts_at` at `t0` and `active_weeks: 4`, when the scheduler tick occurs at `t0 + 4 weeks`, then the season transitions to `FREEZING`.

#### SCN-SCC-SCHEDULE-AUTO-FINALIZE-001 — Finalize and create next season

Given a `FREEZING` scheduled season `season-2026-1` with `starts_at` at `t0`, `active_weeks: 4`, and `off_weeks: 1`, when the scheduler tick occurs at `t0 + 5 weeks`, then the season transitions to `FINALIZED` and a new `ACTIVE` season `season-2026-2` is created with `starts_at` at `t0 + 5 weeks`.

### REQ-SCC-011 — Preserve in-progress credit grace during off-season

Credits for contracts that are already `IN_PROGRESS` may still be recorded while the season is `FREEZING`, regardless of whether the season is scheduled or manual.

#### SCN-SCC-SCHEDULE-IN-PROGRESS-GRACE-001 — Keep crediting in-progress contracts during off-season

Given a `FREEZING` scheduled season and a contract in `IN_PROGRESS`, when a confirmation arrives during the off-season, then the credit is recorded.

### REQ-SCC-012 — Do not auto-transition manual seasons

Seasons created without a `schedule` are manual and are not affected by the `SeasonScheduler`. They can only be frozen and finalized by explicit operator calls.

#### SCN-SCC-SCHEDULE-MANUAL-UNAFFECTED-001 — Manual season is not auto-transitioned

Given a manual season without a `schedule`, when the scheduler runs, then the season state does not change.

## Invariants

- At most one season has `State: ACTIVE` at any time.
- A finalized season is never reopened or silently rewritten.
- A new season creates a fresh marker; it does not erase the previous one.
- The `Official` flag is set at creation and never changes.
- A scheduled season's `SeasonID` is derived from its `SeasonSeries` and is immutable.
- The scheduler evaluates transitions using an explicit `Clock` port; manual seasons are never auto-transitioned.
- Standings are projections of the ledger.

## Implementation Notes

- `CreateSeasonHandler` delegates season validation to the `NewActiveSeason` domain constructor.
- `CreateScheduledSeasonHandler` delegates to `NewScheduledSeason` and derives the `SeasonID` from the series.
- `FreezeSeasonHandler` and `FinalizeSeasonHandler` are split into separate files, each owning one command/transition.
- `SeasonScheduler` uses `FreezeSeasonHandler`, `FinalizeSeasonHandler`, and `CreateScheduledSeasonHandler` to advance scheduled seasons.
- The confirmation identity recorded in `MeritCredit` is a `ConfirmationID` value object produced by `MissionContract.Confirm`.

## Related Capabilities

- `mission-contract-fulfillment` produces the `MeritCredit` events recorded here.
