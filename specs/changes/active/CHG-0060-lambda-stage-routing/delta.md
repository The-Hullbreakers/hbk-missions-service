---
change_id: CHG-0060
artifact: delta
status: implemented
updated: 2026-11-26
---

# Delta — Lambda stage-prefix routing

## What Is Changing

### Lambda event boundary (operational)

- New seam `internal/missions/interface/lambdahttp.NewProxy(handler)` returns
  the Lambda handler function `func(context.Context,
  events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)`.
- When `RequestContext.Stage` is a named stage (non-empty, not `$default`),
  the proxy strips an exact `/<stage>` or `/<stage>/…` prefix from the raw
  path (falling back to `RequestContext.HTTP.Path` when `RawPath` is empty)
  before delegating to `httpadapter.NewV2(handler).ProxyWithContext`.
- `cmd/lambda/main.go` starts Lambda with `missionlambda.NewProxy(handler)`
  instead of the raw adapter; all other wiring — including
  `UseExistingTable: true` — is unchanged.

## MODIFIED

None — `canonical_action: not_affected`. The season-current contract and its
authentication already exist; this delta only normalizes the gateway event
shape at the adapter.

## BDD Scenarios

### SCN-STAGE-PREFIX-NORMALIZED — Named stage routes like the clean path

Given `Stage: stg` and `RawPath: /stg/api/v1/missions/seasons/current`, when
the Lambda is invoked, then the router sees
`/api/v1/missions/seasons/current` and the response is `401 Unauthorized`
(auth boundary reached), not a plain 404.

### SCN-STAGE-PREFIX-EXACT — Only the exact stage prefix is stripped

Given `Stage: stg`, when `RawPath` is `/stg-other/api/...` or
`/prod/api/...`, then nothing is stripped and the response is `404`.
Given `Stage: $default`, no stripping occurs regardless of path.

### SCN-STAGE-FALLBACK-PATH — Empty RawPath uses HTTP.Path

Given `Stage: stg`, `RawPath` empty and `HTTP.Path:
/stg/api/v1/missions/seasons/current`, when invoked, then the prefix is
detected from `HTTP.Path`, `RawPath` is normalized, and the response is
`401`.

## HTTP Boundary Changes

- Internal adapter normalization only; public contract unchanged. Responses
  on stage-prefixed invocations now match clean-path responses.

## Error Vocabulary

- Unchanged — no new codes.
