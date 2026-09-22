---
change_id: CHG-0011
artifact: rollout
status: proposed
updated: 2026-08-30
---

# Rollout — Add Mission Definitions And Season Pools

## Deployment Steps

1. Deploy the new service version with backward-compatible DynamoDB reads.
2. Run the database migration (if any) to add `DefinitionID` and `DefinitionVersion` to existing offers and contracts with a default value.
3. Create initial `MissionDefinition` records for the active season.
4. Create the season `MissionPool` and publish the first manual offers.
5. Verify `GET /api/v1/missions/offers` and `POST /api/v1/missions/offers/{offer_id}/accept` behave correctly.

## Backward Compatibility

- Existing `MissionOffer` and `MissionContract` records without `DefinitionID` load with a default value derived from their existing identity.
- The `POST /api/v1/missions/offers` payload changes; clients must be updated before using the new route.

## Rollback

- Revert the service image to the previous version.
- The new `MissionDefinition` and `MissionPool` tables are not read by the old code path, so rollback is safe for offers and contracts created before this change.
- Offers created by the new endpoint are not editable by the old endpoint, so a rollback would make them read-only through the old path.

## Monitoring

- Alert on `NO_ACTIVE_SEASON`, `DEFINITION_NOT_PUBLISHED`, and `MISSION_DEFINITION_ALREADY_ACTIVE` error rates.
- Track offer publish latency from pool/definition.
