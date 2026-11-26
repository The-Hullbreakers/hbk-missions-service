---
change_id: CHG-0060
title: Normalize API Gateway stage-prefixed paths at the Lambda boundary
status: locally_verified
risk: L2
canonical_action: not_affected
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
execution_mode: sequential
---

# Lambda stage-prefix routing

User request authorizes a local boundary fix for the deployed staging Lambda.
Observed behavior: invoking `GET /api/v1/missions/seasons/current` with a
clean `RawPath` returns the expected `401` JSON (the existing authenticated
season-current contract); the same request with a `/stg` stage-prefixed raw
path returns a plain `404` — matching the public response on the named `stg`
stage with a root mapping. AWS documents that HTTP API Gateway may deliver
stage-prefixed raw paths; the Lambda must tolerate both forms.

Scope: L2 operational boundary repair at the Lambda event adapter only. No
IAM, auth, CORS, template, or data changes; no domain/application change.

## Acceptance scenarios

- Named stage prefix: `Stage: stg`, `RawPath: /stg/api/v1/missions/seasons/
  current` routes identically to the clean path (`401 Unauthorized`
  unauthenticated, not 404).
- Clean path on a named stage: unchanged (`401`).
- Fallback: `RawPath` empty uses `RequestContext.HTTP.Path` for prefix
  detection (`401`).
- Exact-prefix only: `/stg-other/...` is NOT stripped (404); `/prod/...` on
  stage `stg` is NOT stripped (404); `$default` stage is never stripped.
- Clean-path invocations and managed-table startup (CHG-0059) unaffected.

## Layer gates / verification

Domain and application: not affected. Boundary BDD:
TEST-LAMBDA-STAGE-ROUTING-001 builds the real composition root
(`UseExistingTable`, no DynamoDB, no Docker) and invokes the new
`lambdahttp.NewProxy` seam with synthetic `APIGatewayV2HTTPRequest` events
across all cases above. Security: normalization is stage-exact and cannot
redirect across stages; auth remains in Go middleware. Rollback: revert the
seam and the `cmd/lambda` call site.
