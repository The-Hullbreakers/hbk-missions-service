---
change_id: CHG-0047
artifact: tasks
status: implemented
updated: 2026-10-20
---

# Tasks — CHG-0047

> Projection-only change: no Application/Domain *behavior* partition, so
> the inside-out order collapses to boundary serialization RED → GREEN →
> BLUE. Each RED still requires evidence review before its GREEN.

## Wave 1 — Boundary RED

- [x] **T-1 — Add failing list-contracts unit test for merit_summary.**
  Owner: agent. Deps: none. Outcome: `GET /api/v1/missions/contracts`
  response shows populated `Offer.Objectives` and
  `Offer.MeritSummary{objective_merits: 500, total_merits: 500}` for
  terms paying 450×1 + 50×1; real-offer objectives take precedence when
  present. Test:
  `tests/unit/missions/interface/http/list_contracts_merit_test.go`.
  Verify: `go test ./tests/unit/missions/interface/http -run Merit` fails
  for the right reason. Maps to
  `internal/missions/domain/mission_contract.go`. Next: T-2.

- [x] **T-2 — Add failing list-contracts http integration test.** Owner:
  agent. Deps: none. Outcome: same assertion through the wired route with
  an in-memory reader (matches existing `contract_*_test.go` doubles).
  Test: `tests/integration/http/contract_merit_test.go`. Verify:
  `go test ./tests/integration/http -run Merit` fails for the right
  reason. Next: evidence review → T-3.

## Wave 2 — GREEN

- [x] **T-3 — Add MissionContract.MarshalJSON hydrating offer objectives
  from frozen terms.** Implemented in
  `internal/missions/domain/mission_contract_projection.go`. Owner: agent. Deps: T-1, T-2 +
  `common-sdd-review-test-evidence` decision. Outcome: minimum GREEN.
  Verify: both RED tests pass; `go test ./...` green. Next: T-4.

## Wave 3 — BLUE and gates

- [x] **T-4 — BLUE refactor while green.** Owner: agent. Deps: T-3.
  Outcome: roster/membership accessors extracted to
  `mission_contract_membership.go` (165→117 lines); suites green.
- [x] **T-5 — Update public docs.** Owner: agent. Deps: T-3. Outcome:
  `docs/API-CONTRACT.md` documents `GET /contracts` filters + embedded
  `Offer.Objectives`/`MeritSummary` derived from frozen `Terms`.
- [x] **T-6 — Run validate script and verify.** Owner: agent. Deps: T-4,
  T-5. Outcome: `tools/validate-sdd-change.sh` → "SDD validation
  passed" (unit+integration+race+coverage 90.8–92.7%); verify approved
  2026-10-20. Next: consolidate into `mission-contract-fulfillment` spec
  and archive.
