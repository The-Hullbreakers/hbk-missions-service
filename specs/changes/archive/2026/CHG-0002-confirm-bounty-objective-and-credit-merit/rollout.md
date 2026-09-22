---
change_id: CHG-0002
artifact: rollout
status: planned
updated: 2026-08-27
---

# Rollout And Rollback

## Rollout Stages

1. Deploy without published offers and validate health/isolation.
2. Create trial season `BOUNTY_HUNTING` without real prizes.
3. Publish a PvP Bounty offer with 100 merits per unique target.
4. Accept contract and confirm four unique targets.
5. Verify ledger, standing and rejection of duplicates.
6. Test transition `ACTIVE` → `FREEZING` and rejection of new credits.
7. Enable official season with reliable integration to Achievements.

## Feature Controls

- publication of objectives;
- acceptance of automatic confirmations;
- accumulation of merits;
- leaderboard projection.

Each control can stop new work without deleting credits or ledger.

## Rollback

- Disable automatic confirmations.
- Maintain reading of contracts and standings.
- Roll back the service version without deleting data.
- Correct merits through adjustments, never editing history.

## Exit Criteria

- Zero duplicates of confirmed target.
- Ledger ↔ exact standing reconciliation.
- Validated tenant isolation.
- Confirmation rollback without loss of existing data.
