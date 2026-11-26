---
change_id: CHG-0048
artifact: design
status: implemented
updated: 2026-10-20
---

# Design — Season Statistics Summary

## Domain

- `SeasonStatistics` (new file `season_statistics.go`):
  `{SeasonID, State, Official, Buckets []SeasonBucketStatistics}`.
- `SeasonBucketStatistics`:
  `{Bucket StandingBucket, Participants int, TotalMerits int,
  Top []LeaderboardEntry}`.
- Reuses `StandingBucket`, `LeaderboardEntry`, `BuildLeaderboard`,
  `CompareStanding`, `ParseTrack`. `BuildLeaderboard` already orders by
  published tie-breaks (REQ-SCC-008); `top` is a slice of its full
  ordering.

## Application

- `GetSeasonStatisticsHandler` (new `get_season_statistics.go`) ports:
  `SeasonReader` (`Get`), `ActiveSeasonsReader` (`ListActive`, only for
  `current`), `StandingsBucketReader` (`ListBuckets`),
  `LeaderboardReader` (`ListStandings`) — all already defined.
- `Handle(seasonID, category, track string, topN int)`:
  1. `seasonID == "current"` → resolve via `ListActive` (0 →
     `ErrNoActiveSeason`, >1 → `ErrActiveSeasonAlreadyExists`); else
     `seasons.Get` (`ErrSeasonNotFound`).
  2. `ListBuckets(seasonID)`; filter by `category`/`track` (same
     `matchesFilter` semantics as standings history).
  3. Per bucket: `ListStandings` → `BuildLeaderboard(standings,
     len(standings), "")` → `Top = entries[:topN]`,
     `Participants = len(standings)`, `TotalMerits = Σ Standing.Total`.
- Invalid `track` → `ErrInvalidTrack`. `top_n` is parsed/validated at the
  boundary (int ≥ 1; default 3, max 50) → boundary emits
  `INVALID_QUERY_PARAMETER`, consistent with existing query validation.

## Boundary / Composition

- `SeasonStatisticsHandler` in `interface/http/season_statistics.go`;
  error mapping reuses `PublicCode`: `ErrSeasonNotFound`/`ErrNoActiveSeason`
  → 404, `ErrActiveSeasonAlreadyExists` → 422, `ErrInvalidTrack` → 400.
- Route in `server.go`:
  `GET /api/v1/missions/seasons/{season_id}/statistics` behind `auth`
  (no admin role — public competitive data like leaderboards).
- Response serializes the domain struct; bucket/entry fields follow the
  standings-history shape.

## Risks

- Cost: per bucket one `ListStandings` scan — acceptable (same as
  `standings/history`, unbounded seasons excluded since buckets only
  exist where credits do).
- `current` keyword collides with a literal season id — reserved already
  by `seasons/current`; treat `season_id=current` as resolution, never as
  a stored id.
