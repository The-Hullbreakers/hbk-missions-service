---
change_id: CHG-0045
title: Extend offer discovery filters, pagination, and facets
status: archived
validation_status: passed
verified_at: 2026-10-19
archive_path: specs/changes/archive/2026/CHG-0045-offer-discovery-filters/
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/active/2026/CHG-0045-offer-discovery-filters/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
context_scope: existing
created: 2026-10-19
updated: 2026-10-19
---

<!-- Outcome: consolidated into mission-board (REQ-MB-010 modified, REQ-MB-013
added); verified via unit + integration suites and validate-sdd-change.sh
(coverage >= 90%). Residual: envelope is a breaking change coordinated with
hbk-web-app. context_scope: existing — no snapshot. -->

# Extend offer discovery filters, pagination, and facets

## Intent

Bring `GET /api/v1/missions/offers` to parity with `hbk-requisition-service`:
multi-value filters, text search, safe pagination with an envelope response,
controlled sorting, a facets endpoint for the UI, and a DynamoDB list
implementation that pushes filters down and removes the N+1 `GetItem` calls.

## Business Value

- Members can search and combine categories/modes to find relevant offers.
- The board UI can render stable pages and facet pickers without fetching the
  whole catalog.
- Offer reads stay cheap: one `Query` with `FilterExpression`, no per-item reads.

## User Stories

- **US-MB-OFFER-FILTER**: As a player, I want to filter offers by several
  categories and modes at once, so I can narrow the board to what fits me.
- **US-MB-OFFER-SEARCH**: As a player, I want to search offer titles and
  descriptions, so I can find a mission I remember.
- **US-MB-OFFER-PAGE**: As a client, I want bounded pages with a total count,
  so I can paginate without scanning the entire catalog.
- **US-MB-OFFER-SORT**: As a player, I want to sort by reward or category,
  so I can prioritize what to play.
- **US-MB-OFFER-FACETS**: As the web app, I want the distinct categories and
  modes present, so I can populate filter pickers.

## Why L2

Public contract evolution (envelope response) and non-trivial list logic, but
read-only, no trust boundary, no transactional writes.

## Scope

- **Domain**: `OfferListQuery`/`OfferListFilter` multi-value category/mode,
  `q`, `limit`/`offset`, `sort` whitelist, `Match`, `Page`, `SortOffers`,
  `Facets` value object.
- **Application**: `ListOffersHandler.Handle` returns `OfferListPage`
  (offers + total + limit + offset); new `ListOfferFacetsHandler`.
- **Infrastructure**: `OfferStore.List` pushdown via `FilterExpression`,
  direct item mapping (no `GetItem`), `ListFacets`.
- **Boundary**: query parsing, envelope response, `GET /offers/facets` route.
- **Docs**: `docs/API-CONTRACT.md` updated.

## Breaking Change

The list response changes from a bare array to
`{offers, total, limit, offset}` on `/api/v1` without a version bump;
coordinated with `hbk-web-app` per Gate 1 approval.

## Non-Goals

- No relevance ranking for `q` (substring match only).
- No pagination cursor/continuation tokens.
- No filter on difficulty/type/subtype beyond what exists today.

## Approvals

- Gate 1: approved 2026-10-19 — modify `mission-board` only; envelope is a
  breaking change accepted without versioning; facets inherit the resolved
  season + status filter (no `q`, no pagination); stable secondary sort by
  `OfferID`.
