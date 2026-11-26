---
change_id: CHG-0040
artifact: tasks
status: implemented
updated: 2026-10-17
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-040-001 | Phase VO + progression invariants on MissionContract | domain-rule | done | domain |
| T-040-002 | Confirmation gated by current phase + phased completion | application-command | done | domain/application |
| T-040-003 | `Phases`/`CurrentPhase` schema + implicit-phase migration | persistence-adapter | done | infrastructure |
| T-040-004 | Detail exposes phases/current_phase; `PHASE_NOT_UNLOCKED` 422 | rest-endpoint | done | interface |

## Requirement to Test to Production Map

- **REQ-MCF-002** → phase progression unit tests, confirmation-gate tests, roundtrip persistence tests → `domain/mission_contract*.go`, `application/confirm_objective.go`, `infrastructure/dynamodb/contract_store*.go`, `interface/http/`.

## Gates

Standard per risk level.
