---
change_id: CHG-0038
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Seasonal Category Competition

- Finalized seasons expose immutable final standings per category/track via `GET .../standings/history`.

## Requirements

### REQ-SCC-016 — Read standings history (added)

The final leaderboard of a finalized season is retrievable; unfinalized seasons return current standings [NEEDS CLARIFICATION] or 422 SEASON_STATE_INVALID.

## BDD Scenarios

### SCN-SCC-HISTORY — Read finalized standings

Given `season-1` is FINALIZED with recorded standings, when a consumer calls standings/history, then it returns 200 with the final ranks.

### SCN-SCC-HISTORY-ACTIVE — Active season history

Given `season-2` is ACTIVE, when its history is requested, then it returns the current projection or 422 per the design decision.

## Error Vocabulary

- May reuse `SEASON_STATE_INVALID` / `SEASON_NOT_FOUND`.
