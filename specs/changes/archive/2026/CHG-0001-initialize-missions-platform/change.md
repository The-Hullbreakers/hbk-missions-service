---
change_id: CHG-0001
title: Initialize the HBK mission board and seasonal competition
status: consolidated
risk: L3
canonical_action: create
change_action: create_delta
archive_path: specs/changes/archive/2026/CHG-0001-initialize-missions-platform/
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_UPDATE_DOCUMENTATION_WORKFLOW
context_scope: initialize
created: 2026-08-26
updated: 2026-08-30
verified:
  - go test -race ./...
  - gofmt -w .
  - go vet ./...
---

# Initialize The HBK Mission Board And Seasonal Competition

## Intent

Create a Star Citizen-inspired contract board where members take on missions alone or as a team, complete confirmable objectives, and earn Season Merits in the corresponding category. Upon completion, Missions determines the ranking and communicates the awarded positions to Achievements.

## Business Value

- Convert organizational activities into visible and measurable contracts.
- Provide competitive specialization by discipline, without an unfair overall score.
- Allow individual and team competition with published rules.
- Preserve seasons and historical recognitions.

## User Stories

- **US-MB-001**: As a member, I want to discover and accept an eligible mission, to participate in without exceeding my active capacity.
- **US-MCF-001**: As a solo or team participant, I want to complete verifiable objectives, to obtain fair merits in my discipline.
- **US-SCC-001**: As a competitor, I want to consult my standing by category and track, to compare my performance during the season.
- **US-SCC-002**: As a winner, I want my final position to become a permanent achievement, to keep in my profile after the reset.

## Why L3

The change introduces multi-tenant isolation, administrative authorization, concurrent acceptance with limits, PvP action adjudication, merit ledger, fixes, public rankings, and trusted integration with Achievements. A defect can leak data, duplicate scores or award the wrong winner.

## Scope

- Pool of available, accepted and historical offers.
- Curated pools that generate manual, scheduled or rotation offers.
- Competitive category, type and subtype of mission.
- Configurable limit of active contracts per player.
- Exclusion of offers whose template is already active for the player.
- Contracts `SOLO`, `TEAM` and `FLEX`.
- Phases, objectives, confirmations, evidence and adjudication.
- Merits by objective and bonuses defined when publishing.
- Individual and team leaderboards by category.
- Seasons, freeze, disputes, completion and history.
- Idempotent communication of awarded positions to Achievements.

## Non-Goals

- Permanent XP, levels or global progression.
- Add categories in a global ranking in v1.
- Own permanent parties, profiles or achievements.
- Transfer items, money or rewards to Inventory.
- Direct synchronization with the Star Citizen runtime.
- Procedural generation of contracts in v1.

## Acceptance Summary

- A player cannot exceed his effective limit of active contracts, even with concurrent acceptances.
- A player does not see and cannot accept another offer from a template that he already has active.
- Each valid objective credits the published merits only once.
- Merits are registered exclusively in the correct season, category and mode.
- Team contracts score the team and freeze its roster at start.
- The closure preserves the final table and a new season starts at zero.
- Achievements receives the awarded positions without being a synchronous dependency of the closure.

## Current Outcome

Gate 1 approved on 2026-08-26. That day it was also approved to expand Mission Board with curated pools, configurable rotation and exclusion by active template. The Gate 1 documentation convergence was recorded at `verification.md` and `docs/INDEX.md`. Gate 2 approved on 2026-08-26. `go.mod` and test `TEST-MB-001` were created. Both the active limit and offer capacity of `TEST-MB-001` are green. A user request, package `internal/missions/domain` was split into `active_limit_policy.go`, `contract.go`, `mission_offer.go` and `acceptance.go`. The first RED/GREEN/BLUE of `T-002` (`TEST-MCF-001`: start contract and freeze team roster) is complete and `LAYER-GATE-DOMAIN` passed.

Since then, progress has been made in T-002–T-006: `MissionContract` stores confirmations, `ConfirmObjective` validates active contract, active season, category, non-empty roster, duplicates, terminal status and unit limit; `MissionContract` supports defensive `TeamID` and `Roster()`, `NewTeamMissionContract`, `Abandon()` and `IsTerminal()`; `ConfirmObjective` uses `TrackTeam` for teams, returns `ErrContractTerminal`/`CONTRACT_TERMINAL` for terminal contracts and `ErrObjectiveLimitReached`/`OBJECTIVE_LIMIT_REACHED` when `UnitLimit` is exceeded; `Objective` supports `Bonus` and `UnitLimit`; `MissionOffer` supports `Mode` (`SOLO`/`TEAM`/`FLEX`) with persistence; `AcceptMissionHandler` reads offers, validates active limit with `DecideAcceptance` and creates contracts SOLO/TEAM; `PublishMissionHandler` publishes offers with mode and validates entries; `OfferStore` persists offers and `ContractStore.ListActiveByPlayer` supports acceptance in DynamoDB; `ConfirmEvidenceHandler`, `PublishMissionHandler`, `AcceptMissionHandler`, `AuthMiddleware` and `PublicCode` expose `POST /api/v1/missions/contracts/{contract_id}/evidence`, `POST /api/v1/missions/offers` and `POST /api/v1/missions/offers/{offer_id}/accept`; `MeritCredit`, `MeritLedger`, `Standing` and `Season` model the ledger; DynamoDB stores, JWT session and composition are implemented and green in unit and integration. T-006 is still pending; ranking and closure are missing.
