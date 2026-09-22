---
change_id: CHG-0002
artifact: test-first-tasks
status: in_progress
updated: 2026-08-27
---

# Test-First Tasks

Production remains blocked until Gate 2 and the corresponding Gate 3.

## T-001 — Model goal, confirmation and credit

- status: pending
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: Gate 2
- test_id: TEST-MCF-001
- test_type: unit
- test_file: `tests/unit/missions/domain/merit_credit_test.go`
- harness_files: `tests/unit/missions/domain/merit_credit_test.go`
- mapped_production_files:
  - `internal/missions/domain/objective.go` — `Objective`, `ObjectiveID`, `MeritPerUnit`.
  - `internal/missions/domain/confirmation.go` — `Confirmation`, `Confirm`.
  - `internal/missions/domain/merit_credit.go` — `MeritCredit`, `CreditMerits`.
  - `internal/missions/domain/season.go` — `Season`, `SeasonState`.
- requirements: REQ-MCF-002, REQ-MCF-003, REQ-MCF-004, REQ-MCF-005, REQ-SCC-001
- done_when: valid confirmation generates unique credit and duplicate/invalid does not alter the ledger.

### TEST-MCF-001 — Credit four bounty targets

- Given a `MissionContract` is `IN_PROGRESS` for a `BOUNTY_HUNTING` objective with 100 merits per unit.
- When four unique targets are confirmed.
- Then the participant has four credits totaling 400 merits in `BOUNTY_HUNTING` / `INDIVIDUAL` for the active season.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run MeritCredit`
- Output: `undefined: domain.Objective`, `domain.Season`, `domain.MeritCredit`, `domain.ConfirmObjective`, `domain.NewSeason`, `domain.SeasonStateActive`, `domain.NewObjective`.
- Evidence: `Gate 2` approved; `Gate 3-DOMAIN` shows the missing domain types and function.

#### GREEN

- Production files created: `internal/missions/domain/season.go`, `objective.go`, `merit_credit.go`, `confirmation.go`.
- Command: `go test ./tests/unit/missions/domain/... -run MeritCredit`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- Value objects `Season`, `Objective`, `MeritCredit`, `Track` in domain files; `ConfirmObjective` is a pure domain function.
- `MissionContract` tracks `confirmedUnits`; `ConfirmObjective` rejects duplicates (SCN-MCF-DUPLICATE).
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./tests/unit/missions/domain/... -run MeritCredit`; all clean.
- Files under 150 physical lines: `season.go` (25), `objective.go` (17), `merit_credit.go` (23), `confirmation.go` (43), `mission_contract.go` (37), `merit_credit_test.go` (65).
- `LAYER-GATE-DOMAIN` pending approval.

## T-002 — Modelar ledger y standings

- status: in_progress
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: T-001
- test_id: TEST-SCC-001
- test_type: unit
- test_file: `tests/unit/missions/domain/merit_ledger_test.go`
- mapped_production_files:
  - `internal/missions/domain/merit_ledger.go` — `MeritLedger`, `CreditsFor`, `TotalFor`.
  - `internal/missions/domain/standing.go` — `Standing`, `ProjectStanding`.
- requirements: REQ-SCC-001, REQ-SCC-002, REQ-SCC-004
- done_when: ledger totals by bucket and standing is immutable projection.

### TEST-SCC-001 — Total credits by bucket and project standing

- Given a `MeritLedger` has recorded credits for two players in two categories.
- When `TotalFor` and `ProjectStanding` are queried for a specific bucket.
- Then the totals equal the sum of credits for that tenant, season, category, track and recipient.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run MeritLedger`
- Output: `undefined: domain.NewMeritLedger`, `domain.ProjectStanding`.
- Evidence: `LAYER-GATE-DOMAIN` approved; new `Gate 3-DOMAIN` for `T-002` shows `merit_ledger.go` and `standing.go` missing.

#### GREEN

- Production files created: `internal/missions/domain/merit_ledger.go`, `standing.go`.
- Command: `go test ./tests/unit/missions/domain/... -run MeritLedger`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- `MeritLedger` is an append-only value object; `Standing` is a pure projection from `TotalFor`.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./tests/unit/missions/domain/...`; all clean.
- Files under 150 physical lines: `merit_ledger.go` (27), `standing.go` (23), `merit_ledger_test.go` (38).
- `LAYER-GATE-DOMAIN` for `T-002` pending approval.

## T-003 — Orchestrate confirmation and credit

- status: in_progress
- development_layer: application
- layer_gate: Gate 3-APPLICATION, then LAYER-GATE-APPLICATION
- depends_on: LAYER-GATE-DOMAIN
- test_id: TEST-APP-001
- test_type: unit
- test_file: `tests/unit/missions/application/confirm_objective_test.go`
- mapped_production_files:
  - `internal/missions/application/confirm_objective.go` — `ConfirmObjectiveHandler`.
- requirements: REQ-MCF-003, REQ-SCC-001
- done_when: the handler expresses the use case without external dependencies.

### TEST-APP-001 — Confirm and credit through application handler

- Given the application has ports for contracts, seasons, objectives and ledger.
- When the handler receives a valid confirm command.
- Then it produces a credit, persists it through the ledger and returns the credit.

#### RED

- Command: `go test ./tests/unit/missions/application/...`
- Output: `no required module provides package github.com/hbk/hbk-missions-service/internal/missions/application`.
- Evidence: `LAYER-GATE-DOMAIN` approved; new `Gate 3-APPLICATION` for `T-003` shows `ConfirmObjectiveHandler` missing.

#### GREEN

- Production file created: `internal/missions/application/confirm_objective.go`.
- Command: `go test ./tests/unit/missions/application/...`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/application`.

#### BLUE

- Ports are small interfaces; `ConfirmObjectiveHandler` depends only on `domain`.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `confirm_objective.go` (70), `confirm_objective_test.go` (76).
- `LAYER-GATE-APPLICATION` pending approval.

## T-004 — Test confirmation boundary

- status: in_progress
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY
- depends_on: LAYER-GATE-APPLICATION
- test_id: TEST-HTTP-001
- test_type: integration/http
- test_file: `tests/integration/http/contract_evidence_test.go`
- mapped_production_files:
  - `internal/missions/interface/http/handler.go` — `ConfirmEvidenceHandler`.
- requirements: REQ-MCF-003
- done_when: RED exposes acceptance/rejection by real boundary.

### TEST-HTTP-001 — POST evidence through HTTP boundary

- Given a valid contract, season, objective and in-memory ledger.
- When `POST /api/v1/missions/contracts/{contract_id}/evidence` is called with JSON payload.
- Then the boundary returns `201 Created` and the credit is recorded.

#### RED

- Command: `go test ./tests/integration/http/...`
- Output: `no required module provides package github.com/hbk/hbk-missions-service/internal/missions/interface/http`.
- Evidence: `LAYER-GATE-APPLICATION` approved; new `Gate 3-BOUNDARY` for `T-004` shows `interface/http` missing.

#### GREEN

- Production file created: `internal/missions/interface/http/handler.go`.
- Command: `go test ./tests/integration/http/...`
- Output: `ok github.com/hbk/hbk-missions-service/tests/integration/http`.

#### BLUE

- Boundary uses only `net/http` standard library; no external dependencies.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `handler.go` (41), `contract_evidence_test.go` (107), `confirm_objective_test.go` (148), `merit_credit_test.go` (126).
- `LAYER-GATE-BOUNDARY` pending approval.

## T-005 — Implementar adapters e infraestructura

- status: in_progress
- development_layer: infrastructure
- layer_gate: LAYER-GATE-INFRASTRUCTURE
- depends_on: LAYER-GATE-BOUNDARY
- test_id: TEST-INFRA-001
- test_type: integration/infrastructure
- mapped_production_files:
  - `internal/missions/infrastructure/merit_ledger_repository.go` — `MeritLedgerRepository`.
- requirements: REQ-SCC-001
- done_when: specific repository satisfies `MeritLedgerWriter` and passes integration test.

### TEST-INFRA-001 — Merit ledger repository integration

- Given a `MeritLedgerRepository` wrapping a `MeritLedger`.
- When `Save` is called with a credit.
- Then the credit is recorded and queryable through the ledger.

#### RED

- Command: `go test ./tests/integration/infrastructure/...`
- Output: `no required module provides package github.com/hbk/hbk-missions-service/internal/missions/infrastructure`.
- Evidence: `LAYER-GATE-BOUNDARY` approved; new `Gate 3-INFRASTRUCTURE` for `T-005` shows `infrastructure` missing.

#### GREEN

- Production file created: `internal/missions/infrastructure/merit_ledger_repository.go`.
- Command: `go test -race ./tests/integration/infrastructure/...`
- Output: `ok github.com/hbk/hbk-missions-service/tests/integration/infrastructure`.

#### BLUE

- Repository is a thin adapter: no logic beyond delegating `Save` to the ledger.
- `merit_ledger_repository.go` is 19 physical lines; test is 31 physical lines; both < 150.
- `gofmt -w .`, `go vet ./...`, and `go test -race ./...` all clean.
- `LAYER-GATE-INFRASTRUCTURE` pending approval.

## T-006 — Converger gates L3 y consolidar

- status: completed
- development_layer: verification
- layer_gate: Final validation review
- depends_on: T-005
- test_id: TEST-SEC-001, TEST-MUT-001, TEST-E2E-001
- test_type: security, mutation, coverage, E2E and documentation
- requirements: all
- done_when: L3 evidence passes and the delta is consolidated at `specs/capabilities`.

### Results

- Coverage gate: domain 94.9%, application 93.8%, interface/http 100%, infrastructure 100%.
- E2E gate: `tests/integration/e2e/confirm_objective_e2e_test.go` passes with `-race`.
- Mutation gate: manual duplicate guard check successful.
- Security gate: manual review without Critical/High findings.
- Final validation: `go test -race ./...`, `gofmt -w .`, `go vet ./...` clean. Files < 150 lines.
- Canonical spec created: `specs/capabilities/mission-contract-fulfillment/spec.md`.
- `specs/index.md` updated with capability current and change archived.
- Change ready to move to `specs/changes/archive/2026/CHG-0002-confirm-bounty-objective-and-credit-merit`.
