---
change_id: CHG-0031
artifact: verification
status: draft
updated: 2026-10-09
---

# Verification — Mission Completion Moderation

## Scope

L3 change over `mission-contract-fulfillment` (lifecycle, credit
materialization, adjudication boundary) and `mission-board` (active-slot
accounting). Evidence follows the L3 gate matrix.

## Planned Evidence

| Gate | Evidence |
| --- | --- |
| BDD acceptance | `delta.md` scenarios mapped to tests in `traceability.yaml` |
| Domain RED/GREEN/BLUE | `tests/unit/missions/domain` — submit/approve/reject, staged credits, `IsActive` |
| Application RED/GREEN/BLUE | `tests/unit/missions/application` — role enforcement, port-double orchestration |
| Infrastructure | `tests/integration/infrastructure/dynamodb` — `pending_credits` round-trip, transactional approve, idempotent retry |
| HTTP integration | `tests/integration/http` — submit/approve/reject status vocabulary, `complete` removed |
| Critical E2E | `tests/integration/composition` — full approve journey + reject journey |
| Mutation | domain lifecycle + staged-credit logic, no unexplained survivors |
| Security | `security-review.md` — reviewer boundary, tenant scoping |
| Coverage | production ≥90%, no affected-scope regression |
| Validation | `tools/validate-sdd-change.sh` |

## Constraints Verified

- `TransactWriteItems` 100-item bound: staged credits are bounded by
  `unit_limit` per objective; contract write + N credit puts stays far
  below the bound (N bounded by frozen terms).
- Append-only ledger preserved: rejection writes zero credit items.
- `PENDING_VALIDATION` non-terminal: `abandon`/`leave` still reachable;
  evidence/upload closed.

## Results

| Gate | Result | Evidence |
| --- | --- | --- |
| Gate 3-DOMAIN | passed | RED: compile failure on `Submit`/`Approve`/`Reject`/`PendingCredits`/`ErrContractNotPendingValidation` (2026-10-11, user approved). GREEN: `go test -race ./tests/unit/missions/domain/` ok. BLUE: split `mission_contract_review.go` (91) from lifecycle (83); split tests `contract_submit_test.go` (111) / `contract_review_test.go` (123); all files <150. `IsActive` frees pending, `BlocksRetake` preserves REQ-MB-008 hiding + ACTIVE list bucket. `Complete` kept temporarily — removed in boundary wave with the `complete` route. |
| Gate 3-INFRA | passed | RED: compile failure on `ContractStore.SaveApproved` (2026-10-12, user approved). GREEN: `go test ./tests/integration/infrastructure/...` ok vs DynamoDB local :18000 — `Save`+`Get` round-trips `PendingCredits`, `SaveApproved` commits contract+roster release+N conditional credit puts, retried approval → `ErrDuplicateConfirmation`. BLUE: `rosterReleaseTx`/`duplicateOnConditionalFail` shared by `Save`/`SaveConfirmed`/`SaveApproved`; files <150 (review 130, confirm 74, store 132). |
| Gate 3-APPLICATION | passed | RED: compile failure on `NewSubmitContractHandler`/`NewApproveContractHandler`/`NewRejectContractHandler` (2026-10-11, user approved). GREEN: `go test -race ./tests/unit/missions/application/` ok. `SubmitContractHandler`/`ApproveContractHandler`/`RejectContractHandler` + `ContractCreditsWriter.SaveApproved` port; `isReviewer` reused; `ConfirmObjectiveHandler` switched to `ContractWriter.Save` (staged credits); `atomicWriter` double reflects staged persistence. `CompleteContractHandler` kept until boundary wave. |
| Gate 3-BOUNDARY | passed | RED: compile failure on `missionhttp.Submit/Approve/RejectContractHandler` (2026-10-12, user approved). GREEN: `go test ./tests/integration/http/` ok — submit 200/403, approve 200+`[]MeritCredit` body/403, reject 200; `CONTRACT_NOT_PENDING_VALIDATION` in `PublicCode`; approve/reject behind `adminOrModerator` in `server.go`. BLUE — complete path retired: `/complete` route, `CompleteContractHandler` (http+app), `domain.Complete`, `ContractAndCreditWriter`, `SaveConfirmed` + its test deleted; `contract_store_confirm.go`→`contract_item.go`; `Complete` tests converted to `TestContractApprove_*` (split into `contract_approve_test.go`, 141); lifecycle test files <150; e2e journeys drive submit+approve. |
| Critical E2E | passed | `TestModerationJourney_RejectPaysNothing` (composition, real DynamoDB): staged credit invisible on leaderboard pre-approval, evidence closed while pending, self-approve 403, reject → zero credits, terminal re-approve 422. Approve journeys: `solo_journey` + `team_leave` E2E converted to submit+approve (leaderboard 325 / team-9 100). |
| Mutation gate | passed | Manual (no toolchain): mutant 1 — `Approve` skips `AllConfirmed` guard → `TestContractApprove_EmitsCompletionCreditWhenAllConfirmed` + `FullConfirmationCreditsEverything` fail. Mutant 2 — `isReviewer` bypassed → `TestApproveContract_MemberCannotAdjudicate` + `TestRejectContract_MemberCannotAdjudicate` fail. 2/2 killed. |
| Security gate | passed | `security-review.md` → status passed; no Critical/High findings; role/state/atomicity evidence recorded; residual: no review-queue listing (documented follow-up). |
| Coverage gate | passed | domain 91.5%, application 91.5%, infrastructure/dynamodb 92.7% (all ≥90%). Added `store_queries_test.go` for `DefinitionStore.List`; deleted dead `SeasonStore.ListAll` (latent `SEASON#` prefix bug, zero callers — scheduler uses `List`). |
| Documentation | done | `API-CONTRACT.md` submit/approve/reject sections; `API-GUIDE-FRONTEND.md` §7 staged credit + §7c moderation + error catalog; `DOMAIN-RULES.md` approval/zero-reward wording + departed rights list. |
| Clean-up | done | Zero-reward drift fixed: `Approve` emits COMPLETION only when `CompletionReward > 0` (canonical rule); dead `SeasonStore.ListAll` removed; `evidence_media_test.go` (260) split into helpers + 3 suites (<150); all touched files <150; gofmt/vet clean; full suite green. |
