---
change_id: CHG-0005
artifact: canonical-delta
status: proposed
updated: 2026-08-30
---

# Canonical Delta

## MODIFIED — Seasonal Category Competition

### REQ-SCC-005 — Create a season with an official-statistics flag

Only one season may be `ACTIVE` at any time. Creating a new season while another is `ACTIVE` is rejected.

#### SCN-SCC-CREATE-REJECT-ALREADY-ACTIVE — Reject creation when a season is already active

Given a season `season-2026-q4` exists with `State: ACTIVE`, when an operator attempts to create `season-2026-q5`, then the request is rejected with `ACTIVE_SEASON_EXISTS` and no new season is persisted.

#### SCN-SCC-CREATE-ALLOW-NON-ACTIVE — Allow creation when no active season exists

Given no season is `ACTIVE`, when an operator creates `season-2026-q5`, then the request succeeds and the new season is `ACTIVE`.

## Invariants

- At most one season has `State: ACTIVE`.
- `FREEZING`, `FINALIZED`, and `ARCHIVED` states do not count against the active-season limit.
