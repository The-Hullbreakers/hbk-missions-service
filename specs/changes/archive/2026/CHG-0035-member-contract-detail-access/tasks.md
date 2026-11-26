---
change_id: CHG-0035
artifact: tasks
status: implemented
updated: 2026-10-17
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-035-001 | Member authorization in GetContractHandler | application-query | done | application |
| T-035-002 | Member projection / field minimization (own credits, no media_uri) | application-query | done | interface |
| T-035-003 | Boundary admits roster members (drop `adminOrModerator`) + HTTP tests | rest-endpoint | done | interface |

## Requirement to Test to Production Map

- **REQ-MCF-022** → member/outsider/reader tests → `application/get_contract.go`, `interface/http/contract_detail.go`, `composition/server.go`.

## Gates

Standard per risk level.
