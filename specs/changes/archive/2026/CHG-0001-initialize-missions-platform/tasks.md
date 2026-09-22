---
change_id: CHG-0001
artifact: test-first-tasks
status: completed
updated: 2026-08-30
---

# Test-First Tasks

Production remains blocked until Gate 2 and the corresponding Gate 3. Concrete paths will be confirmed after the Go bootstrap, without creating hypothetical folders.

## T-001 — Model pool, eligibility and active limit

- status: completed
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: Gate 2 (approved)
- test_id: TEST-MB-001
- test_type: unit
- test_file: `tests/unit/missions/domain/active_limit_test.go`
- harness_files: `go.mod`, `tests/unit/missions/domain/active_limit_test.go`
- mapped_production_files:
  - `internal/missions/domain/active_limit_policy.go` — `ActiveLimitPolicy`.
  - `internal/missions/domain/contract.go` — `ContractState` and `Contract`.
  - `internal/missions/domain/mission_offer.go` — `MissionOffer`.
  - `internal/missions/domain/acceptance.go` — `AcceptanceDecision`, `DecideAcceptance`, `ReasonActiveMissionLimitReached`, `ReasonOfferCapacityExhausted`.
- requirements: REQ-MB-001, REQ-MB-003, REQ-MB-004, REQ-MB-005, REQ-MB-006, REQ-MB-007, REQ-MB-008
- done_when: limit, rotation, active template exclusion, non-destructive reduction and last concurrent capacity are specified and green.

### TEST-MB-001 — Active mission limit

- Given a player already has three active contracts (`ACCEPTED`, `IN_PROGRESS`, `PENDING_VALIDATION`) and the default active limit is three.
- When the player attempts to accept another `BOUNTY_HUNTING` offer.
- Then the decision is rejected with reason `ACTIVE_MISSION_LIMIT_REACHED`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run ActiveMissionLimit`
- Output: `no required module provides package github.com/hbk/hbk-missions-service/internal/missions/domain` (package missing).
- Evidence: `Gate 3-DOMAIN` approved by user on 2026-08-26.

#### GREEN

- Production file created: `internal/missions/domain/active_limit.go`.
- Command: `go test ./tests/unit/missions/domain/... -run ActiveMissionLimit`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- Introduced `ReasonActiveMissionLimitReached` domain constant to replace the raw string in both production and tests.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./tests/unit/missions/domain/... -run ActiveMissionLimit`; all clean.
- Files under 150 physical lines: `active_limit.go` (95), `active_limit_test.go` (45).
- `LAYER-GATE-DOMAIN` passed.

### TEST-MB-001 — Offer capacity exhaustion

- Given an offer with capacity one has already been accepted once.
- When another acceptance is attempted.
- Then the decision is rejected with reason `OFFER_CAPACITY_EXHAUSTED`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run ActiveMissionLimit`
- Output: `active_limit_test.go:59: expected rejection when offer capacity 1 is exhausted`
- Evidence: `Gate 3-DOMAIN` approved implicitly by user instruction to refactor and continue.

#### GREEN

- Updated `MissionOffer` and `Contract` with `OfferID` so capacity is counted per offer.
- Added `ReasonOfferCapacityExhausted` constant and updated `DecideAcceptance` to reject when `offer.Capacity > 0 && offerActiveCount >= offer.Capacity`.
- Updated test mothers to set `OfferID`.
- Command: `go test -race ./tests/unit/missions/domain/... -run ActiveMissionLimit`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- Refactored at the user's request: split `internal/missions/domain/active_limit.go` into `active_limit_policy.go`, `contract.go`, `mission_offer.go` and `acceptance.go`.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./tests/unit/missions/domain/... -run ActiveMissionLimit`; all clean.
- Files under 150 physical lines: `active_limit_policy.go` (35), `contract.go` (23), `mission_offer.go` (18), `acceptance.go` (41), `active_limit_test.go` (67).
- `LAYER-GATE-DOMAIN` passed for offer capacity.

## T-002 — Model contracts, teams and objectives

- status: in_progress
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: T-001
- test_id: TEST-MCF-001
- test_type: unit
- test_file: `tests/unit/missions/domain/mission_contract_test.go`
- red_command: `go test ./tests/unit/missions/domain/... -run MissionContract`
- expected_red: There is no MissionContract, Start method or RosterFrozen.
- red_output: `undefined: domain.NewMissionContract`
- mapped_production_files:
  - `internal/missions/domain/mission_contract.go` — `MissionContract`, `Start`, `RosterFrozen`, `Confirmations`.
  - `internal/missions/domain/confirmation.go` — `ConfirmObjective`.
  - `internal/missions/domain/objective.go` — `Objective`.
  - `internal/missions/domain/errors.go` — sentinel errors.
- requirements: REQ-MCF-001, REQ-MCF-002, REQ-MCF-003, REQ-MCF-004, REQ-MCF-005
- done_when: transitions and solo/team rules protect their invariants.
- outcome: `TEST-MCF-001` to `TEST-MCF-006` green. `MissionContract` supports `TeamID`, `NewTeamMissionContract`, defensive `Roster()`, `Abandon()` and `IsTerminal()`. `ConfirmObjective` distinguishes `ErrContractTerminal`/`CONTRACT_TERMINAL`, uses `TrackTeam` for teams, sums `Bonus`, and rejects confirmations that exceed `UnitLimit` (`OBJECTIVE_LIMIT_REACHED`). `Objective` has `Bonus` and `UnitLimit`; `MissionOffer` has `Mode` (`SOLO`/`TEAM`/`FLEX`) and is persisted to DynamoDB. Phase and multi-objective rules are still missing and use case `AcceptMission`/`PublishOffer`.
- verification: `go test -race ./...` passes; `gofmt` and `go vet` clean.
- next_step: Continue with T-004 (`AcceptMission`/`PublishOffer`) or phase/multi-target rules.

### TEST-MCF-001 — Start a contract and freeze the team roster

- Given a `MissionContract` accepted for a team with roster `["player-1", "player-2"]`.
- When `Start()` is called.
- Then the contract state is `IN_PROGRESS` and `RosterFrozen()` is `true`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run MissionContract`
- Output: `tests/unit/missions/domain/mission_contract_test.go:12:21: undefined: domain.NewMissionContract`
- Evidence: `Gate 3-DOMAIN` approved by user on 2026-08-26.

#### GREEN

- Production file created: `internal/missions/domain/mission_contract.go`.
- Command: `go test ./tests/unit/missions/domain/... -run MissionContract`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- `NewMissionContract` defensively copies the roster slice.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./tests/unit/missions/domain/... -run MissionContract`; all clean.
- Files under 150 physical lines: `mission_contract.go` (33), `mission_contract_test.go` (27).
- `LAYER-GATE-DOMAIN` passed for TEST-MCF-001.

### TEST-MCF-002 — Team contract credits the team track

- Given a `MissionContract` accepted for a team with roster `["player-1", "player-2"]` and `TeamID` `team-1`.
- When `Start()` and `ConfirmObjective(..., "target-1", ...)` are called.
- Then the credit has `Track` `TEAM` and `RecipientID` `team-1`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run TestConfirmObjective_TeamContractCreditsTeamTrack`
- Output: `undefined: domain.NewTeamMissionContract`
- Evidence: `Gate 3-DOMAIN` approved to continue T-002 on 2026-08-30.

#### GREEN

- Production changes: added `TeamID` to `MissionContract`, `NewTeamMissionContract`, updated `NewMissionContractFromState` and `ConfirmObjective` to select `TrackTeam` when `TeamID` is present; persisted `TeamID` in `dynamodb/contract_store.go`.
- Command: `go test ./tests/unit/missions/domain/... -run TestConfirmObjective_TeamContractCreditsTeamTrack`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- `ConfirmObjective` uses `TrackIndividual` and the first roster member for solo contracts, and `TrackTeam` with `TeamID` for team contracts.
- Encapsulated `MissionContract.roster` and exposed `Roster()` as a defensive copy; call sites in `confirmation.go`, `contract_store.go` and `contract_store_test.go` use `Roster()`.
- Added `TestMissionContract_RosterIsDefensivelyCopied` and `TestMissionContract_RosterCopyDoesNotMutateInternalRoster`.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `mission_contract.go` (73), `confirmation.go` (47), `mission_contract_test.go` (88).
- `LAYER-GATE-DOMAIN` passed for TEST-MCF-002.

### TEST-MCF-003 — Reject confirmation on terminal contract

- Given a `MissionContract` in `IN_PROGRESS` is abandoned.
- When `ConfirmObjective(...)` is called.
- Then it returns `ErrContractTerminal` and the public code is `CONTRACT_TERMINAL`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run TestConfirmObjective_RejectsTerminalContract`
- Output: `merit_credit_test.go:101: expected terminal contract to be rejected` (confirm succeeds because terminal check missing).
- Evidence: `Gate 3-DOMAIN` approved to continue T-002 terminal-state work.

#### GREEN

- Production changes: added terminal `ContractState` constants, `IsTerminal()`, `MissionContract.Abandon()`, `ErrContractTerminal` and public code `CONTRACT_TERMINAL`; `ConfirmObjective` returns `ErrContractTerminal` before other state checks.
- Command: `go test ./tests/unit/missions/domain/... -run TestConfirmObjective_RejectsTerminalContract`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- Terminal states grouped in `IsTerminal()` instead of scattered checks.
- `errors.Is` assertion in test; `errors_test.go` and `interface/http/errors.go` updated with `CONTRACT_TERMINAL`.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `contract.go` (34), `mission_contract.go` (80), `confirmation.go` (50), `errors.go` (14), `merit_credit_test.go` (43), `confirm_objective_test.go` (128).
- `LAYER-GATE-DOMAIN` passed for TEST-MCF-003.

### TEST-MCF-004 — Objective bonus is added to each credit

- Given an `Objective` with `MeritPerUnit` 100 and `Bonus` 25.
- When `ConfirmObjective(...)` is called for one unit.
- Then the `MeritCredit.Amount` is 125.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run TestConfirmObjective_IncludesBonus`
- Output: `expected credit amount 125, got 100` (`ConfirmObjective` ignores `Bonus`).
- Evidence: `Gate 3-DOMAIN` approved to continue T-002 objective work.

#### GREEN

- Production changes: added `Bonus` to `Objective`, `NewObjectiveWithBonus`, and `ConfirmObjective` sets `Amount` to `MeritPerUnit + Bonus`.
- Command: `go test ./tests/unit/missions/domain/... -run TestConfirmObjective_IncludesBonus`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- `NewObjective` delegates to `NewObjectiveWithBonus` with zero bonus; no existing call sites needed updating.
- `merit_credit_test.go` exceeded 150 lines, so `ConfirmObjective` tests were split into `confirm_objective_test.go`; `merit_credit_test.go` now contains only the four-bounty total test.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `objective.go` (24), `confirmation.go` (50), `merit_credit_test.go` (43), `confirm_objective_test.go` (128), `mission_contract_test.go` (106).
- `LAYER-GATE-DOMAIN` passed for TEST-MCF-004.

### TEST-MCF-005 — Objective unit limit is enforced

- Given an `Objective` with `UnitLimit` 2.
- When `ConfirmObjective(...)` is called three times for different units.
- Then the third call returns `ErrObjectiveLimitReached` and the public code is `OBJECTIVE_LIMIT_REACHED`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run TestObjective_UnitLimitIsEnforced`
- Output: `expected third confirmation to exceed unit limit` (no limit check).
- Evidence: `Gate 3-DOMAIN` approved to continue T-002 objective work.

#### GREEN

- Production changes: added `UnitLimit` to `Objective`, `NewObjectiveWithBonusAndLimit`, `ErrObjectiveLimitReached` and public code `OBJECTIVE_LIMIT_REACHED`; `ConfirmObjective` counts prior confirmations per objective and rejects when the limit is reached.
- Command: `go test ./tests/unit/missions/domain/... -run TestObjective_UnitLimitIsEnforced`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- `confirm_objective_test.go` exceeded 150 lines; bonus and unit-limit tests moved to `objective_test.go`.
- `NewObjective` and `NewObjectiveWithBonus` delegate to `NewObjectiveWithBonusAndLimit` to avoid constructor churn.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `objective.go` (31), `confirmation.go` (67), `errors.go` (15), `merit_credit_test.go` (43), `confirm_objective_test.go` (108), `objective_test.go` (56), `mission_contract_test.go` (106).
- `LAYER-GATE-DOMAIN` passed for TEST-MCF-005.

### TEST-MCF-006 — MissionOffer carries mode

- Given a call to `NewMissionOffer` and `NewTeamMissionOffer`.
- When the offers are created.
- Then the first has `Mode` `SOLO` and the second `TEAM`.

#### RED

- Command: `go test ./tests/unit/missions/domain/... -run TestMissionOffer`
- Output: `undefined: domain.MissionMode`/`undefined: domain.NewTeamMissionOffer`.
- Evidence: `Gate 3-DOMAIN` approved to continue T-002 offer work.

#### GREEN

- Production changes: added `MissionMode` type with `SOLO`, `TEAM`, `FLEX` constants; added `Mode` to `MissionOffer`; `NewMissionOffer` defaults to `SOLO`; added `NewTeamMissionOffer` and `NewMissionOfferWithMode`; `dynamodb/contract_store.go` persists and loads `Offer.Mode`.
- Command: `go test ./tests/unit/missions/domain/... -run TestMissionOffer`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/domain`.

#### BLUE

- Constructors delegate to `NewMissionOfferWithMode` to avoid duplication.
- `dynamodb/contract_store.go` stores `Mode` as a string and round-trips it.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `mission_offer.go` (38), `mission_offer_test.go` (21), `contract_store.go` (91).
- `LAYER-GATE-DOMAIN` passed for TEST-MCF-006.

## T-003 — Model seasons and merit ledger

- status: planned
- development_layer: domain
- layer_gate: Gate 3-DOMAIN, then LAYER-GATE-DOMAIN
- depends_on: T-002
- test_id: TEST-SCC-001
- test_type: unit
- red_command: `go test ./tests/unit/missions/domain/... -run SeasonMerit`
- expected_red: there are no score buckets, credits, adjustments or completion.
- requirements: REQ-SCC-001, REQ-SCC-002, REQ-SCC-004, REQ-SCC-006
- done_when: separation by category/track, idempotency and freeze are green.

## T-004 — Ormissionar casos de uso

- status: in_progress
- development_layer: application
- layer_gate: Gate 3-APPLICATION, then LAYER-GATE-APPLICATION
- depends_on: LAYER-GATE-DOMAIN
- test_id: TEST-APP-001
- test_type: unit
- red_command: `go test ./tests/unit/missions/application/...`
- expected_red: there are no cases of publication, acceptance, confirmation, ranking or closure.
- requirements: all
- done_when: focused CQRS ports express each decision without external dependencies.
- outcome: `TEST-APP-001` (ConfirmObjective), `TEST-APP-002` (AcceptMission) and `TEST-APP-003` (PublishMission) green. `AcceptMissionHandler` reads the offer, validates active limit with `DecideAcceptance`, distinguishes SOLO/TEAM/FLEX, creates and persists `MissionContract`. `PublishMissionHandler` validates and persists `MissionOffer` with `Mode` (`SOLO`/`TEAM`/`FLEX`). Ranking, closure, and publish/accept endpoints HTTP are missing.
- verification: `go test -race ./tests/unit/missions/application/...` passes; `gofmt` and `go vet` clean.
- next_step: Implement `OfferStore` to DynamoDB and endpoint HTTP `POST /api/v1/missions/offers` to publish, then endpoint to accept.

### TEST-APP-001 — Confirm objective through the application handler

- Already green from prior work; see previous verification.

### TEST-APP-002 — Accept a mission offer for a player or team

- Given a solo offer with capacity and a player with no active contracts.
- When `AcceptMissionHandler.Handle(...)` is called.
- Then only `MissionContract` is created, saved and returned.

#### RED

- Command: `go test ./tests/unit/missions/application/... -run TestAcceptMissionHandler`
- Output: `undefined: application.NewAcceptMissionHandler`.
- Evidence: `Gate 3-APPLICATION` approved to continue T-004.

#### GREEN

- Production changes: added `internal/missions/application/accept_mission.go` with `AcceptMissionHandler`, `OfferReader`, `ActiveContractsReader` ports; `Handle` uses `DecideAcceptance`, creates `MissionContract`/`NewTeamMissionContract` and saves through `ContractWriter`.
- Added `ErrTeamIDRequired`, `ErrActiveMissionLimitReached` and `ErrOfferCapacityExhausted` with public codes `TEAM_ID_REQUIRED`, `ACTIVE_MISSION_LIMIT_REACHED` and `OFFER_CAPACITY_EXHAUSTED`.
- Command: `go test ./tests/unit/missions/application/... -run TestAcceptMissionHandler`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/application`.

#### BLUE

- `AcceptMissionHandler` depends only on application ports; no framework or persistence references.
- `DecideAcceptance` is reused; domain errors are returned from the handler.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `accept_mission.go` (86), `accept_mission_test.go` (107).
- `LAYER-GATE-APPLICATION` passed for TEST-APP-002.

### TEST-APP-003 — Publish a mission offer with mode

- Given a publisher provides `offerID`, `category`, `capacity` and `mode`.
- When `PublishMissionHandler.Handle(...)` is called.
- Then the `MissionOffer` is persisted and returned.

#### RED

- Command: `go test ./tests/unit/missions/application/... -run TestPublishMissionHandler`
- Output: `undefined: application.NewPublishMissionHandler`.
- Evidence: `Gate 3-APPLICATION` approved to continue T-004.

#### GREEN

- Production changes: added `internal/missions/application/publish_mission.go` with `PublishMissionHandler` and `OfferWriter` port; validates inputs, builds `MissionOffer` with `NewMissionOfferWithMode` and saves it.
- Added `ErrOfferInvalid` with public code `INVALID_MISSION_OFFER`.
- Command: `go test ./tests/unit/missions/application/... -run TestPublishMissionHandler`
- Output: `ok github.com/hbk/hbk-missions-service/tests/unit/missions/application`.

#### BLUE

- `PublishMissionHandler` depends only on `OfferWriter`; domain handles offer construction.
- Validation is in the application port; no persistence references.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `publish_mission.go` (45), `publish_mission_test.go` (67).
- `LAYER-GATE-APPLICATION` passed for TEST-APP-003.

## T-005 — Test the public limit and competitive journey

- status: in_progress
- development_layer: boundary
- layer_gate: Gate 3-BOUNDARY, then LAYER-GATE-BOUNDARY-RED
- depends_on: LAYER-GATE-APPLICATION
- test_id: TEST-HTTP-001, TEST-HTTP-002, TEST-HTTP-003, TEST-E2E-001
- test_type: integration/http and critical E2E
- red_command: `go test -tags=integration ./tests/integration/http/...`
- expected_red: the service does not yet expose its real boundary.
- requirements: REQ-MB-002, REQ-MB-003, REQ-MB-006, REQ-MB-007, REQ-MB-008, REQ-MCF-003, REQ-SCC-003, REQ-SCC-005
- done_when: RED test real composition, isolation, limit, custom board, credit and closure.
- outcome: `TEST-HTTP-001` (confirm evidence), `TEST-HTTP-002` (publish offer) and `TEST-HTTP-003` (accept offer) green. `OfferStore` and `ContractStore` implement the application ports, `POST /api/v1/missions/offers` publishes and `POST /api/v1/missions/offers/{offer_id}/accept` creates SOLO/TEAM contracts. Ranking and closure are missing.
- verification: `go test -tags=integration -race ./tests/integration/...` passes (DynamoDB local); `gofmt` and `go vet` clean.
- next_step: Ranking of merits by season/category and season closing.

### TEST-HTTP-001 — Confirm objective evidence through HTTP

- Already green from prior work; see previous verification.

### TEST-HTTP-002 — Publish a mission offer through HTTP

- Given a publisher POSTs an offer with `offer_id`, `category`, `capacity` and `mode` to `POST /api/v1/missions/offers`.
- When the request is authenticated.
- Then the offer is persisted and returned with 201.

#### RED

- Command: `go test -tags=integration ./tests/integration/composition/... -run TestCompositionServer_PublishMissionOffer`
- Output: `undefined: missionhttp.PublishMissionHandler` / `undefined: missiondynamo.NewOfferStore`.
- Evidence: `Gate 3-BOUNDARY` approved to continue T-005.

#### GREEN

- Production changes: added `internal/missions/infrastructure/dynamodb/offer_store.go` implementing `OfferWriter` and `OfferReader`; added `internal/missions/interface/http/publish_mission.go` with `PublishMissionHandler`; updated `internal/missions/composition/server.go` to wire `OfferStore` and `PublishMissionHandler` and expose `POST /api/v1/missions/offers`.
- Command: `go test -tags=integration ./tests/integration/composition/... -run TestCompositionServer_PublishMissionOffer`
- Output: `ok github.com/hbk/hbk-missions-service/tests/integration/composition`.

#### BLUE

- `OfferStore` follows the `PK`/`SK` key pattern of the single-table.
- HTTP handler decodes JSON, uses `PublicCode` for domain errors and returns 201.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `offer_store.go` (60), `publish_mission.go` (interface) (48), `publish_mission_test.go` (composition) (53), `composition/server.go` (68).
- `LAYER-GATE-BOUNDARY-RED` passed for TEST-HTTP-002.

### TEST-HTTP-003 — Accept a mission offer through HTTP

- Given an offer exists and a player POSTs `player_id` and optional `team_id` to `POST /api/v1/missions/offers/{offer_id}/accept`.
- When the request is authenticated.
- Then a contract is created and returned with 201.

#### RED

- Command: `go test -tags=integration ./tests/integration/composition/... -run TestCompositionServer_AcceptMissionOffer`
- Output: `undefined: missionhttp.AcceptMissionHandler` / `undefined: ContractStore.ListActiveByPlayer`.
- Evidence: `Gate 3-BOUNDARY` approved to continue T-005.

#### GREEN

- Production changes: added `ListActiveByPlayer` to `ContractStore`; added `internal/missions/interface/http/accept_mission.go` with `AcceptMissionHandler`; updated `internal/missions/composition/server.go` to wire `AcceptMissionHandler` and expose `POST /api/v1/missions/offers/{offer_id}/accept`.
- Command: `go test -tags=integration ./tests/integration/composition/... -run TestCompositionServer_AcceptMissionOffer`
- Output: `ok github.com/hbk/hbk-missions-service/tests/integration/composition`.

#### BLUE

- `ContractStore.ListActiveByPlayer` uses `Query` for `PK=CONTRACT` and filters by roster/status/category; scans the same `PK` without adding an extra GSI in v1.
- HTTP handler reuses `PublicCode` and delegates to the application handler.
- Ran `gofmt -w .`, `go vet ./...` and `go test -race ./...`; all clean.
- Files under 150 physical lines: `contract_store.go` (111), `accept_mission.go` (interface) (45), `accept_mission_test.go` (composition) (66), `composition/server.go` (75).
- `LAYER-GATE-BOUNDARY-RED` passed for TEST-HTTP-003.

## T-006 — Implementar adapters, interfaces y DI

- status: planned
- development_layer: infrastructure/interface/composition
- layer_gate: LAYER-GATE-INFRASTRUCTURE, INTERFACE and COMPOSITION
- depends_on: LAYER-GATE-BOUNDARY-RED
- test_id: TEST-INFRA-001, TEST-HTTP-001
- test_type: integration/infrastructure and integration/http
- verification_command: `go test -tags=integration ./tests/integration/...`
- requirements: all
- done_when: limits, ledger, rankings and events work through real composition.

## T-007 — Converge documentation and L3 gates

- status: planned
- development_layer: verification
- layer_gate: Final validation review
- depends_on: T-006
- test_id: TEST-SEC-001, TEST-MUT-001, TEST-E2E-001
- test_type: security, mutation, coverage, E2E and documentation
- requirements: all
- done_when: L3 evidence passes, Independent Reviewer approves and the delta is ready for consolidation.
