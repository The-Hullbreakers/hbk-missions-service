---
change_id: CHG-0038
title: Expose finalized standings history
status: completed
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/active/2026/CHG-0038-standings-history/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
context_scope: existing
created: 2026-10-14
updated: 2026-10-14
---
# Expose finalized standings history

## Intent

The contract documents `GET .../standings/history` but nothing persists or serves a final standings snapshot. This change freezes each finalized season's standings and exposes them for read.

## Business Value

- Competitors and ops can audit past season results after finalization.

## User Stories

- **US-SCC-HISTORY**: As a competitor, I want the final standings of a past season, so results remain consultable.

## Why L2

New persistence (snapshot write at finalization or on-demand projection) and a new read surface; no invariant changes to scoring.

## Scope

- **Application**: standings snapshot at finalize, or deterministic recompute from credits (decision in design).
- **Infrastructure**: storage for finalized standings if snapshot approach is chosen.
- **Boundary**: `GET /api/v1/missions/seasons/{season_id}/standings/history`.

## Non-Goals

- Editing history; partial-season snapshots.

## Approvals

- Gate 1: approved 2026-10-17 — D1 recompute from immutable ledger (no snapshot storage); D2 only FINALIZED seasons serve history (`422 SEASON_STATE_INVALID` otherwise); D3 `?category=&track=` optional filters, grouped `{category, track}` buckets, no pagination; D4 `SEASON_NOT_FOUND`→404, `SEASON_STATE_INVALID`→422, bad `track`→400 `INVALID_QUERY_PARAMETER`. Two tasks: T-038-001 application, T-038-002 interface.
