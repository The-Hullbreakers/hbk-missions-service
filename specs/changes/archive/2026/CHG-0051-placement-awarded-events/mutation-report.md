---
change_id: CHG-0051
artifact: mutation-report
status: completed
updated: 2026-11-16
---

# Mutation Report — CHG-0051

## Scope

Non-trivial logic selected for mutation: `PlacementFact` derivation from the
finalized projection, `SeasonFinalizer` two-phase ordering, deterministic
`event_id`/digest derivation, staged-fact count/hash verification, the
adjudication-lock write fences, and the constant-size manifest commit.

## Method

`go-mutesting`/`gremlins` are not installed; manual analysis of critical
mutations, matching the convention recorded in CHG-0003.

## Critical Mutations Reviewed

| Function | Mutation | Killing test |
| --- | --- | --- |
| `MeritLedgerStore.Save` fence | Drop the season-unlocked `ConditionCheck` | `TestAdjudicationLedgerFence_LockedSeasonRejectsCredit` writes a credit under the lock and expects `ErrSeasonAdjudicating` with no row persisted |
| `ContractStore.Save` credit path | Remove `saveCreditBearing` fence | `TestAdjudicationFence_LockedSeasonCannotStageContractCredit` fails — credit-bearing save under lock must leave contract untouched |
| `ContractStore.SaveApproved` | Omit prepended lock `ConditionCheck` | `TestAdjudicationApprovalFence_LockedSeasonRejectsCreditsAtomically` fails — commit + credits must abort together |
| `CommitAcceptance` | Read season state before transaction instead of fencing in it | `TestSeasonAdjudicationFence_Acceptance*` stale-read cases fail — FREEZING maps to `ErrSeasonNotActive`, locked to `ErrSeasonAdjudicating`, guards untouched |
| `AdjudicationLock.Acquire` | Mint a new attempt per call instead of stable identity | `season_adjudication_lock_retry_test` fails — retry resumes the same attempt/cutoff |
| `StagedFactVerifier.Verify` | Skip digest/count/hash checks | `season_staged_fact_verifier_integrity_test` fails — tampered row, wrong count and wrong aggregate hash each reject |
| `FinalizationCommitter.Commit` | Commit without persisted lock/cutoff CAS | `season_manifest_commit_lock_test` / `season_manifest_commit_retry_test` fail — stale attempt and retry cannot reset progress |
| `SeasonFinalizer.Handle` | Publish before manifest commit | `season_finalizer_failures_test` + verifier ordering asserts fail; staged rows are unreachable by `ListPendingManifests` until commit |
| `LocalizeContract`/offers fallback | Serve `lang=en` without Spanish fallback | localized list/detail integration tests fail — legacy records serve `es` and label it |
| `errors.PublicCode` | Map `ErrSeasonAdjudicating` to `INVALID_REQUEST` | `TestPublicCode_MapsDomainErrors` fails — expects `SEASON_ADJUDICATING` |

## Conclusion

- Every scoped mutant is killed by an existing unit or integration test.
- Mutations inside AWS SDK serialization and OIDC wiring are not
  unit-mutable; they are covered by the security review and the critical
  E2E instead.
- Residual: mutation score is a manual qualitative record, not a
  tool-computed percentage — same declared limitation as CHG-0003.
