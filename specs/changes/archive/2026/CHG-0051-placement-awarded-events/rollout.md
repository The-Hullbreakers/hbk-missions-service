---
change_id: CHG-0051
artifact: rollout
status: pending
updated: 2026-10-20
---

# Rollout — CHG-0051

## Order

1. Achievements ingress queue + service deployed first (consumer before
   producer; see CHG-0001 rollout).
2. This change: additive — finalization gains outbox intents + a publisher.
3. Backfill: seasons finalized before deployment have no intents; if
   historical unlocks are wanted, a reconciliation run replays facts with
   their deterministic `event_id` — decide per environment.

## Rollback

- Publisher failure or service rollback: intents stay pending; retry after
  fix. Season finalization itself is unchanged and never blocked by SNS.
- Removing the publisher stops emission without affecting standings.

## Compatibility

Additive event type on the shared topic; consumers that do not filter for
it must already ignore unknown types (Notification validates its registry).
Contracts approve/confirm endpoints gain one new stable 422 code,
`SEASON_ADJUDICATING`, returned only while a season's adjudication lock is
engaged — clients already handle the domain-error envelope; monitor
reject rates during finalization windows.
