---
change_id: CHG-0061
artifact: delta
status: implemented
updated: 2026-11-26
---

# Delta — Staging season statistics route

## What Is Changing

### Deployment template (operational)

- `template.yaml` gains `MissionsSeasonStatisticsRoute`
  (`AWS::ApiGatewayV2::Route`, `<<: *RouteProps`,
  `RouteKey: GET /api/v1/missions/seasons/{season_id}/statistics`) inserted
  immediately before `MissionsFreezeSeasonRoute`, wiring the existing
  authenticated statistics handler (canonical REQ-SCC-017) to the Gateway.

## MODIFIED

None — `canonical_action: not_affected`. The statistics contract and handler
already exist; this delta only restores the deployed boundary route.

## BDD Scenarios

### SCN-STG-STATISTICS-ROUTE — Statistics GET reaches the Lambda

Given the SAM template, when API Gateway routes are declared, then
`MissionsSeasonStatisticsRoute` maps
`GET /api/v1/missions/seasons/{season_id}/statistics` onto the shared Lambda
integration (`*RouteProps`), so the request reaches authentication rather
than a generic Gateway 404.

## HTTP Boundary Changes

- `GET /api/v1/missions/seasons/{season_id}/statistics` — route now reaches
  the existing authenticated handler. Behavior, auth, and error vocabulary
  unchanged.

## Error Vocabulary

- Unchanged — no new codes.
