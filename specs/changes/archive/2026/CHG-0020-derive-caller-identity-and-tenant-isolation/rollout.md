---
change_id: CHG-0020
artifact: rollout
updated: 2026-10-08
---

# Rollout — CHG-0020

## Compatibility

- Request-body identity fields (`tenant_id`, `player_id`, `team_id`) become
  ignored, not rejected: existing clients keep working; API docs mark them
  deprecated.
- No new endpoints; no new dependencies.

## Data Migration

- `DefinitionStore` and `SeasonStore` move to tenant-keyed items.
- Environments: single-table dev/staging. Options recorded for Gate 2 decision:
  1. backfill script rewriting keys before deploy;
  2. clean redeploy (test data is disposable).
- Ledger/contract items: verify tenant is already encoded; add read-time
  assertion only (no rewrite).

## Deployment Order

1. Migrate/verify data keys.
2. Deploy Lambda with derived-identity handlers.
3. Verify: forged-body probe returns derived-identity result; per-tenant
   seasons independent.

## Rollback

- Revert deploy. If option 2 (clean redeploy) was chosen, restore table backup
  taken before migration; old code cannot read tenant-keyed items.
- Stop condition: any production evidence of identity mis-derivation blocks the
  release until fixed.
