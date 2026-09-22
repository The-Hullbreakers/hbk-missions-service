---
change_id: CHG-0007
artifact: design
status: proposed
updated: 2026-08-30
layer_scope:
  domain: affected
  application: affected
  boundary: affected
  infrastructure: affected
  interface: affected
  composition: affected
---

# Design

## Bounded Context

`Missions` owns the `Season` lifecycle. `mission-contract-fulfillment` consumes the lifecycle state when recording merits, but the state transitions belong to `seasonal-category-competition`.

## Aggregates

- `Season`: identity, `State`, `Official`. Freeze and finalize are state transitions on this aggregate.

## Domain Changes

- `internal/missions/domain/season.go`
  - Add `Freeze(season Season) (Season, error)` and `Finalize(season Season) (Season, error)` enforcing:
    - `ACTIVE -> FREEZING`
    - `FREEZING -> FINALIZED`
  - Add `CanAcceptMerits(season Season, contract *MissionContract) bool`:
    - `ACTIVE` always.
    - `FREEZING` only when `contract != nil && contract.State == ContractInProgress`.
    - `FINALIZED`/`ARCHIVED` never.
  - Add `ErrInvalidSeasonStateTransition` and `ErrSeasonNotFound` to `errors.go`.

- `internal/missions/domain/confirmation.go`
  - Replace `season.State != SeasonStateActive` with `!season.CanAcceptMerits(contract)`.
  - Return `ErrSeasonNotActive` when the policy rejects.

## Ports

- `SeasonReader` (`application`): `Get(seasonID string) (domain.Season, error)` (already exists).
- `SeasonWriter` (`application`): `Save(season domain.Season) error` (already exists).

## Application

- `FreezeSeasonHandler`
  - Load season by `seasonID`.
  - Call `domain.Freeze(season)`.
  - Persist and return the season.
- `FinalizeSeasonHandler`
  - Load season by `seasonID`.
  - Call `domain.Finalize(season)`.
  - Persist and return the season.
- Both return `ErrSeasonNotFound` if `Get` fails, and `ErrInvalidSeasonStateTransition` for bad transitions.

## Infrastructure

- `dynamodb.SeasonStore.Get` and `Save` already implement the required ports; no schema or attribute changes.
- `Get` maps a missing item to `domain.ErrSeasonNotFound`.

## Interface

- `POST /api/v1/missions/seasons/{id}/freeze`
- `POST /api/v1/missions/seasons/{id}/finalize`
- Response: `200 OK` with the `Season` JSON.
- Errors:
  - `404 Not Found` with `SEASON_NOT_FOUND`.
  - `422 Unprocessable Entity` with `SEASON_STATE_INVALID`.
- Authentication via existing `AuthMiddleware`.
- `internal/missions/interface/http/errors.go` maps new domain errors to public codes.

## Composition

- `NewServer` creates `FreezeSeasonHandler` and `FinalizeSeasonHandler` using `seasonStore`.
- Register both routes under `/api/v1/missions/seasons/{id}/...` with `AuthMiddleware`.

## Development Sequence

1. Domain RED/GREEN/BLUE: `Freeze`, `Finalize`, `CanAcceptMerits`, and updated `ConfirmObjective` tests.
2. Application RED/GREEN/BLUE: `FreezeSeasonHandler` and `FinalizeSeasonHandler` with port doubles.
3. Boundary RED/GREEN/BLUE: HTTP handlers, `errors.go` mapping, route wiring, integration tests.
4. Verification: `gofmt -w . && go vet ./... && go test -race ./...`.
