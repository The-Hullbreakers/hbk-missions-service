---
change_id: CHG-0030
title: Abandon contract journey end-to-end test
status: archived
risk: L1
canonical_action: modify
change_action: compact_change
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/archive/2026/CHG-0030-abandon-journey-e2e/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
created: 2026-10-09
---

# CHG-0030 — Abandon Journey E2E

## Intent

Add the missing acceptance journey for contract abandonment. The endpoint
is covered in `tests/integration/http` with in-memory stores, but no
composition-level test proves the full flow over real DynamoDB:
accept → start → abandon → slot freed → template reappears on the board.

No production behavior changes. The change adds test evidence for two
already-consolidated scenarios.

## Scenarios Covered (existing canonical)

- `SCN-MCF-ABANDON` — Abandon releases the slot (mission-contract-fulfillment, REQ-MCF-008).
- `SCN-MB-TEMPLATE-RELEASED` — Show after terminal state (mission-board, REQ-MB-008).

## Test

- `TEST-E2E-ABANDON-JOURNEY` — `tests/integration/composition/abandon_journey_e2e_test.go`
  - Command: `go test -v -count=1 -run TestAbandonJourney ./tests/integration/composition/`
  - Journey: create season → definition → publish → offer → accept → start
    → `POST /contracts/{id}/abandon` (200) → `GET /contracts?status=ACTIVE`
    returns empty for the caller → `GET /offers` lists the same template
    again → leaderboard shows no credit from the abandoned contract.
  - Expected RED: none — production behavior already exists; the test must
    pass on first run (test-only change). If it fails, the failure reveals
    a defect routed through the bug workflow.

## Tasks

- **T-01 — Write abandon journey e2e.** Wire the composition test server
  with real DynamoDB-local; assert lifecycle state, active-slot release,
  and board re-visibility. Verification: `go test` passes; file `< 150`
  lines; gofmt/vet clean.
- **T-02 — Converge specs.** Add the test file to the "test suites"
  listings of `mission-contract-fulfillment` and `mission-board`, update
  `specs/index.md`, archive.

## Consolidation Note

- Test passed on first run — expected for a test-only change; no RED meant
  the existing production behavior already satisfied the journey.
- Suite listings are directory-level: `tests/integration/composition` was
  already listed in both canonical specs, so no canonical text changed.
  `specs/index.md` updated; artifacts archived on 2026-10-09.

## Gates

- Test-first: the test IS the deliverable — RED phase verified by running
  it; a RED failure would expose a real bug.
- BLUE: `< 150` physical lines, gofmt/vet.
- Validation: `tools/validate-sdd-change.sh`.

## Rollback

Delete `abandon_journey_e2e_test.go` and revert the spec test-suite refs.
