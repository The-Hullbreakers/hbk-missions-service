---
change_id: CHG-0062
title: Restore the staging offer facets Gateway route
status: locally_verified
risk: L2
canonical_action: not_affected
canonical_paths:
  - specs/capabilities/mission-board/spec.md
execution_mode: sequential
---

# Staging offer facets route

User request authorizes an operational boundary repair: live `GET
/api/v1/missions/offers/facets` returns a generic API Gateway 404; the offers
API is otherwise live (authenticated routes answer 401) and all 42 category
records are verified present. The authenticated facets handler already exists
— canonical mission-board contract `GET /api/v1/missions/offers/facets`
(spec.md:319) returns distinct `categories`/`modes` under the resolved season
and status. Root cause: `template.yaml` never declared the route — same
defect class as CHG-0059/CHG-0061.

Scope: L2 template-only fix. No Go, CORS, IAM, category data, or other route
changes; no canonical/index edits; no claims of authenticated live access.

## Acceptance scenarios

- `template.yaml` declares `MissionsOfferFacetsRoute`
  (`AWS::ApiGatewayV2::Route`, `<<: *RouteProps`,
  `RouteKey: GET /api/v1/missions/offers/facets`) so the request reaches the
  shared Lambda integration and existing authentication instead of a Gateway
  404.
- All neighboring routes unchanged (`/offers/facets` literal is distinct from
  `/offers/{offer_id}` parameterization).

## Layer gates / verification

Domain and application: not affected. Boundary BDD:
TEST-STG-OFFER-FACETS-ROUTE-001 in `tools/test-workflow-contract.sh` asserts
the 5-line route block (type, `*RouteProps` anchor, exact RouteKey) — static
check only. GREEN support: existing unit tests `TestListOfferFacets*` keep
the application/HTTP contract green. Security: auth remains in Go middleware.
Rollback: remove the route resource.
