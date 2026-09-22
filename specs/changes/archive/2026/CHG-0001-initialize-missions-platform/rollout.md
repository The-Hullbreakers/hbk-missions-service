---
change_id: CHG-0001
artifact: rollout
status: planned
updated: 2026-08-26
---

# Rollout And Rollback

## Rollout Stages

1. Deploy without published offers and validate health/isolation.
2. Create taxonomy, pool and a non-awarded trial season.
3. Publish few missions manually and then activate a controlled rotation.
4. Check limits, credits, corrections and leaderboards.
5. Simulate `FREEZING` and completion without issuing real achievements.
6. Enable an official season and integration with Achievements.

## Feature Controls

- publication of offers;
- automatic pool rotation;
- acceptance of contracts;
- automatic confirmation by source;
- accumulation of merits;
- completion and season awards.

Each control can stop new work without deleting contracts or ledger.

## Rollback

- Disable publishing and acceptance.
- Maintain reading of contracts and standings.
- Roll back the service version without deleting data.
- Reprocess pending posts with the same idempotent identity.
- Correct merits through adjustments, never editing history.

## Exit Criteria

- Zero tenant violations or known duplicates.
- Concurrent limit tested.
- Exact ledger ↔ leaderboard reconciliation.
- Closing rehearsal and successful awards retry.
- Approved alarms and runbooks.
