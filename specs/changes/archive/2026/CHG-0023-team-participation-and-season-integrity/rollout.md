---
change_id: CHG-0023
artifact: rollout
updated: 2026-10-08
---

# Rollout — CHG-0023

## Compatibility

- `team_id` on `SOLO` offers now rejects (previously silently converted) —
  clients relying on the buggy behavior get a clear `422 TEAM_NOT_ALLOWED`.
- `create season` with an existing id now `422`s instead of overwriting —
  correct behavior, previously destructive.
- Tie-break ranking may reorder tied entries — data-correct change.

## Data Migration

- None. Existing corrupted season rows (if any were rewritten pre-fix) are an
  operational audit item, not code migration.

## Deployment Order

1. Deploy conditional Put + create-exists check first (stops the bleeding).
2. Team validation and scheduler retry in the same artifact.
3. Verify E2E probes in staging.

## Rollback

- Revert deploy; conditional Put remains safe for old code (old writer simply
  overwrites again — acceptable since rollback restores previous semantics).
