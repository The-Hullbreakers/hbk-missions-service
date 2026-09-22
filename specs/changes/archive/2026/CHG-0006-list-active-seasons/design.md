---
change_id: CHG-0006
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

`Missions` owns the season lifecycle and its query surface.

## Aggregates

- `Season`: identity, `State`, `Official`. Read-only queries return projections of stored seasons.

## Domain Changes

- Add `SeasonStatusFilter` in `internal/missions/domain/season.go`.
  - Values: `ACTIVE`, `INACTIVE`, `ALL`.
  - `Match(season Season) bool` decides whether a season belongs to the bucket.
  - `INACTIVE` matches any season whose `State` is not `ACTIVE`.

## Ports

- New application query port `SeasonReader` in `internal/missions/application/list_seasons.go`.
  - `List(filter domain.SeasonStatusFilter) ([]domain.Season, error)`.

## Application

- `ListSeasonsHandler` receives a `SeasonStatusFilter`.
- It calls `SeasonReader.List(filter)` and returns the slice.
- No command side effects.

## Infrastructure

- `SeasonStore.List(filter domain.SeasonStatusFilter) ([]domain.Season, error)` queries `PK = SEASON` and applies `filter.Match` to each item.
- Existing `ListActive` is left untouched to avoid changing `CreateSeasonHandler`.

## Interface

- HTTP `GET /api/v1/seasons` in `internal/missions/interface/http/list_seasons.go`.
- Parse `?status`; default to `ACTIVE`; validate against `active|inactive|all` (case-insensitive).
- Map unknown status to `400` with `INVALID_QUERY_PARAMETER`.
- Return `200` with JSON array of `Season` objects.

## Composition

- `NewServer` creates `ListSeasonsHandler` using `seasonStore`.
- Register `GET /api/v1/seasons` wrapped with `AuthMiddleware`.

## Development Sequence

1. Domain RED/GREEN/BLUE: `SeasonStatusFilter` and `Match`.
2. Application RED/GREEN/BLUE: `ListSeasonsHandler` with `SeasonReader` double.
3. Boundary RED/GREEN/BLUE: `SeasonStore.List`, HTTP handler, route wiring, integration tests.
