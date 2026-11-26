---
change_id: CHG-0026
title: Add mission content and completion reward
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
archive_path: specs/changes/archive/2026/CHG-0026-add-mission-content-and-completion-reward/
verified_at: 2026-10-09
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
context_scope: existing
created: 2026-10-09
updated: 2026-10-09
---

# Change — Add Mission Content And Completion Reward

## Intent

Mission definitions are operator-authored content that players read on the
board. Today they carry only machine fields (`definition_code`, `category`,
`capacity`, `mode`, scoring rules). This change makes a definition a real
player-facing mission — Star Citizen style — with a required title and
description, named objectives, and an explicit competitive-points reward
credited when a contract reaches `COMPLETED`.

## Business Value

- Players read a mission title and description instead of guessing from a
  template code.
- Operators design missions end to end in one request: content, objectives
  with readable names, and the completion reward.
- The seasonal leaderboard gains a predictable completion bonus on top of
  per-objective merit, making mission completion meaningful competitively.

## User Stories

- **US-0026-01 — Author readable missions:** As an operator, I create a
  definition with `title`, `description`, `category`, `capacity`, `mode`,
  `completion_reward`, and named `objectives` so the board shows real mission
  content.
- **US-0026-02 — Read mission content on the board:** As a member, I see the
  title, description, and reward of each offer so I can choose missions like
  in Star Citizen.
- **US-0026-03 — Earn the completion reward:** As a member, when my contract
  reaches `COMPLETED`, my season standing receives the frozen
  `completion_reward` once, in addition to per-objective merit.

## Risk Classification

**L3** — touches public contract payloads, domain value objects, persistence
schema, and scoring behavior (a new credit kind). Mutation gate and
critical-e2e gate apply.

## Scope

- `MissionDefinition` gains `Title`, `Description` (both required),
  `CompletionReward` (`>= 0`).
- `Objective` gains `Name` (required), `Description` (optional).
- Completing a contract emits a `MeritCredit` with `ObjectiveID: "COMPLETION"`
  for the frozen reward, atomically with the contract state transition.
- HTTP: create-definition and publish-offer payloads and all responses
  propagate the new fields; DynamoDB stores them; legacy items are not a
  concern (no production data yet).

## Non-Goals

- No i18n or catalog indirection; `title`/`description` are free text chosen
  by the operator.
- No changes to objective scoring formulas (`merit_per_unit`, `bonus`,
  `unit_limit` semantics preserved).
- No split of the completion reward across roster members; the credit targets
  the same recipient/track rule as objective credits.

## Decisions

- **Reward model:** general `completion_reward` per definition plus
  per-objective merit/bonus — both coexist.
- **Title/description required:** `422` on empty; no legacy fallback (no live
  items exist).
- **Objective text:** free `name`/`description` per objective (option A),
  stored on the API.
- **Versioning:** publishing a new version may change `Title`/`Description`;
  `Objectives` and `CompletionReward` are frozen per version so active
  contracts never change.

## Sequencing

Independent of CHG-0024 (tenant removal); touches overlapping files
(`create_mission_definition`, offer/definition stores). Either order is valid;
rebase the later change on the earlier merge.

## Approvals

- Gate 1: approved 2026-10-09.
- Gate 2: approved 2026-10-09.
- Verification: `verification.md` — all suites green, `tools/validate-sdd-change.sh` passed.
