---
change_id: CHG-0023
artifact: tasks
status: complete
updated: 2026-10-13
---

# Tasks: CHG-0023 — Team Participation and Season Integrity

Execution: sequential after CHG-0022 consolidates.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-023-001 | Mode legality + `Roster` VO + member slot consumption | domain-rule | done | domain |
| T-023-002 | Season create-exists invariant + scheduler retry marker | domain-rule | done | domain |
| T-023-003 | Team acceptance path with `TeamMembershipReader` port | application-command | done | application |
| T-023-004 | `CreateSeasonHandler` rejects existing id | application-command | done | application |
| T-023-005 | `SeasonScheduler` retries next-season creation | application-command | done | application |
| T-023-006 | Conditional season `Put`; `CompletedMissions` counts only `COMPLETED` | persistence-adapter | done | infrastructure |
| T-023-007 | Stable `422` codes on HTTP boundary | rest-endpoint | done | interface |
| T-023-008 | E2E + gates; converge artifacts | boundary-integration-test | done | boundary |

## Requirement → Test → Production Map

- **REQ-MCF-001** → `team_participation_test.go`, `accept_mission_team_test.go`,
  `accept_team_violation_test.go` → `domain/team_participation.go`,
  `domain/roster.go`, `application/accept_mission.go`,
  `application/accept_mission_team.go`.
- **REQ-SCC-005** → `season_integrity_test.go`, `create_season_exists_test.go`,
  `season_store_create_test.go`, `season_integrity_e2e_test.go` →
  `domain/season.go`, `application/create_season.go`,
  `application/create_scheduled_season.go`, `infrastructure/dynamodb/season_store.go`.
- **REQ-SCC-008** → `merit_ledger_completed_test.go` →
  `infrastructure/dynamodb/merit_ledger_readers.go`,
  `infrastructure/dynamodb/merit_ledger_contracts.go`.
- **REQ-SCC-010** → `season_integrity_test.go`, `season_scheduler_retry_test.go` →
  `domain/season_successor.go`, `application/season_scheduler.go`.

## Deviations

- `SeasonCreator` port (`Create`) split from `SeasonWriter` (`Save`): creation
  is conditional `attribute_not_exists(PK)`; lifecycle transitions keep the
  unconditional upsert. Doubles mimic the same contract.
- `TeamMembershipReader` wired to a deny-all adapter in composition until
  Party Manager integrates — declared teammates are rejected rather than
  trusted; recorded as a security decision.
- `ValidateCanCreateSeason` gained the new-season-id parameter; callers pass
  the derived/requested id.
- `merit_ledger_readers.go` resolves contract states through a new
  `merit_ledger_contracts.go` helper (files kept `< 150` lines);
  `accept_mission.go` team helpers extracted to `accept_mission_team.go`;
  `season_schedule.go` successor logic moved to `season_successor.go`.

## Gates

Same L3 set. First RED: `create_season` with existing finalized id overwrites
it (reproduction).
