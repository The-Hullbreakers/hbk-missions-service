---
capability_id: CAP-MISSION-CONTRACT-FULFILLMENT
title: Mission Contract Fulfillment
status: current
last_consolidated_change: CHG-0011-mission-definitions-and-pools
updated: 2026-08-30
---

# Mission Contract Fulfillment

## Intent

A participant or team completes verifiable objectives inside an accepted mission contract and receives fair seasonal merit in the right category and track.

## Scope

- Progress through sequential contract phases and required/optional objectives.
- Confirm countable units while the contract is `IN_PROGRESS` and the season is `ACTIVE`.
- Credit seasonal merit per tenant, season, category, track, and recipient.
- Reject duplicate confirmations, terminal contracts, and out-of-season credits.

## Non-Goals

- Generic objective rule engine.
- Synchronous integration with Achievements.
- Disputes, manual reviews, or corrections (see CHG-0001).
- Procedural contract generation.

## Key Concepts

- `MissionContract`: accepted offer, roster, phase, and terminal state.
- `Objective`: published, category-bound unit with merit value.
- `Confirmation`: an attempt to credit a concrete unit.
- `ConfirmationID`: composite value object (`contract:objective:unit:recipient`) for idempotent credit identity and duplicate detection.
- `MeritCredit`: immutable ledger fact.
- `MeritLedger`: append-only credits per bucket.
- `Standing`: projection derived from the ledger, not a second source of truth.

## Requirements

### REQ-MCF-001 — Execute solo or team

`SOLO` credits the individual participant; `TEAM` credits the team; `FLEX` selects a mode at acceptance and cannot change later. Each confirmed member of a team contract consumes one active-contract slot in their personal limit.

#### SCN-MCF-TEAM — Freeze an eligible roster

Given all invitees have active capacity, when they confirm and the mission starts, then the roster is frozen and the contract uses the team track.

### REQ-MCF-002 — Progress by phases and objectives

A contract contains sequential phases and required or optional objectives. Objectives can measure completion, quantity, or time and declare their countable units.

#### SCN-MCF-OBJECTIVES — Unlock next phase

Given a phase requires its mandatory objectives, when all are confirmed, then the next phase becomes available without requiring optional objectives.

### REQ-MCF-003 — Confirm countable units

A unit is valid only when the contract was active, the objective was eligible, it occurred inside the window, the recipient was entitled, the evidence or source was accepted, and the supplied season matches the contract's bound season.

#### SCN-MCF-001 — Credit each confirmed bounty

Given a Bounty PvP mission offers 100 merits per unique target, when four valid objectives are confirmed, then the participant receives 400 merits in `BOUNTY_HUNTING` for the active season.

#### SCN-MCF-SEASON-MATCH — Reject confirmation for the wrong season

Given a contract is bound to `season-2026-q4` and a request supplies `season-2026-q3`, when the confirmation is processed, then it returns `SEASON_MISMATCH` and no credit is recorded.

### REQ-MCF-004 — Prevent duplicates and basic farming

The same unit, objective, contract, and recipient are credited at most once. A mission may define cooldown, unit limit, and published repetition restrictions.

#### SCN-MCF-DUPLICATE — Ignore a replay

Given a target already produced a credit, when the same confirmation arrives again, then the ledger and standing do not change.

### REQ-MCF-005 — Preserve terminal outcomes

A contract ends `COMPLETED`, `FAILED`, `ABANDONED`, `EXPIRED`, `CANCELLED`, or `DISQUALIFIED`. No terminal state returns to active and the history keeps reasons and evidence.

#### SCN-MCF-TERMINAL — Reject confirmation on terminal contract

Given a contract is in a terminal state, when a new confirmation arrives, then it is rejected without creating or modifying credits.

## Related Capability

- `seasonal-category-competition` owns the seasonal ledger and standings.

## Acceptance Summary

- A contract preserves the accepted `definition_id` and `definition_version` from the offer.
- A valid confirmation credits exactly the published merits once.
- The same unit, objective, contract, and recipient do not generate a second credit.
- The credit belongs to the right tenant, season, category, and track.
- A `FREEZING` or `FINALIZED` season rejects new credits except for eligible work started in time.
- Individual standing shows the derived ledger total.
- A team credits the team track, not each member's individual standing.

## HTTP Boundary

- `POST /api/v1/missions/contracts/{contract_id}/evidence`
- Request body: `objective_id`, `unit_id`, `season_id`, `tenant_id`.
- Requires `Authorization: Bearer <OIDC token>` from HBK Identity.
- `201 Created` with the credited `MeritCredit` on success.
- `400 Bad Request` for invalid JSON.
- `401 Unauthorized` for missing or invalid token.
- `403 Forbidden` for caller without required role.
- `422 Unprocessable Entity` for domain errors with a stable `code` (`CONTRACT_NOT_FOUND`, `CONTRACT_NOT_IN_PROGRESS`, `SEASON_NOT_ACTIVE`, `SEASON_MISMATCH`, `CATEGORY_MISMATCH`, `ROSTER_EMPTY`, `DUPLICATE_CONFIRMATION`, `INVALID_REQUEST`).

## Stable Production Files

- `internal/missions/domain/confirmation.go`
- `internal/missions/domain/merit_credit.go`
- `internal/missions/domain/merit_ledger.go`
- `internal/missions/domain/mission_contract.go`
- `internal/missions/domain/season.go`
- `internal/missions/domain/objective.go`
- `internal/missions/application/confirm_objective.go`
- `internal/missions/interface/http/handler.go`
- `internal/missions/interface/http/errors.go`
- `internal/missions/interface/http/middleware.go`
- `internal/missions/infrastructure/merit_ledger_repository.go`
- `internal/missions/infrastructure/dynamodb/*.go`
- `internal/missions/infrastructure/session/*.go`
- `internal/missions/composition/*.go`
- `cmd/hbk-missions-service/main.go`

## Test Suites

- `tests/unit/missions/domain`
- `tests/unit/missions/application`
- `tests/unit/missions/interface/http`
- `tests/unit/missions/infrastructure/dynamodb`
- `tests/unit/missions/infrastructure/session`
- `tests/integration/http`
- `tests/integration/infrastructure`
- `tests/integration/infrastructure/dynamodb`
- `tests/integration/infrastructure/session`
- `tests/integration/composition`
- `tests/integration/e2e`

## Implementation Notes

- `MissionContract.Confirm` owns the confirmation policy; `ConfirmObjective` is a thin wrapper that records the resulting `MeritCredit`.
- `ConfirmationID` centralizes the `contract:objective:unit:recipient` identity and replaces ad-hoc string concatenation.
- `MissionContract.Summary` produces an `ActiveContractSummary` used by the acceptance policy in `mission-board`.
- Adapters persist `MissionContract`, `Season`, `Objective`, and `MeritCredit` in DynamoDB using a single table with `PK`/`SK` and a `GSI1` for bucketed ledger queries.
- `cmd/hbk-missions-service/main.go` wires configuration, DynamoDB client, OIDC session, stores, and HTTP routes.
- `AuthMiddleware` validates OIDC ID tokens and injects `callerID` and claims into the request context.
- Domain errors are mapped to stable public `code` values in the HTTP boundary; internal error strings are never exposed.
