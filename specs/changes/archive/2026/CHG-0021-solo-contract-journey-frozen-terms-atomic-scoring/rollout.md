---
change_id: CHG-0021
artifact: rollout
updated: 2026-10-08
---

# Rollout — CHG-0021

## Compatibility

- New endpoints (`start`, `complete`, `abandon`) are additive.
- Contract responses gain `terms` and new `contract_id` format; clients must
  treat `contract_id` as opaque (documented in API contract).

## Data Migration

- Pre-change `ACCEPTED` contracts lack `terms` and use the old `ConfirmationID`
  encoding. Gate 2 decision: backfill snapshot from the recorded
  `definition_version`, or invalidate pre-change non-terminal contracts.
- Dedup markers under the old delimiter must remain parseable (dual-parse or
  keyed migration) so historical credits stay deduplicated.

## Deployment Order

1. Deploy persistence + domain (terms columns optional/nullable).
2. Migrate or invalidate legacy contracts.
3. Enable lifecycle routes.
4. Verify journey E2E in staging.

## Rollback

- Revert deploy; new columns are ignored by old binary except `contract_id`
  format — avoid mixing old/new writers mid-rollout; rollback requires no new
  acceptances having used the new id format, or a dedup-aware fallback.
