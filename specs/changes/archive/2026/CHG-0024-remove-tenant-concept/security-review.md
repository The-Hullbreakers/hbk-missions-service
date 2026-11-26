---
change_id: CHG-0024
artifact: security-review
status: proposed
updated: 2026-10-08
---

# Security Review — Remove Tenant Concept

## Trust Boundaries

- **AuthN**: unchanged — OIDC token via `Authorization` header or
  `access_token` cookie; `AuthMiddleware` still rejects unauthenticated calls.
- **Identity derivation**: unchanged principle — `player_id`, `roles`,
  `team_ids` come only from the validated session; body identity fields remain
  ignored. The `tenant_id` claim is simply never read.
- **AuthZ**: roster membership plus `admin`/`moderator` roles still gate
  evidence submission and lifecycle transitions; removing tenant does not
  widen authority.

## Removed Boundary

The per-tenant isolation boundary is deleted by product decision: the identity
server issues no tenant claim and the platform runs one scope. There is no
"wrong tenant" state to reach.

## Risks

- **Token replay across environments**: a token valid for this service's
  issuer now grants access to the whole (single) mission scope — same as any
  authenticated user before; audience/issuer validation is unchanged.
- **Body `tenant_id` silently ignored**: consistent with the approved
  forgiving-derive decision; documented for consumers.
- **Orphaned tenant-keyed items**: old rows unreachable post-deploy; treated
  as data-hygiene, not a security exposure (no cross-scope leakage exists).

## Findings

- **Critical/High**: none.
- **Informational**: stale tenant-prefixed DynamoDB items become unreachable
  under the new keys (see `rollout.md` for the operator decision); forged
  `player_id`/`tenant_id` body fields remain ignored by every handler —
  covered by `forged_identity_test.go`.
- Authority checks preserved: roster membership and `admin`/`moderator` roles
  still gate evidence and lifecycle transitions; tests verify no widened
  access after the tenant removal.
