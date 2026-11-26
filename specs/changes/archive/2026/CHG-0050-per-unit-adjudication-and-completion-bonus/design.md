---
change_id: CHG-0050
artifact: design
status: proposed
updated: 2026-10-20
---

# Design — Per-Unit Adjudication And Completion Bonus

## Domain

`internal/missions/domain/`:

- New value object `UnitRef{ObjectiveID, UnitID}` — the reviewer-facing
  identity of a staged unit (the `ConfirmationID` also embeds the
  recipient, which the reviewer never supplies).
- `MissionContract` gains `rejectedUnits map[string]bool` keyed by
  `ConfirmationID`, plus `RejectedUnits() []string` (sorted for stable
  projection) and a `RestoreRejectedUnits` rehydrator.
- `Approve(rejected []UnitRef)` in `mission_contract_review.go`:
  1. `PENDING_VALIDATION` guard as today.
  2. Resolve every `UnitRef` to a staged `ConfirmationID`
     (`contract~objective~unit~recipient`); any miss →
     `ErrUnitNotStaged`, nothing mutates.
  3. Move rejected entries out of `pendingCredits`/`confirmedUnits` into
     `rejectedUnits`.
  4. Emit credits for the remaining staged units, then the `COMPLETION`
     credit when `Terms.AllMandatoryFulfilled` and
     `Offer.CompletionReward > 0`.
  5. `COMPLETED` + credits, one result set.
- New `ContractTerms.AllMandatoryFulfilled(count func(objectiveID string) int)`:
  every phase's mandatory objective is satisfied when its confirmed count
  reaches `UnitLimit` (unlimited `UnitLimit == 0` requires ≥ 1). Phase
  progression (`PhaseComplete`, `advancePhases`) keeps the existing ≥ 1
  unit semantics — reaching the next stage only needs the objective
  touched.
- `Approve()` with no rejects stays binary-compatible in behavior with
  today's approve-all.
- Sentinel: `ErrUnitNotStaged` → `422 UNIT_NOT_STAGED`.

## Application

- `ApproveContractHandler.Handle(identity, contractID, rejected
  []domain.UnitRef)` — reviewer role check unchanged; passes the verdict
  to `contract.Approve`; `ContractCreditsWriter.SaveApproved` commits
  contract (state + rejected set) and credits atomically — the port and
  transaction shape are unchanged.

## HTTP Boundary

- `POST /contracts/{id}/approve` body: optional
  `{"rejected_units": [{"objective_id": "...", "unit_id": "..."}]}`;
  empty/absent body = approve all. Malformed refs → `400
  INVALID_REQUEST`; valid shape but unstaged unit → `422
  UNIT_NOT_STAGED`.
- `contractDetail` gains `rejected_units: [{objective_id, unit_id}]`
  parsed from the stored `ConfirmationID`s. The member projection keeps
  only the caller's own rejected units — same scoping as
  `pending_credits`.

## Infrastructure (DynamoDB)

- Contract item persists a new `RejectedUnits` string-set attribute
  (confirmation ids); absent on legacy items → empty set. No migration.
- `SaveApproved` transaction unchanged: contract `PutItem` (with rejected
  set) + conditional credit puts; retry collapses via
  `attribute_not_exists` on `ConfirmationID`.

## Catalog (seeds + seed script)

- `seeds/pvp-missions.json`: per mission, single objective
  (`baja-fps-verificada` / `derribo-verificado` /
  `victoria-duel-verificada`), `UnitLimit` = required count,
  `MeritPerUnit` = old 90% share ÷ limit, `Bonus: 0`,
  `CompletionReward` = old 10% share, `Version: "3"`, reduced `Rules`
  (≈5).
- `seeds/racing-missions.json`: shared `rules` rewritten (5 items);
  generator emits `vuelta-verificada` with `UnitLimit: 5`,
  `MeritPerUnit = reward*9//10//5` (36/72/126/198 — all exact),
  `CompletionReward = reward//10` (20/40/70/110).
- `scripts/seed_missions.py`: allow 1-objective missions
  (`1 <= len <= 5`); racing expansion emits the single per-lap objective
  + `CompletionReward`.

## Error Mapping

- `ErrUnitNotStaged` → `422 UNIT_NOT_STAGED`.

## Concurrency / Atomicity

- Two simultaneous `approve` calls: both see `PENDING_VALIDATION`, the
  conditional write on the contract item (or the duplicate
  `ConfirmationID` puts) collapses them to one materialization — same
  guarantee as today.
- A verdict retry after mid-write failure reproduces the same reject set
  and credit set; rejected-unit recording is part of the contract item,
  so it cannot double-apply.

## Out of Scope (flagged)

- Member projection shows rejected units only for the caller's own
  recipient (team contracts: rejects credit the team, so all members see
  the team-scoped list).
- Frontend verdict UI and the synthetic completion-objective row live in
  web-app change `173-per-unit-review-verdicts`.
