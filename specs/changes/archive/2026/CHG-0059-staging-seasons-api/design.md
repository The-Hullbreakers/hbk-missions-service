---
change_id: CHG-0059
artifact: design
status: implemented
updated: 2026-11-26
---

# Design — Staging seasons API repair

## Domain / Application

- Not affected. The current-season handler and its authorization already
  exist; nothing in domain or application code changes.

## Infrastructure (template.yaml)

- Additive route resource `MissionsCurrentSeasonRoute` before
  `MissionsLeaderboardRoute`, reusing the `*RouteProps` anchor for the API
  and Lambda integration. Authentication remains in the existing Go route
  middleware, not in that anchor. Route ordering vs. `GET /api/v1/seasons`
  is safe: API Gateway matches the most specific literal path.

## Composition boundary

- `Config.UseExistingTable` (bool, default false) documents that table
  provisioning belongs to the infrastructure owner (SAM `IMPORT_COMPLETE`).
- `NewServer` wraps the existing `CreateMissionsTable` call in
  `if !cfg.UseExistingTable { ... }` — the create call, its error wrapping,
  and local/test behavior are otherwise untouched. No branch beyond this
  guard, no mux/auth edits.
- `cmd/lambda/main.go` passes `UseExistingTable: true` unconditionally: every
  deployed stack owns its table via SAM, and the role must not gain
  `dynamodb:CreateTable`. No environment-variable fallback and no extra AWS
  call — one fewer API call at cold start (cost-neutral to cheaper).

## Test strategy

- TEST-STG-CURRENT-SEASON-ROUTE-001 (`tools/test-workflow-contract.sh`):
  greps the 5-line `MissionsCurrentSeasonRoute:` block for the exact
  `RouteKey: GET /api/v1/missions/seasons/current` line — static template
  contract, no AWS calls.
- TEST-STG-MANAGED-TABLE-001 (`tests/integration/composition/
  managed_table_test.go`, package `composition_test`): `httptest` loopback
  DynamoDB stub answers every call with `AccessDeniedException`; a request
  counter proves zero calls at init. The test drives the real composition
  root through `httpadapter.NewV2(...).ProxyWithContext` with an
  `APIGatewayV2HTTPRequest` for `/api/v1/missions/seasons/current` and
  asserts `401 Unauthorized` — proving init succeeded and the request reached
  the auth boundary. No Docker, no shared writes; the test owns its stub.
- Behavioral RED before GREEN: the field is added first as a harness-only
  contract so the test compiles; against unguarded `NewServer` it fails at
  init with the denied `CreateTable` (the observed staging panic), proving
  the test detects the defect.

## Security / cost

- No IAM changes; the deploy/Lambda roles gain no permissions. Existing
  route authorization is reused (Go middleware). The change strictly removes
  an unauthorized provisioning attempt.
