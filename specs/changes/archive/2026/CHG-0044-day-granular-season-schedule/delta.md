---
change_id: CHG-0044
artifact: delta
status: implemented
updated: 2026-10-18
---

# Delta — seasonal-category-competition

## MODIFIED Requirements

### REQ-SCC-009 — Create a scheduled season with an active and off window

An authorized operator can create a season with a fixed active window, optional
off-season window, and a `SeasonSeries` (`prefix` + `iteration`). Windows are
expressed in days via `active_days`/`off_days`; `active_weeks`/`off_weeks`
remain accepted aliases (×7 days) for backward compatibility. Providing both
units in one request returns `422 INVALID_SEASON`.

#### SCN-SCC-SCHEDULE-DAYS — Create a day-granular scheduled season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/seasons` with
`schedule.active_days: 30`, `schedule.off_days: 3`, and `starts_at`, then the
season is persisted `ACTIVE` with a 30-day active window and a 3-day off-season.

#### SCN-SCC-SCHEDULE-MIXED — Reject mixed units

Given a request carrying both `active_weeks` and `active_days`, then the
response is `422 INVALID_SEASON` and no season is persisted.

### REQ-SCC-010 — Auto-freeze, finalize, and start the next season

Unchanged in behavior; windows are computed from the schedule's day durations
(`active_days`, `active_days + off_days`), and scheduled successors inherit the
day-granular schedule.
