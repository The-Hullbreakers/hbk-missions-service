---
change_id: CHG-0003
artifact: canonical-delta
status: proposed
updated: 2026-08-29
---

# Canonical Delta

## MODIFIED — Mission Contract Fulfillment

### REQ-MCF-PROD-001 — Persist contracts, seasons, objectives and credits

The read and write ports in the use case must be deployable with durable storage in DynamoDB, like the rest of the HBK services. The domain does not know technology.

#### SCN-MCF-PERSIST-001 — Credit survives reboot

Given a merit credit was persisted to DynamoDB, when the process is restarted, then `TotalFor` of the receiver remains the same.

### REQ-MCF-PROD-002 — Serve the use case as HTTP

The service exposes a HTTP server with confirmed endpoint, health, and readiness. The composition root injects real adapters in prod conditions and doubles in tests.

#### SCN-MCF-COMPOSITION-001 — Start the service

Given the connection configuration is valid, when `cmd/hbk-missions-service/main.go` is executed, then the server responds on the configured port and `GET /healthz` returns `200`.

### REQ-MCF-PROD-003 — Autenticar y autorizar al llamante

Each HTTP request must carry an OIDC token of HBK Identity with valid identity and tenant. The service rejects untokened or out-of-scope calls, following the pattern in `hbk-sc-tools-service`.

#### SCN-MCF-AUTH-001 — Reject request without identity

Given a request without an identity token, when sent to `POST /api/v1/missions/contracts/{contract_id}/evidence`, then the server responds `401 Unauthorized`.

#### SCN-MCF-AUTH-002 — Rechazar tenant no autorizado

Given a valid token but no access to the body's `tenant_id`, when a confirmation is sent, then the server responds `403 Forbidden`.

### REQ-MCF-PROD-004 — Post public error codes

Boundary `4xx` responses return stable codes (`CONTRACT_NOT_IN_PROGRESS`, `DUPLICATE_CONFIRMATION`, `SEASON_NOT_ACTIVE`, etc.) instead of domain strings.

#### SCN-MCF-ERROR-CODE-001 — Mapear error de dominio

Given a confirmation about an unstarted contract, when the request is sent, then the response `422` contains `{"code":"CONTRACT_NOT_IN_PROGRESS"}` and not `err.Error()`.

## ADDED — Operations and Deployment

### REQ-OPS-001 — Health y readiness

The service exposes `GET /healthz` and `GET /readyz` for orchestrators and balancers.

#### SCN-OPS-HEALTH — Basic health check

Given the server is running, when `/healthz` is queried, it then responds `200 OK` without depending on the database.

## REMOVED

Existing behavior is not eliminated. In-memory adapters are still available for unit and double integration tests.
