---
change_id: CHG-0001
artifact: design
status: approved
updated: 2026-08-26
layer_scope:
  domain: affected
  application: affected
  boundary: affected
  infrastructure: affected
  interface: affected
  composition: affected
---

# Business-Centered Design

## Bounded Context

Missions has offer, contract, confirmed objective, Season Merit and classification. Identity has identity and tenant; Party Manager owns parties; Achievements has achievements; Notification has delivery to the user.

## Aggregates

- `MissionDefinition`: versioned content and publishable policies.
- `MissionPool`: Curated collection and definition rotation policy.
- `MissionOffer`: availability, capacity and repeatability within the pool.
- `MissionContract`: acceptance, mode, roster, phases, objectives and results.
- `Season`: window and policy for categories, tiebreakers and prizes.
- `MeritLedger`: immutable credits and adjustments by season/category/receiver.

The leaderboard is a projection of the ledger. Does not accept direct edits.

## Taxonomy

`MissionCategory` defines the competitive bucket; `MissionType` and `MissionSubtype` describe the contract. The taxonomy is managed as versioned data, not through a class hierarchy.

## Variable Policies

Current ways of validating objectives and calculating merit use explicit strategies behind small interfaces. V1 supports manual confirmation or trusted source and fixed calculation per unit plus optional bonus. A generic rules engine is not created.

## Active Mission Limit

- Default policy per tenant, with optional override per category.
- Initial recommended value: 3 active contracts; The operating range is validated upon deployment.
- Account contracts `ACCEPTED`, `IN_PROGRESS` and `PENDING_VALIDATION`.
- As a team, it counts for each member after confirming participation.
- Reducing the limit does not cancel contracts; blocks new acceptances.
- Acceptance revalidates the limit and capacity under concurrency.

## Pool Rotation And Active Template Exclusion

- A reusable definition has stable identity and immutable versions.
- A pool selects approved definitions; does not generate procedural rules.
- `MANUAL`, `SCHEDULED` and `ROTATING_POOL` create offers with frozen terms.
- The board is personalized: it excludes any definition that the player has active.
- The exclusion uses `mission_definition_id`, not a version, to prevent evasion via a new publication.
- In team contracts it applies to each confirmed participant.
- A terminal state releases the exclusion; cooldown and repeatability can still block it.
- The board query and the acceptance command apply the same business policy.

## Team Scoring

The roster freezes at startup. The merits of a `TEAM` contract pertain to the team's competitive identity. Missions may retain individual contributions as statistics, but does not create duplicate individual scores.

## Achievement Boundary

Competition decides the observable result: category, season, track, rank and tier. A provider-neutral port publishes that fact. An injected adapter delivers it to Achievements in a retryable manner; the closing of the season does not depend on a synchronous response.

## Development Sequence

1. Domain: taxonomy, limits, contract, objective, season and ledger.
2. Application: publication, consultation, acceptance, progress, adjudication, and finalization.
3. Boundary RED: API and receipt/publishing of facts.
4. Infrastructure, Interface and Composition with DI owned by the module.
5. Security, concurrency, mutation, coverage, E2E, rollout and review.

## Rollback Principle

Before production there is no state to migrate. Once deployed, seasons and ledger are append-only; code rollback preserves data and disables new acceptance/publishing via configuration. No rollback erases merits or rankings.
