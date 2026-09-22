---
change_id: CHG-0011
artifact: delta
status: proposed
updated: 2026-08-30
---

# Delta — Add Mission Definitions And Season Pools

## What Is Changing

### Mission Board

- New `MissionDefinition` entity: stable identity, versioned content, lifecycle `DRAFT → PUBLISHED → RETIRED`.
- New `MissionPool` entity: curated collection of published `MissionDefinition` versions for a tenant, category, and season, with a publishing policy.
- `MissionOffer` gains `DefinitionID` and `DefinitionVersion`.
- `PublishMission` no longer accepts raw fields (`offer_id`, `category`, `capacity`, `mode`); it receives a `definition_id` (required) and optional `pool_id`, then generates an offer bound to the active season.
- `AcceptMission` now rejects an offer when the player already has an active contract with the same `definition_id`.

### Mission Contract Fulfillment

- `MissionContract` stores `DefinitionID` and `DefinitionVersion` inherited from the offer.
- The active-template exclusion uses the stable `definition_id` across versions.

## BDD Scenarios

### SCN-MB-DEFINITION-CREATE-001 — Create a mission definition

Given an operator provides a unique `definition_id`, `category`, `mode`, and optional objectives, when they call `POST /api/v1/missions/definitions`, then the definition is persisted in `DRAFT` and returned with `201 Created`.

### SCN-MB-DEFINITION-PUBLISH-001 — Publish a definition version

Given a `DRAFT` definition `bounty-template-v1`, when the operator calls `POST /api/v1/missions/definitions/bounty-template-v1/publish`, then it becomes `PUBLISHED`, an immutable version is created, and existing contracts are not affected.

### SCN-MB-POOL-CREATE-001 — Create a season pool

Given an active season `season-2026-q4`, a tenant `hbk`, and a category `BOUNTY_HUNTING`, when the operator calls `POST /api/v1/missions/pools` with approved definition versions, then the pool is persisted and returned with `201 Created`.

### SCN-MB-POOL-PUBLISH-001 — Publish an offer from a pool

Given an active season and a pool containing a published definition, when the operator calls `POST /api/v1/missions/offers` with `definition_id` and `pool_id`, then an offer is generated, bound to the active season, and returned with `201 Created`.

### SCN-MB-POOL-PUBLISH-NO-ACTIVE-001 — Reject publish without active season

Given no season is `ACTIVE`, when the operator calls `POST /api/v1/missions/offers` with a valid `definition_id`, then the response is `422 Unprocessable Entity` with `NO_ACTIVE_SEASON` and no offer is persisted.

### SCN-MB-DEFINITION-NOT-PUBLISHED-001 — Reject offer from unpublished definition

Given a definition is still `DRAFT`, when the operator attempts to publish an offer from it, then the response is `422 Unprocessable Entity` with `DEFINITION_NOT_PUBLISHED`.

### SCN-MB-ACTIVE-TEMPLATE-EXCLUDE-001 — Hide an active definition

Given a player has an `IN_PROGRESS` contract for `definition_id: bounty-template`, when they browse the board or attempt to accept another offer with the same `definition_id`, then the offer is hidden and acceptance is rejected with `MISSION_DEFINITION_ALREADY_ACTIVE`.

### SCN-MCF-CONTRACT-SNAPSHOT-001 — Contract keeps the accepted definition version

Given a player accepts an offer generated from `definition_id: bounty-template, version: 2`, when the operator later publishes `version: 3`, then the contract keeps `version: 2` and credits use the original merits.

## Error Vocabulary

- `DEFINITION_NOT_FOUND`
- `DEFINITION_NOT_PUBLISHED`
- `DEFINITION_VERSION_NOT_FOUND`
- `POOL_NOT_FOUND`
- `NO_ACTIVE_SEASON`
- `MISSION_DEFINITION_ALREADY_ACTIVE`
