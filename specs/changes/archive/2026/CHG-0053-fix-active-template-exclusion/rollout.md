---
change_id: CHG-0053
artifact: rollout
status: pending
updated: 2026-10-26
---

# Rollout — CHG-0053

## Legacy-Data Gap — Strict Invariant Not Enforced Pre-Reconciliation

Contracts accepted before deployment hold no reservation items. For those
contracts the commit-time uniqueness guard has nothing to conflict with:
code-first deployment alone **does not** fully enforce the
per-player+definition invariant on legacy contracts. The decision-path
checks (`DecideAcceptance`/`DecideRosterAcceptance`) still reject
duplicates sequentially, but the atomic guarantee only becomes complete
once every live non-terminal contract carries its reservations.

## Reconciliation Constraint — No Naive Backfill

A scan-then-write backfill is **not** safe and is not claimed to be
idempotent: between the scan read and the write, a contract can reach a
terminal state or lose a member to departure — writing a reservation for
it then resurrects a guard that no release path will ever remove (the
owner contract is already past its release point), permanently blocking a
legitimate retake. A conditional `attribute_not_exists` put does not fix
this; it only prevents overwriting an existing reservation.

Production rollout for the strict invariant is therefore **blocked** until
a race-safe reconciliation is separately approved — e.g., fenced writes
that re-validate contract state and active membership inside the
transaction that creates the reservation, or a deployment window in which
acceptance of legacy contracts is paused. **No migration or backfill is
executed under this change; no shared/live-data reads or writes are
performed for it.** New contracts accepted after deployment are fully
guarded regardless of the reconciliation decision.

## Rollback

- Code rollback is safe: nothing reads reservation items; orphaned items
  are inert.
- Caveat: orphaned reservations must be reconciled before any later
  re-deploy that relies on a clean guard, or stale reservations could
  block legitimate retakes. Cleanup uses the same owner-conditioned
  delete as release paths — a mismatched-owner reservation is never
  deleted by cleanup either; it is flagged for review.
- The contract-write fences (departed-set conditions on `Save`,
  `SaveApproved`, `SaveDeparture`) roll back cleanly — removing them
  restores the old unconditional writes; conditions are pure
  `ConditionExpression`s and leave no data footprint.
- Rolling back restores the defects, not data loss.

## Compatibility

- No public contract change: same endpoints, same error codes
  (`MISSION_DEFINITION_ALREADY_ACTIVE` already maps to `422`).
- `PLAYER_ACTIVE` items and roster semantics unchanged; historic
  `ListByPlayer` reads unaffected.
- Legacy contract rows written before the `Departed` attribute existed
  are tolerated by the departed-set fences only when the expected prior
  departed set is empty — matching the "no recorded departures" case.
- Residual risk while unreconciled: inconsistent legacy guards — a legacy
  contract can be raced once; and pre-fix `PENDING_VALIDATION` contracts
  retain `PLAYER_ACTIVE` entries due to the known out-of-scope submit
  defect (recorded in `change.md`). The departed-set fence covers only
  membership divergence — other concurrent contract-state mutations
  (phases, confirmations, invitations racing the same row) have no
  optimistic versioning; classified as residual risk for a later change.

## Production Gate

No production deployment, migration, or shared-data access until Gate 2,
the scoped RED/Gate 3 sequence, and the full L3 gate set converge with
explicit user approval — and, for the strict uniqueness invariant on
legacy contracts, until the race-safe reconciliation plan is separately
approved.

The user approved a local-only cleanup: the two abandoned contracts and
their empty `PLAYER_ACTIVE` record were conditionally removed from
`hbk-missions-local` on `localhost:8000`. Its 72 definitions, 36 offers,
and season row remain. This does not authorize deleting contracts or
bypassing reconciliation in any other environment.
