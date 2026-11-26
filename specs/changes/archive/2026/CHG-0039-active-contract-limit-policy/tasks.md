---
change_id: CHG-0039
artifact: tasks
status: implemented
updated: 2026-10-17
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-039-001 | Policy reader port + persisted policy doc (`PolicyStore` + `Overrides()` accessor) | persistence-adapter | done | infrastructure |
| T-039-002 | UpdateActiveLimitHandler (admin) + `INVALID_POLICY` validation | application-command | done | application |
| T-039-003 | Acceptance reads policy via `ActiveLimitPolicyReader` per call | application-command | done | application |
| T-039-004 | `PUT` endpoint + composition wiring + tests | rest-endpoint | done | interface |

## Requirement to Test to Production Map

- **REQ-MCF-023** → policy update/read/apply tests → `application/`, `infrastructure/dynamodb/`, `interface/http/`.

## Gates

Standard per risk level.
