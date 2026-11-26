---
change_id: CHG-0016
title: Accept access_token HttpOnly cookie at the auth boundary
status: consolidated
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0016-accept-access-token-cookie/
execution_mode: sequential
primary_workflow: WORKFLOW-COMMON_SDD_FIX_BUG_WORKFLOW
context_scope: existing
created: 2026-10-02
updated: 2026-10-02
---

# Accept access_token HttpOnly Cookie at the Auth Boundary

## Defect

- **BUG-001 — Cookie-authenticated browser requests rejected with 401**: `AuthMiddleware.Authenticate` (`internal/missions/interface/http/middleware.go`) reads only the `Authorization` header. HBK Identity issues the OIDC token as an HttpOnly `access_token` cookie, so browser clients that cannot set headers receive `401 UNAUTHORIZED` on every `/api/v1/**` route (reproduced: `GET /api/v1/seasons?status=ACTIVE` via `https://localhost:8443`).

## Classification

- Production defect. Expected behavior: a valid token carried in the `access_token` cookie authenticates the request, matching the sibling-service precedent (`hbk-sc-tools-service` `TokenExtractor`: header first, then `access_token`/`auth_token` cookies).
- Actual: `401` whenever the header is absent, regardless of cookies.

## Root Cause

`Authenticate` short-circuits on empty `Authorization` before any token extraction; no cookie fallback exists.

## Scope

- **FIX**: cookie fallback in `AuthMiddleware.Authenticate` — header keeps precedence; absent header → read `access_token` cookie, validate through the unchanged `session.UserSession.GetUserIDFromHeader` port.
- **NOT CHANGED**: session verification, audience rules, route table, canonical requirement semantics (`401` still applies to missing/invalid credentials).

## Regression Boundary

`tests/unit/missions/interface/http/auth_middleware_test.go` — lowest owning level (interface layer).

## Risks

- Cookie is an ambient credential → CSRF note recorded for mutating routes; Identity cookies are expected `SameSite`/`Secure`. No new secrets or trust stores.

## Rollback

Revert the middleware fallback; header-only contract remains valid.

## Gates

- Gate 1 (plan approval): approved 2026-10-02.
- Gate 3: required after RED before `middleware.go` production change.
- Security gate: recorded — trust-boundary touch; mitigation = unchanged token verification, documented SameSite expectation.
- Coverage gate: affected scope `>= 90%`.
