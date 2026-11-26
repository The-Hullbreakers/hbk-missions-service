---
change_id: CHG-0053
title: Fix active-template exclusion on member departure and concurrent acceptance
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0053-fix-active-template-exclusion/
archive_path: specs/changes/archive/2026/CHG-0053-fix-active-template-exclusion/
source_revision: efa42234f4242e2680b92b024e72d8f6d24c68ed
source_revision_note: reviewed worktree contained uncommitted user/CHG-0051 changes; no commit claimed
context_scope: none
context_scope_reason: repository holds no specs/context snapshot store
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
created: 2026-10-21
updated: 2026-10-26
---

# Change — Fix Active-Template Exclusion On Departure And Concurrency

## Intent

Restore the guarantees `REQ-MB-008` already states: a member who leaves a
team contract stops being bound to that definition (the template reappears
and the old contract no longer consumes their personal active-mission
slot), and one player can never hold two live contracts of the same stable
definition — not even through concurrent acceptances, a different offer
version of the same definition, or roster membership in two teams.

## Business Value

- Departed members regain fair access to missions instead of being locked
  out by a contract they no longer participate in.
- Competitive integrity holds under concurrency: duplicate same-definition
  contracts cannot be minted by racing the acceptance path.
- Players get a distinct, explainable rejection
  (`MISSION_DEFINITION_ALREADY_ACTIVE`) instead of a misleading limit error.

## Defects

### BUG-0053-01 — Departed member keeps the exclusion and the slot

- Classification: production defect.
- Expected (canonical `REQ-MB-008`): once a member leaves (`REQ-MCF-013`),
  offers of that template reappear on their board and the contract stops
  consuming their active-mission slot.
- Actual: `ListOffersHandler` builds the hidden set from
  `ListByPlayer`, whose roster membership includes departed members kept
  as history — the board never re-shows the template. Independently,
  `ContractStore.Save` releases `PLAYER_ACTIVE` entries only on terminal
  states, so a non-terminal leave leaves the departed member's
  `contract_ids` entry behind — a retake can still fail with
  `ACTIVE_MISSION_LIMIT_REACHED` even though `ListActiveByPlayer` already
  excludes departed members from the decision read.
- Regression anchor: `REG-0053-01`.

### BUG-0053-02 — Same-definition uniqueness is a check-then-act decision

- Classification: production defect (concurrency).
- Expected (canonical `REQ-MB-008` / `REQ-MB-004`): acceptance revalidates
  the exclusion rule to prevent bypass and concurrency; a player never
  holds two active contracts of one stable definition.
- Actual: `DecideAcceptance`/`DecideRosterAcceptance` run before the
  commit; `CommitAcceptance`'s conditional transaction guards only the
  active-set **count** (`size(contract_ids) < limit`). Two concurrent
  acceptances of the same definition — same player, two different offer
  versions, or the same player as roster teammate in two team acceptances
  — all pass the decision and all commit while under the limit.
  Additionally `DecideRosterAcceptance` checks only teammate limits, never
  their active definitions, so even a non-concurrent team accept can seat
  a member who already holds the template.
- Regression anchor: `REG-0053-02`.

### BUG-0053-04 — Stale ordinary `Save` resurrects a departed member

- Classification: production defect (concurrency, write integrity).
- Expected (canonical `REQ-MCF-013`/`REQ-MCF-010`): a committed departure
  is durable roster history; no contract write may silently undo it.
- Actual: `ContractStore.Save` writes a non-terminal contract with an
  unconditional `PutItem`. A snapshot fetched before a leave — an
  in-flight confirmation, invitation, or phase write — overwrites the
  committed departure, restoring the leaver to the active roster while
  their reservation and `PLAYER_ACTIVE` slot are already released.
  Reproduced by `TestContractStore_StaleSaveDoesNotRestoreDepartedMember`
  (`evidence/t13-stale-save-red.txt`, security finding FIND-0053-02).
- Regression anchor: `REG-0053-04`.

### BUG-0053-05 — Stale `SaveApproved` may overwrite departure and mint credits

- Classification: production defect (concurrency, write integrity).
- Expected (canonical `REQ-MCF-010`/`REQ-MCF-020`): adjudication commits
  the contract state and staged credits atomically against the roster
  truth the reviewer adjudicated.
- Root-cause hypothesis: `SaveApproved` puts the contract item
  unconditionally, so a review started on a pre-leave snapshot may
  overwrite the departure and materialize credits for a stale roster.
  The planned boundary RED must establish that this path executes.
- Regression anchor: `REG-0053-05`.

### Pre-existing defect recorded, out of scope

`submit` → `PENDING_VALIDATION` never releases the member's
`PLAYER_ACTIVE` entry (non-terminal `Save` performs a plain put), although
canonical `REQ-MB-003`/`REQ-MCF-019` state the active slot frees at
submit. Distinct defect; recorded for a future change — **not fixed here,
not weakened**.

## User Stories

- **US-0053-01 — Departed member regains the template:** As a player who
  left a team contract, I want that definition's offers to reappear on my
  board and a fresh acceptance to not count my old contract toward my
  personal limit, while my former teammates keep it hidden.
- **US-0053-02 — One live contract per definition per player:** As a
  player, I must never be able to hold two live contracts of the same
  definition — concurrently, via another offer version, or as a roster
  teammate — while unrelated players accept freely; once my contract ends
  or I leave, I can retake; a contract awaiting review still blocks my
  retake regardless of whether my active-mission limit has room.

## Regression Requirements

- **REG-0053-01 — Departure visibility and retake:** the board and the
  acceptance path treat departed members as no longer bound to the
  definition; the departure write frees the member's reservation and
  active-set entry atomically.
- **REG-0053-02 — Same-definition uniqueness under concurrency:** the
  acceptance commit enforces per-player+definition uniqueness
  transactionally, for solo and roster-teammate paths, across offer
  versions.
- **REG-0053-03 — Stale parallel departure conflicts:** a leave committed
  on a stale snapshot cancels atomically instead of resurrecting an
  already-departed member; the persisted departed set must equal the set
  the leaver saw.
- **REG-0053-04 — Stale ordinary save cannot undo a departure:** every
  contract write path is fenced on the exact prior departed set, so an
  in-flight non-departure write conflicts rather than restoring a leaver.
- **REG-0053-05 — Stale approval cannot overwrite roster truth:**
  adjudication requires the persisted departed set to match the
  adjudicated snapshot; a mismatch cancels the transaction atomically —
  no contract overwrite, no materialized credits.

## Risk Classification

**L3** — changes the concurrency/idempotency semantics of the atomic
acceptance transaction and the membership-scoped release of competitive
participation guards. Per constitution and risk matrix: BDD, per-layer
RED/GREEN/BLUE, security review, coverage, mutation, critical E2E,
rollout/rollback, and independent review.

## Approach

- Introduce an independent **per-player+definition reservation** written
  in the same conditional transaction as the contract put and each roster
  member's `PLAYER_ACTIVE` update. The reservation is owned by the
  contract id.
- A reservation conflict maps to `ErrMissionDefinitionAlreadyActive`,
  kept distinct from active-limit conflicts
  (`ErrActiveMissionLimitReached`/`ErrTeammateLimitReached`).
- The reservation survives `PENDING_VALIDATION` (it still hides the
  template). Releases are member-scoped and ownership-safe: a terminal
  save frees only members still active on the contract (a departed
  member's reservation may already belong to a newer attempt), a
  non-terminal leave frees the departing member alone — atomically with
  the lifecycle write — and a release that finds a different owner
  conflicts instead of deleting it.
- The leave write also frees the departed member's `PLAYER_ACTIVE`
  contract id so retake cannot fail at the personal limit.
- The board's exclusion source becomes active membership (non-departed,
  non-terminal); historic `ListByPlayer` roster semantics are preserved
  for contract history.

## Behavior Delta Summary

- ADDED: transactional per-player+definition uniqueness guarantee;
  departure-scoped release of board exclusion, reservation, and
  active-set entry; `MISSION_DEFINITION_ALREADY_ACTIVE` as a commit-time
  outcome.
- MODIFIED: `REQ-MB-008` (exclusion binds to active membership; enforces
  uniqueness atomically, including teammates and versions),
  `REQ-MB-004` (the single conditional transaction also serializes
  same-definition acceptances),
  `REQ-MCF-013` (every departure — including the last member's — commits
  through the focused departure write that releases the member's guards;
  the write is fenced on the prior departed set),
  `REQ-MCF-010`/`REQ-MCF-020` (ordinary saves and approvals are fenced on
  the stored departed set so stale writes conflict atomically).
- REMOVED: none.
- Scenarios: `SCN-MB-EXCLUDE-*`/`SCN-MB-DEPART-*` (mission-board delta)
  plus `SCN-MCF-LEAVE-LAST-RETAKE`, `SCN-MCF-STALE-SAVE`,
  `SCN-MCF-STALE-APPROVAL` (mission-contract-fulfillment delta) —
  full text in `delta.md`.

## Non-Goals

- No fix for the `submit` → `PENDING_VALIDATION` `PLAYER_ACTIVE` release
  defect (recorded above; separate change).
- No change to `PLAYER_ACTIVE` count/limit semantics or roster history;
  `ListByPlayer` stays.
- No migration or backfill execution, and no writes to shared/live data —
  the rollout section specifies the strategy for review only.
- No new public endpoint, error code, or contract field;
  `MISSION_DEFINITION_ALREADY_ACTIVE` already maps publicly to `422`.
- Archived change folders and canonical specs remain untouched until
  consolidation.

## Sequencing

Sequential, one agent. Canonical targets are `mission-board` and
`mission-contract-fulfillment` (Gate 1 expansion approved 2026-10-21).
Coexists with in-flight CHG-0051, which separately modifies
`mission-contract-fulfillment` (adjudication-lock fence on
`SaveApproved` and related adapters): the two changes append conditions
rather than replace each other's work; consolidation order is
independent and no CHG-0051 files are edited by this change.

## Approvals

- Gate 1: approved 2026-10-21 — user approved creating this active change
  and later consolidating `specs/capabilities/mission-board/spec.md`.
  Gate 1 expansion approved 2026-10-21 — canonical ownership extended to
  `specs/capabilities/mission-contract-fulfillment/spec.md` for the
  departure-fence and stale-write requirements.
- Gate 2: approved for the mission-board scope (T-1..T-9 RED/GREEN
  complete). Gate 2 expansion **approved 2026-10-21** — the
  mission-contract-fulfillment mapped scope (`Save`/`SaveApproved`
  fences, last-member departure routing) ran its sequential RED
  sequence (`TEST-APP-LVE-002`, `TEST-INF-REL-007/008`,
  `TEST-E2E-MB-003`, all captured in `evidence/`).
- Gate 3 expansions: **Gate 3-APPLICATION-EXPANSION** approved
  2026-10-21 (leave routing + constructor); **Gate
  3-BOUNDARY-EXPANSION** approved 2026-10-26 (adapter prior-state
  fence, `Save`/`SaveApproved` departed-set fences, composition
  wiring — satisfies the Gate 3-MCF-EXPANSION mapped scope).
- L3 gate convergence 2026-10-26: full isolated-environment validator
  exit 0 incl. race + S3 (`t17-isolated-validator.txt`); critical E2E
  9/9 (`t16-critical-e2e.txt`); mutation 10/10 killed
  (`mutation-results.txt`); coverage gate passed — lead-measured
  3198/3552 (90.03%) on `evidence/coverage-project-combined.out`;
  security approved **for local code only** 2026-10-26 (no shared-env
  rollout). User-authorized local-only cleanup executed: two ABANDONED
  contracts + one empty `PLAYER_ACTIVE` row removed from
  `hbk-missions-local` (109 rows remain); no rollout approved
  elsewhere (`rollout.md`).
- Final verify-spec: **approved 2026-10-26** — delta consolidated into
  `mission-board` (REQ-MB-004, REQ-MB-008) and
  `mission-contract-fulfillment` (REQ-MCF-010, REQ-MCF-013,
  REQ-MCF-020); change archived to
  `specs/changes/archive/2026/CHG-0053-fix-active-template-exclusion/`
  on source revision `efa42234f4242e2680b92b024e72d8f6d24c68ed` plus
  the reviewed dirty worktree (no commit claimed).
- Context snapshot gate: `none` — this repository has no maintained
  `specs/context/` convention; the two canonical capabilities, API docs,
  and archived change preserve the durable contract without creating
  a new AI snapshot merely for this defect.
