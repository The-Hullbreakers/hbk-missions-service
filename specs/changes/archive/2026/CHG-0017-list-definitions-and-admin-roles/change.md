---
change_id: CHG-0017
title: List mission definitions and guard admin endpoints by role
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/archive/2026/CHG-0017-list-definitions-and-admin-roles/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
context_scope: existing
created: 2026-10-05
updated: 2026-10-05
---

# List Mission Definitions and Guard Admin Endpoints by Role

## Intent

Expose `GET /api/v1/missions/definitions` so administrators can see existing
definitions and their published versions before creating offers, and enforce
role-based authorization on administrative endpoints so only `admin`/`moderator`
callers can mutate definitions, seasons, and offers.

## Business Value

- The admin panel can list definitions and versions instead of forcing blind ID entry.
- Administrative mutations are protected server-side, not only by UI hiding.

## User Stories

- **US-MD-LIST**: As an admin, I want `GET /api/v1/missions/definitions`, so I can pick an existing definition and version when publishing offers.
- **US-ADMIN-ROLE**: As a platform owner, I want admin endpoints to reject callers without `admin`/`moderator` roles, so members cannot create or mutate catalog data.

## Why L2

New read endpoint plus authorization guard on existing admin routes; no domain
model change, no contract break for players (player-facing GETs stay
authenticated-only).

## Scope

In:
- `DefinitionStore.List` DynamoDB query on `PK=DEFINITION`.
- Application `ListDefinitionsHandler` + `DefinitionListReader` port.
- `GET /api/v1/missions/definitions` HTTP handler.
- `RequireRole` HTTP middleware using `session.UserSession.HasRole`.
- Wiring in `composition/server.go`.

Out:
- Frontend admin panel redesign (tracked in hbk-web-app change).
- Deriving tenant_id from JWT on mutation bodies (separate follow-up).

## Defect: BUG-0017-01 — Gateway route missing for list definitions

- **Reported:** 2026-10-08 — `GET /api/v1/missions/definitions` returned
  `403 {"message":"Missing Authentication Token"}` for a valid `admin` token.
- **Expected:** admin/moderator callers reach `ListDefinitionsHandler` and get 200.
- **Actual:** shared API Gateway rejected the request; Lambda never invoked.
- **Classification:** infrastructure defect (L1) in `template.yaml`.
- **Root cause:** CHG-0017 wired `GET /api/v1/missions/definitions` in
  `composition/server.go` but no `AWS::ApiGatewayV2::Route` was declared, so the
  gateway had no route mapping for it. All sibling routes existed; the GET
  definitions route was omitted.
- **Regression evidence (RED):** `grep 'GET /api/v1/missions/definitions'
  template.yaml` returned no match while `server.go:69` served the path.
- **Fix:** added `MissionsListDefinitionsRoute` (`RouteKey: GET
  /api/v1/missions/definitions`) targeting `MissionsAPIIntegrationV2`.
- **Verification (GREEN):** `sam validate -t template.yaml` — valid; route count
  for the RouteKey = 1.
- **Residual risk:** deployed environments need re-deploy of the stack for the
  route to materialize; verify with a live request after deploy.
