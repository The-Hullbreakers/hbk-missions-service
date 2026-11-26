---
change_id: CHG-0041
artifact: tasks
status: implemented
updated: 2026-10-17
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-041-001 | Dispute entity + guards in leave/abandon | domain-rule | done | domain |
| T-041-002 | RaiseDispute/ResolveDispute/ListDisputes handlers | application-command | done | application |
| T-041-003 | Dispute persistence (`Disputes` item attribute) | persistence-adapter | done | infrastructure |
| T-041-004 | Dispute endpoints + error codes | rest-endpoint | done | interface |

## Requirement to Test to Production Map

- **REQ-MCF-024** → dispute lifecycle + guard tests → `domain/`, `application/`, `infrastructure/dynamodb/`, `interface/http/`.

## Gates

Standard per risk level.
