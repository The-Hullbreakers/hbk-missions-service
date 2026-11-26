---
change_id: CHG-0021
artifact: tasks
updated: 2026-10-08
---

# Tasks: CHG-0021 — Solo Journey, Frozen Terms, Atomic Scoring

Execution: sequential after CHG-0020 consolidates.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-021-001 | `ContractTerms` VO + definition objectives/reward model | domain-rule | done | domain |
| T-021-002 | `MissionContract` state machine: start/complete/abandon/expire | domain-rule | done | domain |
| T-021-003 | Attempt-unique `ContractID` + delimiter-safe `ConfirmationID` | domain-rule | done | domain |
| T-021-004 | Acceptance captures `ContractTerms` and attempt identity | application-command | done | application |
| T-021-005 | Lifecycle use cases (start/complete/abandon) | application-command | done | application |
| T-021-006 | Confirm against snapshot + atomic `ContractAndCreditWriter` port | application-command | done | application |
| T-021-007 | Persist terms, `Bonus`, `UnitLimit`; transactional confirm write | persistence-adapter | done | infrastructure |
| T-021-008 | Lifecycle endpoints + responses | rest-endpoint | done | interface |
| T-021-009 | Composition wiring | composition-root | done | composition |
| T-021-010 | Journey E2E + mutation/coverage gates; converge artifacts | boundary-integration-test | done | boundary |

## Requirement → Test → Production Map

- **REQ-MCF-008** → domain state-machine tests, app lifecycle tests, http lifecycle tests → `domain/mission_contract.go`, `application/start_contract.go`, `application/complete_contract.go`, `application/abandon_contract.go`, `interface/http/*`, `composition/server.go`.
- **REQ-MCF-009 / REQ-MB-013 / REQ-MB-005** → snapshot immutability tests, wrong-objective rejection → `domain/mission_definition.go`, `domain/contract_terms.go`, `application/accept_mission.go`, `application/confirm_objective.go`.
- **REQ-MCF-010** → failure-injection atomicity tests (unit + infrastructure) → `application/confirm_objective.go`, `infrastructure/dynamodb/*`, `domain/merit_ledger.go`.
- **REQ-MCF-011** → attempt-id uniqueness/history tests → `application/accept_mission.go`, `domain/mission_contract.go`.
- **REQ-MCF-004** → unit-limit enforcement + `ConfirmationID` parse tests → `domain/confirmation_id.go`, `infrastructure/dynamodb/objective_store.go`.

## Gates

Same L3 gate set as CHG-0020. First RED: domain — `Start` on `ACCEPTED`
contract yields `IN_PROGRESS` (currently impossible).
