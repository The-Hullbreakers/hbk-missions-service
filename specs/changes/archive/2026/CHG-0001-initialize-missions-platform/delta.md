---
change_id: CHG-0001
artifact: canonical-delta
status: approved
updated: 2026-08-26
---

# Canonical Delta

## ADDED — Mission Board

### REQ-MB-001 — Publish a pool of offers

An authorized operator publishes versioned offers with issuer, category, type, subtype, mode, difficulty, window, objectives, merits, repeatability and capacity. An offer can be unlimited or have a maximum number of acceptances.

#### SCN-MB-POOL — Show a published offer
Given a published offer is within your window, when an eligible member opens the board, then the offer appears with its current rules and availability.

### REQ-MB-002 — Discover eligible missions

The member consults `AVAILABLE` offers from their tenant and filters by category, type, subtype, mode, difficulty and availability. The result explains why an offer is not eligible.

#### SCN-MB-DISCOVERY — Isolate and filter the pool
Given there are offers in several tenants and categories, when the member filters their board, they only receive offers from their tenant that meet the filters.

### REQ-MB-003 — Limitar contratos activos

Each tenant configures a default limit of active contracts per player and can define overrides per category. `ACCEPTED`, `IN_PROGRESS` and `PENDING_VALIDATION` count; Unconfirmed invitations and terminal statuses do not count.

#### SCN-MB-001 — Reject an acceptance over the limit

Given a player has already reached his effective limit, when he attempts to accept another eligible offer, then the acceptance is rejected without creating a contract or consuming the offer's capacity.

#### SCN-MB-002 — Keep contracts when reducing limit

Given an operator reduces the limit below a player's current contracts, when the policy goes into effect, then his contracts are not canceled and he cannot accept another until he falls below the limit.

### REQ-MB-004 — Consistent acceptance

Acceptance validates tenant, eligibility, window, repeatability, capacity and limit in a single decision. Two concurrent accepts cannot exceed the limit or consume the same last capacity.

#### SCN-MB-CONCURRENT-ACCEPTANCE — Protect the last place
Given a personal or offer spot, when two acceptances compete, then exactly one gets the contract and the other receives an explainable rejection.

### REQ-MB-005 — Congelar reglas prometidas

The contract retains the version of objectives, merits and policies in force when accepted. Editing a definition creates a new version and does not alter existing contracts.

#### SCN-MB-VERSION-SNAPSHOT — Respect the accepted promise
Given a player accepted a published version, when the operator publishes another version, then the contract retains original objectives and merits.

### REQ-MB-006 — Administrar pools de plantillas

A `MissionPool` groups approved definitions by tenant, season, and category. Configure cadence, number of offers, duration, capacity, weights and repetition; It does not invent objectives or rewards outside of the templates.

#### SCN-MB-ROTATION — Rotate predefined offers
Given an active pool must publish ten offers every thirty minutes, when it starts another cycle, it then creates offers from eligible definitions with frozen rules and closes the unaccepted ones from the expired cycle.

### REQ-MB-007 — Support publishing modes

An offer is published using `MANUAL`, `SCHEDULED`, or `ROTATING_POOL`. As the season enters `FREEZING` new rotations cease; Closing an offer does not cancel accepted contracts.

#### SCN-MB-PUBLICATION-MODE — Preserve a contract when rotating
Given a player accepted an offer from the previous cycle, when the board publishes the next cycle, then the contract continues until its deadline even if the offer is no longer available.

### REQ-MB-008 — Hide an active template

If a player participates in a contract `ACCEPTED`, `IN_PROGRESS` or `PENDING_VALIDATION` of a stable definition, the board excludes all offers of that same `mission_definition_id`, even of another version. Acceptance revalidates the rule to avoid bypass and concurrency.

#### SCN-MB-ACTIVE-TEMPLATE — Do not show or resume the template
Given a player has the X template active, when they view the board or tries to accept another offer from X, he then does not see it and the direct acceptance is rejected without consuming capacity.

#### SCN-MB-TEMPLATE-RELEASED — Show after terminal state
Given the contract for template X has ended, when cooldown and replayability allow another attempt, then the player sees eligible offers for X.

## ADDED — Mission Contract Fulfillment

### REQ-MCF-001 — Run alone or as a team

`SOLO` credits the individual participant; `TEAM` credits the team; `FLEX` chooses a mode upon acceptance and cannot change afterwards. Each confirmed member of a team contract consumes one spot in their personal limit.

#### SCN-MCF-TEAM — Freeze an eligible roster
Given all guests have active capacity, when they confirm and start the mission, then the roster is frozen and the contract uses the team track.

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

#### SCN-MCF-TERMINAL — Keep a dropout
Given a contract is active, when the participant abandons it, it then ends without new merits and remains visible in the history.

## ADDED — Seasonal Category Competition

### REQ-SCC-001 — Register Season Merits

Each credit belongs to exactly one tenant, season, category, mode and recipient. There is no permanent XP or user level.

#### SCN-SCC-CATEGORY-CREDIT — Credit the correct bucket
Given a Bounty Hunting objective was confirmed, when credited, it only increases the Bounty status of the corresponding season and track.

### REQ-SCC-002 — Separar clasificaciones

Each category has an individual and independent team leaderboard. Team points are not automatically duplicated among team members.

#### SCN-SCC-SEPARATE-TRACKS — Avoid duplicate team scoring
Given a team contract generates merits, when standings are updated, then the team increases and not the individual standings of each member.

### REQ-SCC-003 — Proyectar leaderboards auditables

The table shows position, merits and tiebreakers based on the ledger. A correction rightly adds an adjustment and does not rewrite the original credit.

#### SCN-SCC-CORRECTION — Fix without hiding history
Given a credit was awarded in error, when a reviewer applies a correction, then the total changes through an auditable adjustment.

### REQ-SCC-004 — Congelar y finalizar

Moving from `ACTIVE` to `FREEZING` ceases new merits, resolves eligible validations and disputes, and then publishes an immutable `FINALIZED` ranking.

#### SCN-SCC-FREEZE — Lock new score
Given a season is `FREEZING`, when a new action arrives outside of eligible work, it does not generate merits.

### REQ-SCC-005 — Entregar puestos a Achievements

Missions determines the awarded positions and tiers and communicates idempotent facts. Achievements has the visual definition, materialization and permanence of the achievement.

#### SCN-SCC-AWARD — Retry a awarded placement
Given a winner has already been determined, when submission to Achievements is retried, then the same position is communicated without creating another logical prize.

#### SCN-SCC-001 — Start a season without deleting the previous one

Given a season ended with a user in first place, when the next one starts, then their new leaderboard starts at zero and the previous standings remain searchable.

### REQ-SCC-006 — Aplicar desempates publicados

The season freezes before starting its ordered criteria. V1 uses merits, missions completed, unique objectives, and time at which the result was reached.

#### SCN-SCC-TIE — Resolve a deterministic tie
Given two participants have the same merits, when the table is finalized, then the published tiebreakers are applied in order.

## MODIFIED

There are no previous canonical capabilities.

## REMOVED

No se elimina comportamiento existente.
