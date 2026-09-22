---
change_id: CHG-0006
artifact: canonical-delta
status: proposed
updated: 2026-08-30
---

# Canonical Delta

## MODIFIED — Seasonal Category Competition

### REQ-SCC-006 — List seasons by lifecycle status

Authorized consumers can list seasons and filter them by the business buckets `active`, `inactive`, and `all`. The default filter is `active` so the mission board can immediately display the current competitive window.

#### SCN-SCC-LIST-ACTIVE-DEFAULT — List active seasons by default

Given seasons `season-2026-q4` is `ACTIVE` and `season-2026-q3` is `FINALIZED`, when a consumer calls `GET /api/v1/seasons` without `status`, then the response contains only `season-2026-q4` with `200 OK`.

#### SCN-SCC-LIST-ACTIVE — List active seasons explicitly

Given the same seasons, when a consumer calls `GET /api/v1/seasons?status=active`, then the response contains only `season-2026-q4`.

#### SCN-SCC-LIST-INACTIVE — List non-active seasons

Given the same seasons, when a consumer calls `GET /api/v1/seasons?status=inactive`, then the response contains only `season-2026-q3`.

#### SCN-SCC-LIST-ALL — List every season

Given the same seasons, when a consumer calls `GET /api/v1/seasons?status=all`, then the response contains both seasons.

#### SCN-SCC-LIST-EMPTY — Return empty when no season matches

Given no seasons exist, when a consumer calls `GET /api/v1/seasons?status=active`, then the response is `200 OK` with an empty array.

#### SCN-SCC-LIST-INVALID-STATUS — Reject unknown status values

Given any seasons, when a consumer calls `GET /api/v1/seasons?status=closed`, then the response is `400 Bad Request` with `INVALID_QUERY_PARAMETER`.

## Related Requirements

- `REQ-SCC-005 — Create a season with an official-statistics flag` provides the data surfaced by this query.

## Invariants

- At most one season has `State: ACTIVE` (unchanged from `REQ-SCC-005`).
- The `Official` flag is set at creation and never changes (unchanged).
