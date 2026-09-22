---
change_id: CHG-0009
artifact: test-first-tasks
status: active
updated: 2026-08-30
---

# Test-First Tasks

## T-001 — Introduce ConfirmationID value object

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: Gate 2
- test_id: TEST-DOMAIN-VO-001
- test_type: unit
- test_file: `tests/unit/missions/domain/confirmation_id_test.go`
- harness_files: `tests/unit/missions/domain/confirmation_test.go` (characterization)
- mapped_production_files:
  - `internal/missions/domain/confirmation_id.go` — `ConfirmationID`, `NewConfirmationID`, `ParseConfirmationID`, `String`.
  - `internal/missions/domain/confirmation.go` — `ConfirmObjective` delegates to `MissionContract.Confirm`.
  - `internal/missions/domain/mission_contract_confirm.go` — `MissionContract.Confirm` uses `ConfirmationID`.
- requirements: REQ-MCF-003, REQ-MCF-004
- done_when: confirmation identity is encapsulated and duplicate detection still works.
- verification: `go test -race ./tests/unit/missions/domain/...` passes; `gofmt` and `go vet` clean. Files under 150 lines.
- next_step: T-002

### TEST-DOMAIN-VO-001 — Construct a valid confirmation ID

- Given `contract-1`, `objective-1`, `target-1`, `player-1`.
- When `NewConfirmationID` is called.
- Then it returns a stable `String() == "contract-1:objective-1:target-1:player-1"`.

### TEST-DOMAIN-VO-002 — Reject empty parts

- Given any empty part in the four identifiers.
- When `NewConfirmationID` is called.
- Then it returns an error.

### TEST-DOMAIN-VO-003 — Preserve duplicate rejection

- Given two calls with the same identifiers.
- When both are recorded on a contract.
- Then the second returns `ErrDuplicateConfirmation`.

## T-002 — Centralize contract confirmation as entity method

- status: completed
- development_layer: domain
- layer_gate: LAYER-GATE-DOMAIN
- depends_on: T-001
- test_id: TEST-DOMAIN-CONFIRM-001
- test_type: unit
- test_file: `tests/unit/missions/domain/confirm_objective_test.go`
- mapped_production_files:
  - `internal/missions/domain/mission_contract.go` — `MissionContract.Confirm`.
  - `internal/missions/domain/confirmation.go` — `ConfirmObjective` delegates to the entity.
- requirements: REQ-MCF-003, REQ-MCF-004
- done_when: the entity owns the confirmation policy and all existing tests still pass.
- verification: `go test -race ./tests/unit/missions/domain/...` passes.
- next_step: T-003

### TEST-DOMAIN-CONFIRM-001 — MissionContract confirms a valid unit

- Given a started contract, matching objective, active season.
- When `contract.Confirm(...)` is called.
- Then it returns a `MeritCredit` and records the confirmation.

### TEST-DOMAIN-CONFIRM-002 — MissionContract rejects duplicate

- Given a contract with an existing confirmation.
- When the same unit is confirmed again.
- Then it returns `ErrDuplicateConfirmation`.

## T-003 — Replace generic Contract with ActiveContractSummary

- status: completed
- development_layer: domain
- layer_gate: LAYER-GATE-DOMAIN
- depends_on: T-002
- test_id: TEST-DOMAIN-ACCEPT-001
- test_type: unit
- test_file: `tests/unit/missions/domain/active_limit_test.go`
- mapped_production_files:
  - `internal/missions/domain/acceptance.go` — `ActiveContractSummary`, `DecideAcceptance`.
  - `internal/missions/domain/contract.go` — `ActiveContractSummary` value object.
  - `internal/missions/domain/mission_contract.go` — `MissionContract.Summary`.
- requirements: REQ-MB-003
- done_when: acceptance decision uses a named summary value object.
- verification: `go test -race ./tests/unit/missions/domain/...` passes.
- next_step: T-004

## T-004 — Move input validation to domain constructors

- status: completed
- development_layer: domain
- layer_gate: LAYER-GATE-DOMAIN
- depends_on: T-003
- test_id: TEST-DOMAIN-CONSTRUCT-001
- test_type: unit
- test_file: `tests/unit/missions/domain/mission_offer_test.go`, `tests/unit/missions/domain/season_test.go`
- mapped_production_files:
  - `internal/missions/domain/mission_offer.go` — `NewValidMissionOffer` and `MissionMode.IsValid`.
  - `internal/missions/domain/season.go` — `NewActiveSeason` validating constructor.
- requirements: REQ-MB-001, REQ-SCC-005
- done_when: application handlers no longer manually validate offer/season fields.
- verification: `go test -race ./tests/unit/missions/domain/...` passes.
- next_step: T-005

## T-005 — Refactor application handlers and split freeze/finalize

- status: completed
- development_layer: application
- layer_gate: Gate 3-APPLICATION, then LAYER-GATE-APPLICATION
- depends_on: LAYER-GATE-DOMAIN
- test_id: TEST-APP-REFACTOR-001
- test_type: unit
- test_file: `tests/unit/missions/application/*_test.go`
- mapped_production_files:
  - `internal/missions/application/accept_mission.go` — uses `ActiveContractSummary`.
  - `internal/missions/application/publish_mission.go` — delegates validation to domain.
  - `internal/missions/application/create_season.go` — delegates validation to domain.
  - `internal/missions/application/freeze_season.go` — only `FreezeSeasonHandler`.
  - `internal/missions/application/finalize_season.go` — only `FinalizeSeasonHandler`.
- requirements: REQ-MB-003, REQ-SCC-004, REQ-SCC-005
- done_when: handlers are thin, one use case per file, and all unit tests pass.
- verification: `go test -race ./tests/unit/missions/application/...` passes.
- next_step: T-006

## T-006 — Verify full integration and coverage

- status: completed
- development_layer: integration
- layer_gate: LAYER-GATE-APPLICATION, then integration gates
- depends_on: T-005
- test_id: TEST-INTEGRATION-REFACTOR-001
- test_type: integration/http, integration/composition, integration/infrastructure
- test_files: all `tests/integration/**` suites
- mapped_production_files: all application and domain files plus `internal/missions/composition/server.go`
- requirements: cross-cutting
- done_when: `go test -race ./...` passes, `gofmt -w .` and `go vet ./...` clean, coverage >= 90% on affected files.
- verification:
  - `go test -race ./...` passed.
  - `gofmt -w .` and `go vet ./...` clean.
  - Domain unit coverage: 96.5% (`go test -race -coverpkg=./internal/missions/domain -coverprofile=coverage-domain.out ./tests/unit/missions/domain`).
  - Application unit coverage: 96.9% (`go test -race -coverpkg=./internal/missions/application -coverprofile=coverage-app.out ./tests/unit/missions/application`).
- next_step: T-007

## T-007 — Converge documentation

- status: completed
- development_layer: documentation
- layer_gate: documentation review
- depends_on: T-006
- requirements: cross-cutting
- done_when: `docs/ARCHITECTURE.md` and active change spec reflect new value objects and file layout.
- verification: `docs/ARCHITECTURE.md` updated with `Domain Model And Value Objects` section.
- next_step: verify, consolidate, and archive.
