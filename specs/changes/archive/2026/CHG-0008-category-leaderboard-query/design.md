---
change_id: CHG-0008
artifact: design
status: approved
updated: 2026-08-30
---

# Design

## Domain

- `Leaderboard` is a projection of `[]Standing` ordered by descending merits and deterministic tiebreakers.
- `TieBreakPolicy` v1 sort by: (1) merits, (2) missions completed, (3) unique objectives, (4) time to reach the total. In v1 it can be reduced to merits + `ConfirmationID` for determinism.
- `LeaderboardCursor` encodes opaque last seen `RecipientID`/total for stable paging.
- Existing `Track` (`INDIVIDUAL` / `TEAM`) determines the bucket.

## Application

- `ListLeaderboardHandler` (query) with ports:
  - `SeasonReader.Get(seasonID)` — validates existence.
  - `LeaderboardReader.ListStandings(tenantID, seasonID, category, track, limit, cursor, minMerits) ([]Standing, nextCursor, error)`.
  - `StandingProjector.Project(ledger, ...)` — wrapper for `ProjectStanding`.
- Validate `track` by parsing it to `domain.Track`; fails with `ErrInvalidTrack` mapped to `INVALID_QUERY_PARAMETER`.
- `min_merits` applies in the reader to discard standings with `Total < minMerits`.

## Infrastructure

- `MeritLedgerStore` adds:
  - `ListByBucket(tenantID, seasonID, category, track) ([]MeritCredit, error)` to collect credits from the bucket.
  - `ListRecipients(...)` to obtain unique containers.
- New GSI2 in `CreateMissionsTable`:
  - `GSI2PK`: `CREDIT#<tenantID>#<seasonID>#<category>#<track>`
  - `GSI2SK`: `RECIPIENT#<recipientID>#<confirmationID>`
  - Allows credits to be grouped by bucket and then by recipient.
- Alternative: reuse GSI1 with a new prefix `GSI1PK = CREDIT#...` without container and add GSI2 for individual `TotalFor`.

## Interface

- New `internal/missions/interface/http/leaderboard.go` with `LeaderboardHandler`.
- Parse `track`, `limit` (default 20, max 100), `cursor` and `min_merits` (default 0).
- Returns DTO with `items` and `next_cursor`; does not present evidence.

## Composition

- `server.go` cablea `ListLeaderboardHandler` y registra ruta:
  - `GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}`
- Inyecta `MeritLedgerStore` como `LeaderboardReader`.

## API Contract

```
GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}?track=individual|team&limit=20&cursor=<opaque>&min_merits=0
```

Response:

```json
{
  "items": [
    {
      "rank": 1,
      "recipient_id": "player-1",
      "track": "INDIVIDUAL",
      "merits": 150,
      "tie_break_facts": {
        "missions_completed": 3,
        "unique_objectives": 4
      }
    }
  ],
  "next_cursor": ""
}
```
