# Design — List Mission Offers and Player Contracts

## Domain

### New Value Objects

- `OfferListFilter`: parsed query filters — `SeasonID` (empty = active season), `Status` (default `AVAILABLE`), `Category`, `Type`, `Mode`, `Difficulty`. `ParseOfferListFilter(raw url.Values)` returns typed errors on unknown values.
- `ContractListFilter`: `SeasonID`, `Status` (`ACTIVE` = non-terminal bucket, or explicit `ContractState`). `ParseContractListFilter(raw)`.
- `ContractStateBucket` or reuse `ContractState.IsTerminal()` for `status=ACTIVE` semantics.

### Policy Reuse

- `REQ-MB-008` hiding: reuse `ActiveContractSummary` — the list use case loads the player's active summaries and excludes matching `DefinitionID`s. No new domain rule.

## Application (CQRS read ports — one operation each)

- `ListOffersHandler`
  - Ports: `OfferListReader.List(tenantID string, filter domain.OfferListFilter) ([]domain.MissionOffer, error)`; `ActiveContractsReader` (existing, all categories); `ActiveSeasonsReader` (existing) to resolve empty `season_id`.
  - `Handle(tenantID, playerID string, filter domain.OfferListFilter)` → resolve season → list offers → drop hidden definitions → return.
- `ListPlayerContractsHandler`
  - Port: `PlayerContractsReader.ListByPlayer(tenantID, playerID string, filter domain.ContractListFilter) ([]domain.MissionContract, error)`.
  - `Handle(tenantID, playerID string, filter domain.ContractListFilter)`.

## Infrastructure

- `OfferStore.List(tenantID, filter)`: add `TenantID` attribute on `Save` (`PK = TENANT#<tenant>#OFFER` or `GSI1PK = TENANT#<tenant>` with `GSI1SK = OFFER#<id>`). Minimal upstream fix: persist `TenantID` and filter on it; apply remaining filters via `FilterExpression` / post-filter.
- `ContractStore.ListByPlayer(tenantID, playerID, filter)`: persist `TenantID` on `Save`; query `PK=CONTRACT` + filter `Roster CONTAINS playerID`, tenant, season, and status bucket.
- Key-schema decision recorded at implementation; prefer GSI1 tenant-partitioning if redeployable, else `FilterExpression` on `TenantID` (still correct, cheaper to add).

## Interface (thin)

- `interface/http/list_offers.go`: `GET /api/v1/missions/offers` — parse query via `domain.ParseOfferListFilter` (`400` on error), read tenant/player from session via middleware context, call `app.Handle`, `200` + JSON array.
- `interface/http/list_contracts.go`: `GET /api/v1/missions/contracts` — same pattern.

## Composition / IaC

- `composition/server.go`: wire stores as new readers, add two `mux.Handle` routes behind `auth.Authenticate`.
- `template.yaml`: `MissionsListOffersRoute` (`GET /api/v1/missions/offers`), `MissionsListContractsRoute` (`GET /api/v1/missions/contracts`).

## Backward Compatibility

- Read-only additions; no request/response contract change on existing endpoints.
- Existing items lack `TenantID` attribute → list may exclude legacy rows. Recorded as known limitation; single-tenant deployment makes impact negligible. Migration not required.

## Tests

- Unit: `tests/unit/missions/application/list_offers_test.go`, `list_player_contracts_test.go` (hand-written port doubles); `tests/unit/missions/interface/http/list_offers_test.go`, `list_contracts_test.go`; `tests/unit/missions/domain/offer_list_filter_test.go`, `contract_list_filter_test.go`.
- Integration: `tests/integration/infrastructure/dynamodb/offer_store_list_test.go`, `contract_store_list_test.go`; `tests/integration/http/mission_board_test.go`, `player_contracts_test.go`.
