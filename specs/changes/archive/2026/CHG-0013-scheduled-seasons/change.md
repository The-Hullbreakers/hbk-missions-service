---
change_id: CHG-0013
title: Schedule seasons with automatic transitions
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/2026/CHG-0013-scheduled-seasons/
archive_path: specs/changes/archive/2026/CHG-0013-scheduled-seasons/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CODE_QUALITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
context_scope: existing
context_snapshot_gate: not_applicable
context_snapshot_reason: The repository does not maintain a local AI context snapshot convention; durable context is provided by the global common rules.
verified_at: 2026-09-02
source_revision: uncommitted-working-tree-2026-09-02
created: 2026-09-02
updated: 2026-09-02
---

# Schedule Seasons With Automatic Transitions

## Intent

Allow an operator to schedule a competitive season by defining its active duration and an optional off-season gap. The system automatically freezes the season when the active window ends, finalizes it after the off-season, and starts the next season in the same series without manual intervention.

## Business Value

- Operators do not need to manually freeze, finalize, and recreate seasons every cycle.
- Predictable competitive windows (e.g. four weeks active, one week off) reduce operational overhead.
- The existing single-active-season rule and in-progress-credit grace period are preserved.
- No `MissionPool` or procedural generation is required.

## User Stories

- **US-SCC-009**: As an operator, I want to create a season with a fixed active duration and optional off-season, so the system closes and replaces it automatically.
- **US-SCC-010**: As an operator, I want the next season in a series to start after the off-season, so I do not have to recreate it manually.
- **US-SCC-011**: As a member, I want in-progress contracts to keep earning merits during the off-season, so a small grace period is preserved.

## Why L3

The change alters the critical season lifecycle (`ACTIVE → FREEZING → FINALIZED → next ACTIVE`), introduces a background scheduler with concurrency concerns, and affects the public `POST /api/v1/seasons` contract and persistence. It therefore requires security review, mutation testing, and critical E2E evidence.

## Scope

- **Domain**: add `StartsAt` and `SeasonSchedule` to `Season`; add `SeasonSeries` value object; implement `ShouldFreeze`, `ShouldFinalize`, and `NextScheduledSeason`; update state-transition invariants.
- **Application**: add `CreateScheduledSeasonHandler`; add `SeasonScheduler` with a `Clock` port and idempotent catch-up logic.
- **Boundary**: extend `POST /api/v1/seasons` to accept `schedule` fields; include schedule in season responses.
- **Infrastructure**: extend `SeasonStore` to persist and load `StartsAt`, `Schedule` (active weeks, off weeks, series), and `List` for scheduler queries.
- **Composition**: wire `CreateScheduledSeasonHandler`, start `SeasonScheduler` as a background goroutine, and add `scheduler_tick_minutes` to `Config`.
- **Tests**: unit tests for domain and scheduler, integration tests for HTTP and composition, mutation and critical E2E gates.

## Non-Goals

- Migrate existing manual seasons to scheduled seasons.
- Add a `PENDING` or `SCHEDULED` pre-active state; the first scheduled season is `ACTIVE` from creation.
- Allow arbitrary cron expressions or overlapping windows.
- Change the active-template exclusion or offer publication semantics.
- Expose a public endpoint to force the scheduler.

## Acceptance Summary

- A season can be created with `schedule.active_weeks`, `schedule.off_weeks`, `schedule.series.prefix`, and `schedule.series.iteration`.
- The season `starts_at` is recorded and `SeasonID` is derived as `{prefix}-{iteration}`.
- At `starts_at + active_weeks` the scheduler freezes the season.
- At `starts_at + active_weeks + off_weeks` the scheduler finalizes it and creates the next season as `ACTIVE` with the same prefix and `iteration + 1`.
- Existing manual seasons without schedule data continue to work and must not be auto-transitioned.
- Existing freeze/finalize endpoints remain as manual overrides.
- `go test -race ./...`, `gofmt`, `go vet`, and `tools/validate-sdd-change.sh` pass.

## Verification Record

- **validation_status**: passed
- **reviewer**: agent
- **verified_at**: 2026-09-02
- **source_revision**: uncommitted-working-tree-2026-09-02
- **canonical_paths**:
  - `specs/capabilities/seasonal-category-competition/spec.md`
- **intended_archive_path**: `specs/changes/archive/2026/CHG-0013-scheduled-seasons/`
- **commands**:
  - `tools/validate-sdd-change.sh`
  - `go test -race ./tests/unit/... ./tests/integration/...`
- **coverage**:
  - domain 95.0%
  - application 91.6%
  - dynamodb infrastructure 91.0%
- **residual_risk**: None. Manual seasons remain unaffected and future scheduler tuning can be adjusted through `SCHEDULER_TICK_MINUTES`.
