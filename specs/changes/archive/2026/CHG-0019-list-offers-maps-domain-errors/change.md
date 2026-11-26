---
change_id: CHG-0019
title: Map domain errors correctly in list offers endpoint
status: archived
risk: L1
canonical_action: modify
change_action: compact_change
change_path: specs/changes/active/2026/CHG-0019-list-offers-maps-domain-errors/
canonical_paths:
  - specs/capabilities/mission-board/spec.md
archive_path: specs/changes/archive/2026/CHG-0019-list-offers-maps-domain-errors/
verified_at: 2026-10-08
validation_status: passed
source_revision: HEAD
context_scope: none
created: 2026-10-08
updated: 2026-10-08
---

# Map Domain Errors Correctly In List Offers Endpoint

## Defect Analysis

- **BUG-0019-LIST-OFFERS-500** — `GET /api/v1/missions/offers?status=AVAILABLE` returns `500 INTERNAL_ERROR` when the use case rejects with a known domain error.
- **Expected**: domain rejections return `422` with the stable public code (e.g. `NO_ACTIVE_SEASON`), consistent with `publish_mission.go` and the freeze/finalize boundary style.
- **Actual**: `list_offers.go` maps every application error to `500 INTERNAL_ERROR`.
- **Classification**: production defect. `interface/http/list_offers.go` does not distinguish domain rejections from infrastructure failures; `ErrNoActiveSeason`/`ErrActiveSeasonAlreadyExists` from `resolveActiveSeason` surface as 500.
- **Root cause**: missing domain-error mapping at the HTTP boundary; `PublicCode` already knows the codes but is never consulted on the `app.Handle` error path.
- **Impact**: consumers cannot tell a business rejection (no active season) from a platform failure; `NO_ACTIVE_SEASON` is unreachable on this endpoint.
- **Risk L1**: local contract fix; sibling handlers already establish the pattern.

## Compact Canonical Delta

`REQ-MB-010` — clarify that listing offers resolves `season_id` to the single active season and returns `422` with `NO_ACTIVE_SEASON` when no season is `ACTIVE`, and `ACTIVE_SEASON_EXISTS` when the invariant is broken. Infrastructure failures remain `500 INTERNAL_ERROR`.

## Regression

- **REG-0019-NO-ACTIVE-SEASON** — boundary regression: with no `ACTIVE` season, the response is `422` with `NO_ACTIVE_SEASON` instead of `500`.
- **TEST-0019** — `tests/unit/missions/interface/http/list_offers_test.go`: `TestListOffersHandler_ReturnsNoActiveSeason` asserts `422` + `NO_ACTIVE_SEASON` with an empty `fakeActiveSeasons`.
- **Expected RED**: `500 INTERNAL_ERROR`.

## Fix Mapping

- `internal/missions/interface/http/list_offers.go` — map recognized domain errors to `422` + `PublicCode`; keep `500` for unrecognized failures.

## Non-Goals

- `list_contracts.go` and `list_definitions.go` share the same all-500 pattern; tracked as follow-up, not fixed here.
- No new error codes or response envelope.

## Evidence

- RED (pre-fix): `TestListOffersHandler_ReturnsNoActiveSeason` failed with `500` instead of `422`.
- GREEN (post-fix): `go test ./tests/unit/missions/interface/http/ -run TestListOffersHandler -v` — all pass, `422` + `NO_ACTIVE_SEASON`.
- Neighboring suites: full unit and integration suites pass; coverage domain 94.0%, application 93.5%, dynamodb 89.9%.
- `tools/validate-sdd-change.sh`: SDD validation passed.
- Rollback: revert `internal/missions/interface/http/list_offers.go`.
