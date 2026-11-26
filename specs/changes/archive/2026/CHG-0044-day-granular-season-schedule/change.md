---
change_id: CHG-0044
title: Day-granular season schedule
status: completed
risk: L2
canonical_action: modify
change_action: create_delta
canonical_path: specs/capabilities/seasonal-category-competition/spec.md
primary_workflow: go-sdd-implement-change
supporting_workflows: []
updated: 2026-10-18
---

# CHG-0044 — Day-granular season schedule

## Intent

Allow season schedules to be expressed in days (`active_days`, `off_days`) so an
operator can create e.g. a 30-day season with a 3-day off-season. The existing
week fields remain as accepted aliases (weeks ×7) so stored seasons and current
clients keep working.

## Scope

- Domain: `SeasonSchedule` day fields + duration computation.
- Application: `CreateScheduledSeasonCommand` day parameters.
- Interface: `schedule` request accepts days; mixed units rejected.
- Infrastructure: `ActiveDays`/`OffDays` attributes; legacy week items deserialize.

## Gates

- Gate 1: approved 2026-10-18 — D1 `SeasonSchedule{ActiveDays,OffDays}` with
  `active_weeks`/`off_weeks` as ×7 aliases; D2 durations computed from days;
  D3 mixed units in one request → `422 INVALID_SEASON`; D4 persistence stores
  days and deserializes legacy weeks without migration.

## Verification

- Unit + integration + composition + e2e suites: pass (2026-10-18).
- `tools/validate-sdd-change.sh`: pass. Coverage: domain 90.4%, application
  91.5%, dynamodb 93.0%.
- RED evidence: day constructors undefined → `TestNewSeasonScheduleDays`;
  `ActiveDays` command fields undefined; `active_days` request → 422;
  day season round-trip → `nil` schedule. All GREEN after implementation.
- Iteration `0` allowed only on the day path (`season-0`); the legacy week
  path keeps rejecting `iteration <= 0` (existing `RejectsZeroIteration`).
