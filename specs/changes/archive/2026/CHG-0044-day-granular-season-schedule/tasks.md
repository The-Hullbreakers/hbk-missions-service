---
change_id: CHG-0044
artifact: tasks
status: implemented
updated: 2026-10-18
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-044-001 | Day-granular `SeasonSchedule` + week alias | domain-rule | done | domain |
| T-044-002 | `CreateScheduledSeasonHandler` accepts days | application-command | done | application |
| T-044-003 | `schedule` request: days + mixed-unit rejection | rest-endpoint | done | interface |
| T-044-004 | DynamoDB `ActiveDays`/`OffDays` + legacy weeks | persistence-adapter | done | infrastructure |

## Requirement to Test to Production Map

- **REQ-SCC-009** → domain schedule + HTTP create tests → `domain/`, `interface/http/`.
- **REQ-SCC-010** → scheduler duration tests → `domain/`, `application/`.

## Gates

Standard per risk level.
