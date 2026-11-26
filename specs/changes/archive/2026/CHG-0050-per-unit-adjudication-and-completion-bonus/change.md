---
change_id: CHG-0050
title: Adjudicate each staged unit and gate the completion bonus on full fulfillment
status: proposed
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/CHG-0050-per-unit-adjudication-and-completion-bonus/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-20
updated: 2026-10-20
---

# Change — Per-Unit Adjudication And Completion Bonus

## Intent

A mission that asks for N verifiable results (kills, derribos, duel wins,
laps) pays per verified unit, not per artificial "package + review" pair.
The reviewer adjudicates each staged unit individually — every unit that
passes review earns its credit. Only when **every** required unit passes
does the contract also earn the frozen `completion_reward`, which players
see as the mission's completion objective. The catalog shrinks each
mission's rules to the few essential, verifiable ones.

## Business Value

- Real progress pays early and fairly: 8 of 10 verified kills earn 8 unit
  credits instead of an all-or-nothing outcome.
- Moderators reject only the invalid evidence; a bad clip no longer
  voids a whole package nor grants unearned merits.
- The completion bonus becomes a real objective the player can aim at —
  displayed in the objective list — instead of an invisible ledger field.
- Fewer, concrete rules are readable on the offer card and enforceable by
  review; today's 14–20 rule lists are neither.

## User Stories

- **US-0050-01 — Earn per verified unit:** As a participant, I want each
  approved kill/lap/duel video to pay its own merits so partial honest
  progress is rewarded.
- **US-0050-02 — Adjudicate per unit:** As a moderator, I want to reject
  individual staged units while approving the rest so one invalid clip
  neither cancels the contract nor gets paid.
- **US-0050-03 — Completion objective and bonus:** As a participant, I
  want the mission to list a completion objective so that when every
  required unit is approved I also earn the mission-complete bonus.
- **US-0050-04 — Concise mission rules:** As a participant, I want a
  short list of essential rules so I immediately know what counts.

## Risk Classification

**L3** — changes adjudication semantics on a privileged boundary
(reviewer decides per staged unit), moves when competitive value enters
the ledger (per-unit materialization instead of package-level), tightens
the completion predicate that guards the bonus credit, and ships a new
reward model across the production mission catalog. Authorization,
adjudication integrity, idempotent atomic writes, and competitive
standings are all in scope.

## Approach

- `POST .../approve` gains an optional `{"rejected_units": [{objective_id,
  unit_id}]}` verdict list. Every staged unit not listed commits its
  credit; listed units are recorded on the contract as rejected and pay
  nothing. Missing body keeps the current approve-all behavior.
- The `COMPLETION` credit fires only when every mandatory objective is
  **fully satisfied** — confirmed (non-rejected) unit count reaches
  `unit_limit` (unlimited objectives require at least one unit). This is
  evaluated after rejected units are dropped.
- Rejected units stay recorded on the contract for audit and disputes;
  the contract detail projects them.
- Catalog remodel (`seeds/`): one countable objective per mission —
  `unit_limit` = required kills/derribos/victorias/laps and
  `merit_per_unit` = 90% of the old package split per unit — plus
  `completion_reward` = the old verification share (~10%), keeping every
  mission's total merits identical. Rule lists are rewritten to ~5
  essential verifiable statements. Catalog version bumps to `3`.
- Companion web-app change `173-per-unit-review-verdicts` renders the
  per-unit verdict controls and the synthetic completion objective row.

## Behavior Delta Summary

- ADDED: per-unit verdicts on `approve` (`rejected_units`), rejected-unit
  record on the contract, `UNIT_NOT_STAGED` error, completion predicate
  requiring full objective satisfaction.
- MODIFIED: REQ-MCF-020 (adjudication is per unit, approval commits the
  approved subset), REQ-MCF-012 (completion reward requires full
  fulfillment, evaluated post-verdict), REQ-MCF-022 (detail exposes
  `rejected_units`).
- REMOVED: none — `reject` (whole-contract `DISQUALIFIED`) stays for
  fraud; package-level approve-all remains the default verdict.

## Non-Goals

- No incremental multi-session adjudication: the reviewer submits one
  verdict per contract; the contract completes in that write.
- No return-for-correction loop: a partially approved contract completes
  with the approved subset; the member may raise a dispute or re-accept
  the mission for a fresh attempt (unchanged retake semantics — see
  Decisions).
- No cross-contract evidence deduplication in domain: reusing a kill
  across attempts stays a moderation rule, as today.
- No achievements/notification side-effects on partial approval.
- Web-app rendering work is tracked by the companion change, not here.

## Decisions

- **Rejects are reviewer input, not staged state.** Units stay "staged"
  until the verdict write; `rejected_units` on approve names what failed.
  A staged unit can only be rejected once, atomically with the credits it
  produces.
- **Unknown rejected units fail the whole call.** A verdict referencing a
  unit that was never staged resolves `422 UNIT_NOT_STAGED` and writes
  nothing — a mistyped verdict must never silently pay or drop the wrong
  units.
- **Completion means full satisfaction.** One confirmed unit per
  objective can no longer unlock the bonus once objectives count real
  units (`unit_limit: 10`). Evaluated after rejects so a rejected unit
  blocks the bonus.
- **Retake semantics unchanged.** `COMPLETED` stays terminal and releases
  the template exclusion; a new attempt freezes fresh terms and its own
  full-satisfaction bar. Bonus farming requires completing a whole new
  contract — the same exposure the system already had.
- **Racing units are laps.** Each of the five laps is one verifiable
  unit; the single tanda video is referenced by every lap unit. A
  rejected lap pays the valid laps without the completion bonus.
- **Totals are conserved.** Per-unit merits plus the new completion
  reward equal today's headline totals (e.g., `pvp-live-fps-10`:
  54 × 10 + 60 = 600).

## Sequencing

Self-contained; no other active change touches
`mission-contract-fulfillment`. Lands independently of
CHG-0043 (catalog deployment tooling) but the seed remodel must re-run
the catalog sync after merge.

## Approvals

- Gate 1: approved 2026-10-20 — product owner approved the plan: per-unit
  adjudication, apply the unit model wherever it fits (PvP kills,
  derribos, duels, racing laps), conserved totals, completion bonus shown
  as a synthetic objective row in the UI.
