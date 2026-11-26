---
change_id: CHG-0022
artifact: tasks
updated: 2026-10-08
---

# Tasks: CHG-0022 — Atomic Acceptance Capacity and Limits

Execution: sequential after CHG-0021 consolidates.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-022-001 | Extend `DecideAcceptance`: total-limit + season-state eligibility | domain-rule | done | domain |
| T-022-002 | `AtomicAcceptanceWriter` port + reason mapping in accept handler | application-command | done | application |
| T-022-003 | Transactional capacity/limit/contract write in `ContractStore` | persistence-adapter | done | infrastructure |
| T-022-004 | Stable rejection codes on the HTTP boundary | rest-endpoint | done | interface |
| T-022-005 | Concurrency E2E + gates; converge artifacts | boundary-integration-test | done | boundary |

## Requirement → Test → Production Map

- **REQ-MB-003/004/014** → domain decision-matrix tests, concurrent-write
  infrastructure tests, http concurrency tests → `domain/acceptance*.go`,
  `application/accept_mission.go`, `infrastructure/dynamodb/contract_store.go`,
  `interface/http/errors.go`.

## Gates

Same L3 set. First RED: concurrent capacity-one acceptance leaves two contracts
(current reproduction).
