# Delta — List Mission Offers and Player Contracts Over HTTP

## Affected Canonical Specs

- `specs/capabilities/mission-board/spec.md` (MODIFIED)
- `specs/capabilities/mission-contract-fulfillment/spec.md` (MODIFIED)

## Summary

Add two read-only HTTP endpoints backed by new application query use cases and DynamoDB list adapters: offer discovery for the member's board, and listing of the member's contracts.

## Changed Behavior

### ADDED

- `REQ-MB-010` — `GET /api/v1/missions/offers` returns `AVAILABLE` offers for the caller's tenant, filterable by `season_id` (defaults to the active season), `category`, and `mode`. `MissionOffer` has no `type`/`difficulty` fields, so those filters were dropped during domain RED. Offers whose stable `mission_definition_id` the player already has active are excluded (`REQ-MB-008`).
- `REQ-MC-LIST` — `GET /api/v1/missions/contracts` returns the authenticated player's contracts (roster membership), filterable by `season_id` and `status`.
- Persistence stores honor `tenantID` on list paths (fixing the current `_ = tenantID` ignore).

### MODIFIED

- `mission-board` HTTP Boundary gains `GET /api/v1/missions/offers`.
- `mission-contract-fulfillment` HTTP Boundary gains `GET /api/v1/missions/contracts`.

### REMOVED

- None.

## BDD Acceptance Scenarios

### SCN-MB-LIST-AVAILABLE

Given offers exist for season `season-2026-q4` in the caller's tenant, when a member calls `GET /api/v1/missions/offers?status=AVAILABLE`, then the response is `200` with the offers of the active season only.

### SCN-MB-LIST-TENANT-ISOLATION

Given offers exist in tenant `tenant-b`, when a member of `tenant-a` calls `GET /api/v1/missions/offers`, then no `tenant-b` offer is returned.

### SCN-MB-LIST-HIDES-ACTIVE-TEMPLATE

Given the player has an `ACCEPTED` contract on definition `bounty-template#v1`, when they call `GET /api/v1/missions/offers`, then no offer with `definition_id=bounty-template` appears.

### SCN-MB-LIST-FILTER

Given offers in categories `BOUNTY` and `EXPLORATION`, when a member calls `GET /api/v1/missions/offers?category=BOUNTY`, then only `BOUNTY` offers are returned.

### SCN-MB-LIST-INVALID-FILTER

Given an unknown `status` or `category` value, when a member calls the endpoint, then the response is `400` with a typed error code.

### SCN-MC-LIST-OWN-CONTRACTS

Given the player is in the roster of two contracts and another player has one, when the member calls `GET /api/v1/missions/contracts`, then only their two contracts are returned.

### SCN-MC-LIST-FILTER-SEASON

Given contracts in `season-2026-q3` and `season-2026-q4`, when the member calls `GET /api/v1/missions/contracts?season_id=season-2026-q4`, then only `season-2026-q4` contracts are returned.

### SCN-MC-LIST-FILTER-STATUS

Given contracts in `ACCEPTED` and `COMPLETED` states, when the member calls `GET /api/v1/missions/contracts?status=ACTIVE`, then only non-terminal contracts are returned.

### SCN-MB-LIST-UNAUTHENTICATED

Given no valid session token, when a caller hits either endpoint, then the response is `401`.

## Error Vocabulary

- `INVALID_STATUS_FILTER`
- `INVALID_CATEGORY_FILTER`
- `UNAUTHORIZED` (existing middleware)
