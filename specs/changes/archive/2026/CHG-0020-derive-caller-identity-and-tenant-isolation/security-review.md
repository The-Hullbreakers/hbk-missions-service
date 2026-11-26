---
change_id: CHG-0020
artifact: security-review
security_role: resource-server
status: verified
updated: 2026-10-08
---

# Security Review — CHG-0020

## Role

`resource-server` — the service validates OIDC access tokens via
`AuthMiddleware`/`session.UserSession` and derives `CallerIdentity`
(tenant, player, roles, team memberships) exclusively from validated claims.
L3 requires independent security-owner verification.

## Trust Boundaries

1. **Public HTTP / API Gateway → Lambda**: all caller identity comes from the
   OIDC token validated by `AuthMiddleware`; every request-body field is
   untrusted input.
2. **Application → storage**: tenant key is part of every item identity;
   storage asserts loaded items belong to the caller tenant (defense in depth).
3. **Authority**: evidence submission restricted to roster members and
   `admin`/`moderator`; acceptance always acts as the session player.

## Protocol Decisions

- **Token validation** unchanged: issuer/audience/signature/expiry handled by
  `session.UserSession` (go-oidc); this change consumes claims only.
- **Tenant authorization**: enforced by caller-identity-derived keys on every
  application read/write plus tenant-keyed DynamoDB partitions.
- **Identity from body**: `tenant_id`/`player_id` removed from request
  structs; `acceptMissionRequest.TeamID` is an option field checked against
  session `team_ids`, never a substitute identity.
- **Scheduler**: `domain.NewSystemIdentity` marks an internal actor with no
  tenant; `ListAll` scan is the only cross-tenant read and is not exposed via
  HTTP.

## Findings Addressed

- **Critical**: cross-tenant contract creation and impersonation via body
  `tenant_id`/`player_id` (BUG-0020-01, BUG-0020-02).
- **High**: cross-tenant definition reads (BUG-0020-03).
- **High**: global season storage allowing cross-tenant lifecycle writes
  (BUG-0020-04).

## Automated Controls

| control | command | scope | result |
| --- | --- | --- | --- |
| SEC-CHECK-001 secret scan | `git diff` grep for password/secret/token/key literals | full diff | pass — no credentials in diff |
| SEC-CHECK-002 static analysis | `go vet ./...` | all packages | pass |
| SEC-CHECK-003 dependency audit | `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` (v1.8.0) | go.mod + transitives | pass — no vulnerabilities found |
| SEC-CHECK-004 identity tests | `go test ./tests/unit/missions/...` | domain/application authz | pass |
| SEC-CHECK-005 boundary security | `go test ./tests/integration/http/` forged-identity + outsider tests | HTTP entry | pass |
| SEC-CHECK-006 tenant isolation | `go test ./tests/integration/composition -run TestCrossTenantJourney_Isolation` + dynamodb tenant tests | real stores | pass |

Scanners `gosec`/`gitleaks` not installed; manual diff review + vet +
govulncheck used as alternatives. New deps (`aws-lambda-go`,
`aws-lambda-go-api-proxy`, `go-jose v4.1.4`) are current stable releases
covered by the govulncheck pass.

## Evidence Required Before Close

- Forged-identity RED → GREEN on accept and evidence endpoints — done
  (`forged_identity_test.go`: forged `tenant_id`/`player_id` ignored;
  contract written under session identity).
- `403 FORBIDDEN`/`CALLER_NOT_AUTHORIZED` for non-roster, non-privileged
  evidence — done (unit, http, and E2E).
- Cross-tenant `404`/`422` without existence leak — done (cross-tenant E2E).
- Two tenants with independent `ACTIVE` seasons and isolated standings —
  done (`season_store_tenant_test.go`, cross-tenant E2E).
- No Critical/High residual findings; secrets unchanged — confirmed.

## Residual Risk

- Existing clients still sending identity fields: silently ignored; flagged in
  API docs as deprecated fields.
- `ListAll` scheduler scan reads across tenants; internal-only use, not
  exposed on any route.

## Decision

Passed 2026-10-08. Security-owner verified: no unresolved Critical/High
findings; no exceptions required.
