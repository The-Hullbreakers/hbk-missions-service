---
change_id: CHG-0002
artifact: canonical-delta
status: proposed
updated: 2026-08-27
---

# Canonical Delta

## ADDED — Mission Contract Fulfillment

### REQ-MCF-002 — Progress through phases and objectives

A contract contains sequential phases and mandatory or optional objectives. Objectives can measure compliance, quantity or time and declare their scoring units.

#### SCN-MCF-OBJECTIVES — Unlock the next phase

Given a phase requires its mandatory objectives, when all are confirmed, then the next phase becomes available without requiring optional objectives.

### REQ-MCF-003 — Confirmar unidades puntuables

A unit is only valid when the contract was active, the target was eligible, it occurred in window, the recipient was eligible, and the evidence or source was accepted.

#### SCN-MCF-001 — Credit each confirmed bounty

Given a Bounty PvP mission offers 100 merits per single objective, when four valid objectives are confirmed, then the participant receives 400 merits in `BOUNTY_HUNTING` for the active season.

### REQ-MCF-004 — Prevent duplicates and basic farming

The same unit, target, contract and recipient are credited at most once. The mission can define cooldown, unit limit and published replay restrictions.

#### SCN-MCF-DUPLICATE — Ignore a replay

Given a target has already produced a credit, when the same confirmation arrives again, then neither the ledger nor the standing changes.

### REQ-MCF-005 — Conservar resultados terminales

A contract ends `COMPLETED`, `FAILED`, `ABANDONED`, `EXPIRED`, `CANCELLED`, or `DISQUALIFIED`. No terminal returns to active and the history preserves reasons and evidence.

#### SCN-MCF-TERMINAL — Reject confirmation in terminal contract

Given a contract is in a terminal state, when a new confirmation arrives, it is then rejected without creating or modifying credits.

## ADDED — Seasonal Category Competition

### REQ-SCC-001 — Register Season Merits

Each credit belongs to exactly one tenant, season, category, mode and recipient. There is no permanent XP or user level.

#### SCN-SCC-CATEGORY-CREDIT — Credit the correct bucket

Given a Bounty Hunting objective was confirmed, when credited, it only increases the Bounty status of the corresponding season and track.

### REQ-SCC-002 — Separar clasificaciones

Each category has an individual and independent team leaderboard. Team points are not automatically duplicated among team members.

#### SCN-SCC-SEPARATE-TRACKS — Avoid duplicate team scoring

Given a team contract generates merits, when standings are updated, then the team increases and not the individual standings of each member.

### REQ-SCC-004 — Congelar y finalizar

Moving from `ACTIVE` to `FREEZING` ceases new merits, resolves eligible validations and disputes, and then publishes an immutable `FINALIZED` ranking.

#### SCN-SCC-FREEZE — Bloquear nuevo puntaje

Given a season is `FREEZING`, when a new action arrives outside of eligible work started on time, it does not generate merit.

## MODIFIED

There are no previous published canonical capabilities. `CHG-0001` maintains ownership of platform initialization; this change deepens the US-MCF-001 use case.

## REMOVED

No se elimina comportamiento existente.
