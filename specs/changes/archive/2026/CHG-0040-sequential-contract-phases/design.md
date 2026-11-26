---
change_id: CHG-0040
artifact: design
updated: 2026-10-14
---

# Design — Sequential Phases

## Model Sketch

`Phase { PhaseID, Order, Objectives []Objective, HasUnlocked }` — objectives
keep their current shape; phases group them. `MissionContract` gains
`phases []Phase` and `currentPhase`.

## Invariants

- Phase order is immutable after contract creation.
- Confirmation allowed only for objectives of the current phase (or any
  unlocked phase? [NEEDS CLARIFICATION — decide whether earlier phases stay
  open]).
- A phase unlocks when its mandatory objectives are all confirmed.

## Migration

Existing flat contracts map to one implicit phase with all objectives —
keeps old items valid without data migration.
