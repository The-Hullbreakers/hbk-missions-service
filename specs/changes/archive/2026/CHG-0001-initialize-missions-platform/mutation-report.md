---
change_id: CHG-0001
artifact: mutation-testing
status: planned
updated: 2026-08-26
workflow: WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
---

# Mutation Testing Plan

There is no execution or score: the production and Gate 2 do not yet exist.

## Critical Mutation Scope

- comparison `active_count < effective_limit`;
- selection of category override versus default;
- states that count as assets;
- exclusion due to stable template identity, even between versions;
- stop rotations during `FREEZING`;
- validation of capacity and repeat/cooldown;
- uniqueness of target/confirmation/credit;
- membership to tenant, season, category and track;
- roster freeze;
- separation of team and individual merits;
- rejection of credits during `FREEZING`/`FINALIZED`;
- order of tiebreakers;
- idempotence of placement awards.

## Gate

- Run mutation testing after GREEN/BLUE and coverage.
- No surviving mutant in critical invariants is left unexplained.
- Do not weaken, delete or attach tests to improve the score.
- Register tool, version, command, scope, score, survivors and decision.
- Every valid survivor opens a task/test before final review.
