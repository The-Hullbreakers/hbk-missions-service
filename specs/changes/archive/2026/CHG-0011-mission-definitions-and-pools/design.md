---
change_id: CHG-0011
artifact: design
status: proposed
updated: 2026-08-30
---

# Design — Add Mission Definitions And Season Pools

## Overview

Definitions are reusable, versioned templates. Pools are season-bound collections of approved definitions. Manual publication selects a definition (and optionally a pool), validates the active season, and produces a `MissionOffer` with frozen `DefinitionID`/`DefinitionVersion`.

## Domain Model

- `MissionDefinition`: `DefinitionID`, `Category`, `Mode`, `Capacity`, `Repeatability`, `Cooldown`, `LifecycleState` (`DRAFT`, `PUBLISHED`, `RETIRED`), `Version`.
- `MissionPool`: `PoolID`, `TenantID`, `Category`, `SeasonID`, `Entries` (`[]PoolEntry`), `PublishPolicy`.
- `PoolEntry`: `DefinitionID`, `Version`, `Weight`.
- `MissionOffer`: adds `DefinitionID`, `DefinitionVersion`.
- `MissionContract`: adds `DefinitionID`, `DefinitionVersion`.
- `OfferPublication`: domain service that creates a `MissionOffer` from a definition/pool and the active season.
- `DecideAcceptance`: now checks active contracts for the same `DefinitionID`.

## Application Layer

- `CreateMissionDefinitionHandler`
- `PublishMissionDefinitionHandler`
- `CreateMissionPoolHandler`
- `PublishMissionFromPoolHandler` (replaces raw `PublishMissionHandler`)
- `AcceptMissionHandler` updated to consider `DefinitionID`.

## Boundary

- `POST /api/v1/missions/definitions`
- `POST /api/v1/missions/definitions/{definition_id}/publish`
- `POST /api/v1/missions/pools`
- `GET /api/v1/missions/pools`
- `POST /api/v1/missions/offers` (new payload: `definition_id`, `pool_id`)

## Infrastructure

- DynamoDB stores for `MissionDefinition` and `MissionPool`.
- `OfferStore`/`ContractStore` add `DefinitionID` and `DefinitionVersion` attributes.
- Existing items without these attributes load with a sentinel default value.

## Concurrency

- Definition publishing is a single-row state transition.
- Pool creation and offer generation are independent.
- Acceptance concurrency is protected by the existing `DecideAcceptance` policy and conditional persistence (optimistic concurrency on offer capacity).

## Backward Compatibility

- Legacy `MissionOffer` and `MissionContract` records that lack `DefinitionID` receive a default derived from their existing identity, so the board and fulfillment continue to work.
- The `POST /api/v1/missions/offers` payload changes; clients must migrate to the new `definition_id`/`pool_id` contract.
