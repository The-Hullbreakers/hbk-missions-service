---
change_id: CHG-0037
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Seasonal Category Competition

- `GET /api/v1/missions/seasons/current` returns the single ACTIVE season or 404.
- `GET .../leaderboards/{category}/me` returns the caller's standing (rank, total, tie-break facts) or 404 when absent.

## Requirements

### REQ-SCC-014 — Read the current season (added)

The current ACTIVE season is retrievable without filtering a list.

### REQ-SCC-015 — Read own standing (added)

A competitor reads their own leaderboard position for a season/category/track.

## BDD Scenarios

### SCN-SCC-CURRENT — Current season

Given `season-1` is ACTIVE, when a consumer calls GET /missions/seasons/current, then it returns 200 with `season-1`.

### SCN-SCC-CURRENT-NONE — No active season

Given no ACTIVE season exists, when current is requested, then it returns 404 NO_ACTIVE_SEASON.

### SCN-SCC-ME — Own standing

Given player-1 has 150 merits in BOUNTY_HUNTING INDIVIDUAL for season-1, when they call .../leaderboards/BOUNTY_HUNTING/me?track=individual, then it returns 200 with their rank and total.

## Error Vocabulary

- Reuses `NO_ACTIVE_SEASON` — 404; `STANDING_NOT_FOUND` — 404 (new).
