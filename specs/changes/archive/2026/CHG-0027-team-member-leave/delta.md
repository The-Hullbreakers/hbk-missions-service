---
change_id: CHG-0027
artifact: delta
status: proposed
updated: 2026-10-09
---

# Delta — Team Member Leave

## What Is Changing

### Mission Contract Fulfillment

- A roster member may **leave** a non-terminal contract individually without
  ending it for the remaining members. Departed members lose the right to
  start, confirm, complete, or leave again. When the last active member
  leaves, the contract transitions to `ABANDONED`.

### Mission Board

- A departed member no longer counts the contract toward their active-slot
  limit, and offers of the same `definition_code` stop being hidden from
  them (`REQ-MB-008` applies to members still active on the contract).

## ADDED

### REQ-MCF-013 — Leave a team contract without terminating it

`POST /api/v1/missions/contracts/{contract_id}/leave` removes the caller
from the contract's active membership when the contract is non-terminal and
the caller is a current (non-departed) roster member. The response is
`200`. An outsider or an already-departed member resolves as
`403 FORBIDDEN`; a terminal contract resolves as `422` with
`CONTRACT_TERMINAL`. When the caller is the last active member, the
contract transitions to `ABANDONED` — the solo-contract leave equals
abandon.

#### SCN-MCF-LEAVE-MEMBER — Member exits, contract survives

Given a `TEAM` contract in `IN_PROGRESS` with roster `[p1, p2]`, when `p1`
leaves, then the contract stays `IN_PROGRESS`, `p1` appears in the departed
set, and `p2` may still submit evidence.

#### SCN-MCF-LEAVE-BLOCKED — Departed loses contract rights

Given `p1` left a team contract, when `p1` submits evidence, starts,
completes, or leaves again, then every operation is rejected with
`403 FORBIDDEN`.

#### SCN-MCF-LEAVE-LAST — Last member abandons

Given a team contract whose only remaining member is `p2`, when `p2`
leaves, then the contract is `ABANDONED` and terminal.

#### SCN-MCF-LEAVE-OUTSIDER — Outsider cannot leave

Given `p3` was never in the roster, when `p3` calls leave, then the response
is `403 FORBIDDEN` and the contract is unchanged.

## MODIFIED

### REQ-MCF-008 — Progress a contract through its lifecycle

In addition to start/complete/abandon/expire, a roster member may **leave**
individually. Lifecycle transitions (`start`, `complete`, `evidence`) are
restricted to members still active on the contract — departures are
excluded before the roster-member check passes.

### REQ-MB-008 — Hide an active template

A contract hides its `definition_code` only for members still active on it.
Once a member leaves, offers of that template reappear on their board and
the contract stops consuming their active-mission slot.

## HTTP Boundary Changes

- `POST /api/v1/missions/contracts/{contract_id}/leave` — new endpoint.
  `200` on success; `401` unauthenticated; `403` outsider or already
  departed; `422` for `CONTRACT_TERMINAL`.
- No request body.

## Persistence Changes

- Contract item gains a `Departed` string-set attribute — list of player ids
  who left; absent means empty. No migration needed (no live items).
