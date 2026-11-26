---
change_id: CHG-0027
artifact: rollout
status: draft
updated: 2026-10-09
---

# Rollout — Team Member Leave

## Deployment

- Single service, additive change: new endpoint + new contract attribute.
- `Departed` is absent on existing items → reads as empty set; no migration.
- No flag needed — behavior is strictly additive for team contracts; solo
  leave is equivalent to abandon.

## Backward Compatibility

- Existing contracts without `Departed` behave exactly as before.
- `abandon` is unchanged — still available as the contract-level exit.
- API is additive: new endpoint, no modified request/response shapes.

## Rollback

- Revert the change. Persisted `Departed` attributes on items would be
  ignored by the reverted readers — but since this change introduces the
  attribute, a rollback window with departures would leave stale data.
  Acceptable: departures are rare, the attribute is ignored harmlessly, and
  a re-deploy of this change interprets them correctly again.

## Verification On Deploy

- `POST /contracts/{id}/leave` on a team contract returns `200`; survivors
  can still submit evidence; leaver gets `403` on evidence.
