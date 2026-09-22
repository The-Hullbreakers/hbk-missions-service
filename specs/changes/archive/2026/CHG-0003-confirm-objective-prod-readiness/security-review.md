---
change_id: CHG-0003
artifact: security-review
status: completed
updated: 2026-08-29
---

# Security Review

## Trust Boundaries

- **Client → HTTP boundary**: TLS terminated in ingress; token OIDC `Authorization: Bearer` of HBK Identity.
- **Auth middleware → handler**: validation with `go-oidc` + `golang-jwt`; context with `callerID`, `tenantID`, `roles`.
- **Handler → application**: validated input; body `tenantID` cross-checked against the token's `tenantID`.
- **Application → persistence**: DynamoDB with strict keys and attributes, without query injection.

## Threats and Mitigations

| Threat | Mitigation | Evidence |
| --- | --- | --- |
| Forge merit request | OIDC token + tenant scope check | `TEST-AUTH-001`, `TEST-AUTH-002` |
| Injection in queries | DynamoDB with controlled keys/attributes; not concatenated strings | `go vet`, `gosec` (manual) |
| Information disclosure | Public error codes, no `err.Error()` in response | `TEST-ERROR-001` |
| Privilege escalation | Roles/scope check in middleware | `TEST-AUTH-002` |
| Replay/duplicate | Domain `ConfirmObjective` idempotency preserved | `TestMeritCredit_DuplicateTargetIsRejected` |

## Findings

- No Critical/High findings.
- `gosec` is not installed; manual review performed.
- No hardcoded secrets; environment variables for OIDC and DynamoDB.
- `JWTUserSession` validates issuer, client_id and audience; roles are extracted from claims.
- `ConfirmEvidenceHandler` does not expose internal error messages; returns public codes.
- DynamoDB uses `PutItem` and `GetItem` with strict keys; no string concatenation.
