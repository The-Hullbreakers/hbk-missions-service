---
change_id: CHG-0003
title: Make confirm objective and credit merit prod-ready
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
change_path: specs/changes/active/CHG-0003-confirm-objective-prod-readiness/
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
archive_path: specs/changes/archive/2026/CHG-0003-confirm-objective-prod-readiness/
verified_at: 2026-08-29
source_revision: CHG-0003-confirm-objective-prod-readiness
context_scope: existing
created: 2026-08-29
updated: 2026-08-29
---

# Make Confirm Objective And Credit Merit Prod-Ready

## Intent

Take the verified US-MCF-001 use case and make it a runnable, multi-tenant, observable production service with durable persistence, caller identity, and safe error responses.

## Business Value

- Merit credits survive process restarts and can be audited.
- Only authenticated and authorized callers can credit merits.
- Operators can deploy the missions service as a standalone HTTP service.
- Public error responses do not leak internal domain state.

## User Stories

- **US-MCF-PROD-001**: As an operator, I want the service to start with a real database, so that the credits persist and can be audited.
- **US-MCF-PROD-002**: As a player, I want only authorized calls to credit merits on my behalf, so that no one fakes my score.
- **US-MCF-PROD-003**: As an operator, I want to deploy the service as a process HTTP, to integrate it with other boundaries of HBK.
- **US-MCF-PROD-004**: As a client of API, I want public error codes, to handle failures without depending on internal strings.

## Why L3

The change touches the trust boundary (authn/authz), data durability (persistence), deployment, and public contract. A defect can expose data, allow unauthorized credits, lose merit records, or leak internal state.

## Scope

- Replace in-memory read/write ports with DynamoDB adapters, like the rest of HBK.
- Add `cmd/hbk-missions-service/main.go` composition root with DI.
- Add authn/authz middleware to the HTTP boundary using HBK Identity OIDC tokens, following the pattern of `hbk-sc-tools-service`.
- Map domain errors to public API error codes.
- Add health/observability endpoints (health, ready).
- Provide `Dockerfile` and `docker-compose.yml` for local deployment.
- Update canonical spec with persistence, auth, and deployment contract.

## Non-Goals

- Distributed transactions or event sourcing.
- Full OAuth/OIDC provider implementation (use existing identity token from HBK Identity).
- Leaderboards UI or Achievements integration.
- Procedural contract generation.

## Acceptance Summary

- The service starts from `main.go` and serves `POST /api/v1/missions/contracts/{contract_id}/evidence`.
- Unauthenticated or untenanted requests return `401`.
- Unauthorized tenant or player scope returns `403`.
- A valid request with real persistence survives process restart and the same total is returned.
- Domain errors return a stable public `code` and never `err.Error()`.
- The existing `go test -race ./...` suite keeps passing with in-memory adapters for unit tests.
- Integration tests run with Testcontainers against a real PostgreSQL.

## Current Outcome

Gate 1 approved on August 29, 2026. The active change artifacts are created.
