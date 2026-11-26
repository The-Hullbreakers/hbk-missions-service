---
change_id: CHG-0050
artifact: tasks
updated: 2026-10-20
---

# Tasks: CHG-0050 — Per-Unit Adjudication And Completion Bonus

Execution: sequential, inside-out (Application RED → Domain → Application
GREEN → Boundary → Infrastructure → Composition → gates).

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-050-001 | Approve handler accepts `[]UnitRef` verdict and forwards to domain | application-command | done | application-red |
| T-050-002 | `MissionContract.Approve(rejected)` drops rejected units, emits approved subset; `AllMandatoryFulfilled` gates the completion credit; `rejectedUnits` recorded | domain-rule | done | domain |
| T-050-003 | Approve application GREEN: verdict plumbing + role guard | application-command | done | application |
| T-050-004 | `POST /approve` parses optional `rejected_units`; map `UNIT_NOT_STAGED`; detail projects `rejected_units` | rest-endpoint | done | interface |
| T-050-005 | Persist `RejectedUnits` attribute + rehydration; rejected entries dropped from pending save | persistence-adapter | done | infrastructure |
| T-050-006 | Composition unchanged; verify wiring compiles | composition-wiring | done | composition |
| T-050-007 | Seeds remodel: per-unit objectives, `CompletionReward`, reduced rules, `Version: "3"`; seed script allows 1-objective + emits racing per-lap units | catalog-seed | done | boundary |
| T-050-008 | Critical E2E (partial-verdict approval journey), mutation gate, coverage, docs, converge | boundary-integration-test | done | boundary |

## Requirement → Test → Production Map

- **REQ-MCF-020** → domain verdict tests + app approve-verdict tests +
  http body/UNIT_NOT_STAGED tests + infra atomic verdict tests →
  `domain/mission_contract_review.go`, `application/review_contract.go`,
  `interface/http/contract_lifecycle.go`, `interface/http/errors.go`,
  `infrastructure/dynamodb/contract_store_review.go`.
- **REQ-MCF-012** → full-fulfillment bonus tests (unit_limit reached vs
  ≥1, rejected unit blocks) → `domain/contract_terms*.go`,
  `domain/mission_contract_review.go`.
- **REQ-MCF-022** → detail projection rejected-units tests (reviewer +
  member scoped) → `interface/http/contract_detail_projection.go`.
- **REQ-MCF-004** → rejected unit released from confirmed set →
  `domain/mission_contract_review.go`.
- **Catalog** → seed unit/integration tests → `seeds/*.json`,
  `scripts/seed_missions.py`.

## Gates

L3 set: scoped RED → Gate 3 per task; mutation gate on
`mission_contract_review.go` + `contract_terms_phases.go` verdict/
fulfillment logic; security gate (reviewer verdict trust boundary);
critical E2E (partial-approval journey); coverage ≥ 90% on touched
production scope; `tools/validate-sdd-change.sh`; documentation
(DOMAIN-RULES.md, API-CONTRACT.md, seeds/README.md); clean-up pass.

## First RED

Application: `ApproveContractHandler.Handle` rejects a verdict naming an
unstaged unit → expected compile/RED failure (`UnitRef` does not exist
yet); the demanded domain cycle then opens.
