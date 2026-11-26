---
change_id: CHG-0024
artifact: delta
status: consolidated
updated: 2026-10-08
---

# Delta — Remove Tenant Concept

## What Is Changing

### Mission Board

- The board, catalog reads, offer publishing, and acceptance operate in a
  single scope; no tenant filter or tenant key exists.
- `tenant_id` is no longer resolved from the session; only `player_id`, roles,
  and `team_ids` are derived.

### Mission Contract Fulfillment

- Contracts, evidence authority, and lifecycle transitions are roster- and
  role-scoped only; merit credits no longer carry a tenant bucket.

### Seasonal Category Competition

- Seasons are stored and transitioned in one global scope; the single-`ACTIVE`
  invariant is global; leaderboards are keyed by season, category, and track
  only.

## REMOVED

### REQ-MB-012 — Read catalog data only inside the caller tenant

Tenant scoping of definition lookups is removed entirely; a definition
identifier resolves in the single shared catalog or not at all.

### REQ-SCC-013 — Store and transition seasons per tenant

Tenant-keyed season persistence and per-tenant transitions are removed; the
scheduler's cross-tenant enumeration (`ListAll`) collapses to a single-scope
list.

### SCN-SCC-LEADERBOARD-007 — Tenant boundary denied

No tenant boundary exists; `TENANT_BOUNDARY_DENIED` is retired.

### SCN-SEC-ACCEPT-CROSS-TENANT — No cross-tenant contract

Superseded: an offer id resolves in the single catalog or returns
`OFFER_NOT_FOUND`.

### SCN-SEC-SEASON-TENANT — Tenant seasons are independent

Superseded: there is one season scope.

### SCN-SEC-DEF-TENANT — Foreign definition is invisible

Superseded: unknown ids return `DEFINITION_NOT_FOUND`.

## MODIFIED

### REQ-MB-002 — Discover eligible missions

A member queries `AVAILABLE` offers and filters by category, type, subtype,
mode, difficulty, and availability. The result explains why an offer is not
eligible. Tenant isolation no longer applies.

### REQ-MB-003 — Limit active contracts

The active-contract limit is a single platform default with optional category
overrides (no per-tenant configuration).

### REQ-MB-004 — Consistent acceptance

Acceptance validates eligibility, window, repeatability, capacity, and limit in
a single decision; tenant validation is removed.

### REQ-MB-010 — List mission offers

`GET /api/v1/missions/offers` returns offers in the single scope; responses no
longer carry `tenant_id`.

### REQ-MB-011 — Accept an offer only as the authenticated member

`POST /api/v1/missions/offers/{offer_id}/accept` resolves `player_id` (and
`team_ids`) from the caller session; identity fields in the body are ignored.
A caller can never produce a contract for another player. `tenant_id` is no
longer part of identity or persistence.

### REQ-MCF-006 — List player contracts

`GET /api/v1/missions/contracts` returns the contracts where the authenticated
player is in the roster; tenant scoping is removed.

### REQ-MCF-007 — Confirm evidence only within caller authority

Submission is authorized when the session player is in the contract roster or
the caller holds `admin`/`moderator`; the tenant qualifier is removed.
`403 FORBIDDEN` unchanged.

### REQ-MCF-008 — Progress a contract through its lifecycle

Lifecycle transitions are restricted to roster members; the contract-tenant
qualifier is removed.

### REQ-SCC-001 — Record seasonal merits

Each credit belongs to exactly one season, category, track, and recipient.

### REQ-SCC-005 — Create a season with an official-statistics flag

The "season already exists" and `ACTIVE_SEASON_EXISTS` checks are global.

### REQ-SCC-007 — Query public leaderboards by season, category and track

The response never exposes evidence or contracts; tenant-boundary language and
`TENANT_BOUNDARY_DENIED` are removed.

## BDD Scenarios

### SCN-SEC-ACCEPT-DERIVE — Identity comes from the session

Given an authenticated member `player-1`, when they accept an offer while
sending `player_id: "player-9"` and `tenant_id: "tenant-b"` in the body, then
the contract is created for `player-1` and stores no tenant.

### SCN-SEC-EVIDENCE-ROSTER — Roster member confirms

Given `player-1` is in the contract roster, when they submit evidence for that
contract, then the credit is recorded for `player-1` in the season, category,
and track bucket.

### SCN-SEC-EVIDENCE-OUTSIDER — Outsider cannot confirm

Given `player-2` is not in the contract roster and holds no privileged role,
when they submit evidence for the contract, then the response is `403` with
`FORBIDDEN` and no credit is recorded.

### SCN-REM-TENANT-AUTH — Authenticate without a tenant claim

Given an OIDC token carrying `player_id`, `roles`, and `team_ids` but no
`tenant_id`, when the caller hits any authenticated route, then the request
succeeds and `CallerIdentity` is built without a tenant.

### SCN-REM-TENANT-GLOBAL-SEASON — One active season for everyone

Given season `season-2026-q4` is `ACTIVE`, when any authenticated operator
creates another season, then the request is rejected with
`ACTIVE_SEASON_EXISTS` regardless of caller.

## Error Vocabulary

- `TENANT_BOUNDARY_DENIED` — retired.
- `FORBIDDEN` — `403`, unchanged for authority failures.
- `OFFER_NOT_FOUND` / `DEFINITION_NOT_FOUND` / `SEASON_NOT_FOUND` — `404`, single scope.
- `ACTIVE_SEASON_EXISTS` — `422`, now global.
