---
change_id: CHG-0055-local-season-preview-scenarios
change_title: Create local season preview scenarios
status: archived
risk: L3
context_scope: none
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0055-local-season-preview-scenarios/
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
execution: sequential
max_parallel_agents: 1
archive_path: specs/changes/archive/2026/CHG-0055-local-season-preview-scenarios/
verified_at: 2026-11-23
source_revision: efa4223+reviewed-uncommitted-diff
validation_status: passed
reviewer: user-approved-final-review-2026-11-23
---

# CHG-0055 — Create local season preview scenarios

## User story

US-0055-001 — Preview seasonal experiences: As a product reviewer, I want
reproducible local off-season, fresh-start and mid-season data so I can inspect
the existing mission and competition screens without altering live local data.

## Observed starting state (read-only, 2026-11-23)

- Loopback DynamoDB Local `http://localhost:8000`: `hbk-users` has 61 user
  metadata rows, 53 active and 8 inactive; no personal details are recorded here.
- `hbk-missions-local` has `local-training` ACTIVE, 36 offers, 72 definitions,
  and no merit credits. `hbk-missions` is a separate table.
- Existing catalog bootstrap in `scripts/seed_local_missions.py` seeds offers;
  it does not generate seasonal progress or selectable preview datasets.

## Scope and decision

- Add local-only, deterministic scenario loading for three isolated tables:
  `hbk-missions-preview-off-season`, `hbk-missions-preview-fresh-start`,
  `hbk-missions-preview-mid-season`.
- Read only active user identifiers from the local identity table. Never write
  to identity, the original missions tables, cloud endpoints, or user files.
- Preview is the default; explicit `--apply` creates missing preview tables and
  inserts only scenario-owned items. Existing unexpected or divergent data
  fails closed; no automatic deletion, overwrite, or reset.
- Reuse the existing published catalogs for offers and definitions; generated
  contracts/credits must have compatible keys and fields for existing readers.
- Provide a clear switch command for the local service's table environment
  variable, not an automatic edit to the web app or its credentials.
- "Screenshot" means a selectable data snapshot, not an image capture.

## Risk, gates, and rollback

L3 applies because this writes persistence fixtures linked to existing user
identities and must protect against accidental overwrites, wrong endpoints,
and duplicate credits. Security role: `none` for application identity behavior;
the script crosses a local data-access boundary and reads only public
user identity fields needed for recipient matching. No secrets or personal
data in logs, snapshots, fixtures, or Git.

Gate 1 approved by user on 2026-11-23 for these three canonical owners.
Gate 2 and scoped Gate 3 for pure scenario modules approved
2026-11-23; local-resource Gate 3 and scoped BUG-0055-001 regression
Gate 3 approved 2026-11-23; GREEN/BLUE recorded in verification.md.
Final verification and consolidation approved by user 2026-11-23.

Rollback means returning `DYNAMODB_TABLE_NAME` (Lambda) or the local executable's
configured table to its previous value. Preview tables remain intact; removing
tables is a separate explicitly approved destructive operation, not a script
feature. The original tables are never mutated by this change.

Conditional artifacts: L3 governed evidence is required; a separate
`parallel-tracks.md` is not (sequential execution). Routing stays inline:
Go repository workflow governs this Python local tooling and existing
unit/infrastructure suites; no Go product boundary changes. No context
snapshot is justified (`context_scope: none`).

## Current evidence

- Pure in-memory model check (no DB writes): 72 definitions, 36 offers,
  12 contracts, 64 credits across PVP_AIR/PVP_DUEL/PVP_FPS/RACING;
  every credit's amount and lookup keys verified against its contract.
- Unit suite GREEN 10/10 via `python3 -m unittest discover -s
  tests/unit/seeds -p 'test_season_scenarios.py' -v`; Gate 1, Gate 2,
  and Gate 3 pure-tool scope approved 2026-11-23.
- T-0055-004 fixture strategy approved by user: separate persistent
  fresh-start and mid-season preview tables on localhost,
  no cleanup/deletion. See
  `handoffs/context-checkpoints/CHECKPOINT-20261123-001.md`.
- BUG-0055-001 — Reject empty-userinfo preview endpoints /
  REG-0055-001 — Empty userinfo must be rejected: regression RED
  added to TEST-0055-001; `validate_target('http://@localhost:8000',
  'mid-season')` expected None, actual fixed table name; classified
  production security guard defect mapped only to
  `scripts/season_scenario_model.py::validate_target`; no canonical
  intent/risk change; scoped Gate 3 approved; regression GREEN 10/10
  and local integration 2/2.
- All three local preview tables loaded; Python unit 14/14, affected
  Python coverage 91%, five selected mutants killed, Go real-adapter
  and HTTP preview checks passed, repository-native validator passed
  once under explicit container reset authorization; original local
  data unchanged. Final human review approved and canonical
  consolidation completed 2026-11-23.

## Consolidation outcome

The reviewer approved final evidence and the three named canonical owners on 2026-11-23. REQ-0055-001/002 became current local-preview requirements REQ-SCC-019, REQ-MB-018, and REQ-MCF-027. `specs/index.md` points each to this archived change. The entire active folder moved to `specs/changes/archive/2026/CHG-0055-local-season-preview-scenarios/` with no active duplicate. `context_scope: none`: no durable AI snapshot was justified; the canonical capability specs and local preview guide own the reusable behavior. No user data, original table, or preview tables were deleted.
