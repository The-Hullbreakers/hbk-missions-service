---
change_id: CHG-0053
artifact: design
status: archived
updated: 2026-10-26
---

# Design — CHG-0053

## Root Cause

- **BUG-0053-01a (visibility):** `ListOffersHandler.Handle` builds the
  hidden-definition set from `PlayerContractsReader.ListByPlayer`, which
  keeps departed members — the roster is preserved as history
  (`IncludesPlayer` ignores `departed`). The membership-aware reader
  (`ListActiveByPlayer`, excludes departed) exists but is not used by the
  board.
- **BUG-0053-01b (slot):** `ContractStore.Save` releases `PLAYER_ACTIVE`
  entries only for terminal contracts. A non-terminal `leave` save is a
  plain put, so the departed member's `contract_ids` entry stays and can
  block a retake at the personal limit.
- **BUG-0053-02 (uniqueness):** the same-definition check lives in
  `DecideAcceptance`/`DecideRosterAcceptance` — reads evaluated before
  the commit. `CommitAcceptance`'s transaction conditions only
  `size(contract_ids) < limit`; nothing in the write detects a duplicate
  definition, so racing (or roster-smuggled) duplicates both commit.
  `DecideRosterAcceptance` also never inspects teammates' active
  definitions, so even sequential team accepts can seat an already-bound
  member.

## Decision — Per-Player+Definition Transactional Reservation

An independent reservation item, **owned by the contract id**, guards one
live contract per (player, definition). It is written and released only
inside conditional transactions; the read-path decisions stay as the
explainable first check, and the reservation is the commit-time backstop
that closes the race.

- Item: `PK = "PLAYER_DEF_ACTIVE#<player_id>"`, `SK = "<definition_code>"`
  — player and definition are separate key components, never concatenated
  into one ambiguous string. Attribute `contract_id` carries the guard
  owner separately.
- Effective definition key: the contract/offer `DefinitionCode`, falling
  back to `OfferID` when empty — the same fallback `Summary()` and the
  board list already apply to legacy offers without a definition code.
- **Reserve:** `CommitAcceptance` appends one conditional `Put` per roster
  member (`attribute_not_exists(PK) AND attribute_not_exists(SK)`),
  alongside the contract `Put` and the existing `PLAYER_ACTIVE`
  `ADD … size < limit` updates — one `TransactWriteItems`, no new round
  trip.
- **Hold:** the reservation survives `PENDING_VALIDATION` — submit does
  not release it (matches `BlocksRetake`).
- **Release, ownership-safe and member-scoped:**
  - `leave` on a non-terminal contract: release **only the departing
    member's** reservation and `DELETE` their `PLAYER_ACTIVE` contract
    id, atomically with the contract put.
  - Terminal `Save`/`SaveApproved`: release only for members **still
    active** on the contract (`ActiveMembers`), never for departed
    members — a departed member may already hold a newer reservation for
    the same definition from a retake, and the terminal write of the old
    contract must never touch it. Departed members were already released
    at their leave. The `PLAYER_ACTIVE` set `DELETE` may still cover the
    whole roster (idempotent — deleting an absent id is a no-op); the
    reservation release may not, because it is keyed by definition, not
    by contract membership.
  - Every reservation release is a conditional `Delete`:
    `attribute_not_exists(PK) OR contract_id = :cid` — it succeeds when
    the reservation is absent or owned by the contract being saved, and
    is refused when a **different** contract owns it. A mismatched owner
    is never deleted and never cleaned up. A stale terminal `Save` that
    trips the condition cancels the transaction and returns a conflict
    error — it must not silently mutate or destroy the newer reservation.
- **Stale-write guard (REG-0053-03):** the `SaveDeparture` contract `Put`
  is conditional — `attribute_exists(PK)`, `#state = :state` (a
  non-terminal leave does not change state), and the persisted
  `Departed` list must equal the exact set the leaver saw:
  `size(#departed) = :departedBefore` plus `contains(#departed, :memberN)`
  for every member in `Departures() - departingPlayerID`. The legacy
  `attribute_not_exists(#departed)` branch is tolerated only when that
  prior set is empty — a missing attribute can never satisfy a
  non-empty expectation, and an equal count of different members
  conflicts the same as a changed set. A snapshot that predates another
  member's committed departure cancels the whole transaction — no
  contract overwrite and no guard release. The condition lives on the
  `Put` itself; the transaction keeps exactly one operation per key.
- `PLAYER_ACTIVE` count semantics unchanged: the set keeps counting
  contract ids against the effective limit; the reservation guards a
  different invariant (uniqueness), not the count.

## Error Policy

`TransactionCanceledException` cancellation indices distinguish the
conflict: a reservation `Put` failure maps to
`ErrMissionDefinitionAlreadyActive`; a `PLAYER_ACTIVE` update failure
maps to `ErrActiveMissionLimitReached` (requester) /
`ErrTeammateLimitReached` as today. A reservation-release `Delete`
failure on lifecycle writes surfaces as a conflict error — the write is
aborted rather than allowed to remove a foreign-owned guard. Public
mapping already exists: `MISSION_DEFINITION_ALREADY_ACTIVE` → `422` via
`PublicCode`; no new error type.

## Application — Fixed Interfaces

- `ListOffersHandler` takes the existing `ActiveContractsReader` port and
  builds the hidden set from `ListActiveByPlayer` (non-terminal contracts
  where the player is a non-departed member — including
  `PENDING_VALIDATION`, matching `BlocksRetake`). The
  `PlayerContractsReader.ListByPlayer` port stays untouched for contract
  history.
- `AcceptMissionHandler` keeps its orchestration: the requester's
  `DecideAcceptance` stays; the roster pass delegates per-member checks
  to the domain decision below.
- `LeaveContractHandler` gains a focused departure writer port:

  ```go
  type DepartureWriter interface {
      SaveDeparture(contract *domain.MissionContract, departingPlayerID string) error
  }
  ```

  After `contract.Leave(identity)` (existing authorization/membership
  checks unchanged), the handler routes: contract terminal (last member
  left) → existing `ContractWriter.Save` terminal path; non-terminal →
  `SaveDeparture(contract, identity.PlayerID())`, which persists the
  departure and releases only that member's reservation and
  `PLAYER_ACTIVE` entry in one transaction.

## Domain — Fixed Decision

- Keep the existing `DecideRosterAcceptance(offer, seasonState,
  memberActive, policy)` signature. It delegates to `DecideAcceptance`
  for every member: a definition conflict returns
  `ReasonMissionDefinitionAlreadyActive`, and a limit failure retains
  `ReasonTeammateLimitReached`. The requester is checked separately by
  `AcceptMissionHandler` first, preserving
  `ReasonActiveMissionLimitReached` for their limit.
- No new domain types — the reservation is a persistence guard, not
  business vocabulary.

## Infrastructure

- `contract_store_accept.go`: add reservation `Put`s; extend cancellation
  mapping by item index.
- `contract_store.go` / `contract_store_review.go`: extend release
  transaction builders to delete each still-active member's reservation
  with the owner condition beside the `PLAYER_ACTIVE` `DELETE`; add the
  `SaveDeparture` adapter implementing the focused port above.

## Expansion — Contract-Write Fences (Gate 1 expansion, approved 2026-10-21)

Gate 1 ownership now also covers `mission-contract-fulfillment`
(`REQ-MCF-013`/`REQ-MCF-010`/`REQ-MCF-020`). Settled direction:

- **Domain:** the accepted roster remains immutable history;
  `MissionContract.Departures` remains the mutable record of exits.
  No new domain behavior is added.
- **Application:** `LeaveContractHandler` captures `previousState` before
  `contract.Leave(...)` and **always** routes through the focused port —
  `SaveDeparture(contract, departingPlayerID, previousState)` — for both
  non-terminal and last-member terminal exits. The generic
  `ContractWriter` dependency is dropped from the constructor, which
  becomes `NewLeaveContractHandler(contracts, departure)`. The T-3 guard
  test expecting last-member `Save` is superseded by this behavior;
  its evidence is preserved, not deleted.
- **`SaveDeparture` adapter:** takes the caller-supplied prior `State`
  rather than deriving it; the contract `Put` keeps the same-set fence
  (existing row + stored prior state + `Departures() − leaver` exact
  set). For a terminal last-member leave it also releases the leaver's
  reservation and `PLAYER_ACTIVE` id in the same transaction — earlier
  departures were released at their own leaves.
- **`Save` (non-departure):** the contract `Put` becomes conditional on
  the stored departed set matching the candidate's `Departures()`
  (`size` + `contains`, order-independent; a missing `Departed`
  attribute is tolerated only when the expected set is empty). Creating
  a brand-new contract row stays allowed only when the departed set is
  empty (`attribute_not_exists(PK)` branch); an existing row with a
  changed departed set can never be overwritten.
- **`SaveApproved`:** the contract `Put` requires the row to exist with
  the same departed-set match (the approval always operates on an
  existing contract); only credit-item conditional failures map to
  `ErrDuplicateConfirmation` — the index-scoped mapping is preserved.
- **No general versioning** is introduced for other concurrent state
  changes (phase, confirmations, invitations racing the same row);
  recorded as residual risk, not addressed by this change.
- CHG-0051 separately fences `SaveApproved` with an adjudication lock on
  the same file — conditions are appended, never replaced, and no
  CHG-0051 artifact is edited by this change.

## Interface / Composition

- No HTTP changes required — handler and `PublicCode` already cover the
  error. Composition wires the same stores plus `DepartureWriter` to
  `LeaveContractHandler` and `ActiveContractsReader` to
  `ListOffersHandler`.

## Boundary Tests (DynamoDB Local)

Regression evidence must fail pre-fix: two parallel `CommitAcceptance`
calls — same player, spare limit, different `definition_version` of the
same `definition_code` — yield one success and one
`ErrMissionDefinitionAlreadyActive`; two different players both commit.
Release regressions cover terminal save (active members only),
departed-member leave, stale-save/reforeign-owner refusal, and
`PENDING_VALIDATION` retention. See `tasks.md` T-7/T-8.

## Rollback

Code rollback leaves orphan reservation items that are inert (nothing
reads them) but must be reconciled before any later re-deploy relies on a
clean guard. See `rollout.md` — production rollout is blocked on a
race-safe reconciliation plan for legacy contracts.
