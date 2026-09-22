---
change_id: CHG-0006
title: List seasons by lifecycle status
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
change_path: specs/changes/archive/2026/CHG-0006-list-active-seasons/
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
archive_path: specs/changes/archive/2026/CHG-0006-list-active-seasons/
verified_at: 2026-08-30
source_revision: HEAD
context_scope: existing
created: 2026-08-30
updated: 2026-08-30
---

# List Seasons By Lifecycle Status

## Intent

Allow authorized consumers to list seasons filtered by lifecycle status so the mission board can display the currently active season and, when needed, inspect historical or transitioning seasons.

## Business Value

- The board can render the current active season before assigning missions.
- Operators can review non-active seasons (`FREEZING`, `FINALIZED`, `ARCHIVED`) through the same endpoint.
- Keeps the season surface consistent with `POST /api/v1/seasons`.

## User Story

- **US-SCC-006**: As a mission board, I want to consult the seasons filtered by status, to show the current active season and access the history if necessary.

## Why L2

New public read endpoint and query parameter contract; adds an application query, a new port, and a DynamoDB scan/filter in infrastructure. No identity, security, or external integration changes.

## Scope

- Domain: add `SeasonStatusFilter` value object and matching rule.
- Application: `ListSeasonsHandler` with a `SeasonReader` port.
- Infrastructure: `SeasonStore.List` filtering by `PK = SEASON` and the status filter.
- Interface: HTTP `GET /api/v1/seasons?status=active|inactive|all`.
- Composition: wire the route in `NewServer`.
- Tests: unit (domain, application, interface) and integration (composition, infrastructure).
- Documentation: update `docs/API-CONTRACT.md`.

## Non-Goals

- Pagination, sorting, or search by `official_statistics`.
- Multi-tenant filtering; the current `Season` model is not scoped by tenant.
- Exact state filtering (e.g., `FREEZING` only); `inactive` is a business bucket for all non-active states.
- Changing `POST /api/v1/seasons` behavior or response shape.

## Acceptance Summary

- `GET /api/v1/seasons` without `status` returns only `ACTIVE` seasons (default `active`).
- `?status=active` returns `ACTIVE` seasons.
- `?status=inactive` returns seasons with `State != ACTIVE`.
- `?status=all` returns every season.
- Unknown `status` returns `400` with `INVALID_QUERY_PARAMETER`.
- Empty result returns `200` with an empty array.

## Current Outcome

Gate 1 and Gate 2 approved on 2026-08-30. T-001–T-006 completed with RED/GREEN/BLUE evidence.
`gofmt -w .`, `go vet ./...`, and `go test -race ./...` all pass.
`docs/API-CONTRACT.md` updated to include `GET /api/v1/seasons{?status=active|inactive|all}`.
The active change is ready for `common-sdd-verify-spec` and canonical consolidation.
