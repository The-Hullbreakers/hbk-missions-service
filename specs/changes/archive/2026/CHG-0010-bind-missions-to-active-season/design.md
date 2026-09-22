---
change_id: CHG-0010
artifact: design
status: active
updated: 2026-08-30
---

# Design — Bind Mission Offers To The Active Season

## Domain

### `MissionOffer`

```go
type MissionOffer struct {
    OfferID  string
    Category string
    Capacity int
    Mode     MissionMode
    SeasonID string
}
```

`NewValidMissionOffer(offerID, category string, capacity int, mode MissionMode, seasonID string)` adds `seasonID != ""` to the validation gate and returns `ErrOfferInvalid` when missing.

### `MissionContract`

```go
type MissionContract struct {
    ContractID     string
    Offer          MissionOffer
    State          ContractState
    SeasonID       string
    // ... existing fields
}
```

`NewMissionContract` and `NewTeamMissionContract` copy `offer.SeasonID` into the contract.

`NewMissionContractFromState` receives `seasonID` as a new parameter to allow reconstruction from persistence.

### `Confirm`

`MissionContract.Confirm` validates:

1. contract is not terminal.
2. contract is `IN_PROGRESS`.
3. `season.CanAcceptMerits(c)`.
4. `season.SeasonID == c.SeasonID` (new `ErrSeasonMismatch`).
5. objective category matches offer category.
6. roster is non-empty.
7. no duplicate confirmation.
8. objective unit limit is respected.

## Application

### `PublishMissionHandler`

New port:

```go
type ActiveSeasonsReader interface {
    ListActive() ([]domain.Season, error)
}
```

`PublishMissionHandler` now depends on `ActiveSeasonsReader`. On `Handle`:

1. Query `ListActive()`.
2. If no active season, return `ErrNoActiveSeason`.
3. If more than one active season, return `ErrActiveSeasonAlreadyExists` (defensive; single-active rule should prevent this).
4. Build `NewValidMissionOffer(..., active[0].SeasonID)`.
5. Persist and return.

### `ConfirmObjectiveHandler`

No signature change. `ConfirmObjective` continues to receive `seasonID` from the request, but `MissionContract.Confirm` now validates that the loaded `Season` matches the contract's bound `SeasonID`.

### `AcceptMissionHandler`

No change to behavior; the contract constructor automatically inherits the season from the offer.

## Boundary

### `POST /api/v1/missions/offers`

Request stays the same. Response now includes `season_id`.

New public code:

- `NO_ACTIVE_SEASON` mapped from `domain.ErrNoActiveSeason`.

### `POST /api/v1/missions/contracts/{contract_id}/evidence`

Request still accepts `season_id`. New public code:

- `SEASON_MISMATCH` mapped from `domain.ErrSeasonMismatch`.

## Infrastructure

### `OfferStore`

Save includes `SeasonID`:

```go
item["SeasonID"] = &types.AttributeValueMemberS{Value: offer.SeasonID}
```

`Get` returns the stored `SeasonID`; if absent, it returns `domain.MissionOffer` with the empty string (constructor will reject downstream on re-validation).

### `ContractStore`

`Save` persists `SeasonID` from `contract.SeasonID`.

`Get` reconstructs `MissionContract` using `NewMissionContractFromState` with the stored `SeasonID`. If a legacy item has no `SeasonID`, it defaults to the empty string and `Confirm` will reject with `SEASON_MISMATCH` until backfilled or re-accepted.

## Composition

`NewServer` wires `seasonStore` into both `createSeasonApp` and `publishApp`:

```go
publishApp := application.NewPublishMissionHandler(offerStore, seasonStore)
```

## Data Model

- `OFFER` items gain a `SeasonID` string attribute.
- `CONTRACT` items gain a `SeasonID` string attribute.
- No GSI or key changes; `SeasonID` is a non-key attribute.

## Backward Compatibility

- Old offers without `SeasonID` cannot be re-published or accepted into a new contract until a backfill. Existing active contracts read with empty `SeasonID` will fail confirmation with `SEASON_MISMATCH`, which is intentional: the operator must ensure every in-flight contract is tied to the active season before the change is deployed, or a backfill migration must be run.
- The `POST /api/v1/missions/contracts/{contract_id}/evidence` request still accepts `season_id`, so clients are not forced to change immediately.
