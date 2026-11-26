---
change_id: CHG-0037
title: Expose current season and caller standing
status: completed
risk: L1
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/2026/CHG-0037-current-season-and-my-standing/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
context_scope: existing
created: 2026-10-14
updated: 2026-10-14
---
# Expose current season and caller standing

## Intent

Two documented reads are missing: the current active season and the caller's own standing in a leaderboard. Both are plain queries over existing stores.

## Business Value

- The mission board can resolve the active season without listing.
- A player sees their rank without scanning all pages.

## User Stories

- **US-SCC-CURRENT**: As a client, I want the current active season, so the board shows live competition context.
- **US-SCC-ME**: As a competitor, I want my standing in a category, so I can see my rank and totals directly.

## Why L1

Additive read-only endpoints reusing existing readers; no domain or persistence change.

## Scope

- **Application**: `GetCurrentSeasonHandler`, `GetMyStandingHandler` queries.
- **Boundary**: `GET /api/v1/missions/seasons/current`, `GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me`.

## Non-Goals

- Historical standings (CHG-0038).
- Cross-category aggregate standing.

## Approvals

- Gate 1: approved 2026-10-17 — reuse `ActiveSeasonsReader`/`SeasonReader`/`LeaderboardReader`; `NO_ACTIVE_SEASON`→404, `ACTIVE_SEASON_EXISTS`→422, `SEASON_NOT_FOUND`→404; new `ErrStandingNotFound`→404 `STANDING_NOT_FOUND`; rank via `CompareStanding`; team track resolves `RecipientID` against caller `team_ids`.
