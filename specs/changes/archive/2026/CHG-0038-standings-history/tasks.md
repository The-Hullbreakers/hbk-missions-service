---
change_id: CHG-0038
artifact: tasks
status: implemented
updated: 2026-10-17
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-038-001 | `ListStandingsHistoryHandler` + `StandingBucket`/`StandingsHistory` types + `StandingsBucketReader` port + `MeritLedgerStore.ListBuckets` adapter | application-query + persistence-adapter | done | domain→application→infrastructure |
| T-038-002 | `GET /seasons/{id}/standings/history` handler + route + `SEASON_STATE_INVALID` code | rest-endpoint | done | interface |

## Requirement to Test to Production Map

- **REQ-SCC-016** → history read tests → `application/`, `infrastructure/dynamodb/`, `interface/http/`.

## Gates

Standard per risk level.
