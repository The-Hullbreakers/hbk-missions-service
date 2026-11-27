---
change_id: CHG-0062
artifact: delta
status: implemented
updated: 2026-11-26
---

# Delta — Staging offer facets route

## What Is Changing

### Deployment template (operational)

- `template.yaml` gains `MissionsOfferFacetsRoute`
  (`AWS::ApiGatewayV2::Route`, `<<: *RouteProps`,
  `RouteKey: GET /api/v1/missions/offers/facets`) inserted immediately after
  `MissionsListOffersRoute` / before `MissionsListContractsRoute`, wiring the
  existing authenticated facets handler (mission-board contract,
  spec.md:319) to the Gateway.

## MODIFIED

None — `canonical_action: not_affected`. The facets contract and handler
already exist; this delta only restores the deployed boundary route.

## BDD Scenarios

### SCN-STG-OFFER-FACETS-ROUTE — Facets GET reaches the Lambda

Given the SAM template, when API Gateway routes are declared, then
`MissionsOfferFacetsRoute` maps `GET /api/v1/missions/offers/facets` onto the
shared Lambda integration (`*RouteProps`), so the request reaches
authentication rather than a generic Gateway 404.

## HTTP Boundary Changes

- `GET /api/v1/missions/offers/facets` — route now reaches the existing
  authenticated handler. Behavior, auth, and error vocabulary unchanged.

## Error Vocabulary

- Unchanged — no new codes.
