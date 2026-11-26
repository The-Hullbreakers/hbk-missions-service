---
change_id: CHG-0023
title: Validate team participation and repair season integrity
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/archive/2026/CHG-0023-team-participation-and-season-integrity/
execution_mode: sequential
depends_on:
  - CHG-0022
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_FIX_BUG_WORKFLOW
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-08
updated: 2026-10-08
---

# Validate Team Participation and Repair Season Integrity

## Intent

Team contracts only form from eligible, validated participation — mode-legal,
real teammates who count toward their own limits — and the season lifecycle can
never be rewritten after finalization nor stall when next-season creation
fails.

## Business Value

- Team standings reflect real teams; a player cannot convert a solo mission
  into team points or inject phantom teammates.
- Finalized seasons and their `Official` flag are immutable history.
- Scheduled seasons always roll over; a transient failure is retried, not
  silently dropped.
- Leaderboard tie-breaks reflect truly completed missions.

## User Stories

- **US-MCF-TEAM**: As a platform owner, I want team participation validated against mode, membership, and each member's active limit, so team rankings and personal limits stay honest.
- **US-SCC-IMMUTABLE**: As a platform owner, I want finalized seasons and their official flag to be immutable, so competitive history can never be rewritten.
- **US-SCC-ROLLOVER**: As an operator, I want a failed next-season creation retried on later scheduler ticks, so a scheduled series never stalls.
- **US-SCC-TIEBREAK**: As a competitor, I want completed-mission tie-breaks to count only finished contracts, so tied rankings resolve fairly.

## Why L3

Team-mode fraud affects competitive adjudication and authorization;
finalized-season immutability is a constitutional invariant; scheduler
correctness affects lifecycle integrity.

## Defects

- **BUG-0023-01 — Any `team_id` converts SOLO to TEAM**: acceptance ignores mode and builds a roster containing only the requester; no party snapshot, no member confirmations, no teammate limit checks (`application/accept_mission.go:69`).
- **BUG-0023-02 — Finalized seasons are overwritten**: creating a season with an existing finalized `season_id` rewrites it as `ACTIVE`, including the immutable `Official` flag (`application/create_season.go:29`).
- **BUG-0023-03 — Scheduler never retries next-season creation**: if creating the next season fails after finalization, later ticks do not retry (`application/season_scheduler.go:61`).
- **BUG-0023-04 — Tie-break counts unfinished missions**: `CompletedMissions` counts any contract with a credit, so an `IN_PROGRESS` contract counts as completed (`infrastructure/dynamodb/merit_ledger_readers.go:49`).

## Scope

- **Domain**: `MissionMode` legality for team contracts (`TEAM`/`FLEX` only);
  roster = requester + validated teammates; per-member active-limit
  consumption.
- **Application**: team acceptance validates teammate membership (via a
  party/roster port — Party Manager stays external) and each member's limit;
  `CreateSeasonHandler` rejects any existing `season_id`; `SeasonScheduler`
  retries next-season creation until it succeeds.
- **Infrastructure**: tie-break projection counts only `COMPLETED` contracts;
  season `Put` guarded by `attribute_not_exists`.
- **Boundary**: rejection codes for team violations.

## Non-Goals

- Full invitation flow (invite/confirm lifecycle beyond acceptance-time
  validation) — minimal contract: teammates listed at acceptance must be valid
  members with capacity; deferred richer flows documented in design.
- Award tiers, placements, and Achievements delivery at finalization (roadmap).
- Cross-tenant team logic (CHG-0020 already enforces tenant isolation).

## Sequencing

Last in the approved set; builds on CHG-0020..CHG-0022.

## Approvals

- Gate 1: approved 2026-10-08.
