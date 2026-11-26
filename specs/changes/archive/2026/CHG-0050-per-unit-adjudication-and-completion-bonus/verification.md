---
change_id: CHG-0050
artifact: verification
status: converged
updated: 2026-10-20
---

# Verification — Per-Unit Adjudication And Completion Bonus

## Gate Evidence

- [x] Gate 1 — active change approved (2026-10-20).
- [x] Gate 2 — test-first plan approved (2026-10-20, "si").
- [x] Gate 3 per layer — RED evidence:
  - Application RED: `contract_review_verdict_test.go` compile-failed on
    missing `HandleWithVerdict`/`UnitRef`/`ErrUnitNotStaged`.
  - Domain RED: `contract_approve_verdict_test.go` failed before
    `ApproveWithVerdict`/`AllMandatoryFulfilled` existed.
  - Boundary RED: verdict tests returned 200/all-credits before the
    handler parsed `rejected_units`; unknown unit returned 200 instead of
    422.
- [x] `go test ./...` green — 13 packages ok, no failures.
- [x] Coverage: domain 90.2%, application 91.2%, dynamodb 92.9%
  (`tools/validate-sdd-change.sh` profiles).
- [x] Mutation gate (manual mutants on new adjudication logic — all
  killed by `tests/unit/missions/domain`):
  - `count < required → count < 1` (at-least-one completion) → FAIL.
  - `ErrUnitNotStaged → continue` (unknown units ignored) → FAIL.
  - `AllMandatoryFulfilled → true` (bonus on partial) → FAIL.
- [x] Security gate — reviewer verdict boundary: member self-approve
  → 403 (moderation journey E2E); verdict body carries only unit refs —
  no amounts or identities accepted from the caller; rejected units never
  reach the ledger; `UNIT_NOT_STAGED` is a stable public code, internals
  not leaked. No Critical/High findings.
- [x] Critical E2E — `TestVerdictJourney_PartialApprovalPaysAcceptedUnitsOnly`
  through the real composition root: 2 staged units, verdict rejects 1 →
  leaderboard 100 (no 50 bonus), contract COMPLETED, `rejected_units`
  visible to the member.
- [x] `tools/validate-sdd-change.sh` green (gofmt, vet, unit, integration,
  race, coverage).
- [x] Seed catalogs verified: `test_seed_missions` 8 tests ok;
  `test_seed_integration` deployment lifecycle ok against DynamoDB Local
  (versioned immutability + offer capacity preserved, `Version: "3"`).

## Notes

- `TestSoloContractJourney_FrozenTermsEndToEnd` was updated: its fixture
  relied on the old at-least-one completion predicate (unit_limit 2 with
  one confirmed unit). It now confirms both units — the journey semantics
  the test always claimed.
- E2E covers solo SOLO contracts; team-contract verdict projection is
  covered by unit/integration projections (member-scoped
  `rejected_units`).

## Rollback

- Seed catalog republish restores prior definitions under a higher
  version; live contracts keep frozen terms — the verdict-capable approve
  remains backward compatible with approve-all clients (empty body).
