---
change_id: CHG-0024
artifact: rollout
status: proposed
updated: 2026-10-08
---

# Rollout — Remove Tenant Concept

## Deployment

Single deploy; no feature flag. All code paths drop tenant at once.

## Data Decision (required before merge)

Existing DynamoDB items carry tenant-prefixed `PK`/`SK`. Options:

1. **Recreate table** (preferred while pre-production): drop and let the new
   key schema take over; no migration code.
2. **Rewrite keys offline**: batch-scan and re-put items with tenant-free keys;
   only if deployed data must survive.

Decision: **pending operator confirmation** — default is option 1.

## Rollback

Git revert restores tenant-scoped code; any items written with tenant-free
keys after deploy are invisible to the old code path, so rollback requires the
same data decision in reverse. Prefer fixing forward.

## Post-Deploy Checks

- Authenticated request succeeds with a token that has no `tenant_id` claim.
- `GET /api/v1/missions/offers` returns the single-scope board.
- `POST /api/v1/missions/seasons` rejects a second `ACTIVE` season globally.
