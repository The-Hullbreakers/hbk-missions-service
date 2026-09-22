---
change_id: CHG-0011
artifact: tasks
status: proposed
updated: 2026-08-30
---

# Tasks — Add Mission Definitions And Season Pools

## T-001 — Define BDD scenarios for definitions, pools, and manual publish

- **Owner**: agent
- **Wave**: Gate 1 → Gate 2
- **Outcome**: `delta.md` scenarios reviewed and stable.
- **Verification**: All scenarios have Given/When/Then and map to a test.
- **Next**: T-002

## T-002 — Domain RED for MissionDefinition and MissionPool

- **Owner**: agent
- **Wave**: Domain RED
- **Dependencies**: T-001
- **Outcome**: `MissionDefinition`, `MissionPool`, `OfferPublication`, and updated `MissionOffer` constructors fail the first RED tests.
- **Verification**: `go test ./internal/missions/domain` shows expected RED.
- **Next**: T-003

## T-003 — Domain GREEN for definitions and pool publication

- **Owner**: agent
- **Wave**: Domain GREEN
- **Dependencies**: T-002
- **Outcome**: Domain constructors and `OfferPublication` make the domain tests pass.
- **Verification**: `go test ./internal/missions/domain` green.
- **Next**: T-004

## T-004 — Domain BLUE refactor

- **Owner**: agent
- **Wave**: Domain BLUE
- **Dependencies**: T-003
- **Outcome**: Small value objects, < 150 lines per file, no framework leaks.
- **Verification**: `gofmt`, `go vet`, and line-count check pass.
- **Next**: T-005

## T-005 — Application RED for definition/pool handlers

- **Owner**: agent
- **Wave**: Application RED
- **Dependencies**: T-004
- **Outcome**: `CreateMissionDefinitionHandler`, `PublishMissionDefinitionHandler`, `CreateMissionPoolHandler`, and `PublishMissionFromPoolHandler` fail their unit tests.
- **Verification**: `go test ./internal/missions/application` RED.
- **Next**: T-006

## T-006 — Application GREEN

- **Owner**: agent
- **Wave**: Application GREEN
- **Dependencies**: T-005
- **Outcome**: Handlers pass unit tests using hand-written port doubles.
- **Verification**: `go test ./internal/missions/application` green.
- **Next**: T-007

## T-007 — Boundary and composition

- **Owner**: agent
- **Wave**: Boundary/COMPOSITION
- **Dependencies**: T-006
- **Outcome**: HTTP routes and composition wire real handlers and stores.
- **Verification**: Integration tests `tests/integration/composition` and `tests/integration/http` green.
- **Next**: T-008

## T-008 — Update AcceptMission active-template exclusion

- **Owner**: agent
- **Wave**: Application/Domain
- **Dependencies**: T-006
- **Outcome**: `DecideAcceptance` rejects an offer when the player has an active contract with the same `definition_id`.
- **Verification**: Unit and integration tests green.
- **Next**: T-009

## T-009 — Infrastructure persistence

- **Owner**: agent
- **Wave**: Infrastructure
- **Dependencies**: T-007
- **Outcome**: `DefinitionStore`, `PoolStore`, and updated `OfferStore`/`ContractStore` persist and load all fields.
- **Verification**: `tests/integration/infrastructure/dynamodb` green.
- **Next**: T-010

## T-010 — Coverage, mutation, and critical E2E

- **Owner**: agent
- **Wave**: Verification
- **Dependencies**: T-009
- **Outcome**: Coverage >= 90% on affected scope, mutation and E2E evidence recorded.
- **Verification**: `go test -race ./...`, coverage reports, and critical E2E scenarios pass.
- **Next**: T-011

## T-011 — Documentation and rollout

- **Owner**: agent
- **Wave**: Converge
- **Dependencies**: T-010
- **Outcome**: `docs/API-CONTRACT.md` and capability specs updated; rollout steps recorded.
- **Verification**: `verification.md` and `rollout.md` complete.
