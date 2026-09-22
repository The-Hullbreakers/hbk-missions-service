---
change_id: CHG-0010
artifact: delta
status: active
updated: 2026-08-30
---

# Delta — Bind Mission Offers To The Active Season

## What Is Changing

### Mission Board

`MissionOffer` gains a required `SeasonID`. Publishing an offer now requires an active season; otherwise the operation is rejected.

### Seasonal Category Competition

The single-active-season rule is leveraged by `PublishMissionHandler`. A new query port exposes the active season to the publishing use case.

### Mission Contract Fulfillment

`MissionContract` stores the `SeasonID` inherited from its offer. `Confirm` rejects a confirmation whose supplied season does not match the contract's season.

## BDD Scenarios

### SCN-MB-PUBLISH-ACTIVE-001 — Publish an offer when an active season exists

Given a season `season-2026-q4` exists with `State: ACTIVE`,
when an operator calls `POST /api/v1/missions/offers` with a valid offer,
then the response is `201 Created` and the returned offer contains `season_id: "season-2026-q4"`.

### SCN-MB-PUBLISH-REJECT-NO-ACTIVE-001 — Reject offer creation without an active season

Given no season is `ACTIVE`,
when an operator calls `POST /api/v1/missions/offers`,
then the response is `422 Unprocessable Entity` with code `NO_ACTIVE_SEASON` and no offer is persisted.

### SCN-MB-PUBLISH-REJECT-EMPTY-SEASON-001 — Reject an offer with an empty season reference

Given a season `season-2026-q4` is `ACTIVE`,
when the domain constructor `NewValidMissionOffer` is called with an empty `season_id`,
then it returns `ErrOfferInvalid`.

### SCN-MCF-CONFIRM-SEASON-001 — Confirm a unit against the contract's season

Given a contract exists for an offer bound to `season-2026-q4` with `State: IN_PROGRESS`,
when a reviewer calls `POST /api/v1/missions/contracts/{contract_id}/evidence` with `season_id: "season-2026-q4"`,
then a `MeritCredit` is recorded with `SeasonID: "season-2026-q4"`.

### SCN-MCF-CONFIRM-SEASON-MISMATCH-001 — Reject confirmation for the wrong season

Given a contract is bound to `season-2026-q4`,
when a reviewer calls evidence with `season_id: "season-2026-q3"`,
then the response is `422 Unprocessable Entity` with code `SEASON_MISMATCH` and no credit is recorded.

### SCN-MCF-ACCEPT-INHERIT-001 — Contract inherits the offer's season

Given a published offer with `season_id: "season-2026-q4"`,
when a player accepts that offer,
then the created contract contains `SeasonID: "season-2026-q4"`.

## Error Vocabulary

- `NO_ACTIVE_SEASON` — returned when `POST /api/v1/missions/offers` is called and no season is `ACTIVE`.
- `SEASON_MISMATCH` — returned when a confirmation request supplies a `season_id` different from the contract's bound season.
