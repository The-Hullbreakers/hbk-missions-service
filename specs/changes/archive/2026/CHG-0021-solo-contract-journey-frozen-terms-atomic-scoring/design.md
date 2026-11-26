---
change_id: CHG-0021
artifact: design
updated: 2026-10-08
---

# Design — Solo Journey, Frozen Terms, Atomic Scoring

## Canonical Layer Scope

Domain → Application → Boundary(HTTP) → Infrastructure → Interface →
Composition.

## Domain

- `ContractTerms` value object: ordered objectives (`ObjectiveID`, type,
  category, unit, `MeritValue`, `Bonus`, `UnitLimit`) frozen at acceptance;
  serialized onto `MissionContract`.
- `MissionContract` state machine: `ACCEPTED → IN_PROGRESS → COMPLETED |
  ABANDONED | EXPIRED` (plus existing terminal states). Transitions validate
  actor (roster) and precondition (required objectives confirmed for complete).
- `ContractID` gains an attempt suffix (e.g. `tenant#offer#player#attempt` —
  delimiter chosen to be unambiguous and storage-safe).
- `ConfirmationID` rebuilt with a safe delimiter (e.g. `~`) so it parses
  unambiguously; stored ids migrated or dual-parsed.
- Objective model gains `Bonus` and `UnitLimit` round-trip.

## Application

- `StartContractHandler`, `CompleteContractHandler`, `AbandonContractHandler`
  use cases (one command each); expiry evaluated lazily on load/write and by
  scheduler hook (decision recorded at Gate 2).
- `AcceptMissionHandler` builds `ContractTerms` from the definition version.
- `ConfirmObjectiveHandler` resolves objective from `ContractTerms`, enforces
  unit limit, and commits via a single `ContractAndCreditWriter` port
  (transactional write abstraction — provider-neutral port, atomic adapter).

## Boundary — HTTP

- `POST /api/v1/missions/contracts/{contract_id}/start|complete|abandon`.
- Evidence/confirmation contract unchanged; contract responses include
  `contract_id`, `state`, `terms`.

## Infrastructure

- `MissionContract` items persist `terms`; `ObjectiveStore` items persist
  `Bonus`/`UnitLimit` (fixes BUG-0021-04).
- `TransactWriteItems` for contract-update + ledger-append + dedup marker
  (fixes BUG-0021-03).
- Unit-limit counter derived from credits grouped by `ConfirmationID` prefix
  (fixed parsing, BUG-0021-05).

## Test Strategy

- Domain: state machine partitions, snapshot immutability, `ConfirmationID`
  round-trip with colon-containing ids.
- Application: lifecycle handlers with port doubles; atomic-commit failure
  injection (writer spy fails → no partial state, retry idempotent).
- Infrastructure: transaction rollback proof on injected credit failure;
  Bonus/UnitLimit round-trip; attempt-id uniqueness.
- HTTP: full journey accept→start→confirm→complete E2E (critical).
- Mutation: state machine and atomic-write branches.

## Risk And Rollback

- **Risk**: in-flight `ACCEPTED` contracts created before this change lack
  `terms` — migration: reject with `CONTRACT_STATE_INVALID` + operator notice,
  or backfill from definition version (decide at Gate 2).
- **Rollback**: state-machine and terms are additive columns; rollback restores
  previous binary (old rows still readable).
