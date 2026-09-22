---
change_id: CHG-0009
title: Refactor domain and application layers with value objects and SOLID
status: completed
risk: L2
canonical_action: modify
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/archive/2026/CHG-0009-refactor-domain-application-value-objects/
archive_path: specs/changes/archive/2026/CHG-0009-refactor-domain-application-value-objects/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_REFACTOR_CODE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_REFACTOR_LIFECYCLE_WORKFLOW
  - WORKFLOW-COMMON_SDD_UPDATE_DOCUMENTATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CODE_QUALITY_GATE_WORKFLOW
context_scope: existing
created: 2026-08-30
updated: 2026-08-30
---

# Refactor Domain And Application Layers With Value Objects And SOLID

## Intent

Improve the domain and application layers by introducing meaningful value objects, removing primitive obsession, clarifying single responsibilities, and aligning every handler with SOLID/Clean Architecture without changing observable behavior.

## Business Value

- Reduce the risk of invalid identifiers and inconsistent string concatenation (e.g. `ConfirmationID`).
- Make the domain model easier to extend with new rules and new value objects.
- Keep handlers, ports, and adapters small and focused, each with a single reason to change.
- Preserve the existing public HTTP contract and all existing acceptance tests.

## Why L2

The change is a behavior-preserving refactor across domain and application. It touches multiple capabilities but does not introduce new business rules or public endpoints. A mistake could break existing tests or public error codes, but the contract, persistence schema, and auth boundary remain stable.

## Scope

- `internal/missions/domain` — value objects, entity methods, and policy types.
- `internal/missions/application` — use-case handlers and ports.
- `tests/unit/missions/domain` — new and updated characterization tests.
- `tests/unit/missions/application` — new and updated characterization tests.
- `docs/ARCHITECTURE.md` and affected capability notes if structure or ownership changes.

## Non-Goals

- No change to HTTP routes, DTOs, or status codes.
- No change to DynamoDB table schema or key design.
- No new use cases or business rules.
- No change to the invariants in `mission-board`, `mission-contract-fulfillment`, or `seasonal-category-competition`.

## Acceptance Summary

- All existing unit and integration tests pass unchanged in observable behavior.
- New value objects have explicit unit tests and valid/invalid construction paths.
- Application handlers remain thin and delegate domain decisions to the domain layer.
- Every affected maintained source, test, and spec file stays under 150 physical lines.
- `go test -race ./...`, `gofmt -w .`, and `go vet ./...` are clean.
- Domain and application coverage on affected files remains >= 90%.

## Planned Value Objects

- `ConfirmationID` — centralizes `contractID:objectiveID:unitID:recipient` identity and validation.
- `ActiveContractSummary` — replaces the generic `Contract` type used in acceptance decisions.
- Validating constructors for `MissionOffer`, `Season`, and `MissionContract` where the application currently performs manual checks.

## Verification

- `go test -race ./...` passed.
- `gofmt -w .` and `go vet ./...` clean.
- `find internal/missions tests -name '*.go' | xargs wc -l | sort -n | tail -25` shows every affected file under 150 physical lines.
- Domain unit coverage: 96.5% (`go test -race -coverpkg=./internal/missions/domain -coverprofile=coverage-domain.out ./tests/unit/missions/domain`).
- Application unit coverage: 96.9% (`go test -race -coverpkg=./internal/missions/application -coverprofile=coverage-app.out ./tests/unit/missions/application`).

## SOLID Targets

- **SRP** — one concept per domain file; one use case per application file; split `freeze` and `finalize` into separate files.
- **OCP** — extend confirmation/acceptance rules by adding domain policy/value objects, not by editing handler logic.
- **LSP** — `ActiveContractSummary` expresses the real abstraction consumed by acceptance policy.
- **ISP** — keep each application port a single command or query operation.
- **DIP** — handlers depend on small domain-driven ports; adapters implement those ports.
