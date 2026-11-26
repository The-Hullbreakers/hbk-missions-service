---
change_id: CHG-0022
artifact: rollout
updated: 2026-10-08
---

# Rollout — CHG-0022

## Compatibility

- Endpoint and request shape unchanged; only new stable `422` codes added.

## Data Migration

- Player active counters bootstrapped lazily (`if_not_exists`) on first write;
  offer `remaining_capacity` derived from existing capacity field — verify
  existing offers carry it or backfill during deploy.

## Deployment Order

1. Deploy transaction-capable store + handler together (one artifact).
2. Verify concurrency harness in staging.

## Rollback

- Revert deploy; counter items are inert to the old code path.
