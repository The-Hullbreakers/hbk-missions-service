---
change_id: CHG-0007
artifact: canonical-delta
status: proposed
updated: 2026-08-30
---

# Canonical Delta

## MODIFIED — Seasonal Category Competition

### REQ-SCC-004 — Freeze and finalize

When a season moves from `ACTIVE` to `FREEZING`, new merits stop except for eligible work started in time. After disputes resolve, a `FINALIZED` immutable standing is published. Only `Official` seasons produce official standings; unofficial seasons may still be finalized for display but do not generate permanent achievements.

#### SCN-SCC-FREEZE-ACTIVE — Freeze the active season

Given a season `season-2026-q4` exists with `State: ACTIVE`, when an operator calls `POST /api/v1/missions/seasons/season-2026-q4/freeze`, then the season is persisted with `State: FREEZING` and returned with `200 OK`.

#### SCN-SCC-FREEZE-REJECT — Reject freeze from non-active state

Given a season `season-2026-q3` exists with `State: FINALIZED`, when an operator calls `POST /api/v1/missions/seasons/season-2026-q3/freeze`, then the response is `422 Unprocessable Entity` with `code: SEASON_STATE_INVALID` and no change is persisted.

#### SCN-SCC-FINALIZE-FREEZING — Finalize a freezing season

Given a season `season-2026-q4` exists with `State: FREEZING`, when an operator calls `POST /api/v1/missions/seasons/season-2026-q4/finalize`, then the season is persisted with `State: FINALIZED` and returned with `200 OK`.

#### SCN-SCC-FINALIZE-REJECT — Reject finalize from non-freezing state

Given a season `season-2026-q4` exists with `State: ACTIVE`, when an operator calls `POST /api/v1/missions/seasons/season-2026-q4/finalize`, then the response is `422 Unprocessable Entity` with `code: SEASON_STATE_INVALID` and no change is persisted.

#### SCN-SCC-FREEZE-ACCEPT-IN-PROGRESS — Allow confirmation for started contracts while freezing

Given season `season-2026-q4` is `FREEZING` and contract `contract-1` is `IN_PROGRESS`, when a unit is confirmed, then a `MeritCredit` is recorded and the contract is persisted.

#### SCN-SCC-FREEZE-REJECT-NOT-STARTED — Reject confirmation for contracts not in progress

Given season `season-2026-q4` is `FREEZING` and contract `contract-1` is `ACCEPTED` (not started), when a unit is confirmed, then the response is `CONTRACT_NOT_IN_PROGRESS` and no credit is recorded.

#### SCN-SCC-FINALIZED-REJECT-CONFIRM — Reject any confirmation after finalization

Given season `season-2026-q4` is `FINALIZED`, when any unit is confirmed, then the response is `SEASON_NOT_ACTIVE` and no credit is recorded.

### MODIFIED — Merit recording during non-active season

`ConfirmObjective` in `mission-contract-fulfillment` must accept `FREEZING` seasons when the contract is `IN_PROGRESS`, and continue to reject `FINALIZED` and `ARCHIVED` seasons.

## Related Requirements

- `REQ-SCC-005 — Create a season with an official-statistics flag` provides the `Official` flag that freeze/finalize preserve.
- `REQ-SCC-006 — List seasons by lifecycle status` exposes `FREEZING`/`FINALIZED` through the existing `inactive` bucket.

## Invariants

- At most one season has `State: ACTIVE` (unchanged from `REQ-SCC-005`).
- A finalized season is never reopened or silently rewritten.
- The `Official` flag is set at creation and never changes.
- Standings are projections of the ledger (unchanged).
