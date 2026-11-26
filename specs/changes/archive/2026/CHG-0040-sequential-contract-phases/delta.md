---
change_id: CHG-0040
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Contract Fulfillment

- Contracts contain ordered phases; each phase declares mandatory and optional objectives.
- Objectives confirmable only in the current phase; completing mandatory objectives unlocks the next phase.
- Migration: flat objective lists become a single implicit phase.

## Requirements

### REQ-MCF-002 — Progress by phases and objectives (modified)

A contract is a sequence of phases; mandatory objective completion unlocks the next phase; out-of-phase confirmation is rejected.

## BDD Scenarios

### SCN-MCF-PHASE-UNLOCK — Mandatory completion unlocks next phase

Given phase 1 requires objective `obj-1`, when `obj-1` is confirmed, then phase 2 objectives become eligible.

### SCN-MCF-PHASE-OPTIONAL — Optional objectives do not gate

Given phase 1 has an optional objective unmet, when all mandatory are confirmed, then phase 2 unlocks.

### SCN-MCF-PHASE-ORDER — Out-of-phase confirmation rejected

Given the contract is in phase 1, when a unit for a phase-2 objective is submitted, then it returns 422 PHASE_NOT_UNLOCKED.

## Error Vocabulary

- `PHASE_NOT_UNLOCKED` — 422 (new).
