---
change_id: CHG-0052
artifact: delta
status: proposed
updated: 2026-10-20
---

# Delta — Remove Offer Capacity

## Requirements

### REQ-MB-001 — Publish a mission offer — MODIFIED

An authorized operator publishes versioned offers with issuer, category,
type, subtype, mode, difficulty, window, objectives, merits, and
repeatability. Offers carry no acceptance cap — every eligible member may
accept while the offer is published.

### REQ-MB-003 — Limit active contracts — MODIFIED

Unchanged except wording: a rejection over the limit creates no contract
(the removed phrase "or consuming offer capacity" is dropped).

### REQ-MB-004 — Consistent acceptance — MODIFIED

Acceptance validates eligibility, window, repeatability, and limit in a
single decision, then commits the contract and each roster member's
active-set entry in **one conditional transaction**. Concurrent
acceptances serialize: losers receive the explainable rejection
(`ACTIVE_MISSION_LIMIT_REACHED` or `SEASON_NOT_ACTIVE`) with no partial
writes. `OFFER_CAPACITY_EXHAUSTED` no longer exists.

#### SCN-MB-CONCURRENT-ACCEPTANCE — Protect the last slot — MODIFIED

Given the player has one active-set slot remaining, when two acceptances
compete, then exactly one obtains the contract and the other receives an
explainable rejection. (Offer-level slot contention is gone.)

### REQ-MB-010 — List mission offers — MODIFIED

`status` accepts `AVAILABLE` (default, offers in the resolved season
scope minus active-template exclusions) or `ALL`; no capacity filter.

### REQ-MB-013 — Publish definitions — MODIFIED

`POST /api/v1/missions/definitions` no longer accepts `capacity`;
`POST .../publish` no longer accepts `capacity`. Payloads carrying the
field are ignored for forward compatibility (unknown JSON fields remain
non-fatal per existing decoder behavior).

## Invariants — MODIFIED

- Concurrent acceptances serialize through a conditional transaction — a
  check-then-write race cannot overspend the player's active limit.
- Reaching a terminal contract state releases the player's active-set
  slot.
- (Removed) "Offer capacity is consumed atomically via
  `remaining_capacity`."

## Implementation Notes — MODIFIED

- `ContractStore.CommitAcceptance` runs `TransactWriteItems`: `Put`
  contract + `Update` the per-player `PLAYER_ACTIVE` StringSet item
  (`ADD contract_ids` conditioned `size() < limit`).
  `ConditionalCheckFailed` maps to `ErrActiveMissionLimitReached` by
  cancellation index.
- `NewValidMissionOffer` no longer takes a capacity argument.
