---
change_id: CHG-0027
artifact: tasks
status: proposed
updated: 2026-10-09
---

# Tasks — Team Member Leave

## Wave 1 — Domain

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-01 | Add `departed` set + `Leave` + `ActiveMembers`/`IsDeparted`; tighten `CanSubmitEvidence` | domain | — | `go test ./tests/unit/missions/domain/` |
| T-02 | Extend `NewMissionContractFromState` with `departed` param | domain | T-01 | same suite compiles |

RED evidence: `TestMissionContract_Leave*` — member leaves and contract
stays alive; departed cannot confirm/start/complete/leave; last member
leaving abandons; outsider rejected.

## Wave 2 — Application

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-03 | Add `LeaveContractHandler` (ContractReader + ContractWriter) | application | T-01 | `go test ./tests/unit/missions/application/` |

RED evidence: `TestLeaveContractHandler_*` — leave persists; not-found /
outsider / terminal propagate.

## Wave 3 — Boundary + Infrastructure

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-04 | HTTP `POST /contracts/{id}/leave` handler + error mapping | interface | T-03 | `go test ./tests/unit/missions/interface/http/` + `tests/integration/http/` |
| T-05 | Persist `Departed` on contract item + `mapToContract` reload | infrastructure | T-01 | `go test ./tests/integration/infrastructure/dynamodb/` |

## Wave 4 — Composition + E2E

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-06 | Wire `LeaveContractHandler` in `server.go` + route | composition | T-03, T-04 | `go test ./tests/integration/composition/` |
| T-07 | Team journey e2e: member leaves, survivor completes, reward credits | e2e | T-06 | `tests/integration/e2e/` or composition journey |

## Wave 5 — Gates

| ID | Task |
| --- | --- |
| T-08 | BLUE — file sizes < 150 lines, gofmt/vet clean |
| T-09 | Docs — `API-CONTRACT.md`, `API-GUIDE-FRONTEND.md`, `DOMAIN-RULES.md` |
| T-10 | `tools/validate-sdd-change.sh` + fill `verification.md` |
| T-11 | Security review (roster authority) + mutation consideration |
| T-12 | Consolidate canonical specs, index, archive |
