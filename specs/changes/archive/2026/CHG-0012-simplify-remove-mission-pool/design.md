# Design — Simplify Mission Publishing By Removing MissionPool

## Domain

### Aggregates And Value Objects

- `MissionDefinition`: keeps its existing lifecycle (`DRAFT` → `PUBLISHED` → `RETIRED`).
- `MissionPool`: removed.
- `PoolEntry`: removed.
- `MissionPublishMode`: removed.
- `MissionOffer`: unchanged; continues to carry `DefinitionID` and `DefinitionVersion`.
- `MissionContract`: unchanged; continues to inherit and preserve `DefinitionID`/`DefinitionVersion`.

### Removed Domain Errors

- `ErrPoolInvalid`
- `ErrPoolNotFound`

## Application

### Removed Ports And Handlers

- `PoolReader` / `PoolWriter`.
- `CreateMissionPoolHandler`.

### Renamed/Refactored Handler

- `PublishMissionFromPoolHandler` → `PublishMissionHandler`.
- Signature becomes `Handle(tenantID, offerID, definitionID, version string, capacity int)`.
- Logic:
  1. List active seasons; fail with `ErrNoActiveSeason` if none.
  2. Fetch the published definition by `definitionID` and optional `version`; default to latest published if version empty.
  3. Build `MissionOffer` from definition with the given `capacity` (default to definition capacity if `<= 0`).
  4. Save offer through `OfferWriter`.

## Boundary

### Removed

- `POST /api/v1/missions/pools`
- `CreateMissionPoolHandler` HTTP adapter

### Modified

- `POST /api/v1/missions/offers` request DTO no longer has `pool_id`.
- Public error mapping loses `INVALID_MISSION_POOL` and `POOL_NOT_FOUND`.

## Infrastructure

### Removed

- `internal/missions/infrastructure/dynamodb/pool_store.go`
- Pool-related table items / `GSI` (if any).

### Kept

- `DefinitionStore` unchanged.
- `OfferStore` unchanged.
- `ContractStore` unchanged.

## Composition

- `internal/missions/composition/server.go` no longer wires `PoolStore`, `createPoolApp`, or the `POST /api/v1/missions/pools` route.
- `PublishMissionHandler` takes `OfferWriter`, `DefinitionReader`, `ActiveSeasonsReader` only.

## Backward Compatibility

- Existing offers and contracts with `DefinitionID` continue to work.
- No pool data exists in production, so no migration is required.
- The legacy fallback for pre-definition offers remains (`DefinitionID = OfferID`, `DefinitionVersion = "legacy"`).
