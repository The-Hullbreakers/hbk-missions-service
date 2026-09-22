---
change_id: CHG-0010
artifact: test-first-tasks
status: active
updated: 2026-08-30
---

# Test-First Tasks

## T-001 — Add `SeasonID` to `MissionOffer`

- status: pending
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: Gate 2
- test_id: TEST-DOMAIN-OFFER-001
- test_type: unit
- test_file: `tests/unit/missions/domain/mission_offer_test.go`
- mapped_production_files:
  - `internal/missions/domain/mission_offer.go` — add `SeasonID`, update `NewValidMissionOffer`.
- requirements: REQ-MB-001
- done_when: `NewValidMissionOffer` rejects empty `SeasonID` and accepts a valid one.
- verification: `go test -race ./tests/unit/missions/domain/...` passes; `gofmt` and `go vet` clean.
- next_step: T-002

### TEST-DOMAIN-OFFER-001 — Reject offer with empty season

Given an offer with `offer_id="offer-1"`, `category="BOUNTY_HUNTING"`, `capacity=10`, `mode="SOLO"`, `season_id=""`.
When `NewValidMissionOffer` is called.
Then it returns `ErrOfferInvalid`.

### TEST-DOMAIN-OFFER-002 — Accept offer with valid season

Given an offer with `season_id="season-2026-q4"`.
When `NewValidMissionOffer` is called.
Then it returns a `MissionOffer` with `SeasonID: "season-2026-q4"`.

## T-002 — Publish offer requires active season

- status: pending
- development_layer: application
- layer_gate: Gate 3-APPLICATION, then LAYER-GATE-APPLICATION
- depends_on: T-001
- test_id: TEST-APP-PUBLISH-001
- test_type: unit
- test_file: `tests/unit/missions/application/publish_mission_test.go`
- mapped_production_files:
  - `internal/missions/application/publish_mission.go` — add `ActiveSeasonsReader` port, reject no active season.
- requirements: REQ-MB-001
- done_when: `PublishMissionHandler` returns `ErrNoActiveSeason` when no active season exists and binds the active season otherwise.
- verification: `go test -race ./tests/unit/missions/application/...` passes.
- next_step: T-003

### TEST-APP-PUBLISH-001 — Reject publish without active season

Given `ListActive()` returns an empty list.
When `PublishMissionHandler.Handle` is called.
Then it returns `ErrNoActiveSeason`.

### TEST-APP-PUBLISH-002 — Publish with active season

Given `ListActive()` returns `[{SeasonID:"season-2026-q4", State:ACTIVE, Official:true}]`.
When `PublishMissionHandler.Handle` is called with a valid offer.
Then the saved offer has `SeasonID: "season-2026-q4"`.

## T-003 — Contract inherits offer season and confirms only matching season

- status: pending
- development_layer: domain
- layer_gate: LAYER-GATE-DOMAIN
- depends_on: T-001
- test_id: TEST-DOMAIN-CONFIRM-001
- test_type: unit
- test_file: `tests/unit/missions/domain/mission_contract_confirm_test.go`
- mapped_production_files:
  - `internal/missions/domain/mission_contract.go` — add `SeasonID` field, update constructors.
  - `internal/missions/domain/mission_contract_confirm.go` — add `ErrSeasonMismatch` validation.
- requirements: REQ-MCF-003, REQ-MCF-004
- done_when: contract has the offer's season and rejects a mismatched confirmation season.
- verification: `go test -race ./tests/unit/missions/domain/...` passes.
- next_step: T-004

### TEST-DOMAIN-CONFIRM-001 — Contract inherits offer season

Given an offer with `SeasonID: "season-2026-q4"`.
When `NewMissionContract` is called.
Then the contract has `SeasonID: "season-2026-q4"`.

### TEST-DOMAIN-CONFIRM-002 — Reject confirmation with mismatched season

Given a started contract bound to `season-2026-q4` and a season `season-2026-q3`.
When `contract.Confirm` is called with `season-2026-q3`.
Then it returns `ErrSeasonMismatch`.

### TEST-DOMAIN-CONFIRM-003 — Accept confirmation with matching season

Given a started contract bound to `season-2026-q4` and an active season `season-2026-q4`.
When `contract.Confirm` is called.
Then it returns a `MeritCredit` with `SeasonID: "season-2026-q4"`.

## T-004 — Persist `SeasonID` in offer and contract stores

- status: pending
- development_layer: infrastructure
- layer_gate: LAYER-GATE-APPLICATION
- depends_on: T-002, T-003
- test_id: TEST-INFRA-OFFER-001
- test_type: unit/integration
- test_file: `tests/integration/infrastructure/dynamodb/offer_store_test.go`, `tests/integration/infrastructure/dynamodb/contract_store_test.go`
- mapped_production_files:
  - `internal/missions/infrastructure/dynamodb/offer_store.go`
  - `internal/missions/infrastructure/dynamodb/contract_store.go`
- requirements: REQ-MB-001, REQ-MCF-003
- done_when: round-trip of offer and contract preserves `SeasonID`.
- verification: `go test -race ./tests/integration/infrastructure/...` passes.
- next_step: T-005

## T-005 — Update HTTP boundary and public error codes

- status: pending
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY, then LAYER-GATE-BOUNDARY
- depends_on: T-002, T-003
- test_id: TEST-HTTP-PUBLISH-001
- test_type: unit
- test_file: `tests/unit/missions/interface/http/publish_mission_test.go`, `tests/unit/missions/interface/http/handler_test.go`
- mapped_production_files:
  - `internal/missions/interface/http/publish_mission.go`
  - `internal/missions/interface/http/handler.go`
  - `internal/missions/interface/http/errors.go`
- requirements: REQ-MB-001, REQ-MCF-003
- done_when: `POST /api/v1/missions/offers` returns `NO_ACTIVE_SEASON`; `POST /api/v1/missions/contracts/{contract_id}/evidence` returns `SEASON_MISMATCH`; response DTOs include `season_id`.
- verification: `go test -race ./tests/unit/missions/interface/http/...` passes.
- next_step: T-006

## T-006 — Wire composition and run integration suite

- status: pending
- development_layer: composition
- layer_gate: LAYER-GATE-APPLICATION
- depends_on: T-004, T-005
- test_id: TEST-INTEGRATION-001
- test_type: integration/http, integration/composition
- test_file: `tests/integration/http/publish_mission_test.go`, `tests/integration/composition/publish_mission_test.go`
- mapped_production_files:
  - `internal/missions/composition/server.go`
- requirements: cross-cutting
- done_when: `go test -race ./...` passes, `gofmt` and `go vet` clean, coverage >= 90% on affected files.
- verification:
  - `go test -race ./...` passes.
  - `gofmt -w .` and `go vet ./...` clean.
  - Coverage >= 90% on `internal/missions/domain`, `internal/missions/application`, and changed infrastructure files.
- next_step: T-007

## T-007 — Converge canonical specs and documentation

- status: pending
- development_layer: documentation
- layer_gate: documentation review
- depends_on: T-006
- requirements: cross-cutting
- done_when: affected capability specs include new scenarios and error vocabulary; `docs/API-CONTRACT.md` updated.
- verification: `specs/capabilities/mission-board/spec.md`, `specs/capabilities/seasonal-category-competition/spec.md`, and `specs/capabilities/mission-contract-fulfillment/spec.md` reflect the new acceptance scenarios.
- next_step: verify, consolidate, and archive.
