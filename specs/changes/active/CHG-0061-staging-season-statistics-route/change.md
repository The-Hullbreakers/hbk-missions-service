---
change_id: CHG-0061
title: Restore the staging season statistics Gateway route
status: locally_verified
risk: L2
canonical_action: not_affected
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
execution_mode: sequential
---

# Staging season statistics route

User request authorizes an operational boundary repair: live `GET
/api/v1/missions/seasons/{season_id}/statistics?top_n=3` returns a generic
API Gateway 404 with no CORS while `OPTIONS` returns 204 with the correct
origin; `get-routes` confirms the route is absent from the deployed API. The
authenticated Go handler already exists (composition server mux, canonical
REQ-SCC-017). Root cause: `template.yaml` never declared the route — same
class of defect as CHG-0059.

Scope: L2 template-only fix. No Go, CORS, IAM, or other route changes; no
canonical/index edits; no claims of authenticated live access.

## Acceptance scenarios

- `template.yaml` declares `MissionsSeasonStatisticsRoute`
  (`AWS::ApiGatewayV2::Route`, `<<: *RouteProps`,
  `RouteKey: GET /api/v1/missions/seasons/{season_id}/statistics`) so the
  request reaches the shared Lambda integration and its existing
  authentication instead of a Gateway 404.
- All neighboring routes unchanged.

## Layer gates / verification

Domain and application: not affected. Boundary BDD:
TEST-STG-SEASON-STATISTICS-ROUTE-001 in `tools/test-workflow-contract.sh`
asserts the 5-line route block (type, `*RouteProps` anchor, exact RouteKey)
— static check, no AWS calls. GREEN support: existing unit tests
`TestGetSeasonStatistics*` prove the application contract the route exposes.
Security: existing route auth is Go middleware; `*RouteProps` reuses the
same integration. Rollback: remove the route resource.
