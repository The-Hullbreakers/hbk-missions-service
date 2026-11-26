---
change_id: CHG-0025
artifact: tasks
status: proposed
updated: 2026-10-09
---

# Tasks — Rename Definition Identifier To Definition Code

Layers run inside-out: Domain → Application → Boundary (HTTP/DynamoDB) →
Composition. Each task lists its RED test, the production it unlocks, and the
verification command.

## Wave 1 — Domain

- **T-01 — Rename definition identifier in domain value objects**
  - Test: `tests/unit/missions/domain/mission_definition_test.go` (+ offer/summary usages) — assert `DefinitionCode` field, `NewMissionDefinition` signature, empty-code `ErrDefinitionInvalid`.
  - Production: `internal/missions/domain/{mission_definition,mission_offer,contract,mission_contract,acceptance}.go`.
  - Verify: `go test ./tests/unit/missions/domain/`.

## Wave 2 — Application (waits for domain green)

- **T-02 — Rename definition identifier through application ports and handlers**
  - Tests: `tests/unit/missions/application/{create_mission_definition_test,publish_mission_definition_test,publish_mission_test,accept_mission_terms_test,accept_mission_atomic_test,list_offers_visibility_test,list_definitions_test}.go` — ports use `definitionCode`; visibility/lookup keyed by code.
  - Production: `internal/missions/application/{create_mission_definition,publish_mission_definition,publish_mission,accept_mission_attempt,list_offers}.go`.
  - Verify: `go test ./tests/unit/missions/application/`.

## Wave 3 — Boundary (waits for LAYER-GATE-APPLICATION)

- **T-03 — HTTP contract renames body and path fields**
  - Tests: `tests/unit/missions/interface/http/{create_mission_definition_test,caller_identity_context_test}` + `tests/integration/http/list_definitions_test.go` — request `definition_code`, path `{definition_code}`, response `DefinitionCode`; `422` on empty code.
  - Production: `internal/missions/interface/http/{create_mission_definition,publish_mission,publish_mission_definition}.go`.
  - Verify: `go test ./tests/unit/missions/interface/http/ ./tests/integration/http/`.
- **T-04 — DynamoDB attribute rename with legacy read fallback**
  - Tests: `tests/integration/infrastructure/dynamodb/{definition_store_test,offer_store_test,contract_store_test,contract_store_list_test}.go` — items write `DefinitionCode`; items with only `DefinitionID` still resolve.
  - Production: `internal/missions/infrastructure/dynamodb/{definition_store,definition_store_queries,offer_store,contract_store_mapping,contract_store_confirm}.go`.
  - Verify: `go test ./tests/integration/infrastructure/dynamodb/`.
- **T-05 — Composition and e2e journeys use definition_code**
  - Tests: `tests/integration/composition/{publish_mission_test,solo_journey_e2e_test}.go` — full create → publish → offer → accept journey carries `definition_code`.
  - Production: `internal/missions/composition/server.go`.
  - Verify: `go test ./tests/integration/composition/ ./tests/integration/e2e/`.

## Wave 4 — Gates and convergence

- **T-06 — BLUE cleanup** — mechanical rename verified green; every maintained file `< 150` physical lines.
- **T-07 — Documentation** — update `API-CONTRACT.md`, `API-GUIDE-FRONTEND.md`, `DOMAIN-RULES.md` to `definition_code`.
- **T-08 — Validation** — `tools/validate-sdd-change.sh`; converge `delta.md`, record `verification.md`, consolidate and archive.

## Requirement → Task → Test Map

- `definition_code` contract fields → T-03 → http unit + `tests/integration/http/list_definitions_test.go`.
- `DefinitionCode` domain semantics → T-01 → `mission_definition_test.go`.
- Port renames and lookup-by-code → T-02 → application suite.
- Legacy attribute read → T-04 → dynamodb suite.
- End-to-end code propagation → T-05 → composition e2e.
