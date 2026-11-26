---
change_id: CHG-0031
artifact: delta
status: consolidated
updated: 2026-10-09
---

# Delta — Mission Completion Moderation

## What Is Changing

### Mission Contract Fulfillment

- Evidence confirmation during `IN_PROGRESS` records the confirmed unit on
  the contract (with its optional `media_uri`) but writes **no
  `MeritCredit`** — credits materialize only at reviewer approval.
- `submit` transitions `IN_PROGRESS` → `PENDING_VALIDATION` at any
  confirmation level (full or partial); the roster member's active slot is
  released at this point.
- `approve` (`admin`/`moderator`, `PENDING_VALIDATION`) → `COMPLETED` and
  atomically writes every confirmed-unit credit plus the
  `completion_reward` when all required objectives are confirmed.
- `reject` (`admin`/`moderator`, `PENDING_VALIDATION`) → `DISQUALIFIED`
  terminal; zero credits enter the ledger.
- `PENDING_VALIDATION` is non-terminal but closed to evidence, upload
  grants, and `submit`/`leave` still apply per lifecycle rules
  (evidence/upload reject it; abandon/leave allowed — last-member leave
  abandons).

### Mission Board

- `PENDING_VALIDATION` no longer counts toward the player's active-contract
  slot limit.
- It still hides the template's offers (REQ-MB-008) so a pending attempt
  cannot be re-taken.

### Seasonal Category Competition

- Unchanged — standings stay a pure projection of the ledger; credits
  simply arrive only after approval.

## MODIFIED

### REQ-MCF-003 — Confirm countable units

A valid confirmation marks the unit confirmed on the contract and returns
the computed `MeritCredit` projection for the response; the credit is
**not** persisted to the ledger at this point. All other validation rules
(season, objective, window, duplicate, media prefix/existence) are
unchanged. `ConfirmationID` semantics are unchanged.

### REQ-MCF-008 — Progress a contract through its lifecycle

`IN_PROGRESS` → `PENDING_VALIDATION` via `submit` (roster member, any
confirmation level). `PENDING_VALIDATION` → `COMPLETED` via reviewer
`approve`, or → `DISQUALIFIED` via reviewer `reject`. `PENDING_VALIDATION`
is non-terminal: `abandon` and `leave` still apply; evidence and upload
grants reject it with `CONTRACT_NOT_IN_PROGRESS`. Terminal states remain
irreversible.

### REQ-MCF-012 — Completion grants the frozen reward

The `COMPLETION` reward credit is written at `approve`, not at submit, and
only when all required objectives are confirmed at approval time. It
commits in the same atomic transaction as the deferred unit credits. The
deterministic completion identity is unchanged (idempotent retries).

### REQ-MB-003 — Limit active contracts

`PENDING_VALIDATION` no longer counts toward the player's active-contract
limit — the slot is released at `submit`. `ACCEPTED` and `IN_PROGRESS`
still count.

### REQ-MB-008 — Hide an active template

`PENDING_VALIDATION` still hides the template's offers from the member's
board and blocks re-acceptance of the same `definition_code`.

### Non-Goals (capability level)

"Disputes, manual reviews, or corrections (see CHG-0001)" is removed —
manual review is now core to fulfillment. Disputes/corrections remain out.

## ADDED

### REQ-MCF-019 — Submit a contract for review

`POST /api/v1/missions/contracts/{contract_id}/submit` transitions an
`IN_PROGRESS` contract to `PENDING_VALIDATION` for a roster member still
active on it (`403 FORBIDDEN` outsider/departed; `422
CONTRACT_NOT_IN_PROGRESS` otherwise). No minimum confirmation threshold.

### REQ-MCF-020 — Adjudicate a pending contract

`POST /api/v1/missions/contracts/{contract_id}/approve|reject` require
`admin`/`moderator` in the contract's tenant (`403 FORBIDDEN` otherwise)
and `PENDING_VALIDATION` state (`422 CONTRACT_NOT_PENDING_VALIDATION`).
`approve` returns `200` with the written credits; `reject` returns `200`.

## BDD Scenarios

### SCN-MCF-SUBMIT — Member submits for review

Given an `IN_PROGRESS` contract for `player-1`, when they call `submit`,
then the contract is `PENDING_VALIDATION` and their active slot is freed.

### SCN-MCF-SUBMIT-PARTIAL — Partial submission allowed

Given a contract with one of two required objectives confirmed, when a
member submits, then it is `PENDING_VALIDATION` awaiting review.

### SCN-MCF-SUBMIT-OUTSIDER — Outsider cannot submit

Given `player-2` outside the roster, when they call `submit`, then the
response is `403 FORBIDDEN` and the contract is unchanged.

### SCN-MCF-PENDING-CLOSED — No evidence while pending

Given a `PENDING_VALIDATION` contract, when a member submits evidence or
requests an upload grant, then it is rejected with
`CONTRACT_NOT_IN_PROGRESS`.

### SCN-MCF-APPROVE-FULL — Approval credits everything

Given a `PENDING_VALIDATION` contract with all required objectives
confirmed and `completion_reward: 200`, when a moderator approves, then
one atomic write records all unit credits plus the `COMPLETION` credit and
the contract is `COMPLETED`.

### SCN-MCF-APPROVE-PARTIAL — Approval credits only confirmed units

Given a pending contract with one confirmed unit and no full completion,
when a moderator approves, then only that unit's credit is written, no
`COMPLETION` credit exists, and the contract is `COMPLETED`.

### SCN-MCF-REJECT — Rejection credits nothing

Given a pending contract with confirmed units, when a moderator rejects,
then the contract is `DISQUALIFIED` and the standing gains zero merits.

### SCN-MCF-REVIEW-MEMBER-DENIED — Roster member cannot adjudicate

Given `player-1` owns the pending contract but holds no privileged role,
when they call `approve`, then the response is `403 FORBIDDEN`.

### SCN-MCF-REVIEW-WRONG-STATE — Adjudication needs pending state

Given an `IN_PROGRESS` contract, when a moderator calls `approve`, then it
is rejected with `422 CONTRACT_NOT_PENDING_VALIDATION`.

### SCN-MCF-REVIEW-IDEMPOTENT — Retried approval credits once

Given an approval whose write failed mid-transaction, when it is retried,
then exactly one credit set exists — duplicate `ConfirmationID`s make
repeated unit writes no-ops.

### SCN-MB-PENDING-FREES-SLOT — Pending releases the active slot

Given `player-1` at their active limit submits a contract, when it becomes
`PENDING_VALIDATION`, then a new offer can be accepted immediately.

### SCN-MB-PENDING-HIDES-TEMPLATE — Pending still blocks re-take

Given a `PENDING_VALIDATION` contract for definition `d`, when the member
lists the board, then no offer of `d` appears.

## Error Vocabulary

- `CONTRACT_NOT_PENDING_VALIDATION` — `422`
- `FORBIDDEN` — `403` (roster member adjudicating, outsider submitting)
