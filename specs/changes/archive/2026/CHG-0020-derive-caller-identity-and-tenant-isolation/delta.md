---
change_id: CHG-0020
artifact: delta
status: proposed
updated: 2026-10-08
---

# Delta — Derive Caller Identity and Isolate Tenant State

## What Is Changing

### Mission Board

- Acceptance derives `tenant_id`, `player_id`, and team participation from the
  authenticated session; body values for identity fields are ignored.
- Definition reads are scoped to the caller tenant at the storage level.

### Mission Contract Fulfillment

- Evidence submission derives identity from the session and is allowed only for
  a caller with authority over the contract: roster member, or `admin`/`moderator`.

### Seasonal Category Competition

- Seasons are stored and transitioned per tenant; the single-`ACTIVE` invariant
  applies within each tenant.

## ADDED

### REQ-MB-011 — Accept an offer only as the authenticated member

`POST /api/v1/missions/offers/{offer_id}/accept` resolves `tenant_id` and
`player_id` from the caller session. Identity fields in the request body are
ignored. A caller can never produce a contract for another tenant or player.

### REQ-MCF-007 — Confirm evidence only within caller authority

`POST /api/v1/missions/contracts/{contract_id}/evidence` resolves tenant and
player from the session. Submission is authorized only when the session player
is in the contract roster or the caller holds `admin`/`moderator`; otherwise
`403` with `CALLER_NOT_AUTHORIZED`.

### REQ-MB-012 — Read catalog data only inside the caller tenant

All definition lookups, including internal lookups performed during acceptance
and confirmation, are scoped to the caller tenant. A definition identifier from
another tenant resolves as not found.

### REQ-SCC-013 — Store and transition seasons per tenant

Season persistence keys include the tenant. The single-`ACTIVE` invariant,
freeze, finalize, and scheduler transitions operate within one tenant's seasons.

## MODIFIED

### REQ-MB-002 — Discover eligible missions

The board and all catalog reads (offers, definitions, active-template
exclusion) are additionally guaranteed to be tenant-scoped at the storage
level, not only by handler filtering.

### REQ-SCC-005 — Create a season with an official-statistics flag

"Season already exists" and the `ACTIVE_SEASON_EXISTS` check are evaluated per
tenant; creating a season cannot read or overwrite another tenant's season.

## BDD Scenarios

### SCN-SEC-ACCEPT-DERIVE — Identity comes from the session

Given an authenticated member `player-1` of `tenant-a`, when they accept an
offer while sending `tenant_id: "tenant-b"` and `player_id: "player-9"` in the
body, then the contract is created under `tenant-a` for `player-1`.

### SCN-SEC-ACCEPT-CROSS-TENANT — No cross-tenant contract

Given the same member, when they attempt to accept an offer owned by
`tenant-b`, then the offer resolves as not found and no contract is created.

### SCN-SEC-EVIDENCE-ROSTER — Roster member confirms

Given `player-1` is in the contract roster, when they submit evidence for that
contract, then the credit is recorded for `player-1` in their tenant's bucket.

### SCN-SEC-EVIDENCE-OUTSIDER — Outsider cannot confirm

Given `player-2` is not in the contract roster and holds no privileged role,
when they submit evidence for the contract, then the response is `403` with
`CALLER_NOT_AUTHORIZED` and no credit is recorded.

### SCN-SEC-DEF-TENANT — Foreign definition is invisible

Given a definition exists in `tenant-b`, when a `tenant-a` member or operator
reads or references it, then it resolves as not found.

### SCN-SEC-SEASON-TENANT — Tenant seasons are independent

Given `tenant-a` and `tenant-b` each operate seasons, when `tenant-b` freezes
or creates a season, then `tenant-a`'s seasons and standings are unchanged.

## Error Vocabulary

- `CALLER_NOT_AUTHORIZED` — `403`
- `TENANT_BOUNDARY_DENIED` — `403` (existing, reused)
- `OFFER_NOT_FOUND` / `DEFINITION_NOT_FOUND` / `SEASON_NOT_FOUND` — `404` for foreign-tenant identifiers (no existence leak)
