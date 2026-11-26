---
change_id: CHG-0015
title: List mission offers and player contracts over HTTP
status: active
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0015-list-offers-and-contracts/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_AWS_LAMBDA_REST_WORKFLOW
context_scope: existing
created: 2026-09-29
updated: 2026-09-29
---

# List Mission Offers and Player Contracts Over HTTP

## Intent

Expose the already-specified mission-board discovery and contract listing through two read-only HTTP endpoints so clients (e.g. `MissionsPage`) no longer need to know `offerId`/`contractId` upfront.

## Business Value

- Members can browse available offers for the active season instead of typing identifiers.
- Members can see their own contracts and their state without manual input.
- Tenant and active-template isolation invariants are exercised at the real HTTP boundary.

## User Stories

- **US-MB-LIST**: As a member, I want `GET /api/v1/missions/offers` filtered by season and status, so I can discover missions I can accept.
- **US-MC-LIST**: As a member, I want `GET /api/v1/missions/contracts` filtered by season and status, so I can see my contracts and their progress.

## Why L2

Read-only public boundary evolution of two existing capabilities. No domain lifecycle change; the change adds query use cases, DynamoDB list adapters, HTTP handlers, and API Gateway routes. Tenant isolation is a trust-boundary concern and is covered by targeted tests; no new credentials or secrets.

## Scope

- **ADDED**: `GET /api/v1/missions/offers` (filters: `season_id`, `status`, `category`, `mode`; `type`/`difficulty` dropped — `MissionOffer` lacks those fields).
- **ADDED**: `GET /api/v1/missions/contracts` (filters: `season_id`, `status`; always scoped to the caller in the roster).
- **NOT ADDED**: `GET /seasons/current` — redundant; `GET /api/v1/seasons?status=ACTIVE` already exists and defaults to `ACTIVE`.
- **FIX (approved)**: honor `tenantID` in `OfferStore`/`ContractStore` list paths instead of ignoring it (upstream minimal fix, not a workaround).

## Layer Gates

- `LAYER-GATE-DOMAIN`: filter value objects and offer-hiding rule reuse `REQ-MB-008` policy.
- `LAYER-GATE-APPLICATION`: `ListOffersHandler`, `ListPlayerContractsHandler` with one-operation reader ports.
- `LAYER-GATE-INFRASTRUCTURE`: `OfferStore.List`, `ContractStore.ListByPlayer` (tenant-scoped).
- `LAYER-GATE-INTERFACE`: thin handlers in `interface/http/` + routes in `composition/server.go` + `AWS::ApiGatewayV2::Route` entries in `template.yaml`.

## Gates

- Gate 1 (plan approval): approved 2026-09-29 — two GET endpoints, no `seasons/current`, tenantID fix included.
- Gate 2 (test approval): pending — artifacts and first RED to be shown.
- Gate 3 per layer: RED evidence required before mapped production files.
- Security gate: not selected; read-only endpoints, existing auth middleware enforces authentication, tenant isolation covered by tests.
- Coverage gate: selected per L2 non-trivial list/filter logic — `>= 90%` on affected production scope.
- Mutation/critical-E2E: not selected for L2.
