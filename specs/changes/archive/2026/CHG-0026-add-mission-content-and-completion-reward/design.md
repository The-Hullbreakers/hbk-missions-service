---
change_id: CHG-0026
artifact: design
status: proposed
updated: 2026-10-09
---

# Design — Add Mission Content And Completion Reward

## Domain

### `MissionDefinition` (`internal/missions/domain/mission_definition.go`)

```go
type MissionDefinition struct {
    DefinitionCode   string
    Title            string   // required
    Description      string   // required
    Category         string
    Capacity         int
    Mode             MissionMode
    Objectives       []Objective
    CompletionReward int      // >= 0
    State            DefinitionState
    Version          string
}
```

`NewMissionDefinition(code, title, description, category string, capacity int, mode MissionMode, completionReward int, objectives []Objective)`:
- rejects empty `code`, `title`, `description`, `category`; `capacity < 0`;
  invalid `mode`; `completionReward < 0` → `ErrDefinitionInvalid`.
- objectives validated element-wise: empty `objective_id` or `name` →
  `ErrDefinitionInvalid`.

### `Objective` (`internal/missions/domain/objective.go`)

```go
type Objective struct {
    ObjectiveID  string
    Name         string   // required in definitions
    Description  string   // optional
    Category     string
    MeritPerUnit int
    Bonus        int
    UnitLimit    int
}
```

Constructor gains `name`/`description`; `ContractTerms` snapshots keep them so
frozen contract rendering stays stable.

### `MissionOffer` (`internal/missions/domain/mission_offer.go`)

Gains `Title`, `Description`, `CompletionReward`, copied from the definition at
`NewMissionOfferFromDefinition` — the offer freezes the published version's
content.

### Completion credit (`mission_contract_lifecycle.go`, `mission_contract_confirm.go`)

`Complete(identity)` returns `(MeritCredit, error)`:
- same guards as today (roster authorization, terminal/state checks,
  `AllConfirmed`);
- on success sets `ContractCompleted` and returns a `MeritCredit` with
  `ObjectiveID: "COMPLETION"`, `Amount: c.Offer.CompletionReward`,
  `SeasonID: c.SeasonID`, `Category: c.Offer.Category`, recipient/track via the
  existing `recipient()`/`track()` helpers, and a `ConfirmationID` derived from
  `(contractID, "COMPLETION", contractID, recipient)` so a retry is a no-op
  write, not a second credit;
- `CompletionReward == 0` → `MeritCredit.Amount == 0`; the application layer
  skips recording zero-amount credits.

## Application

- **`create_mission_definition.go`** — `Handle(code, title, description,
  category, capacity, mode, completionReward, objectives)`; forwards to
  `domain.NewMissionDefinition`.
- **`publish_mission.go`** — unchanged signature; `MissionOffer` now carries
  the content fields (populated inside the domain constructor).
- **`complete_contract.go`** — swaps `ContractWriter` for the existing
  `ContractAndCreditWriter` port (`SaveConfirmed`) so the `COMPLETED`
  transition and reward credit commit atomically; when `Amount == 0` the plain
  `Save` path is used. Constructor signature gains no new dependencies —
  `ContractStore` already implements `SaveConfirmed`.
- **`list_definitions.go` / `list_offers.go`** — unchanged; content fields flow
  through the domain types.

## Interface (HTTP)

- `create_mission_definition.go` — request gains `title`, `description`
  (required), `completion_reward` (optional int, default 0);
  `objectiveRuleRequest` gains `name` (required), `description` (optional).
  Empty `title`/`description`/`name` reach the domain and map to `422`.
- `list_definitions.go`, `list_offers.go` handlers — domain structs serialize
  directly; JSON exposes `Title`, `Description`, `CompletionReward`,
  `Objectives[].Name/Description`.

## Infrastructure (DynamoDB)

- `definition_store.go` — item gains `Title`, `Description`,
  `CompletionReward` attributes; objectives attribute gains `Name`/`Description`
  inside `objectivesToAttribute`.
- `definition_store_queries.go` — `definitionFromItem` reads them; no legacy
  fallback needed (no live items).
- `offer_store.go` — persists and restores `Title`, `Description`,
  `CompletionReward`.
- `contract_store_confirm.go` / `contract_store_mapping.go` — contract item
  gains `Title`, `Description`, `CompletionReward` on the embedded offer so a
  reloaded contract can still emit its completion credit.
- `creditItem` (merit ledger) — unchanged; `ObjectiveID: "COMPLETION"` keys
  idempotency via `attribute_not_exists` in `SaveConfirmed`.

## Composition

- `server.go` — `NewCompleteContractHandler` now receives the
  `ContractAndCreditWriter` (the same `contractStore`); no new stores.

## Test Doubles

Application unit tests use hand-written port doubles; `CompleteContractHandler`
tests get a recording `ContractAndCreditWriter` fake asserting the completion
`MeritCredit`.

## Idempotency And Concurrency

- The completion credit reuses the `ConfirmationID` uniqueness mechanism:
  `NewConfirmationID(contractID, "COMPLETION", contractID, recipient)` is
  deterministic, so a double `Put` under `attribute_not_exists` yields
  `ErrDuplicateConfirmation`-equivalent cancel — mapped to success since the
  contract is already `COMPLETED` (terminal-state check rejects earlier).
- Zero reward: no credit emitted, ledger untouched.
