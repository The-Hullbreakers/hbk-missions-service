---
change_id: CHG-0008
title: Query public leaderboards by season, category and track
status: consolidated
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/CHG-0008-category-leaderboard-query/
archive_path: specs/changes/archive/2026/CHG-0008-category-leaderboard-query/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_UPDATE_DOCUMENTATION_WORKFLOW
context_scope: existing
created: 2026-08-30
updated: 2026-08-30
verified:
  - go test -race ./...
  - gofmt -w .
  - go vet ./...
  - coverage > 90% for affected files
---

# Query Public Leaderboards By Season, Category And Track

## Intent

Allow members and authorized consumers to consult the leaderboard of a season, category and track (`INDIVIDUAL` / `TEAM`), obtaining position, merits and public tie-breaking facts without exposing evidence or private data.

## Business Value

- Close the competitive journey: a contract accumulates merits and the participant can see where they are.
- Maintain a single source of truth: the leaderboard is a projection of the ledger, not an edited table.
- Isolate standings by tenant, season and track to avoid duplication between teams and individuals.

## User Stories

- **US-SCC-003**: As a competitor, I want to see the leaderboard of a category and track, to know my position and that of others in the season.
- **US-SCC-004**: As an operator, I want the leaderboard to not leak evidence or data from other tenants.

## Why L2

The change exposes public competition data, introduces a paginated REST contract, and requires new read access to the ledger (potentially a new GSI in DynamoDB). It does not alter adjudication, credit, or finalization, but an error can filter standings between tenants or degrade performance with incorrect pagination.

## Scope

- Endpoint `GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}`.
- Filters: `track` (mandatory), `limit`, `cursor`, `min_merits`.
- Descending merit order with deterministic tiebreakers v1.
- Stable pagination without numerical offset.
- Strict isolation per tenant; nonexistent season returns `SEASON_NOT_FOUND`.

## Non-Goals

- Edit the leaderboard directly.
- Support arbitrary sorting by fields not calculated in v1.
- Cross-season or cross-tenant global ranking.
- Materialize standings as mutable entities separate from the ledger.

## Acceptance Summary

- The leaderboard always requires `track=individual|team`.
- Invalid `track` values ​​return `INVALID_QUERY_PARAMETER`.
- `limit` and `cursor` control pagination; `min_merits` filters participants with at least that total.
- The response never includes evidence, contracts or private tiebreaker identifiers.
- A tenant cannot read standings from another tenant.
- The projection is calculated from the ledger and respects `Track` (`INDIVIDUAL` vs `TEAM`).
