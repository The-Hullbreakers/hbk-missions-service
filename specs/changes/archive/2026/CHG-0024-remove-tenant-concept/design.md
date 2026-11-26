---
change_id: CHG-0024
artifact: design
status: proposed
updated: 2026-10-08
---

# Design — Remove Tenant Concept

## Identity Model

`CallerIdentity` keeps `playerID`, `roles`, `teamIDs`; the `tenantID` field,
its validation, and `TenantID()` are removed. `NewCallerIdentity(playerID,
roles, teamIDs)` requires only a non-empty `playerID`. `NewSystemIdentity()`
takes no tenant. `CallerIdentityFromContext` reads `player_id` (via session
`userID`), `roles`, and `team_ids` only; a missing `tenant_id` claim is no
longer an error. `tenantFromContext` is deleted.

## Domain

- `MeritCredit`, `Standing`, `Season`, `MissionDefinition`, `MissionOffer`,
  `MissionContract` lose tenant fields.
- Contract id format becomes `offer:player:attempt` (tenant segment dropped);
  `ConfirmationID` delimiter scheme is unchanged.
- Lifecycle guards (`Start`, `Complete`, `Abandon`) check roster membership and
  roles only.
- `DecideAcceptance` drops the tenant validation input.

## Application

Every handler signature that takes `CallerIdentity` keeps it (player, roles,
teams still needed) but no tenant value flows to ports. Season handlers treat
existence and `ACTIVE_SEASON_EXISTS` as global checks. The scheduler works over
the single season set.

## Infrastructure (DynamoDB)

- Key builders in `season_store*.go`, `definition_store*.go`, `offer_store.go`,
  `contract_store*.go`, `merit_ledger_*.go` drop the tenant segment from `PK`,
  `SK`, and `GSI1`.
- Store constructors/methods stop accepting a tenant parameter.
- `ListAll` on the season store remains but returns the single scope.

## Interface (HTTP)

- `middleware.go`: no `tenant_id` claim read; `claims` map still stored for
  roles/teams.
- Handlers stop passing tenant to use cases; response DTOs drop `tenant_id`.
- `TENANT_BOUNDARY_DENIED` mapping removed from `errors.go`.

## Composition

`composition/` wiring drops tenant parameters; `cmd/*` unchanged otherwise.

## Data / Rollout

Existing tables may hold tenant-prefixed keys; `rollout.md` records the
deploy-time decision (fresh table or rewritten keys). No runtime migration
code is added (YAGNI for a pre-production single-scope service).

## What Survives Unchanged

- Session-derived `player_id`, `roles`, `team_ids`; body identity fields still
  ignored.
- Roster membership and `admin`/`moderator` authority for evidence and
  lifecycle.
- Atomic acceptance transaction, frozen terms, duplicate-confirmation
  idempotency, single-`ACTIVE` season (now global), scheduler clock port.
