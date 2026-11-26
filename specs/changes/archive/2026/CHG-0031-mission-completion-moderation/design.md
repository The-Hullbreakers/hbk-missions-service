---
change_id: CHG-0031
artifact: design
status: proposed
updated: 2026-10-09
---

# Design — Mission Completion Moderation

## Domain

`internal/missions/domain/`:

- `ContractState` already carries `ContractPendingValidation`
  (`"PENDING_VALIDATION"`, `contract.go`); no new state needed.
- `ContractState.IsActive()` drops `PENDING_VALIDATION` — the slot frees at
  submit (decision 3). `IsTerminal()` unchanged; pending stays
  non-terminal.
- `MissionContract` gains a `pendingCredits map[string]domain.MeritCredit`
  keyed by `ConfirmationID` — confirmations stage the computed credit
  (amount, media, recipient resolved at confirm time) without writing the
  ledger.
- `confirm()` keeps all guards; instead of returning a credit for the
  atomic write it records `pendingCredits[confirmationID]` and returns the
  projection for the response.
- New transitions in `mission_contract_lifecycle.go`:
  - `Submit(identity)` — roster member, `IN_PROGRESS` →
    `PENDING_VALIDATION`; zero confirmations allowed.
  - `Approve(identity)` — caller must hold `admin`/`moderator` (checked in
    application via `identity.HasRole`), `PENDING_VALIDATION` →
    `COMPLETED`; returns the staged unit credits plus the
    `COMPLETION` credit when `Terms.AllConfirmed`.
  - `Reject(identity)` — same guards → `DISQUALIFIED`; returns no credits.
- `Complete(identity)` is removed — the semantic no longer exists.
- `canAdjudicate` is split: roster rule (`CanSubmitEvidence`) vs reviewer
  rule (`identity.HasRole("admin") || HasRole("moderator")`).

New sentinel: `ErrContractNotPendingValidation` (→ `422
CONTRACT_NOT_PENDING_VALIDATION`).

## Application

- `SubmitContractHandler` — replaces `CompleteContractHandler`: `Get` →
  `Submit` → `writer.Save` (no credit involved).
- `ApproveContractHandler` — `Get` → reviewer-role check → `Approve` →
  new port `ContractCreditsWriter.SaveApproved(contract, credits)` writing
  the state transition plus every credit in one transaction.
- `RejectContractHandler` — `Get` → reviewer-role check → `Reject` →
  `writer.Save`.
- `ConfirmObjectiveHandler` switches from `ContractAndCreditWriter` to
  `ContractWriter` — `Save` persists the staged confirmation; handler
  still returns the credit projection for the `201` echo.

## HTTP Boundary

`interface/http/`:

- `contract_lifecycle.go`: `SubmitContractHandler` (same `transition`
  adapter); `ApproveContractHandler`/`RejectContractHandler` —
  `approve` returns `200` with `{credits: [...]}`, `reject` returns
  `200`. `CompleteContractHandler` route removed.
- `errors.go`: `ErrContractNotPendingValidation` →
  `CONTRACT_NOT_PENDING_VALIDATION`.
- `handler.go`: `complete` route replaced by `submit`, `approve`,
  `reject`.

## Infrastructure

`infrastructure/dynamodb/`:

- Contract item gains `pending_credits` (list of serialized credit
  attributes); absent = empty.
- `ContractCreditsWriter` adapter: `TransactWriteItems` — conditional
  contract `Put` (state must be `PENDING_VALIDATION`) plus one conditional
  credit `Put` (`attribute_not_exists` on `ConfirmationID`) per staged
  credit. `ConditionalCheckFailed` on the contract →
  `ErrContractNotPendingValidation`; on a credit → that credit is skipped
  (idempotent retry).
- Bound: `TransactWriteItems` supports 100 items — unit-limit caps keep
  pending credits far below; documented in verification.

## Composition

`server.go`/`config.go`: wire `SubmitContractHandler`,
`ApproveContractHandler`, `RejectContractHandler`; drop
`CompleteContractHandler`; update route table.

## Test Strategy

- Domain unit: `Submit`/`Approve`/`Reject` guards and transitions,
  `IsActive` change, staged-credit materialization, partial approve.
- Application unit: reviewer-role enforcement, handler orchestration with
  port doubles.
- Infrastructure: `pending_credits` round-trip; transactional approve
  writes contract + N credits; conditional-failure mapping.
- HTTP integration: `submit`/`approve`/`reject` status vocabulary;
  removed `complete` route 404s.
- Critical E2E: accept → start → evidence → submit → approve → standing
  reflects points; and the reject journey → zero points.

## Rollback

Routes are additive except `complete` removal — breaking API change,
flagged in rollout. `pending_credits` is attribute-additive (no
migration). Rollback restores `complete` and re-points confirm at the
ledger; contracts stuck mid-review on rollback are enumerated in
`rollout.md`.
