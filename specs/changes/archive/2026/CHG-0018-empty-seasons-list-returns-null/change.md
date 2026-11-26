---
change_id: CHG-0018
title: Return empty array instead of null for empty season list
status: archived
risk: L1
canonical_action: modify
change_action: compact_change
change_path: specs/changes/active/2026/CHG-0018-empty-seasons-list-returns-null/
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
archive_path: specs/changes/archive/2026/CHG-0018-empty-seasons-list-returns-null/
verified_at: 2026-10-07
validation_status: passed
source_revision: HEAD
context_scope: none
created: 2026-10-07
updated: 2026-10-07
---

# Return Empty Array Instead Of Null For Empty Season List

## Defect Analysis

- **BUG-0018-EMPTY-LIST-NULL** — `GET /api/v1/seasons?status=active` returns `null` when no season matches the filter.
- **Expected**: `200 OK` with `[]` per `SCN-SCC-LIST-EMPTY` (canonical spec is correct).
- **Actual**: `200 OK` with `null`, breaking consumers that iterate the response.
- **Classification**: production defect. The DynamoDB `SeasonStore` leaves `var seasons []domain.Season` nil when the query returns no items, and the HTTP handler encodes that nil slice, which `encoding/json` serializes as `null`.
- **Root cause**: the wire contract "collection = JSON array" is enforced nowhere; a nil slice propagates from the store through the application handler to the encoder.
- **Impact**: any consumer iterating `data.map(...)`/`for`/`len` on the response crashes or needs null checks.
- **Risk L1**: local contract fix, existing requirement already specifies expected behavior.

## Compact Canonical Delta

No canonical change required: `SCN-SCC-LIST-EMPTY` already states `200 OK` with an empty array. Consolidation records the correction evidence only.

## Regression

- **REG-0018-EMPTY-LIST** — regression partition at the public HTTP boundary: when the reader returns no seasons, the response body must be `[]`.
- **TEST-0018** — `tests/unit/missions/interface/http/list_seasons_test.go`: `TestListSeasonsHandler_ReturnsEmptyArrayWhenNoSeasons` asserts body `[]` with a reader returning nil.
- **Expected RED**: body is `null`.

## Fix Mapping

- `internal/missions/interface/http/list_seasons.go` — normalize nil slice to empty before encoding, guaranteeing the collection contract for any adapter.
- `internal/missions/infrastructure/dynamodb/season_store.go` — initialize results with `make` so adapters do not propagate nil slices.

## Non-Goals

- No change to status filter vocabulary or pagination.
- No new response envelope.

## Evidence

- RED (pre-fix): `TestListSeasonsHandler_ReturnsEmptyArrayWhenNoSeasons` failed with body `"null\n"`.
- GREEN (post-fix): `go test ./tests/unit/missions/interface/http/ -run TestListSeasonsHandler -v` — all pass, body `"[]\n"`.
- Neighboring suites: full unit suite and integration suites pass; coverage domain 94.0%, application 93.5%, dynamodb 89.9%.
- `tools/validate-sdd-change.sh`: SDD validation passed.
- Rollback: revert the two mapped production files.
