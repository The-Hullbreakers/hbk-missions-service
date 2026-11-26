---
change_id: CHG-0031
artifact: tasks
updated: 2026-10-09
---

# Tasks: CHG-0031 — Mission Completion Moderation

Execution: sequential after CHG-0023 consolidates.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-031-001 | Stage credits on confirm; `IsActive` frees `PENDING_VALIDATION` | domain-rule | done | domain |
| T-031-002 | `Submit`/`Approve`/`Reject` transitions; drop `Complete` | domain-rule | done | domain |
| T-031-003 | `SubmitContractHandler`, `ApproveContractHandler`, `RejectContractHandler`; confirm handler writes staged credit | application-command | done | application |
| T-031-004 | `pending_credits` persistence; `ContractCreditsWriter` transactional approve | persistence-adapter | done | infrastructure |
| T-031-005 | `submit`/`approve`/`reject` routes, `complete` removal, error codes | rest-endpoint | done | interface |
| T-031-006 | Composition wiring + route table | composition-wiring | done | composition |
| T-031-007 | Critical E2E (approve + reject journeys); gates; converge | boundary-integration-test | done | boundary |

## Requirement → Test → Production Map

- **REQ-MCF-019** → submit domain/app/http tests →
  `domain/mission_contract_lifecycle.go`, `application/submit_contract.go`,
  `interface/http/contract_lifecycle.go`.
- **REQ-MCF-020** → approve/reject domain/app/http/infra tests →
  `domain/mission_contract_lifecycle.go`, `application/approve_contract.go`,
  `application/reject_contract.go`,
  `infrastructure/dynamodb/*` (transactional writer).
- **REQ-MCF-003** → confirm-stages-credit tests →
  `domain/mission_contract_confirm.go`, `application/confirm_objective.go`.
- **REQ-MCF-012** → approve-writes-completion tests →
  `domain/mission_contract_lifecycle.go`.
- **REQ-MB-003/008** → `IsActive`/board-filter tests →
  `domain/contract.go`, `tests/unit/missions/domain/active_limit_test.go`.

## Gates

L3 set: scoped RED → Gate 3 per task, mutation gate (domain adjudication
logic), security gate (reviewer trust boundary), critical E2E, coverage
≥90%, `tools/validate-sdd-change.sh`, documentation update, clean-up pass.
First RED: domain test — `Submit` on `IN_PROGRESS` yields
`PENDING_VALIDATION` and frees the slot.
