---
change_id: CHG-0004
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

`Missions` owns `Season` lifecycle and boundaries. Identity owns authentication and tenant claim. The existing `AuthMiddleware` injects the caller.

## Aggregates

- `Season`: identity, state, and `Official` flag. The aggregate root is the season itself; there are no child entities.

## Domain Changes

- Add `Official bool` to `Season`.
- Keep `NewSeason(id, state)` defaulting `Official: true` for existing callers.
- Add `NewSeasonWithOfficial(id, state, official)` for explicit creation.
- Keep `SeasonState` unchanged.

## Ports

- `SeasonWriter` (application): `Save(season domain.Season) error`
- `SeasonReader` (application): `Get(seasonID string) (domain.Season, error)`

Both are already implemented by `dynamodb.SeasonStore`.

## Application

- `CreateSeasonHandler` with `SeasonReader` and `SeasonWriter`.
- Input: `tenantID`, `seasonID`, `official`.
- Validation: `seasonID` not empty; default `official = true`.
- On success, persist and return `Season`.
- On invalid input, return `ErrSeasonInvalid`.

## Infrastructure

- `dynamodb.SeasonStore` adds `Official` attribute to the item.
- Loading a missing `Official` attribute defaults to `true` for backward compatibility.
- Add `getBool` helper to `attribute_helpers.go` or inline read with default.

## Interface

- `POST /api/v1/seasons`
- Request: `{ "tenant_id": "...", "season_id": "...", "official_statistics": true }`
- Success: `201 Created` with the created `Season`.
- Validation error: `422 Unprocessable Entity` with `{ "code": "INVALID_SEASON" }`.
- Authentication via existing `AuthMiddleware`.

## Composition

- `NewServer` receives a `SeasonStore` and a `CreateSeasonHandler`.
- Register `POST /api/v1/seasons` route.

## Development Sequence

1. Domain RED/GREEN/BLUE: `Season` with `Official` and constructors.
2. Application RED/GREEN/BLUE: `CreateSeasonHandler` and unit tests.
3. Boundary RED/GREEN/BLUE: HTTP handler, `SeasonStore` update, integration test.
