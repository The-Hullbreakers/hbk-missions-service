---
change_id: CHG-0060
artifact: design
status: implemented
updated: 2026-11-26
---

# Design — Lambda stage-prefix routing

## Domain / Application

- Not affected. Routing, authorization, and handlers are untouched.

## Boundary (interface/lambdahttp)

- New package `internal/missions/interface/lambdahttp` with
  `NewProxy(handler http.Handler)` — a thin adapter seam around
  `httpadapter.NewV2(handler).ProxyWithContext`. This is the only place that
  knows about API Gateway event shapes; the HTTP mux stays gateway-agnostic.
- Normalization rule: when `Stage` is a named stage (`!= ""` and
  `!= "$default"`), compute `path = RawPath`, falling back to
  `RequestContext.HTTP.Path` when `RawPath` is empty. If `path == "/"+stage`,
  `RawPath` becomes `"/"`; if `path` starts with `"/"+stage+"/"`, the prefix
  is trimmed. Anything else — including similar-looking prefixes
  (`/stg-other`) and foreign stages (`/prod` on `stg`) — passes through
  unchanged and resolves to the router's 404.
- `cmd/lambda/main.go` swaps the direct `httpadapter.NewV2(handler).
  ProxyWithContext` registration for `missionlambda.NewProxy(handler)`; no
  other wiring changes (`UseExistingTable` preserved).

## Test strategy

- TEST-LAMBDA-STAGE-ROUTING-001 (`tests/integration/composition/
  lambda_stage_routing_test.go`, package `composition_test`): builds the real
  composition root with `UseExistingTable: true` (no DynamoDB client needed —
  init makes zero AWS calls; no Docker) and drives the proxy with synthetic
  `APIGatewayV2HTTPRequest` events. Table-driven cases assert `401` for the
  normalized/fallback/clean cases and `404` for non-exact prefixes, a foreign
  stage, and verify `$default` never strips.
- Behavioral RED first: `NewProxy` is introduced as a pass-through harness so
  the test compiles; stage-prefixed and fallback cases then fail with 404 vs
  expected 401, proving detection before normalization is implemented.

## Security / cost

- Prefix matching is exact-segment only and keyed to the event's own stage;
  it cannot make a route reachable under a different stage's path. Auth is
  unchanged (Go middleware). One string comparison per invocation — cost
  neutral.
