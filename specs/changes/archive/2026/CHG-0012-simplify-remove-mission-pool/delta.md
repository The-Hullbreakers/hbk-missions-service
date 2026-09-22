# Delta — Simplify Mission Publishing By Removing MissionPool

## Affected Canonical Specs

- `specs/capabilities/mission-board/spec.md` (MODIFIED)

## Summary

Remove `MissionPool` from the public contract, domain model, application, boundary, and infrastructure. Mission offers are published directly from a published `MissionDefinition`.

## Changed Behavior

### REMOVED

- `MissionPool` domain aggregate.
- `PoolEntry`, `MissionPublishMode` value objects.
- `POST /api/v1/missions/pools` public endpoint.
- `PoolStore` persistence adapter and DynamoDB table usage.
- Pool-related public error codes `INVALID_MISSION_POOL` and `POOL_NOT_FOUND`.

### MODIFIED

- `POST /api/v1/missions/offers` no longer accepts `pool_id`; it accepts `definition_id` (and optional `definition_version`) and `capacity`.
- `PublishMissionFromPoolHandler` becomes `PublishMissionHandler`; it no longer takes a `PoolReader` and ignores `poolID`.
- `MissionBoard` canonical spec: remove `MissionPool` from key concepts, requirements, and HTTP boundary.

### ADDED

- None.

## BDD Acceptance Scenarios

### SCN-MB-PUBLISH-FROM-DEFINITION

Given a published definition `bounty-template#v1` and an active season, when an operator publishes an offer with `definition_id=bounty-template` and `capacity=10`, then the offer is bound to the active season and carries `definition_id` and `definition_version`.

### SCN-MB-REJECT-PUBLISH-FROM-DRAFT

Given a definition in `DRAFT` state, when an operator tries to publish an offer from it, then the request is rejected with `DEFINITION_NOT_PUBLISHED`.

### SCN-MB-OFFER-WITHOUT-POOL

Given the operator does not provide `pool_id`, when the offer is published, then it is accepted and persisted without requiring a pool.

### SCN-MB-ACTIVE-TEMPLATE-EXCLUSION-PRESERVED

Given a player has an active contract for `bounty-template#v1`, when they browse the board, then no other offer with `definition_id=bounty-template` is shown.

## Error Vocabulary

- `DEFINITION_NOT_FOUND`
- `DEFINITION_NOT_PUBLISHED`
- `NO_ACTIVE_SEASON`
- `OFFER_INVALID`
- `MISSION_DEFINITION_ALREADY_ACTIVE`
- `ACTIVE_MISSION_LIMIT_REACHED`
- `OFFER_CAPACITY_EXHAUSTED`
