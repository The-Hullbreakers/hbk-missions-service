---
capability_id: CAP-SEASONAL-CATEGORY-COMPETITION
title: Seasonal Category Competition
status: current
last_consolidated_change: CHG-0055-local-season-preview-scenarios
updated: 2026-11-23
---

# Seasonal Category Competition

## Intent

Track, freeze, and publish competitive standings per season, category, and track, using the ledger of seasonal merits as the single source of truth.

## Scope

- Record every merit credit under exactly one season, category, track, and recipient.
- Separate individual and team leaderboards per category.
- Freeze and finalize seasons so no new credit can be added after the window closes.
- Derive standings from the ledger; do not maintain a second source of truth.

## Non-Goals

- Permanent player XP or level.
- Cross-season global ranking.
- Dispute resolution or manual adjustments (see CHG-0001).

## Key Concepts

- `MeritCredit`: immutable record of season, category, track, recipient, amount, and confirmation identity.
- `MeritLedger`: append-only collection of credits.
- `Standing`: projection of ledger totals for a bucket.
- `Season`: a competitive window with `SeasonState`, plus an `Official` flag that decides whether the season counts for official statistics. A season may have a `StartsAt` timestamp and a `SeasonSchedule`.
- `SeasonSchedule`: value object defining `active_days`, optional `off_days`, and a `SeasonSeries` (`prefix` + `iteration`) that derives the `SeasonID` (`{prefix}-{iteration}`). `active_weeks`/`off_weeks` remain accepted aliases (×7 days) for backward compatibility; day-granular series allow `iteration` 0.
- `NewActiveSeason`: validating constructor for an `ACTIVE` manual season.
- `NewScheduledSeason`: validating constructor for an `ACTIVE` scheduled season bound to a `StartsAt` and a `SeasonSchedule`.
- `SeasonState`: `ACTIVE`, `FREEZING`, `FINALIZED`, `ARCHIVED`.
- `Track`: `INDIVIDUAL` or `TEAM`.

## Requirements

### REQ-SCC-001 — Record Seasonal Merits

Each credit belongs to exactly one season, category, track, and recipient. No permanent user XP exists.

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

When finalization starts, the starter conditionally acquires an internal adjudication lock and mints `AdjudicationCutoff` + attempt id exactly once (a retry keeps them; the public state stays `FREEZING` throughout). While the lock is engaged, every season-bound merit-credit write and tie-break-relevant contract mutation MUST carry the lock check in the same write — rejected once engaged (`SEASON_ADJUDICATING`). Finalization stages immutable placement facts per REQ-SCC-018, verifies the staged count/hash, and commits one guarded constant-size transaction: `FREEZING` -> `FINALIZED` plus the deterministic MANIFEST. A pre-commit crash leaves the season locked with non-publishable staged rows; an internal scheduler sweep lists locks past a threshold for operator reconcile — resume always reuses the same attempt id and cutoff.

#### SCN-SCC-LOCK-001 — Credit racing the lock is rejected
Given a season whose adjudication lock is engaged, when a merit-credit write or tie-break-relevant mutation targets it, then it is rejected — a racing write either lands before the lock or is rejected, never both.

#### SCN-SCC-LOCK-002 — FREEZING still credits eligible in-progress work
Given a `FREEZING` season whose adjudication lock is NOT yet engaged, when an eligible in-progress unit is confirmed or a pending review approved, then the credit is recorded normally — freeze is unchanged.

#### SCN-SCC-LOCK-003 — Orphaned locked season is reconcilable
Given adjudication stopped mid-way leaving the season locked, when the operator reconciliation runs, then it is identified and can resume the same attempt id and cutoff.

### REQ-SCC-005 — Create a season with an official-statistics flag

An authorized operator can create a season and decide whether it counts for official standings. The default is official. The flag is set at creation and never changes. Creation is refused when a season with the same `season_id` already exists (`SEASON_ALREADY_EXISTS`); no state or `Official` flag of an existing season can change through creation. The "season already exists" and `ACTIVE_SEASON_EXISTS` checks are global — one platform-wide season scope.

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

#### SCN-SCC-CREATE-EXISTS — Reject overwrite of an existing season

Given `season-x` exists as `FINALIZED` with `Official: true`, when an operator creates `season-x` again, then the response is `422` with `SEASON_ALREADY_EXISTS` and the stored season is unchanged.

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

Authorized consumers can query the public leaderboard for a given season, category, and track. The response shows rank, total merits, and public tie-break facts only; it never exposes evidence or contracts.

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

#### SCN-SCC-LEADERBOARD-008 — Season not found

Given season `season-unknown` does not exist, when its leaderboard is requested, then it returns `404 Not Found` with `SEASON_NOT_FOUND`.

#### SCN-SCC-LEADERBOARD-009 — Empty leaderboard

Given a season has no credits for the requested category and track, when it is requested, then it returns `200 OK` with an empty array and empty `next_cursor`.

### REQ-SCC-008 — Apply published tie-breaks

When participants are tied on total merits, the leaderboard resolves their order by published public facts: completed missions, then unique objectives, then recipient id. `CompletedMissions` counts only contracts that reached `COMPLETED` — contracts in `IN_PROGRESS` or any other state with credits do not count. These criteria are frozen before the season begins.

#### SCN-SCC-TIE — Resolve a deterministic tie

Given two participants have the same merits, when the leaderboard is finalized, then the published tie-break facts are applied in order.

#### SCN-SCC-TIE-COMPLETED — Count only finished contracts

Given two participants tied on merits where one has an `IN_PROGRESS` credited contract and the other a `COMPLETED` credited contract, when the leaderboard applies tie-breaks, then the participant with the `COMPLETED` contract ranks first.

### REQ-SCC-009 — Create a scheduled season with active and off weeks

An authorized operator can create a season with a fixed active window, optional off-season window, and a `SeasonSeries` (`prefix` + `iteration`). Windows are expressed in days via `active_days`/`off_days`; `active_weeks`/`off_weeks` remain accepted aliases (×7 days). Providing both units in one request returns `422 INVALID_SEASON`. The `SeasonID` is derived as `{prefix}-{iteration}`, and the season starts at the provided `starts_at` timestamp. Manual seasons continue to use an explicit `season_id`.

#### SCN-SCC-SCHEDULE-CREATE-001 — Create a scheduled season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/seasons` with `schedule.series.prefix: "season-2026"`, `schedule.series.iteration: 1`, `schedule.active_weeks: 4`, `schedule.off_weeks: 1`, and `starts_at: "2026-09-02T00:00:00Z"`, then the season is persisted as `ACTIVE` with `SeasonID: "season-2026-1"` and `201 Created`.

#### SCN-SCC-SCHEDULE-DAYS-001 — Create a day-granular scheduled season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/seasons` with `schedule.active_days: 30`, `schedule.off_days: 3`, and `starts_at`, then the season is persisted `ACTIVE` with a 30-day active window and a 3-day off-season.

#### SCN-SCC-SCHEDULE-MIXED-001 — Reject mixed units

Given a request carrying both `active_weeks` and `active_days`, when it is processed, then it returns `422` `INVALID_SEASON` and no season is persisted.

#### SCN-SCC-SCHEDULE-CREATE-REJECT-001 — Reject scheduled season with zero active weeks

Given any caller, when they call `POST /api/v1/seasons` with `schedule.active_weeks: 0`, then the response is `422 Unprocessable Entity` with `INVALID_SEASON` and no season is persisted.

### REQ-SCC-010 — Auto-freeze, finalize, and start the next season

A `SeasonScheduler` evaluates seasons against a `Clock`. It freezes an `ACTIVE` scheduled season when the active window (`active_days`, or `active_weeks × 7` for week-created schedules) has elapsed, finalizes a `FREEZING` scheduled season when the off-season window has elapsed, and creates the next `ACTIVE` season in the series. A `FINALIZED` scheduled season whose successor is absent records pending next-season creation, and every scheduler tick retries it until it succeeds; the failed tick is observable. Manual seasons are ignored by the scheduler.

#### SCN-SCC-SCHEDULE-AUTO-FREEZE-001 — Freeze when active window ends

Given a scheduled season `season-2026-1` with `starts_at` at `t0` and `active_weeks: 4`, when the scheduler tick occurs at `t0 + 4 weeks`, then the season transitions to `FREEZING`.

#### SCN-SCC-SCHEDULE-AUTO-FINALIZE-001 — Finalize and create next season

Given a `FREEZING` scheduled season `season-2026-1` with `starts_at` at `t0`, `active_weeks: 4`, and `off_weeks: 1`, when the scheduler tick occurs at `t0 + 5 weeks`, then the season transitions to `FINALIZED` and a new `ACTIVE` season `season-2026-2` is created with `starts_at` at `t0 + 5 weeks`.

#### SCN-SCC-RETRY-NEXT — Retry failed next-season creation

Given a finalized scheduled season whose next-season creation failed once, when the scheduler ticks again, then the next season is created `ACTIVE` and the series resumes.

### REQ-SCC-011 — Preserve in-progress credit grace during off-season

Credits for contracts that are already `IN_PROGRESS` may still be recorded while the season is `FREEZING`, regardless of whether the season is scheduled or manual.

#### SCN-SCC-SCHEDULE-IN-PROGRESS-GRACE-001 — Keep crediting in-progress contracts during off-season

Given a `FREEZING` scheduled season and a contract in `IN_PROGRESS`, when a confirmation arrives during the off-season, then the credit is recorded.

### REQ-SCC-012 — Do not auto-transition manual seasons

Seasons created without a `schedule` are manual and are not affected by the `SeasonScheduler`. They can only be frozen and finalized by explicit operator calls.

#### SCN-SCC-SCHEDULE-MANUAL-UNAFFECTED-001 — Manual season is not auto-transitioned

Given a manual season without a `schedule`, when the scheduler runs, then the season state does not change.

### REQ-SCC-013 — Store and transition seasons in a single scope

Season persistence is single-scope. The single-`ACTIVE` invariant, freeze, finalize, and scheduler transitions operate on the global season list.

#### SCN-REM-TENANT-GLOBAL-SEASON — One active season for everyone

Given season `season-2026-q4` is `ACTIVE`, when any authenticated operator creates another season, then the request is rejected with `ACTIVE_SEASON_EXISTS` regardless of caller.

### REQ-SCC-014 — Read the current season

`GET /api/v1/missions/seasons/current` returns `200` with the single `ACTIVE` season. When none exists it returns `404` with `NO_ACTIVE_SEASON`; a broken single-active invariant returns `422` `ACTIVE_SEASON_EXISTS`.

#### SCN-SCC-CURRENT — Current season

Given `season-1` is `ACTIVE`, when a consumer calls `GET /missions/seasons/current`, then it returns `200` with `season-1`.

#### SCN-SCC-CURRENT-NONE — No active season

Given no `ACTIVE` season exists, when current is requested, then it returns `404` `NO_ACTIVE_SEASON`.

### REQ-SCC-015 — Read own standing

`GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me?track=individual|team` returns `200` with the caller's ranked entry (rank, total, tie-break facts) in the bucket. The individual track resolves by caller `player_id`; the team track resolves the recipient against the caller's `team_ids`. A missing season returns `404` `SEASON_NOT_FOUND`, an invalid `track` returns `400` `INVALID_QUERY_PARAMETER`, and a caller absent from the bucket returns `404` `STANDING_NOT_FOUND`.

#### SCN-SCC-ME — Own standing

Given `player-1` has 150 merits in BOUNTY_HUNTING INDIVIDUAL for `season-1`, when they call `.../leaderboards/BOUNTY_HUNTING/me?track=individual`, then it returns `200` with their rank and total.

#### SCN-SCC-ME-ABSENT — No standing yet

Given `player-9` has no merits in the bucket, when they call `.../me`, then it returns `404` `STANDING_NOT_FOUND`.

### REQ-SCC-016 — Read standings history

`GET /api/v1/missions/seasons/{season_id}/standings/history` returns `200` with the final ranked standings of a `FINALIZED` season, grouped by `{category, track}` bucket with the same rank/total/tie-break projection as the leaderboard. History is recomputed from the immutable ledger — no separate snapshot is stored. Optional `category` and `track` query parameters narrow the buckets. Only `FINALIZED` seasons serve history; other states return `422` `SEASON_STATE_INVALID`, a missing season returns `404` `SEASON_NOT_FOUND`, and an invalid `track` returns `400` `INVALID_QUERY_PARAMETER`.

#### SCN-SCC-HISTORY — Read finalized standings

Given `season-1` is `FINALIZED` with recorded standings, when a consumer calls `standings/history`, then it returns `200` with the final ranks grouped by bucket.

#### SCN-SCC-HISTORY-ACTIVE — Active season history rejected

Given `season-2` is `ACTIVE`, when its history is requested, then it returns `422` `SEASON_STATE_INVALID` — live standings stay on the leaderboard endpoint.

### REQ-SCC-017 — Read season statistics summary

`GET /api/v1/missions/seasons/{season_id}/statistics` returns `200` with the season header (`SeasonID`, `State`, `Official`) and one entry per `{category, track}` bucket carrying `participants` (recipients with credits), `total_merits` (bucket sum), and `top` — the first `top_n` ranked entries using the same public leaderboard projection (rank, recipient_id, track, total, completed_missions, unique_objectives). Optional `category` and `track` narrow the buckets; `top_n` defaults to `3`, max `50`. `season_id=current` resolves the single `ACTIVE` season. The summary is recomputed from the ledger for any season state — there is no FINALIZED-only restriction. Errors: `SEASON_NOT_FOUND` (404), `NO_ACTIVE_SEASON` (404, `current` with no active season), `ACTIVE_SEASON_EXISTS` (422, broken single-active invariant), invalid `track` or `top_n` → `400 INVALID_QUERY_PARAMETER`. A season with no credits returns `200` with an empty `buckets` array. It never exposes evidence, contract identifiers, or unit identifiers.

#### SCN-SCC-STATS-SUMMARY — Read grouped statistics

Given `season-1` has `BOUNTY_HUNTING` `INDIVIDUAL` credits for `player-1` (150, 2 completions) and `player-2` (100), when a consumer calls `GET /seasons/season-1/statistics`, then it returns `200` with a `{BOUNTY_HUNTING, INDIVIDUAL}` bucket: `participants=2`, `total_merits=250`, and `top[0]` = `player-1` rank 1.

#### SCN-SCC-STATS-CURRENT — Resolve current season

Given `season-1` is the single `ACTIVE` season, when a consumer calls `GET /seasons/current/statistics`, then it returns `200` with `SeasonID=season-1`.

#### SCN-SCC-STATS-TOP-N — Limit top entries per bucket

Given a bucket with 5 participants, when `top_n=2` is requested, then `top` carries only ranks 1–2 and `participants` still reports 5.

#### SCN-SCC-STATS-FILTER — Narrow by track

Given buckets on both tracks, when `track=team` is requested, then only `TEAM` buckets are returned.

#### SCN-SCC-STATS-ACTIVE — Live season works

Given `season-1` is `ACTIVE`, when its statistics are requested, then it returns `200` with the live projection (unlike `standings/history`).

#### SCN-SCC-STATS-EMPTY — Season without credits

Given `season-2` exists with no credits, when statistics are requested, then it returns `200` with `buckets: []`.

#### SCN-SCC-STATS-NOT-FOUND — Unknown season

Given `season-x` does not exist, when its statistics are requested, then it returns `404` `SEASON_NOT_FOUND`.

#### SCN-SCC-STATS-INVALID — Bad query parameters

Given any season, when `track=squad` or `top_n=abc` is sent, then it returns `400` `INVALID_QUERY_PARAMETER`.

### REQ-SCC-018 — Publish adjudicated placements on official finalization

When a season with `Official: true` is adjudicated and committed to `FINALIZED`, Missions emits one `missions.season.placement-awarded` v1.0 fact per ranked entry of every `{category, track}` bucket — at any scale. Immutable facts are staged during adjudication, verified, and committed as a durable MANIFEST intent with the transition; the publisher emits each staged fact and tracks progress until drained. Each fact carries `season_id`, `category`, `track`, `recipient_id`, `rank`, `final_merits`; TEAM facts add `member_ids` — the union of frozen rosters of exactly the contracts whose credits contributed to that ranked entry (uncredited roster members and current membership are never used). Unofficial seasons emit nothing. `occurred_at` is the adjudication cutoff minted when the lock engaged — never the publish time; republication repeats the same `event_id`. Which ranks produce achievements is Achievements' decision, never Missions'.

#### SCN-SCC-AWARD-001 — Official season emits ranked facts
Given an official `FINALIZED` season where `player-1` ranks 1 in `BOUNTY_HUNTING` `INDIVIDUAL`, when the manifest drains, then a `placement-awarded` fact exists with `rank: 1`, merits, and cutoff `occurred_at`.

#### SCN-SCC-AWARD-002 — Unofficial season emits nothing
Given an unofficial season reaches `FINALIZED`, when finalization completes, then no `placement-awarded` fact is emitted for any bucket.

#### SCN-SCC-AWARD-003 — Team fact carries credited frozen rosters only
Given team `team-alpha` ranks 1 in `BOUNTY_HUNTING` `TEAM` where contract `c-1` contributed credits and `c-2` (same team, same season) none, when the fact is emitted, then `member_ids` lists exactly `c-1`'s frozen roster — `c-2`'s members are excluded.

#### SCN-SCC-AWARD-004 — Fact emission survives publication failure
Given a manifest publish fails mid-drain, when the publisher retries, then published facts repeat the same `event_id` — no duplicates.

#### SCN-SCC-AWARD-005 — Large standings finalize and publish fully
Given an official `FREEZING` season with 200 ranked entries, when the operator finalizes, then the snapshot is staged (beyond a single transaction), the manifest commits `FINALIZED`, and all 200 facts are eventually published.

#### SCN-SCC-AWARD-006 — Interrupted staging resumes the same snapshot
Given a crash after staging, when finalization retries, then attempt id/cutoff are reused, identical digests replay as no-ops, the manifest commits once under its deterministic key — nothing publishable before.

#### SCN-SCC-AWARD-007 — Divergent staged content is an integrity error
Given a staged fact row exists whose digest differs from the freshly derived fact for the same key, when finalization retries, then the write is refused — never overwritten; the attempt blocks for reconciliation.

### REQ-SCC-019 — Preview local seasonal standings safely

Local-only preview datasets model off-season (no active season or credits), fresh-start (one active unofficial season without credits), and mid-season (one active unofficial season with category-scoped individual merit credits for active local members). Each preview occupies a separate fixed local missions dataset; the original missions dataset and identity records are never changed. Loading defaults to a read-only preview, requires explicit apply for inserts, rejects non-loopback destinations or unexpected/conflicting rows, and repeats unchanged data without duplicating credits. No preview merits become official results.

#### SCN-SCC-LOCAL-PHASES — Compare seasonal standings without replacing local history

Given a reviewer has an existing local season, when they select each local preview phase, then the current-season and standings readers show that phase while the existing season remains available on its original dataset.

### REQ-SCC-020 — Rank competitors within one season

Sum merits across all categories for each distinct recipient on the
INDIVIDUAL track in one season. Rank by total descending, then distinct
completed contract count, then distinct category/objective pairs, then
recipient ID as the deterministic final tie-break. A competitor with
credits in multiple categories appears once; TEAM credits and other
seasons never contribute. Participants are distinct credited recipients,
not the sum of category participant counts. Existing category rankings
are unchanged.

#### SCN-SCC-OVERALL-001 — Reward progress across categories

Given Mira has 120 merits in air combat and 80 in racing while Kai has
190 in air combat in the same season, when a member views the overall
individual standings, then Mira leads with 200 merits and Kai follows
with 190.

#### SCN-SCC-OVERALL-002 — Keep tracks and seasons separate

Given Mira also has team merits and merits from an earlier season, when
a member views this season's individual standings, then those merits do
not change Mira's overall position.

#### SCN-SCC-OVERALL-003 — Resolve ties predictably

Given two competitors have equal merits, when their overall positions
are determined, then distinct completed contracts, category/objective
pairs, and finally stable recipient identity determine their order.

### REQ-SCC-021 — Show my honest position

The same overall result contains the authenticated member's position,
derived from their session identity, even when they are outside the
requested top page. A member without credited merits has no rank, not
rank zero. No unrelated identity, evidence, contracts, or private
metadata is returned.

#### SCN-SCC-OVERALL-004 — Show a position outside the top page

Given Kai is ranked 23rd and the top page shows ten members, when Kai
views the overall standings, then Kai sees their true position 23
alongside the top ten.

#### SCN-SCC-OVERALL-005 — Do not invent a first position

Given a new season has no credited merits, when a member views the
overall standings, then the page is empty and their position is absent.

#### SCN-SCC-OVERALL-006 — Preserve safe access and navigation

Given a season has multiple pages of ranked competitors, when an
authenticated member requests subsequent pages, then rows retain their
deterministic order without duplicates; an unauthenticated member sees
no standings.

## Invariants

- At most one season has `State: ACTIVE` at any time.
- A finalized season is never reopened or silently rewritten.
- A new season creates a fresh marker; it does not erase the previous one.
- The `Official` flag is set at creation and never changes.
- A scheduled season's `SeasonID` is derived from its `SeasonSeries` and is immutable.
- The scheduler evaluates transitions using an explicit `Clock` port; manual seasons are never auto-transitioned.
- Standings are projections of the ledger.
- Overall standings aggregate only INDIVIDUAL-track credits of the
  requested season; a contract counts as completed once even when
  credited in several categories.
- Staged placement facts are never publishable until the digest-verified manifest commits; an adjudication attempt id and cutoff are minted once per lock.
- While the adjudication lock is engaged, every season-bound credit or tie-break-relevant write carries the lock check in the same transaction.

## Implementation Notes

- `CreateSeasonHandler` delegates season validation to the `NewActiveSeason` domain constructor.
- `CreateScheduledSeasonHandler` delegates to `NewScheduledSeason` and derives the `SeasonID` from the series.
- `FreezeSeasonHandler` and `FinalizeSeasonHandler` are split into separate files, each owning one command/transition.
- `SeasonScheduler` uses `FreezeSeasonHandler`, `FinalizeSeasonHandler`, and `CreateScheduledSeasonHandler` to advance scheduled seasons.
- `GetSeasonStatisticsHandler` recomputes the REQ-SCC-017 summary from the ledger per `{category, track}` bucket, resolving `current` to the single `ACTIVE` season; `season_statistics.go` (domain) owns the projection types and `interface/http/season_statistics.go` maps parameters and error codes.
- `GetOverallStandingsHandler` recomputes the REQ-SCC-020/021 result per `{season}` from INDIVIDUAL credit facts; `overall_standings.go` (domain) owns `ProjectOverallStandings`, `overall_credit_facts.go` (dynamodb) reads the facts, and `interface/http/overall_standings.go` maps limits, cursor, and error codes at `GET /api/v1/missions/seasons/{id}/standings/overall`.
- The confirmation identity recorded in `MeritCredit` is a `ConfirmationID` value object produced by `MissionContract.Confirm`.

## Related Capabilities

- `mission-contract-fulfillment` produces the `MeritCredit` events recorded here.
