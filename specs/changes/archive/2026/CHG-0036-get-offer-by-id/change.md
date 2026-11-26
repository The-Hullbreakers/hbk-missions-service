---
change_id: CHG-0036
title: Read a single offer by ID
status: completed
risk: L1
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/active/2026/CHG-0036-get-offer-by-id/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
context_scope: existing
created: 2026-10-14
updated: 2026-10-14
---
# Read a single offer by ID

## Intent

The API contract documents querying an offer by ID, but only the list endpoint exists. This change adds the single-offer read.

## Business Value

- Clients can refresh one offer (capacity, availability) without listing the whole board.

## User Stories

- **US-MB-OFFER-GET**: As a player, I want to fetch one offer by ID, so I can check availability before accepting.

## Why L1

Additive read-only endpoint with an existing port (`OfferReader.Get`); no domain or persistence change.

## Scope

- **Application**: `GetOfferHandler` query.
- **Boundary**: `GET /api/v1/missions/offers/{offer_id}` returning the same offer projection as the list items.

## Non-Goals

- Filtering or embedding contract counts.

## Approvals

- Gate 1: approved 2026-10-16 — reuse `OfferReader.Get`; new sentinel `ErrOfferNotAvailable`→404 `OFFER_NOT_AVAILABLE`; `OfferStore.Get` maps missing item to the sentinel; member-facing `GET /offers/{offer_id}` behind auth. Two tasks: T-036-001 (sentinel+store+code), T-036-002 (handler+route).
