---
change_id: CHG-0031
title: Gate contract completion behind reviewer moderation
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/archive/2026/CHG-0031-mission-completion-moderation/
execution_mode: sequential
depends_on:
  - CHG-0023
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-09
updated: 2026-10-09
---

# Change — Mission Completion Moderation

## Intent

A finished contract no longer credits merits on its own: a roster member
submits it for review, and only a `admin`/`moderator` decision turns the
accumulated evidence into seasonal points — or discards it entirely.

## Business Value

- Competitive standings only reflect work a reviewer judged valid; fraud or
  weak evidence never reaches the ledger.
- Rejection needs no ledger reversals: credits are written once, at
  approval, preserving the append-only ledger invariant.
- Partial fulfillment is administrable: the reviewer decides, and only
  confirmed units are credited.
- Players recover their active slot the moment the contract goes under
  review — moderation latency does not block new missions.

## User Stories

- **US-0031-01 — Submit for review:** As a roster member, I want to send my
  finished contract to review so my evidence is adjudicated, whether I
  completed all objectives or only part of them.
- **US-0031-02 — Approve or reject:** As a moderator, I want to approve a
  pending contract so confirmed units and the completion reward credit the
  recipient, or reject it so nothing is credited.
- **US-0031-03 — No unearned points:** As a platform owner, I want a
  rejected or still-pending contract to contribute zero merits to any
  standing so competitive results stay trustworthy.

## Risk Classification

**L3** — introduces a new privileged adjudication boundary
(`admin`/`moderator`-only transitions), changes when competitive value
enters the ledger, and alters the terminal state machine of contracts.
Authorization, adjudication integrity, and an existing constitutional-style
invariant (append-only ledger) are all touched.

## Approach

- Reuse the existing `PENDING_VALIDATION` contract state (already defined
  in `contract.go`, currently unused by transitions) as the review queue
  state.
- `POST .../submit` (roster member, `IN_PROGRESS`) → `PENDING_VALIDATION`;
  the contract stops consuming the member's active slot.
- `POST .../approve` and `POST .../reject` (`admin`/`moderator` only,
  `PENDING_VALIDATION` only): approval writes every confirmed-unit credit
  plus the `completion_reward` (only when all required objectives are
  confirmed) in one atomic transaction → `COMPLETED`; rejection writes
  nothing → `DISQUALIFIED`.
- Evidence confirmation during `IN_PROGRESS` records the unit on the
  contract — the `MeritCredit` is materialized only at approval.
- Discovery consequence: `mission-board` REQ-MB-003/008 currently count
  `PENDING_VALIDATION` toward the active limit. Approved decision 3
  requires excluding it from the slot count while it still hides the
  template (prevents duplicate pending attempts).

## Behavior Delta Summary

- ADDED: `PENDING_VALIDATION` transitions (`submit`, `approve`, `reject`);
  reviewer-only adjudication; deferred credit materialization; new error
  codes.
- MODIFIED: REQ-MCF-003 (confirmation no longer writes a credit),
  REQ-MCF-008 (lifecycle gains the review hop), REQ-MCF-012 (reward moves
  to approval), REQ-MB-003/008 (pending validation frees the slot, still
  hides the template), Non-Goal "no manual reviews" removed.
- REMOVED: `POST .../complete` (replaced by `submit` + reviewer decision).

## Non-Goals

- No review queue listing endpoint — reviewers act on known contract ids;
  a `GET .../pending-validation` listing is a follow-up if moderation
  volume requires it.
- No rejection reason taxonomy or appeal flow — a single terminal
  `DISQUALIFIED` outcome; disputes stay out of scope.
- No partial-approval grading by the reviewer — approval credits exactly
  what evidence confirmed; nothing more.

## Decisions

- **Credits deferred to approval, not reversed.** The ledger stays
  append-only and standings remain a pure projection; rejection is a
  contract write with zero ledger items.
- **`PENDING_VALIDATION` releases the slot but hides the template.** A
  player can take other missions while awaiting review but cannot stack a
  second pending attempt of the same definition.
- **`complete` endpoint removed.** `submit` does not lie about the
  outcome — the contract is not completed until a reviewer says so.
- **Submission allowed at any confirmation level.** Zero-confirmed
  submissions are valid; the reviewer decides. This keeps "parcial"
  administrable without extra states.

## Sequencing

Must run after CHG-0023 consolidates — it modifies the same canonical
specs (`mission-contract-fulfillment` lifecycle, `mission-board` active
limit). No overlap with CHG-0014/0017.

## Approvals

- Gate 1: approved 2026-10-09 (decisions 1-4 confirmed by product owner:
  full/partial moderation, reject = zero points, slot release at pending,
  pending excluded from tie-break count).
