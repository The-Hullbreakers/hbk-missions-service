---
change_id: CHG-0022
artifact: verification
updated: 2026-10-08
---

# Verification — CHG-0022

## Status

`verified` — all L3 gates green.

## Evidence Plan

| Check | Command / Method | Result |
| --- | --- | --- |
| Unit suites | `go test ./tests/unit/...` | pass — 5 packages |
| HTTP integration | `go test ./tests/integration/http/...` | pass |
| Infrastructure integration | `go test ./tests/integration/infrastructure/...` | pass |
| Composition + E2E | `go test ./tests/integration/composition ./tests/integration/e2e` | pass — incl. `TestConcurrentAcceptance_*` |
| Race | `go test -race -p=1 ./tests/...` | pass |
| SDD validation | `tools/validate-sdd-change.sh` | pass |
| Coverage gate | validate script coverage section | pass — domain 90.6%, application 91.5%, dynamodb 92.2% |
| Mutation gate | manual mutants (gremlins not installed) | pass — 4/4 killed: season guard, capacity condition, limit condition, reason mapping |
| Critical E2E | `TestConcurrentAcceptance_LastOfferSlotSerializes` | pass — 2 racers on one slot → one 201 + one 422 |

## Concurrency Evidence

- `TestCommitAcceptance_ConcurrentCapacityRace`: two goroutines on a
  capacity-1 offer → one commits, other gets `ErrOfferCapacityExhausted`.
- `TestCommitAcceptance_ConcurrentLimitRace`: player one below limit, two
  concurrent accepts → one commits, other `ErrActiveMissionLimitReached`.
- `TestContractStore_TerminalSaveReleasesActiveSlot`: terminal save frees the
  per-player slot; re-accept succeeds.
- `TestConcurrentAcceptance_ClosedSeasonRejects`: FINALIZED season →
  422 `SEASON_NOT_ACTIVE`, no contract, no capacity consumed.

## Reproduction Evidence (pre-fix)

- Two players could accept a capacity-one offer (check-then-write race).
- Two concurrent acceptances exceeded the player's limit.
- Limit bypassed by spreading acceptances across categories
  (`ListActiveByPlayer` was category-filtered).
- Offers bound to FINALIZED/FREEZING seasons could be accepted.
