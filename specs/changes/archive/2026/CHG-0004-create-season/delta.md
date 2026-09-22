---
change_id: CHG-0004
artifact: canonical-delta
status: proposed
updated: 2026-08-30
---

# Canonical Delta

## MODIFIED — Seasonal Category Competition

### REQ-SCC-005 — Create a season with an official-statistics flag

An authorized operator can create a season and decide whether it counts for official standings. The default is official. The flag is persisted and respected by future standings and achievement logic.

#### SCN-SCC-CREATE-OFFICIAL — Create an official season

Given no season `season-2026-q4` exists, when an operator creates it with `official_statistics: true`, then the season is persisted as `ACTIVE` and `Official: true` and returned with `201 Created`.

#### SCN-SCC-CREATE-UNOFFICIAL — Create a practice season

Given no season `practice-2026-q4` exists, when an operator creates it with `official_statistics: false`, then the season is persisted as `ACTIVE` and `Official: false` and returned with `201 Created`.

#### SCN-SCC-CREATE-DEFAULT — Create a season with default official flag

Given no season `season-default` exists, when an operator creates it without `official_statistics`, then the season is persisted as `ACTIVE` and `Official: true` and returned with `201 Created`.

#### SCN-SCC-CREATE-REJECT-EMPTY — Reject an empty season id

Given the operator sends an empty `season_id`, when the create request is processed, then it returns `INVALID_SEASON` with `422 Unprocessable Entity` and no season is persisted.

### MODIFIED REQ-SCC-004 — Freeze and finalize

When a season moves from `ACTIVE` to `FREEZING`, new merits stop except for eligible work started in time. After disputes resolve, a `FINALIZED` immutable standing is published. Only `Official` seasons produce official standings; unofficial seasons may still be finalized for display but do not generate permanent achievements.

#### SCN-SCC-OFFICIAL-FREEZE — Official seasons produce standings

Given an `Official` season is `FREEZING`, when all eligible credits are recorded, then the final standing counts toward official history.

#### SCN-SCC-UNOFFICIAL-FREEZE — Unofficial seasons do not produce official standings

Given an `Official: false` season is `FREEZING`, when all eligible credits are recorded, then the final standing is available for display but is not sent to Achievements as an official reward.

## Invariants

- A finalized season is never reopened or silently rewritten.
- A new season creates a fresh marker; it does not erase the previous one.
- The `Official` flag is set at creation and never changes.
