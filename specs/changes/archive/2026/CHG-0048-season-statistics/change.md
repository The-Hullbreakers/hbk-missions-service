---
change_id: CHG-0048
title: Expose season statistics summary grouped by category and track
status: completed
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/CHG-0048-season-statistics/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
context_scope: existing
created: 2026-10-20
updated: 2026-10-20
---
# Expose season statistics summary grouped by category and track

## Intent

Frontend needs one read to render a season dashboard — current or past —
without N leaderboard calls. This change adds a per-season statistics
summary grouped by `{category, track}` bucket with top recipients.

## Business Value

- One call powers season dashboards, player cards, and archive views.
- Works on `ACTIVE` seasons (live) unlike `standings/history`
  (FINALIZED-only).

## User Stories

- **US-SCC-STATS**: As a player, I want a season summary with the top
  recipients per category and track, so I can see competition state at a
  glance.

## Why L2

New public read surface over an existing canonical capability; data is a
projection of the immutable `MeritLedger` (no second source of truth, no
scoring invariant change, no trust/identity change).

## Scope

- **Application**: `GetSeasonStatisticsHandler` — resolves `current` or
  explicit `season_id`, lists buckets, builds ranked entries, slices top-N.
- **Domain**: `SeasonStatistics` / `SeasonBucketStatistics` projection
  types; reuse `StandingBucket`, `LeaderboardEntry`, `BuildLeaderboard`,
  `ParseTrack`.
- **Boundary**: `GET /api/v1/missions/seasons/{season_id}/statistics`
  with `category`, `track`, `top_n` filters; `season_id=current` resolves
  the single ACTIVE season.
- **Infrastructure**: none — reuses `MeritLedgerStore.ListBuckets` /
  `ListStandings` and `SeasonStore` reads.

## Non-Goals

- No pagination inside buckets (`top_n` caps each list; the leaderboard
  endpoint remains the paginated deep-dive).
- No aggregated season totals across categories beyond per-bucket facts.
- No change to `standings/history` FINALIZED-only semantics.

## Approvals

- Gate 1: approved 2026-10-20 — D1 `season_id=current` resolves the
  single ACTIVE season (`NO_ACTIVE_SEASON` 404); D2 response groups by
  `{category, track}` bucket with `participants`, `total_merits`, and
  `top` ranked entries (same public leaderboard projection); D3 optional
  `category`, `track`, `top_n` (default 3, max 50); D4
  `SEASON_NOT_FOUND`→404, `NO_ACTIVE_SEASON`→404, bad `track`/`top_n`→400
  `INVALID_QUERY_PARAMETER`. Serves `ACTIVE`, `FREEZING`, `FINALIZED`,
  `ARCHIVED` — no state restriction.
