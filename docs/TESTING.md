# Testing Strategy

## Business-Critical Properties

1. Active limit includes individual contracts and team memberships.
2. Two concurrent acceptances do not exceed limit or capacity.
3. Two offers from the same template produce a maximum of one active contract per player.
4. The board hides an active template and the command rejects bypass by ID.
5. A terminal state releases the template subject to cooldown/repeatability.
6. Reducing the limit does not cancel existing work.
7. A published version does not change accepted contracts.
8. A single target credits once.
9. Replay of confirmation, credit or placement award is idempotent.
10. Merits never cross tenant, season, category or track.
11. Team points are not duplicated in members.
12. Ledger and standings always reconcile.
13. Freeze stops rotations/credits and waits for eligible decisions.
14. A new season starts from scratch and preserves history.
15. Failure of Achievements does not reverse standings.

## Unit Scope

Domain tests limits, statuses, roster, objectives, credits, adjustments, tiebreakers and closure. Application tests orchestration with manual doubles for outgoing ports.

## Integration Scope

- `integration/http`: real boundary for pool, acceptance, contracts and standings.
- `integration/infrastructure`: real local persistence, concurrency, ledger and projection.
- Message boundary: reliable confirmations and idempotent placement awards.

Each scope runs from a clean state and independently.

## Critical E2E

Minimum journey:

1. Set limit two.
2. Publish Bounty PvP and season.
3. Accept two contracts and reject the third.
4. Confirm objectives, including a replay.
5. Verify merits and leaderboard.
6. End one to free capacity.
7. Accept another contract.
8. Freeze/end season.
9. Retry placement award without duplicating achievement.
10. Start next season at zero preserving the previous one.

## L3 Gates

Production coverage and critical scope ≥90%, mutation testing, cross-tenant security, concurrency, E2E, rollback and independent Reviewer are mandatory before consolidating.
