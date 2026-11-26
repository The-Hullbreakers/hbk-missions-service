---
change_id: CHG-0020
artifact: tasks
updated: 2026-10-08
---

# Tasks: CHG-0020 — Derive Caller Identity and Isolate Tenant State

Execution: sequential, one agent. Order follows domain → application →
infrastructure → interface → composition.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-020-001 | Add `CallerIdentity` VO + `MissionContract.CanSubmitEvidence` | domain-rule | done | domain |
| T-020-002 | Accept mission consumes `CallerIdentity`, ignores body identity | application-command | done | application |
| T-020-003 | Confirm objective authorizes roster/privileged caller | application-command | done | application |
| T-020-004 | Season handlers scope create/freeze/finalize to caller tenant | application-command | done | application |
| T-020-005 | Tenant-key `DefinitionStore` reads/writes | persistence-adapter | done | infrastructure |
| T-020-006 | Tenant-key `SeasonStore` + per-tenant single-active | persistence-adapter | done | infrastructure |
| T-020-007 | Middleware builds `CallerIdentity`; handlers stop decoding identity fields | rest-endpoint | done | interface |
| T-020-008 | Wire composition; verify middleware ordering | composition-root | done | composition |
| T-020-009 | Boundary + critical E2E evidence; converge artifacts | boundary-integration-test | done | boundary |

## Requirement → Test → Production Map

- **REQ-MB-011 / REQ-MCF-007** → tests/unit/missions/domain (identity VO, roster rule), tests/unit/missions/application (forged-identity accept/evidence), tests/integration/http (forged body → derived identity; 403 outsider) → production: `domain/caller_identity.go`, `domain/mission_contract.go`, `application/accept_mission.go`, `application/confirm_objective.go`, `interface/http/accept_mission.go`, `interface/http/handler.go`, `interface/http/middleware.go`.
- **REQ-MB-012** → tests/integration/infrastructure (cross-tenant definition lookup) → production: `infrastructure/dynamodb/definition_store.go`.
- **REQ-SCC-013 / REQ-SCC-005** → tests/integration/infrastructure (two tenants, independent active seasons), tests/unit/missions/application (tenant-scoped create/freeze) → production: `infrastructure/dynamodb/season_store.go`, `application/create_season.go`, `application/freeze_season.go`, `application/finalize_season.go`, `application/season_scheduler.go`.

## Gates

- Gate 2: approved 2026-10-08 (identity from OIDC claims; clean redeploy for
  season/definition keys; preserve CHG-0017 middleware behavior).
- T-020-001 (2026-10-08): RED — build failed, `undefined:
  domain.NewCallerIdentity` / `contract.CanSubmitEvidence undefined`.
  Gate 3-DOMAIN approved 2026-10-08. GREEN — `go test
  ./tests/unit/missions/domain/` ok; BLUE — all unit suites ok, `go vet`
  and `gofmt` clean, `caller_identity.go` 52 lines. LAYER-GATE-DOMAIN: passed
  for the identity scope. Next: T-020-002 application RED.
- T-020-002 (2026-10-08): RED — signature mismatch on `Handle` +
  `undefined: domain.ErrNotTeamMember`. Gate 3-APPLICATION approved.
  GREEN — `go build ./...`, all unit suites ok. BLUE — vet/gofmt clean,
  all touched files <150 lines. Interface compile fix included
  (`middleware.CallerIdentityFromContext`, body identity fields removed).
- T-020-003 (2026-10-08): RED — `Handle` signature mismatch +
  `undefined: domain.ErrCallerNotAuthorized` in
  `confirm_objective_identity_test.go` (roster-member, outsider, foreign
  tenant, privileged-role partitions). Gate 3-APPLICATION approved.
  GREEN — build + all unit suites ok; authorization runs before
  season/objective loads so foreign tenants never reach them;
  `SeasonReader.Get` unchanged (tenant keying lands in T-020-005/006).
  BLUE — vet/gofmt clean, all touched files <150 lines. Evidence endpoint
  now returns 403/FORBIDDEN for unauthorized callers.
- T-020-004 (2026-10-08): RED — `season.TenantID` undefined; `ListActive`,
  `SeasonReader.Get`, `SeasonListReader.List`, `ListAll`, and
  `Handle(identity, …)` signature mismatches in `season_tenant_scope_test.go`.
  Gate 3-APPLICATION approved. GREEN — `Season.TenantID` added;
  `NewSystemIdentity` for internal actors; ports tenant-scoped
  (`SeasonReader.Get`, `ActiveSeasonsReader.ListActive`,
  `SeasonListReader.List`); new `GlobalSeasonListReader.ListAll` reserved for
  the scheduler which now acts per tenant via system identity; create/scheduled
  handlers stamp `TenantID`; freeze/finalize/list/leaderboard consume
  `CallerIdentity`; season store filters/loads by tenant and persists
  `TenantID` (physical keying lands in T-020-006); all interface handlers
  derive tenant from session — `tenant_id` removed from request bodies
  (publish offer, definitions, create season). BLUE — `go vet`/`gofmt` clean,
  all unit suites pass, touched files <150 lines (`season_store_queries.go`,
  `season_tenant_scope_doubles_test.go` split out).
- T-020-005/006 (2026-10-08): RED — behavioral failures against
  DynamoDB-local: `Get`/`GetVersion` return foreign-tenant rows,
  `List` leaks all tenants' definitions, same `DefinitionID`/`SeasonID`
  rows clobber across tenants (PK/SK not tenant-keyed).
  `DefinitionListReader.List(tenantID)` port change compiles.
  `ListAll` already spans tenants (scheduler) — passes.
  Gate 3-INFRASTRUCTURE approved. GREEN — definitions keyed
  `DEFINITION#<tenant>` / `<id>` (+ `#<version>`) with `TenantID` persisted;
  seasons keyed `SEASON#<tenant>` / `<id>`; `ListAll` scans `SEASON#`
  partitions for the scheduler; `NextScheduledSeason` propagates `TenantID`;
  `CreateMissionDefinitionHandler` stamps caller tenant; composition tests
  seed tenant-scoped rows. BLUE — vet/gofmt clean, all unit + integration
  suites pass, touched files <150 lines (`definition_store_queries.go`
  split out).
- T-020-007/008 (2026-10-08): done — `AuthMiddleware.Authenticate` stores
  claims; `CallerIdentityFromContext` rebuilds the domain identity; every
  interface handler consumes it (no body `tenant_id`/`player_id` decode
  remains; `acceptMissionRequest.TeamID` is an option field, not identity).
  Composition verified: all routes behind `auth.Authenticate`; admin routes
  additionally behind `RequireRole`. Boundary evidence:
  `tests/integration/http/forged_identity_test.go` — forged `tenant_id`/
  `player_id` in bodies ignored (contract written under session identity),
  non-roster outsider gets 403/FORBIDDEN.
- T-020-009 (2026-10-08): done — critical E2E
  `tests/integration/composition/cross_tenant_e2e_test.go` (REG-0020-01):
  tenant-B cannot accept tenant-A offers (422), submit evidence on tenant-A
  contracts (403), read tenant-A leaderboard (404), or publish with
  tenant-A definitions (422); tenant-B sees zero seasons; tenant-A contract
  intact. `tools/validate-sdd-change.sh` green: unit + integration + race
  pass; coverage domain 92.2%, application 94.0%, dynamodb infra 92.1%.
- Gates run: coverage ≥90% (domain 92.2, app 94.0, dynamodb 92.1);
  mutation via gremlins `-D HEAD` — 42 mutants, 100% killed after adding
  malformed-body and single-string-claim tests; critical E2E
  (`TestCrossTenantJourney_Isolation`) green; security-review.md recorded
  (role: resource-server); clean-up gate green (all maintained Go files
  <150 lines; test doubles/helpers split); docs converged
  (`API-GUIDE-FRONTEND.md`, `API-CONTRACT.md`, canonical spec evidence
  body). Pending: security-owner Phase 6 verification + reviewer handoff.
- Layer gates: DOMAIN → APPLICATION → BOUNDARY per constitution sequence.
