---
change_id: CHG-0021
artifact: delta
status: consolidated
updated: 2026-10-08
---

# Delta — Solo Journey, Frozen Terms, Atomic Scoring

## What Is Changing

### Mission Board

- Definitions carry concrete objectives and reward rules; acceptance embeds a
  `ContractTerms` snapshot (objectives, merit values, unit limits, bonuses) at
  the accepted `definition_version`.
- Every acceptance creates a new contract attempt identity; a completed,
  abandoned, expired, or failed attempt never blocks or rewrites another.

### Mission Contract Fulfillment

- Contracts expose lifecycle transitions: start, complete, abandon, and
  automatic expiry at the offer window end.
- Confirmation executes against the contract's frozen `ContractTerms`, not a
  global objective lookup.
- Contract state and `MeritCredit` commit in one atomic write; a retried
  confirmation after a partial failure is idempotent, not a silent loss.
- `ConfirmationID` uses a delimiter that cannot collide with generated ids.
- Objective `Bonus` and `UnitLimit` persist and are enforced.

## ADDED

### REQ-MCF-008 — Progress a contract through its lifecycle

An `ACCEPTED` contract can be started by a roster member (`IN_PROGRESS`),
finished once its required objectives are confirmed (`COMPLETED`), abandoned by
the player (`ABANDONED`), or expired when the window closes (`EXPIRED`).
Terminal states are irreversible and keep reason and evidence history.

### REQ-MCF-009 — Score against frozen contract terms

Confirmation resolves the objective and merit value from the contract's
`ContractTerms` snapshot captured at acceptance. Later definition or offer
edits do not change what an existing contract pays.

### REQ-MCF-010 — Commit confirmation and credit atomically

Persisting the confirmation and appending the `MeritCredit` is one atomic
operation. If the write fails, no partial state remains; a retry of the same
logical confirmation produces the same single credit.

### REQ-MCF-011 — Give every attempt its own identity

Each acceptance generates a unique `ContractID` (attempt identity), so
re-accepting an offer after a terminal attempt creates a new record instead of
overwriting history.

### REQ-MB-013 — Publish definitions with objectives and reward rules

A `MissionDefinition` declares its objectives with type, category, countable
unit, merit value, optional `bonus`, and optional `unit_limit`. Acceptance
copies them into the contract terms.

## MODIFIED

### REQ-MB-005 — Freeze promised rules

The contract keeps not only `definition_id`/`definition_version` but the full
objective and reward snapshot effective at acceptance.

### REQ-MCF-004 — Prevent duplicates and basic farming

Unit limits and bonuses are actually persisted and enforced; duplicate
detection uses a `ConfirmationID` that always parses.

## BDD Scenarios

### SCN-MCF-START — Start an accepted contract

Given player `player-1` holds an `ACCEPTED` contract, when they start it, then
the contract is `IN_PROGRESS` and objectives become confirmable.

### SCN-MCF-COMPLETE — Finish with required objectives confirmed

Given all required objectives are confirmed, when the player completes the
contract, then it is `COMPLETED`, slots are released, and history is preserved.

### SCN-MCF-ABANDON — Abandon releases the slot

Given an `IN_PROGRESS` contract, when the player abandons it, then it is
`ABANDONED` and no longer counts toward the active limit.

### SCN-MCF-FROZEN — Republished definition does not alter a live contract

Given a contract was accepted at version 1, when the operator republishes
version 2 with different merits, then the live contract still scores version-1
objectives and merits.

### SCN-MCF-WRONG-OBJECTIVE — Reject an objective not in the snapshot

Given a contract's snapshot has objective `obj-a`, when evidence references
`obj-z` (same category, different objective), then it is rejected with
`OBJECTIVE_NOT_IN_CONTRACT` and no credit is recorded.

### SCN-MCF-ATOMIC-RETRY — Retry after ledger failure

Given a confirmation failed mid-write, when the same confirmation is retried,
then it completes once and the standing reflects exactly one credit.

### SCN-MCF-ATTEMPT — Re-accept keeps both attempts

Given `player-1` abandoned a contract for offer `off-1`, when they accept it
again, then a second contract with a distinct `contract_id` is created and the
abandoned attempt remains queryable.

### SCN-MCF-UNIT-LIMIT — Enforce the unit cap

Given an objective with `unit_limit: 2` and two confirmations recorded, when a
third confirmation arrives, then it is rejected with `UNIT_LIMIT_EXCEEDED`.

## Error Vocabulary

- `CONTRACT_STATE_INVALID` — `422` (illegal transition)
- `OBJECTIVE_NOT_IN_CONTRACT` — `422`
- `UNIT_LIMIT_EXCEEDED` — `422`
- `CONTRACT_EXPIRED` — `422`
- `DUPLICATE_CONFIRMATION` — `422` (existing)
