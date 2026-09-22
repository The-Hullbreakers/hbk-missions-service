---
change_id: CHG-0003
artifact: design
status: proposed
updated: 2026-08-29
layer_scope:
  domain: affected (error typing, contract state reconstruction)
  application: affected (ContractWriter port, handler saves contract)
  boundary: affected
  infrastructure: affected
  interface: affected
  composition: affected
---

# Business-Centered Design

## Bounded Context

`MissionContract Fulfillment` remains the proprietary capability. No new capability is introduced.

## Adapters

- `ContractStore`, `SeasonStore`, `ObjectiveStore`, `MeritLedgerWriter` are application ports.
- DynamoDB adapters are added under `internal/missions/infrastructure/dynamodb/`.
- Unit tests keep doubles in memory.
- The integration tests use DynamoDB Local (container) following the `hbk-sc-tools-service` pattern.

## Composition

- `cmd/hbk-missions-service/main.go` reads configuration, creates `DynamoDBClient` with `aws-sdk-go-v2`, creates repositories, injects `ConfirmObjectiveHandler` and bootstraps `http.Server`.
- `internal/missions/composition/server.go` mounts handlers, middleware and routes.

## Auth Boundary

- Middleware `AuthMiddleware` in `internal/missions/interface/http/middleware.go`.
- Extract header token `Authorization: Bearer <token>` (and optionally cookies, as in `hbk-sc-tools-service`).
- Validates OIDC token against the HBK Identity issuer using `github.com/coreos/go-oidc/v3` and `github.com/golang-jwt/jwt/v5`, following the `internal/catalogs/infrastructure/session/jwt_user_session.go` pattern.
- Injects `callerID`, `tenantID` and `roles` into `context.Context`.
- Handler extracts `tenantID` from the context and compares it with the `tenant_id` of the body.
- Supported roles: `roles`, `tenant_roles`, `realm_access.roles`, `resource_access.<clientId>.roles`.

## Error Mapping

- `internal/missions/interface/http/errors.go` maps domain errors to public codes.
- `ConfirmEvidenceHandler` never returns `err.Error()` directly.

## Persistence Principles

- Single DynamoDB table with prefix per entity (`PK`/`SK` or `entityType`).
- Entities: `contract`, `season`, `objective`, `merit_credit`.
- `merit_credit` insert-only; `TotalFor` performs `Query` with `GSI` per bucket (`tenant`, `season`, `category`, `track`, `recipient`).
- No migrations; schema-free using `PutItem` and `Query` with typed attributes.
- Domain remains pure; only adapters know AWS.

## Rollback

- Rollback preserves existing credits.
- The acceptor of new confirmations by feature flag/config is disabled.
