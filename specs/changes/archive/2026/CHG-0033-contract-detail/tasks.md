---
change_id: CHG-0033
artifact: tasks
status: complete
updated: 2026-10-13
---

# Tasks — CHG-0033 Contract Detail

Execution mode: sequential, one agent. No domain or infrastructure RED —
the change reuses `ContractReader` and the existing `ContractStore.Get`.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-033-001 | `GetContractHandler` with reviewer guard over `ContractReader` | application-query | done | application |
| T-033-002 | `ContractDetailHandler` + `GET /contracts/{contract_id}` + `contractDetail` DTO | rest-endpoint | done | interface |
| T-033-003 | Composition wiring + docs + converge | composition-wiring | done | composition |

## Requirement → Test → Production Map

- **REQ-MCF-022** → app unit tests (reviewer/member/missing), boundary
  200/403/404 tests →
  `application/get_contract.go`,
  `interface/http/contract_detail.go`,
  `composition/server.go`.

## RED Plan

| Task | Test | Layer | Expected RED |
| --- | --- | --- | --- |
| T-033-001 | `tests/unit/missions/application/get_contract_test.go` | unit | compile: `NewGetContractHandler` undefined |
| T-033-002 | `tests/integration/http/contract_detail_test.go` | boundary | compile: `ContractDetailHandler` undefined |

## Verification

- Unit + integration suites green; `validate-sdd-change.sh` passes.
- Docs updated: `API-CONTRACT.md`, `API-GUIDE-FRONTEND.md`.
