---
change_id: CHG-0048
artifact: delta
status: implemented
updated: 2026-10-20
---

# Delta

## What Is Changing

### Seasonal Category Competition

- Every season (any state) exposes a statistics summary grouped by
  `{category, track}` via `GET .../seasons/{season_id}/statistics`;
  `season_id=current` resolves the single `ACTIVE` season.

## Requirements

### REQ-SCC-017 — Read season statistics summary (added)

`GET /api/v1/missions/seasons/{season_id}/statistics` returns `200` with
the season header (`SeasonID`, `State`, `Official`) and one entry per
`{category, track}` bucket carrying `participants` (recipients with
credits), `total_merits` (bucket sum), and `top` — the first `top_n`
ranked entries using the same public leaderboard projection (rank,
recipient_id, track, total, completed_missions, unique_objectives).
Optional `category` and `track` narrow the buckets; `top_n` defaults to
`3`, max `50`. `season_id=current` resolves the single `ACTIVE` season.
The summary is recomputed from the ledger for any season state — there is
no FINALIZED-only restriction. Errors: `SEASON_NOT_FOUND` (404),
`NO_ACTIVE_SEASON` (404, `current` with no active season),
`ACTIVE_SEASON_EXISTS` (422, broken single-active invariant), invalid
`track` or `top_n` → `400 INVALID_QUERY_PARAMETER`. A season with no
credits returns `200` with an empty `buckets` array. It never exposes
evidence, contract identifiers, or unit identifiers.

## BDD Scenarios

### SCN-SCC-STATS-SUMMARY — Read grouped statistics

Given `season-1` has `BOUNTY_HUNTING` `INDIVIDUAL` credits for `player-1`
(150, 2 completions) and `player-2` (100), when a consumer calls
`GET /seasons/season-1/statistics`, then it returns `200` with a
`{BOUNTY_HUNTING, INDIVIDUAL}` bucket: `participants=2`,
`total_merits=250`, and `top[0]` = `player-1` rank 1.

### SCN-SCC-STATS-CURRENT — Resolve current season

Given `season-1` is the single `ACTIVE` season, when a consumer calls
`GET /seasons/current/statistics`, then it returns `200` with
`SeasonID=season-1`.

### SCN-SCC-STATS-TOP-N — Limit top entries per bucket

Given a bucket with 5 participants, when `top_n=2` is requested, then
`top` carries only ranks 1–2 and `participants` still reports 5.

### SCN-SCC-STATS-FILTER — Narrow by track

Given buckets on both tracks, when `track=team` is requested, then only
`TEAM` buckets are returned.

### SCN-SCC-STATS-ACTIVE — Live season works

Given `season-1` is `ACTIVE`, when its statistics are requested, then it
returns `200` with the live projection (unlike `standings/history`).

### SCN-SCC-STATS-EMPTY — Season without credits

Given `season-2` exists with no credits, when statistics are requested,
then it returns `200` with `buckets: []`.

### SCN-SCC-STATS-NOT-FOUND — Unknown season

Given `season-x` does not exist, when its statistics are requested, then
it returns `404` `SEASON_NOT_FOUND`.

### SCN-SCC-STATS-INVALID — Bad query parameters

Given any season, when `track=squad` or `top_n=abc` is sent, then it
returns `400` `INVALID_QUERY_PARAMETER`.

## Error Vocabulary

- Reuses `SEASON_NOT_FOUND`, `NO_ACTIVE_SEASON`, `ACTIVE_SEASON_EXISTS`,
  `INVALID_QUERY_PARAMETER`. No new codes.
