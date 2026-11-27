---
change_id: CHG-0059
artifact: delta
status: implemented
updated: 2026-11-26
---

# Delta — Staging seasons API repair

## What Is Changing

### Deployment template (operational)

- `template.yaml` gains `MissionsCurrentSeasonRoute`
  (`AWS::ApiGatewayV2::Route`, `<<: *RouteProps`,
  `RouteKey: GET /api/v1/missions/seasons/current`) before
  `MissionsLeaderboardRoute`, restoring the API Gateway route to the existing
  authenticated handler.

### Composition (operational boundary, no domain/application change)

- `missioncomposition.Config` gains `UseExistingTable bool`: when set,
  `NewServer` skips the `CreateMissionsTable` provisioning call so Lambda
  init succeeds against a SAM-owned table the role may not provision.
- `cmd/lambda/main.go` sets `UseExistingTable: true` unconditionally — SAM
  owns the table in every deployed environment. No env fallback, no extra
  network call, no permission changes.
- Without the flag (tests, local), the existing create-and-fail behavior is
  preserved verbatim.

## MODIFIED

None — `canonical_action: not_affected`. The current-season read contract
already exists in seasonal-category-competition; this delta only wires the
deployed boundary back to it.

## BDD Scenarios

### SCN-STG-CURRENT-SEASON-ROUTE — Route reaches the Lambda

Given the SAM template, when API Gateway routes are declared, then
`MissionsCurrentSeasonRoute` maps `GET /api/v1/missions/seasons/current` onto
the shared Lambda integration (`*RouteProps`).

### SCN-STG-MANAGED-TABLE — Managed table starts without provisioning

Given a DynamoDB endpoint that denies `CreateTable` (AccessDenied) and
`UseExistingTable: true`, when the Lambda initializes and a `GET
/api/v1/missions/seasons/current` request arrives, then `NewServer` returns no
error, zero DynamoDB calls are made at init, and the response is `401
Unauthorized` — the request reached authentication instead of an init panic.

### SCN-STG-LOCAL-PROVISION — Local behavior preserved

Given `UseExistingTable` unset, when `NewServer` initializes, then
`CreateMissionsTable` is attempted and its errors propagate exactly as before.

## HTTP Boundary Changes

- `GET /api/v1/missions/seasons/current` — route now reaches the existing
  authenticated handler instead of a Gateway 404. Behavior, auth, and error
  vocabulary unchanged.

## Error Vocabulary

- Unchanged — no new codes.
