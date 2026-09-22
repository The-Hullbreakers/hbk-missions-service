---
capability_id: CAP-MISSION-BOARD
title: Mission Board
status: current
last_consolidated_change: CHG-0012-simplify-remove-mission-pool
updated: 2026-09-01
---

# Mission Board

## Intent

Publish, discover, and accept eligible mission offers so members can take solo or team contracts while respecting active-mission limits and template-exclusion rules.

## Scope

- Publish versioned offers with issuer, category, type, subtype, mode, difficulty, window, objectives, merits, repeatability, and capacity.
- Discover available offers filtered by tenant, category, type, subtype, mode, difficulty, and availability.
- Enforce a per-player active-mission limit with optional category overrides.
- Hide any offer whose stable definition the player already has active.
- Preserve accepted contract terms even when the offer or definition is updated later.

## Non-Goals

- Generic rule engine for objectives.
- Procedural contract generation.
- Cross-season or global ranking.
- Manual reviews or disputes (see `mission-contract-fulfillment`).

## Key Concepts

- `MissionDefinition`: stable, versioned contract template.
- `MissionOffer`: published availability, capacity, and repeatability.
- `MissionMode`: `SOLO`, `TEAM`, `FLEX`.
- `ActiveLimitPolicy`: tenant default and optional category overrides.
- `DecideAcceptance`: single decision that validates tenant, eligibility, window, repeatability, capacity, and active limit.
- `ActiveContractSummary`: read-only value object consumed by `DecideAcceptance`.

## Requirements

### REQ-MB-001 — Publish a mission offer

An authorized operator publishes versioned offers with issuer, category, type, subtype, mode, difficulty, window, objectives, merits, repeatability, and capacity. An offer can be unlimited or have a maximum number of acceptances.

#### SCN-MB-POOL — Show a published offer

Given a published offer is inside its window, when an eligible member opens the board, then the offer appears with its current rules and availability.

### REQ-MB-002 — Discover eligible missions

A member queries `AVAILABLE` offers from their tenant and filters by category, type, subtype, mode, difficulty, and availability. The result explains why an offer is not eligible.

#### SCN-MB-DISCOVERY — Isolate and filter the pool

Given offers exist in several tenants and categories, when the member filters their board, then they only receive offers from their tenant that match the filters.

### REQ-MB-003 — Limit active contracts

Each tenant configures a default active-contract limit per player and can define overrides per category. `ACCEPTED`, `IN_PROGRESS`, and `PENDING_VALIDATION` count; unconfirmed invitations and terminal states do not.

#### SCN-MB-001 — Reject acceptance over the limit

Given a player already reached their effective limit, when they attempt to accept another eligible offer, then the acceptance is rejected without creating a contract or consuming offer capacity.

#### SCN-MB-002 — Preserve contracts when lowering the limit

Given an operator lowers the limit below a player's current contracts, when the policy becomes effective, then their contracts are not cancelled and they cannot accept another until they are below the limit.

### REQ-MB-004 — Consistent acceptance

Acceptance validates tenant, eligibility, window, repeatability, capacity, and limit in a single decision. Two concurrent acceptances cannot exceed the limit or consume the same last slot.

#### SCN-MB-CONCURRENT-ACCEPTANCE — Protect the last slot

Given one personal or offer slot remains, when two acceptances compete, then exactly one obtains the contract and the other receives an explainable rejection.

### REQ-MB-005 — Freeze promised rules

The contract keeps the version of objectives, merits, and policies active at acceptance. Editing a definition creates a new version and does not alter existing contracts.

#### SCN-MB-VERSION-SNAPSHOT — Respect the accepted promise

Given a player accepted a published version, when the operator publishes another version, then the contract keeps the original objectives and merits.

### REQ-MB-009 — Bind offers to the active season

A published offer is bound to the single active season. If no season is `ACTIVE`, the publish operation is rejected. The offer and all contracts created from it carry the same `season_id`.

#### SCN-MB-PUBLISH-ACTIVE — Publish with active season

Given a season `season-2026-q4` is `ACTIVE`, when an operator calls `POST /api/v1/missions/offers`, then the response is `201 Created` and the returned offer contains `season_id: "season-2026-q4"`.

#### SCN-MB-PUBLISH-NO-ACTIVE — Reject publish without active season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/missions/offers`, then the response is `422 Unprocessable Entity` with code `NO_ACTIVE_SEASON` and no offer is persisted.

### REQ-MB-008 — Hide an active template

If a player participates in an `ACCEPTED`, `IN_PROGRESS`, or `PENDING_VALIDATION` contract from a stable definition, the board excludes every offer of the same `mission_definition_id`, even from another version. Acceptance revalidates the rule to prevent bypass and concurrency.

#### SCN-MB-ACTIVE-TEMPLATE — Do not show or retake the template

Given a player has active template X, when they browse the board or attempt to accept another offer of X, then they do not see it and direct acceptance is rejected without consuming capacity.

#### SCN-MB-TEMPLATE-RELEASED — Show after terminal state

Given the contract for template X ended, when cooldown and repeatability allow another attempt, then the player sees eligible offers of X again.

## HTTP Boundary

- `POST /api/v1/missions/definitions`
- `POST /api/v1/missions/definitions/{definition_id}/publish`
- `POST /api/v1/missions/offers`
- `POST /api/v1/missions/offers/{offer_id}/accept`

## Invariants

- A player never exceeds their effective active-mission limit.
- An accepted contract preserves the offer terms that existed at acceptance.
- The board always filters by caller tenant and never exposes another tenant's offers.
- A stable definition already active for a player is hidden from that player.

## Implementation Notes

- Offer validation lives in the domain constructor `NewValidMissionOffer`; the application layer no longer checks offer fields manually.
- `AcceptMissionHandler` builds `ActiveContractSummary` values via `MissionContract.Summary` for the `DecideAcceptance` policy.

## Related Capabilities

- `mission-contract-fulfillment` executes accepted contracts.
- `seasonal-category-competition` owns seasons and standings.
