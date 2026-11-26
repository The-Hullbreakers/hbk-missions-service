---
change_id: CHG-0020
artifact: design
updated: 2026-10-08
---

# Design — Derive Caller Identity and Isolate Tenant State

## Canonical Layer Scope

Domain → Application → Boundary(HTTP) → Infrastructure → Interface →
Composition, per `RULE-COMMON_INSIDE_OUT_DEVELOPMENT`. Outer tasks wait for
`LAYER-GATE-APPLICATION`.

## Domain

- New value object `CallerIdentity`: `TenantID`, `PlayerID`, `Roles`,
  `TeamIDs`. Provider-neutral; built at the boundary from session claims.
- Roster-authority rule on `MissionContract`: `CanSubmitEvidence(identity)`
  is true for roster members; privileged roles are evaluated in application.
- No new aggregates; `MissionContract` gains the authority query.

## Application

- `AcceptMissionHandler` signature takes `CallerIdentity` instead of trusting
  `AcceptMissionCommand.TenantID/PlayerID/TeamID`; command keeps offer id and
  option fields only.
- `ConfirmObjectiveHandler` takes `CallerIdentity`; authorizes roster member or
  `admin`/`moderator` before confirmation.
- Season handlers (`create`, `freeze`, `finalize`, scheduler lookups) operate on
  `identity.TenantID`.
- Ports stay one-operation each; readers that resolve definitions/offers gain a
  tenant-scoped lookup signature.

## Boundary — HTTP

- `AuthMiddleware` already parses the OIDC token; it now builds
  `CallerIdentity` and stores it in request context under one accessor.
- Handlers stop decoding `tenant_id`, `player_id`, `team_id` from bodies.
  Body schemas keep the fields optional-ignored for compatibility
  (approved forgiving contract); `docs/API-CONTRACT.md` marks them deprecated.
- Foreign-tenant resource ids return the same `404` as missing resources — no
  existence oracle.

## Infrastructure

- `DefinitionStore`: item `PK`/`SK` include `TenantID`; every read/write is
  tenant-keyed. Fixes BUG-0020-03.
- `SeasonStore`: item `PK`/`SK` include `TenantID`; `List`/`GetActive`/
  single-active checks run within a tenant. Fixes BUG-0020-04.
- Contract and ledger keys already carry tenant in the contract id but reads
  must assert the loaded item's tenant equals the caller tenant (defense in
  depth).
- Key-migration decision recorded in `rollout.md`: single dev/staging table,
  backfill script vs clean redeploy.

## Interface And Composition

- `internal/missions/interface/http/accept_mission.go`, `handler.go`
  (evidence), season handlers: read `CallerIdentity` from context.
- `composition/server.go`: unchanged shape; middleware wiring order verified so
  `CallerIdentity` is always present before handlers.

## Test Strategy

- Domain: `CallerIdentity` invariants, `CanSubmitEvidence` partitions
  (member/outsider/admin).
- Application: hand-written port doubles; accept/confirm with forged identity
  fields; per-tenant season transitions.
- Infrastructure (DynamoDB local container): cross-tenant reads return not
  found; two tenants hold independent `ACTIVE` seasons.
- HTTP integration: forged-body accept and evidence reproduce RED; after GREEN
  return derived identity results; `403` for non-roster evidence.
- Critical E2E: authenticated tenant-A member cannot create/credit anything in
  tenant-B across accept → evidence → leaderboard.

## Risk And Rollback

- **Risk**: clients sending identity in bodies — ignored, no break.
- **Risk**: existing rows under old global keys — migration plan in
  `rollout.md`; reads are additive-safe only after migration.
- **Rollback**: revert deploy; new keys are tenant-scoped so old code would not
  read them — rollback requires restoring from pre-migration backup or
  maintaining dual-write during transition (decision in `rollout.md`).
