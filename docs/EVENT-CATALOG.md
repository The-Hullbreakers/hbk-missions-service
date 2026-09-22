# Proposed Event Catalog

All events include identity, version, time, correlation/causation, and tenant. Consumers validate schema, source and idempotence.

## Published By Missions

| Event | Meaning |
| --- | --- |
| `missions.offer.rotation-published` | A cycle published its set of predefined offers. |
| `missions.contract.accepted` | An individual or team contract was created. |
| `missions.contract.started` | Modality and roster were frozen. |
| `missions.objective.confirmed` | A valid unit was awarded. |
| `missions.contract.completed` | The contract reached its accepted result. |
| `missions.contract.failed` | It ended unsuccessfully and rightly so. |
| `missions.merit.credited` | Added a credit to the ledger. |
| `missions.merit.adjusted` | An auditable correction was applied. |
| `missions.season.frozen` | Acceptance of new merits was closed. |
| `missions.season.finalized` | The final standings became immutable. |
| `missions.season.placement-awarded` | One recipient qualified for an achievement tier. |

## Placement Award Payload

```json
{
  "seasonId": "2956-S3",
  "category": "BOUNTY_HUNTING",
  "track": "INDIVIDUAL",
  "recipientId": "user-id",
  "rank": 1,
  "tier": "CHAMPION",
  "finalMerits": 18500
}
```

Achievements uses this fact to materialize permanent achievement. The `event_id` and the season/category/track/recipient/tier identity prevent duplicates.

## Consumed By Missions

- Party snapshot/confirmation when the contract uses an existing team.
- Confirmations from explicitly reliable sources for objectives.
- Award materialization result for observability/retry only; does not change standings.

An arrival from a queue is not considered reliable on its own. Source, tenant, contract, target and receiver must match local state.
