---
change_id: CHG-0024
artifact: tasks
status: complete
updated: 2026-10-08
---

# Tasks — Remove Tenant Concept

Layers run inside-out: Domain → Application → Boundary (HTTP/DynamoDB) →
Composition. Each task lists its RED test, the production it unlocks, and the
verification command.

## Wave 1 — Domain

- **T-01 — Simplify CallerIdentity to player, roles, teams**
  - Test: `tests/unit/missions/domain/caller_identity_test.go` — rewrite: no tenant arg, missing tenant claim irrelevant, `NewSystemIdentity()` parameterless.
  - Production: `internal/missions/domain/caller_identity.go`.
  - Verify: `go test ./tests/unit/missions/domain/ -run CallerIdentity`.
- **T-02 — Drop tenant from season, merit, standing, contract, definition, offer value objects**
  - Test: `tests/unit/missions/domain/*_test.go` tenant fixtures removed (`contract_terms_test.go`, `merit_ledger_test.go`, `season*`, `standing*`).
  - Production: `internal/missions/domain/{season,merit_credit,merit_ledger,standing,mission_contract*,mission_definition,mission_offer,season_schedule}.go`.
  - Verify: `go test ./tests/unit/missions/domain/`.
- **T-03 — Roster/role-only lifecycle guards**
  - Test: lifecycle tests assert `Start/Complete/Abandon` authorize by roster and roles without tenant.
  - Production: `internal/missions/domain/mission_contract_lifecycle.go`.
  - Verify: `go test ./tests/unit/missions/domain/ -run Lifecycle`.

## Wave 2 — Application (waits for domain green)

- **T-04 — Rewrite application handlers without tenant plumbing**
  - Tests: `tests/unit/missions/application/` — remove `season_tenant_scope_test.go`, `season_tenant_scope_doubles_test.go`; rewrite `accept_mission*_test.go`, `confirm_objective_identity_test.go`, `freeze_season_test.go`, `finalize_season_test.go`, `list_offers_test.go`, `list_player_contracts_test.go`, `list_definitions_test.go`, `list_seasons_test.go`, `list_leaderboard_test.go`, `publish_mission*_test.go`, `create_season*_test.go` with tenant-free Mothers.
  - Production: `internal/missions/application/*.go` (~15 handlers + `season_scheduler.go`).
  - Verify: `go test ./tests/unit/missions/application/`.

## Wave 3 — Boundary (waits for LAYER-GATE-APPLICATION)

- **T-05 — HTTP: derive identity without tenant claim**
  - Tests: `tests/unit/missions/interface/http/` + `tests/integration/http/` — `forged_identity_test.go` rewritten for player/role/team only; remove cross-tenant cases.
  - Production: `internal/missions/interface/http/middleware.go`, `errors.go`, handlers.
  - Verify: `go test ./tests/unit/missions/interface/http/ ./tests/integration/http/`.
- **T-06 — DynamoDB keys drop tenant segment**
  - Tests: `tests/integration/infrastructure/dynamodb/` — remove `*_tenant_test.go`; rewrite key assertions in `season_store_test.go`, `definition_store*.go`, `offer_store_list_test.go`, `contract_store_list_test.go`, `merit_ledger*`.
  - Production: `internal/missions/infrastructure/dynamodb/*.go`.
  - Verify: `go test ./tests/integration/infrastructure/dynamodb/`.
- **T-07 — Composition and e2e single-scope**
  - Tests: `tests/integration/composition/` — remove `cross_tenant_e2e_test.go`, `tenant_token_session_test.go`; adjust `server_test.go`, `concurrent_acceptance_e2e_test.go`.
  - Production: `internal/missions/composition/`, `cmd/`.
  - Verify: `go test ./tests/integration/composition/ ./tests/integration/e2e/`.

## Wave 4 — Gates and convergence

- **T-08 — BLUE cleanup** — Fowler passes while green; every maintained file `< 150` physical lines.
- **T-09 — Security gate** — `security-review.md`: single-scope trust boundary, no tenant confusion path, secrets unchanged.
- **T-10 — Coverage gate** — `>= 90%` on affected production scope.
- **T-11 — Mutation gate** — surviving mutants on identity derivation and authority checks justified or killed.
- **T-12 — Critical E2E** — authenticated accept → confirm → leaderboard journey in single scope.
- **T-13 — Documentation** — update `AUTHENTICATION-AND-MULTITENANCY.md` (rename/rework), `API-GUIDE-FRONTEND.md`, `API-CONTRACT.md`.
- **T-14 — Validation** — `tools/validate-sdd-change.sh`; converge `traceability.yaml`, `verification.md`, `rollout.md`.
