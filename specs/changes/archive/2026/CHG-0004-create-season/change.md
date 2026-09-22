---
change_id: CHG-0004
title: Create season with official statistics flag
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
change_path: specs/changes/archive/2026/CHG-0004-create-season/
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
archive_path: specs/changes/archive/2026/CHG-0004-create-season/
verified_at: 2026-08-30
source_revision: HEAD
context_scope: existing
created: 2026-08-30
updated: 2026-08-30
---

# Create Season With Official Statistics Flag

## Intent

Allow an authorized operator to create a competitive season and decide whether it counts toward official standings, so the organization can run practice or off-season events without polluting the historical record.

## Business Value

- Run practice seasons and test events without affecting official leaderboards.
- Preserve a clear distinction between official and unofficial competitive windows.
- Keep the existing season lifecycle intact for all season types.

## User Stories

- **US-SCC-003**: As an operator, I want to create a season marked as official, so that it determines permanent standings and achievements.
- **US-SCC-004**: As an operator, I want to create an unofficial season, to practice or test rules out of season without affecting the history.

## Why L2

New administrative endpoint, modification of the `Season` domain model and persistence with backward compatibility. It does not touch identity, authorization beyond the existing middleware or calculation of standings in v1.

## Scope

- Domain: add `Official bool` to `Season` and constructor `NewSeasonWithOfficial`.
- Application: `CreateSeasonHandler` with `SeasonReader`/`SeasonWriter` ports.
- Infrastructure: `SeasonStore` persists and loads `Official` with default `true` for missing items.
- Interface: HTTP `POST /api/v1/seasons`.
- Composition: wire `CreateSeasonHandler` in `NewServer`.
- Tests: unit application, integration through `tests/integration/composition`.

## Non-Goals

- Change `ConfirmObjective`, ranking or closure to ignore unofficial seasons (will be addressed in future changes).
- Add admin role authorization; existing `AuthMiddleware` is reused.
- Support logical deletion or reopening of completed seasons.

## Acceptance Summary

- A season created with `official_statistics: true` is marked official.
- A season created with `official_statistics: false` is marked unofficial.
- Missing `official_statistics` defaults to `true`.
- Empty `season_id` returns `INVALID_SEASON`.
- Existing seasons without the `Official` field load as `true`.

## Current Outcome

Gate 1 and Gate 2 approved on 2026-08-30. T-001, T-002, and T-003 completed with unit and integration tests green. `go test -race ./...`, `gofmt`, and `go vet` pass. `tools/validate-sdd-change.sh` not present in this repository. The active change is ready for `common-sdd-verify-spec` or user review.
