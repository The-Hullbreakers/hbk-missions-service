---
change_id: CHG-0005
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

`Missions` owns the season lifecycle and the single-active invariant.

## Aggregates

- `Season`: identity, state, `Official`. The active-season invariant is checked across the set of seasons.

## Domain Changes

- Add `ErrActiveSeasonAlreadyExists` in `internal/missions/domain/errors.go`.
- Add `domain.ValidateCanCreateSeason(active []Season) error` or a method on a season value: if any `active` has `State == SeasonStateActive`, return `ErrActiveSeasonAlreadyExists`.

## Ports

- `SeasonWriter`: already exists (`Save`).
- `ActiveSeasonsReader`: new application port `ListActive() ([]domain.Season, error)`.

## Application

- `CreateSeasonHandler` receives an `ActiveSeasonsReader` in addition to `SeasonWriter`.
- Before saving, call `reader.ListActive()` and pass the result to the domain validation.
- On `ErrActiveSeasonAlreadyExists`, return it so the HTTP boundary maps it to `ACTIVE_SEASON_EXISTS`.

## Infrastructure

- `SeasonStore.ListActive()` queries DynamoDB with `PK = SEASON` and a `FilterExpression` `State = ACTIVE`.
- Alternatively, use a `GSI` or scan; in v1, with few seasons, a query with filter is acceptable and stays backward-compatible.

## Interface

- `POST /api/v1/seasons` already exists; it returns `422` with `code: ACTIVE_SEASON_EXISTS` when the rule is violated.
- Add the new public code in `errors.go`.

## Composition

- `NewServer` already has `seasonStore`; pass it as both `ActiveSeasonsReader` and `SeasonWriter` to `CreateSeasonHandler`.

## Development Sequence

1. Domain RED/GREEN/BLUE: new sentinel error and validation.
2. Application RED/GREEN/BLUE: `CreateSeasonHandler` uses `ActiveSeasonsReader`.
3. Boundary RED/GREEN/BLUE: `SeasonStore.ListActive()`, `errors.go` mapping, integration test.
