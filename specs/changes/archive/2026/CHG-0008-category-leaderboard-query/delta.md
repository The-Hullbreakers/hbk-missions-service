---
change_id: CHG-0008
artifact: canonical-delta
status: approved
updated: 2026-08-30
---

# Canonical Delta

## ADDED — Seasonal Category Competition

### REQ-SCC-007 — Query public leaderboards by season, category and track

An authorized consumer can consult the public leaderboard of a season and category, separated by track (`INDIVIDUAL` / `TEAM`). The response shows position, total merits and public tie-breaking facts; never evidence, contracts or data from another tenant.

#### SCN-SCC-LEADERBOARD-001 — List individual leaderboard for a category

Given season `season-2026-q4` has credits `BOUNTY_HUNTING` `INDIVIDUAL` for `player-1` (150) and `player-2` (100), when a consumer calls `GET /api/v1/missions/seasons/season-2026-q4/leaderboards/BOUNTY_HUNTING?track=individual`, then it returns `player-1` in position 1 and `player-2` in position 2 with `200 OK`.

#### SCN-SCC-LEADERBOARD-002 — List team leaderboard for a category

Given season `season-2026-q4` has credits `BOUNTY_HUNTING` `TEAM` for `team-alpha` (300) and `team-beta` (200), when a consumer calls `GET /api/v1/missions/seasons/season-2026-q4/leaderboards/BOUNTY_HUNTING?track=team`, then it returns `team-alpha` in position 1 and `team-beta` in position 2 with `200 OK`.

#### SCN-SCC-LEADERBOARD-003 — Reject missing track

Given a consumer calls the endpoint without `track`, when the request is processed, it then returns `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

#### SCN-SCC-LEADERBOARD-004 — Reject invalid track

Given a consumer calls with `track=squad`, when the request is processed, then returns `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

#### SCN-SCC-LEADERBOARD-005 — Paginate with limit and cursor

Given a leaderboard with more than 20 participants, when called with `limit=10&cursor=<page1>`, then returns the next 10 standings and a new cursor.

#### SCN-SCC-LEADERBOARD-006 — Filter by minimum merits

Given participants with 0, 50 and 150 merits, when called with `min_merits=50`, then the response does not include the participant with 0 merits.

#### SCN-SCC-LEADERBOARD-007 — Tenant boundary denied

Given the consumer belongs to `tenant-a`, when it requests `leaderboards` from `tenant-b`, then it returns `403 Forbidden` with `TENANT_BOUNDARY_DENIED`.

#### SCN-SCC-LEADERBOARD-008 — Season not found

Given the season `season-unknown` does not exist, when its leaderboard is queried, it then returns `404 Not Found` with `SEASON_NOT_FOUND`.

#### SCN-SCC-LEADERBOARD-009 — Empty leaderboard

Given a season with no credits for the requested category and track, when queried, then returns `200 OK` with an empty array and `cursor` empty.

## MODIFIED

- `docs/API-CONTRACT.md` is updated with DTO response and final filters.
- DynamoDB scheme: add GSI2 to list credits per bucket (`tenant#season#category#track`) without including `RecipientID` in the hash key.

## REMOVED

- No se elimina comportamiento existente.
