---
change_id: CHG-0026
artifact: tasks
status: proposed
updated: 2026-10-09
---

# Tasks — Add Mission Content And Completion Reward

## Wave 1 — Domain

### T-01 — Extend definition content fields

- RED: `tests/unit/missions/domain/mission_definition_test.go` — create with
  `title`/`description`/`completion_reward`; reject empty title, empty
  description, negative reward; objective without `name` rejected.
- GREEN: `internal/missions/domain/mission_definition.go`,
  `objective.go` — fields, validation, constructor signature.
- BLUE: keep constructor readable; extract param grouping only if needed.
- Verify: `go test ./tests/unit/missions/domain/`

### T-02 — Freeze content into offers and terms

- RED: `mission_offer_test.go` — `NewMissionOfferFromDefinition` copies
  `Title`, `Description`, `CompletionReward`; `contract_terms_test.go` —
  snapshot preserves objective `Name`/`Description`.
- GREEN: `mission_offer.go`, `contract_terms.go`.
- Verify: `go test ./tests/unit/missions/domain/`

### T-03 — Complete emits completion credit

- RED: `mission_contract_test.go` — `Complete` returns `MeritCredit` with
  `ObjectiveID "COMPLETION"`, `Amount == CompletionReward`, season/category/
  recipient/track correct; zero reward returns zero-amount credit; terminal
  and not-fully-confirmed still rejected.
- GREEN: `mission_contract_lifecycle.go` — `Complete` returns the credit
  reusing `recipient()`/`track()`/`NewConfirmationID`.
- Verify: `go test ./tests/unit/missions/domain/`

## Wave 2 — Application

### T-04 — Create definition accepts content

- RED: `create_mission_definition_test.go` — handler forwards title,
  description, completion_reward, named objectives; domain errors propagate.
- GREEN: `application/create_mission_definition.go`.
- Verify: `go test ./tests/unit/missions/application/`

### T-05 — Complete contract records reward atomically

- RED: `complete_contract_test.go` — handler calls `SaveConfirmed` with the
  completion credit; zero reward uses plain `Save`; rosters/team recipient
  asserted.
- GREEN: `application/complete_contract.go` — swap port to
  `ContractAndCreditWriter`.
- Verify: `go test ./tests/unit/missions/application/`

## Wave 3 — Boundary

### T-06 — HTTP create/list payloads

- RED: `tests/unit/missions/interface/http/create_mission_definition_test.go`,
  `list_definitions_test.go`, `tests/integration/http/list_definitions_test.go`
  — request accepts new fields; empty title/description → `422`; responses
  expose `Title`, `Description`, `CompletionReward`, objective `Name`.
- GREEN: `internal/missions/interface/http/create_mission_definition.go`.
- Verify: `go test ./tests/unit/missions/interface/http/ ./tests/integration/http/`

### T-07 — DynamoDB persists and restores content

- RED: `tests/integration/infrastructure/dynamodb/definition_store_test.go`,
  `offer_store_test.go`, `contract_store_test.go` — round-trip `Title`,
  `Description`, `CompletionReward`, objective `Name`/`Description`; contract
  reload still emits the correct completion credit.
- GREEN: `definition_store.go`, `definition_store_queries.go`,
  `offer_store.go`, `contract_store_confirm.go`, `contract_store_mapping.go`,
  objectives attribute helpers.
- Verify: `go test ./tests/integration/infrastructure/dynamodb/`

### T-08 — Composition and e2e journey

- RED→GREEN: `tests/integration/composition/solo_journey_e2e_test.go` —
  full create → publish → offer → accept → confirm → complete journey ends
  with `COMPLETED` contract and standing increased by `completion_reward`.
- `internal/missions/composition/server.go` — wire `ContractAndCreditWriter`
  into `NewCompleteContractHandler`.
- Verify: `go test ./tests/integration/composition/ ./tests/integration/e2e/`

## Wave 4 — Gates

### T-09 — BLUE cleanup

Refactor while green; all maintained files `< 150` physical lines; gofmt/vet.

### T-10 — Documentation convergence

Update `docs/API-CONTRACT.md`, `docs/API-GUIDE-FRONTEND.md`,
`docs/DOMAIN-RULES.md` for content fields and completion reward.

### T-11 — Risk gates and validation

`tools/validate-sdd-change.sh`; L3 → mutation gate on `Complete`/scoring and
critical-e2e evidence on the complete journey; update `verification.md`.

### T-12 — Consolidate and archive

Verify → merge delta into `mission-board` and `mission-contract-fulfillment`
canonical specs → update `specs/index.md` → archive change.
