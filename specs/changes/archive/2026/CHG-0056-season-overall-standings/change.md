---
change_id: CHG-0056-season-overall-standings
change_title: Add season-wide individual standings
status: archived
risk: L3
context_scope: none
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/archive/2026/CHG-0056-season-overall-standings/
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-GO_REST_API_WORKFLOW
execution: sequential
max_parallel_agents: 1
---

# CHG-0056 - Add season-wide individual standings

## Story and decision

US-0056-001 - Compare my season-wide progress: as an authenticated
competitor, I want to see the ranking across all categories of one
season and my own position, so I know where I stand and what to play next.

Gate 1 approved 2026-11-24 for this canonical owner as part of the
competitive missions hub plan. Gate 2 approved 2026-11-24 for the
planned RED tests. TEST-0056-001 Application RED recorded; Gate
3-APPLICATION and implementation are pending.
"Global" means one season across categories on the INDIVIDUAL track,
never cross-season XP or a mixture of individuals and teams. An unofficial
preview season remains visibly unofficial. The category ranking, own
category standing, season statistics, and history stay available.

## Scope and risk

Add an authenticated, read-only overall standings query with a bounded
leaderboard page and caller-derived personal standing. Aggregate the
existing immutable merit ledger rather than writing a second ranking.
Keep existing API responses and readers compatible; no migrations,
identity writes, official rewards, or unrelated frontend changes.

L3: public rankings, authenticated identity, tie-breaks, pagination,
and a new HTTP/SAM route can silently show plausible incorrect scores.
No Go production changes before Application RED, demanded Domain RED,
and the independently reviewed HTTP RED. Tests use owned local resources,
not the persistent preview tables or production data.

## Consumer inventory and removal rule

The frontend still calls category leaderboard and personal standing
from `MissionsLeaderboardTab`, season statistics for category selection,
and history for finalized seasons. HTTP/integration clients also use
those public routes. None is dead or approved for removal now. Before
cleanup, re-audit all in-repo clients, templates, tests, docs, and known
external consumers. Remove a route and its use case, wiring, adapters,
and tests only after an explicitly approved REMOVED delta and a
compatibility decision; never delete a shared reader still in use.

## Rollback and ownership

Roll back only the additive overall route and frontend consumer;
the existing category API remains usable. The backend `template.yaml`
owns the deployed SAM route. Local SAM route generation is a separate
SDD change in `hbk-shared-infrastructure`; UI/client changes belong to
`hbk-web-app`. No durable AI context snapshot is justified.
