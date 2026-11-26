---
change_id: CHG-0037
artifact: tasks
status: implemented
updated: 2026-10-17
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-037-001 | GetCurrentSeasonHandler + `GET /seasons/current` + tests | application-query | done | application→interface |
| T-037-002 | `ErrStandingNotFound` + GetMyStandingHandler + `GET .../leaderboards/{category}/me` + tests | application-query | done | application→interface |

## Requirement to Test to Production Map

- **REQ-SCC-014/015** → handler + HTTP tests → new `get_current_season.go`, `get_my_standing.go` in application + interface.

## Gates

Standard per risk level.
