---
change_id: CHG-0013
artifact: rollout
status: proposed
updated: 2026-09-02
---

# Rollout — Schedule Seasons With Automatic Transitions

## Pre-deployment

1. Ensure existing manual seasons have no `Schedule` attributes so they load as manual.
2. Add `scheduler_tick_minutes` to service configuration (default `1`).
3. Verify DynamoDB table does not require schema changes (attributes added as needed).

## Deployment

1. Deploy new binary with `Season` schedule fields and `SeasonScheduler`.
2. Start the service. The scheduler goroutine begins on `NewServer`.
3. No table migration is required; missing attributes are treated as manual.

## Post-deployment

1. Create the first scheduled season via `POST /api/v1/seasons` with `schedule`.
2. Verify `GET /api/v1/seasons` returns `starts_at` and `schedule`.
3. Wait for the first active window to end and confirm auto-freeze, auto-finalize, and next season creation.
4. Manual freeze/finalize endpoints remain available as overrides.

## Rollback

1. Set `scheduler_tick_minutes: 0` to stop the scheduler.
2. Operators can still manage seasons manually.
3. No data is erased; the rollback only disables automatic transitions.

## Compatibility

- Existing manual seasons are unaffected.
- Existing clients that omit `schedule` in `POST /api/v1/seasons` continue to create manual seasons.
- `GET /api/v1/seasons` adds new optional fields; older clients can ignore them.
