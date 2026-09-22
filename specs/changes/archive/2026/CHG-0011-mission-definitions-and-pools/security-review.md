---
change_id: CHG-0011
artifact: security-review
status: proposed
updated: 2026-08-30
---

# Security Review — Add Mission Definitions And Season Pools

## Trust Boundaries

- HTTP routes are protected by the existing tenant and auth middleware.
- Definition/pool administration is an operator capability; members only discover and accept offers.

## Data Exposure

- Pool endpoints return only the caller's tenant content.
- Definition content does not include internal revision history beyond the published version.
- Contracts do not expose the player's evidence or private identifiers in leaderboards.

## Input Validation

- `definition_id`, `pool_id`, and `season_id` are validated for non-empty and tenant-scoped values.
- `POST /api/v1/missions/offers` rejects unknown `definition_id` or `pool_id`.
- Offer capacity and pool policy values are validated in the domain.

## Concurrency

- Definition lifecycle transitions are single-row updates.
- Pool creation and offer publication are idempotent at the offer identity level.
- `AcceptMission` uses the existing policy to prevent race conditions on capacity and active limits.

## Findings

No Critical or High findings anticipated. Any new findings will be recorded here and resolved before consolidation.
