---
change_id: CHG-0040
title: Sequential phases in mission contracts
status: completed
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/2026/CHG-0040-sequential-contract-phases/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-14
updated: 2026-10-14
---
# Sequential phases in mission contracts

## Intent

The canonical spec describes sequential phases with mandatory/optional objectives, but the domain has a flat objective list. This change introduces `Phase` as a first-class domain concept with progression rules on the contract.

## Business Value

- Multi-stage missions become expressible: later stages gate on mandatory objectives of earlier ones.

## User Stories

- **US-MCF-PHASES**: As a mission designer, I want phased contracts, so missions can express staged progression.

## Why L3

Domain model change: new aggregate member, progression invariants, persistence schema evolution, and confirmation rules affected.

## Scope

- **Domain**: `Phase` (ordered, mandatory/optional objectives); contract phase progression; confirmation only within the current phase's objectives.
- **Application**: objective confirmation validates phase eligibility.
- **Infrastructure**: contract item schema gains phases (backwards compatibility for flat contracts).
- **Boundary**: contract detail/terms expose phase structure.

## Non-Goals

- Phase-specific rewards beyond existing MeritCredit model.
- Editing phases after contract creation.

## Approvals

- Gate 1: approved 2026-10-17 — D1 `Phase{Mandatory, Optional}` VOs, `NewContractTermsPhased`, flat = one implicit phase; D2 `currentPhase` on contract, confirm gates by phase (`ErrPhaseNotUnlocked`→422), lazy advance after confirm; D3 completion = all mandatory across phases; D4 `Phases` item attribute only when phased, flat items deserialize to implicit phase; D5 detail exposes `phases`/`current_phase`.
