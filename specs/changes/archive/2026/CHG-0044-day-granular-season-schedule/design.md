---
change_id: CHG-0044
artifact: design
status: draft
updated: 2026-10-18
---

# Design

## Decisions

- **D1 — VO fields**: `SeasonSchedule` stores `ActiveDays` and `OffDays`.
  `NewSeasonSchedule` gains a days-based constructor (or accepts days);
  `NewSeasonScheduleWeeks`-style alias converts weeks→days for old call sites.
- **D2 — Duration math**: `ActiveDuration`/`CycleDuration` return
  `days * 24h` instead of `weeks * 7 * 24h`.
- **D3 — HTTP contract**: `schedule` accepts `active_days`, `off_days` (preferred)
  or `active_weeks`, `off_weeks` (alias). Both present → `INVALID_SEASON`.
- **D4 — Persistence**: items store `ActiveDays`/`OffDays`. Deserializer reads
  days first, falls back to `ActiveWeeks`/`OffWeeks` ×7 — no migration.

## Open Questions

- None — D4 keeps legacy items readable.
