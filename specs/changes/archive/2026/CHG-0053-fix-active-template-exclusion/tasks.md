---
change_id: CHG-0053
artifact: tasks
status: archived
updated: 2026-10-26
---

# Tasks — CHG-0053

> Inside-out, sequential — one task at a time: Application RED first ->
> demanded Domain -> Application GREEN/BLUE -> Boundary RED (adapter
> regressions **and** public-API critical E2E, all written and run RED
> before outer production) -> one Gate 3-BOUNDARY for the mapped scopes ->
> adapters/composition GREEN -> E2E re-run -> gates -> consolidate.
> Every regression partition must produce actual RED evidence on the
> pre-fix code; do not weaken existing tests or canonical rules.

## Wave 1 — Application RED

- [x] **T-1 — Application RED: departed member sees the template again.**
  Owner: agent. Deps: none. Outcome: `ListOffersHandler` tests fail —
  `TestListOffers_DepartedMemberSeesTemplateAgain` and
  `TestListOffers_ActiveRosterMemberKeepsTemplateHidden` in
  `tests/unit/missions/application/list_offers_visibility_test.go`;
  dual-port `boardContractsStub` added in `list_offers_doubles_test.go`.
  Command: `go test -count=1 ./tests/unit/missions/application -run
  '^TestListOffers_(DepartedMemberSeesTemplateAgain|ActiveRosterMemberKeepsTemplateHidden)$'`.
  Maps to `internal/missions/application/list_offers.go`. RED evidence
  2026-10-21 (`evidence/t1-application-red.txt`): FAIL — departed member's
  board still hides `def-x` (returned 1 offer, expected 2) because the
  handler reads roster history via `ListByPlayer`; guard
  `ActiveRosterMemberKeepsTemplateHidden` PASS pre-fix. Next: Gate
  3-APPLICATION -> T-2 (Gate 3-APPLICATION covers T-1..T-3 RED -> T-6).
- [x] **T-2 — Application RED: roster member already bound is rejected.**
  Deps: T-1. Outcome: `AcceptMissionHandler` tests fail —
  `TestAcceptMissionHandler_RejectsTeammateWithSameDefinitionActive` and
  `TestAcceptMissionHandler_RejectsTeammatePendingValidation` in
  `tests/unit/missions/application/accept_mission_team_test.go`: a team
  acceptance whose roster includes a member blocking the same
  `definition_code` resolves `ErrMissionDefinitionAlreadyActive`;
  `PENDING_VALIDATION` also blocks **while the policy has spare slots**
  (the submit-side `PLAYER_ACTIVE` release is a separate out-of-scope
  defect — the test asserts the block, not the slot state). Command:
  `go test -count=1 ./tests/unit/missions/application -run
  '^TestAcceptMissionHandler_RejectsTeammate(WithSameDefinitionActive|PendingValidation)$'`.
  Maps to `internal/missions/application/accept_mission.go`,
  `accept_mission_team.go`, `internal/missions/domain/team_participation.go`.
  RED evidence 2026-10-21 (`evidence/t2-application-red.txt`): both FAIL —
  acceptance commits with no error (`writer.saved` grows) instead of
  rejecting, proving teammates are never definition-checked. Next: T-3.
- [x] **T-3 — Application RED: leave commits through `SaveDeparture`.**
  Deps: T-2. Outcome: `TestLeaveContract_ReleasesThroughDepartureWriter`
  fails in `tests/unit/missions/application/leave_contract_test.go` —
  non-terminal leave calls `DepartureWriter.SaveDeparture(contract,
  playerID)`; last-member leave routes to the terminal `Save`. Command:
  `go test -count=1 ./tests/unit/missions/application -run
  '^TestLeaveContract_(ReleasesThroughDepartureWriter|LastMemberStillSaves)$'`.
  Maps to `internal/missions/application/leave_contract.go`. RED evidence
  2026-10-21 (`evidence/t3-application-red.txt`): behavioral FAIL —
  `departureCalls` 0, expected 1 (handler still calls `Save`); guard
  `LastMemberStillSaves` PASS pre-fix via `Save`. BLUE retrofit note:
  once `NewLeaveContractHandler` takes `DepartureWriter`, the existing
  `lifecycleStore` doubles (Get+Save only) in the older leave tests must
  be adapted to the focused port — expected at GREEN wiring, not a new
  defect. Gate 3-APPLICATION covers T-1..T-3 RED -> T-6; **approved by
  user 2026-10-21**. ⚠ SUPERSeded (Gate 1 expansion 2026-10-21): the
  `LastMemberStillSaves` guard asserted last-member leave routes to
  generic `Save`; the settled expansion design routes **every** leave
  through `SaveDeparture` (pending revised Gate 2 + new RED) — the
  guard stays green on current code and its evidence is preserved, but
  it will be replaced by `TEST-APP-LVE-002`
  (`LastMemberUsesDepartureWriter`) once the expanded scope is approved.

## Wave 2 — Demanded Domain

- [x] **T-4 — Domain RED: `DecideRosterAcceptance` delegates per member.**
  Deps: T-2 RED. Outcome:
  `TestDecideRosterAcceptance_RejectsMemberWithSameDefinition` fails in
  `tests/unit/missions/domain/active_template_test.go` (subtests
  `IN_PROGRESS` + `PENDING_VALIDATION`) — every member is decided via
  `DecideAcceptance`; any member's definition conflict →
  `ReasonMissionDefinitionAlreadyActive`; guard
  `TestDecideRosterAcceptance_TeammateLimitKeepsTeammateReason` stays
  green. Command:
  `go test -count=1 ./tests/unit/missions/domain -run '^TestDecideRosterAcceptance_(RejectsMemberWithSameDefinition|TeammateLimitKeepsTeammateReason)$'`.
  Maps to `internal/missions/domain/team_participation.go` (unchanged).
  RED evidence 2026-10-21 (`evidence/t4-domain-red.txt`): FAIL —
  `decision.Accepted` true for a bound teammate (limit-only check).
  Gate 3-DOMAIN -> T-5; **approved by user 2026-10-21**.
- [x] **T-5 — Domain GREEN/BLUE + LAYER-GATE-DOMAIN.** Deps: T-4 evidence.
  Outcome: `DecideRosterAcceptance` delegates to `DecideAcceptance` per
  member in `internal/missions/domain/team_participation.go` (signature
  preserved); definition conflict propagates
  `ReasonMissionDefinitionAlreadyActive` and is checked before limit
  (deterministic regardless of map order); member limit resolves
  `ReasonTeammateLimitReached`. Commands: narrow
  `go test -count=1 ./tests/unit/missions/domain -run '^TestDecideRosterAcceptance_'`
  PASS and standalone `go test -count=1 ./tests/unit/missions/domain` PASS
  2026-10-21 (`evidence/t5-domain-narrow-green.txt`,
  `evidence/t5-domain-standalone-green.txt`). File 46 lines, gofmt clean.
  LAYER-GATE-DOMAIN satisfied. Next: T-6 Application GREEN needs its own
  approval.

## Wave 3 — Application GREEN/BLUE

- [x] **T-6 — Application GREEN/BLUE + LAYER-GATE-APPLICATION.** Deps:
  T-1..T-3 RED evidence + T-5. Outcome: `ListOffersHandler` now takes
  `ActiveContractsReader` and reads `ListActiveByPlayer` (history
  `ListByPlayer` port untouched — still serves
  `ListPlayerContractsHandler` and accept history); `LeaveContractHandler`
  gained the focused `DepartureWriter` port and routes non-terminal leave
  to `SaveDeparture`, terminal last-member to `Save`; T-2 was already
  green via the T-5 domain fix. Commands 2026-10-21: focused T-1/T-2/T-3
  `-run` PASS and `go test -count=1 ./tests/unit/missions/application`
  PASS (`evidence/t6-application-green.txt`). Test doubles updated to
  real membership semantics (`boardContractsStub` active-membership read;
  `lifecycleStore`/`lifecycleContractStore`/`disputeStore` gained
  `SaveDeparture` delegates). Known blocker until Gate 3-BOUNDARY:
  `internal/missions/composition` does not compile — the production
  `ContractStore` has no `SaveDeparture` yet and faking it is deferred to
  T-10 (approved scope); `go build ./...` otherwise clean.
  LAYER-GATE-APPLICATION satisfied. Next: T-7 Boundary RED needs its own
  approval.

## Wave 4 — Boundary RED (adapter + public API, all RED pre-fix)

- [x] **T-7 — Boundary RED: concurrent same-definition commit serializes.**
  Deps: T-6. Outcome: new
  `tests/integration/infrastructure/dynamodb/contract_store_reservation_test.go`
  (+ shared `chg0053_helpers_test.go` fixture with per-test dedicated
  `chg0053-*` tables deleted on cleanup). Command:
  `go test -count=1 ./tests/integration/infrastructure/dynamodb -run '^TestCommitAcceptance_(ConcurrentSameDefinitionSerializes|DifferentPlayersSameDefinitionBothCommit)$'`.
  RED evidence 2026-10-21 (`evidence/t7-boundary-red.txt`): behavioral —
  `ConcurrentSameDefinitionSerializes` FAILs (2 commits succeeded,
  expected 1 + `ErrMissionDefinitionAlreadyActive`); guard
  `DifferentPlayersSameDefinitionBothCommit` PASS pre-fix. Review
  corrections applied: fixture mints high-entropy table names and
  requires `ResourceNotFoundException` (not any error) before create —
  cleanup can only delete a table the test created, registered only
  after a successful create; race extracted into assertion-free
  `raceCommits` helper for one-line Act. `go vet` confirms the only
  package error is the intentional `SaveDeparture` gap.
- [x] **T-8 — Boundary RED: member-scoped, ownership-safe release.**
  Deps: T-6. Outcome: `contract_store_release_test.go` gained
  `TestContractStore_TerminalSaveReleasesOnlyActiveMembers`,
  `TestContractStore_StaleReleaseKeepsNewerReservation` (stale terminal
  save with a mismatched owner must conflict and write nothing),
  `TestContractStore_PendingValidationKeepsReservation`;
  `contract_store_leave_test.go` gained
  `TestContractStore_SaveDepartureReleasesDepartedGuards`. Commands:
  release suite and departure test run separately. RED evidence
  2026-10-21: `evidence/t8-release-red.txt` — behavioral FAILs
  (active-member reservation survives terminal save; stale save returns
  nil instead of conflicting) + `PendingValidationKeepsReservation` guard
  PASS pre-fix; `evidence/t8-departure-red.txt` — compile RED exactly on
  `store.SaveDeparture undefined` (mapped missing adapter). NOTE: the
  departure test intentionally fails the dynamodb package build until
  T-10 GREEN implements `SaveDeparture`.
- [x] **T-9 — Boundary RED: public-API critical E2E.** Deps: T-6.
  Outcome: new
  `tests/integration/composition/same_definition_acceptance_e2e_test.go`
  `TestConcurrentAcceptance_SameDefinitionSerializes` (public API: one
  `201`, one `422 MISSION_DEFINITION_ALREADY_ACTIVE`) and
  `departed_retake_e2e_test.go` `TestTeamLeaveJourney_DepartedMemberRetakes`
  (leave -> board re-shows -> retake `201`, slot released under limit 1).
  Command:
  `go test -count=1 ./tests/integration/composition -run '^(TestConcurrentAcceptance_SameDefinitionSerializes|TestTeamLeaveJourney_DepartedMemberRetakes)$'`.
  RED evidence 2026-10-21 (`evidence/t9-composition-red.txt`): compile
  RED — `composition/server.go:52` cannot wire `DepartureWriter` because
  `ContractStore` lacks `SaveDeparture` (the mapped missing adapter; no
  CHG-0051 noise). Review corrections applied: high-entropy table names +
  `ResourceNotFoundException` pre-create check; board decodes PascalCase
  `OfferID` (matching `MissionOffer.MarshalJSON`); race/journey extracted
  into assertion-free `raceAccepts`/`runRetakeJourney` helpers returning
  outcomes and per-call errors; retake seeds `PLAYER_DEF_ACTIVE`
  reservations for both members so a missing reservation release turns
  the retake into a definition conflict; evidence re-captured unchanged.
  **Gate 3-BOUNDARY (single approval for T-7/T-8/T-9 mapped scopes)
  pending user review -> T-10.**

## Wave 5 — Infrastructure / Composition GREEN

- [x] **T-10 — Persistence GREEN: reservation + release.** Deps: T-7..T-9
  RED + Gate 3-BOUNDARY (approved by user 2026-10-21). Outcome: new
  `contract_store_reservation.go` (guard keys, conditional `Put`,
  owner-conditioned member-scoped `Delete`) and
  `contract_store_departure.go` (`SaveDeparture`: terminal/departed
  validation, contract `Put` + departed member's `PLAYER_ACTIVE` DELETE +
  reservation release in one tx); `CommitAcceptance` adds one
  reservation `Put` per roster member with index-based mapping
  (definition conflict preferred over limit, mirroring
  `DecideAcceptance`); terminal `Save`/`SaveApproved` release
  reservations for `ActiveMembers()` only — departed members' guards
  never touched, mismatched owner cancels the tx; read-side methods moved
  to `contract_store_list.go` for the <150-line gate. Characterization:
  `TerminalSaveReleasesActiveSlot` now rejects on a different-definition
  offer so the limit guard stays the asserted signal. Command 2026-10-21:
  `go test -count=1 ./tests/integration/infrastructure/dynamodb -run 'TestCommitAcceptance_|TestContractStore_(TerminalSaveReleasesOnlyActiveMembers|StaleReleaseKeepsNewerReservation|PendingValidationKeepsReservation|SaveDepartureReleasesDepartedGuards)'`
  PASS (`evidence/t10-persistence-green.txt`); all six boundary tests
  green plus `ConcurrentLimitRace`. Next: T-11 composition wiring.
- [x] **T-11 — Composition wiring.** Deps: T-10. Outcome:
  `server.go:52` wires `NewLeaveContractHandler(contractStore,
  contractStore, contractStore)`; `go build ./...` and `go vet ./...`
  clean (one mechanical double fix: `fakePlayerContractsReader` gained
  `ListActiveByPlayer`); `PublicCode` mapping already covers
  `MISSION_DEFINITION_ALREADY_ACTIVE` -> 422 — no HTTP change needed.
  T-9 E2E now GREEN end-to-end 2026-10-21
  (`evidence/t11-composition-green.txt`): `SameDefinitionSerializes` —
  one 201 + one 422 `MISSION_DEFINITION_ALREADY_ACTIVE`;
  `DepartedMemberRetakes` — leave -> board re-shows `offer-x` -> retake
  201 with reservation+slot released under limit 1. New RED partition
  found in approved scope: `TestContractStore_SaveApprovedStaleReleaseIsNotDuplicate`
  FAILs (`evidence/t10-saveapproved-red.txt`) — `SaveApproved` maps a
  stale reservation-owner conflict to `ErrDuplicateConfirmation`.
  Partition Gate 3 approved 2026-10-21; GREEN fix: `creditStart` index
  recorded before the credit loop and `duplicateOnConditionalFail` now
  maps only credit-range conditional failures —
  `SaveApprovedRejectsDuplicateRetry` still passes, stale guard owner
  fails closed as a wrapped error (`evidence/t10-saveapproved-green.txt`,
  file 135 lines). Next: T-12.

## Wave 6 — Post-GREEN Evidence, Gates, Convergence

- [x] **T-12 — Re-run critical E2E + integration suite post-GREEN.**
  Deps: T-11. Outcome: T-9 scenarios green on final code; neighboring
  boundary suites green. Command:
  `go test ./tests/integration/composition ./tests/integration/http ./tests/integration/infrastructure/dynamodb`.
  Verified 2026-10-21: domain + application unit suites, all three
  integration packages (`-p=1`), `go build ./...` and `go vet ./...`
  all pass on DynamoDB Local :18000 — raw output in
  `evidence/t12-{domain,application,dynamodb,composition,http,build-vet}.txt`.
  Next: T-13.
- [~] **T-13 — L3 gates.** Deps: T-12. Outcome: mutation
  (`mutation-report.md`), coverage >= 90% domain/application + affected
  infra, security-review findings resolved, clean-up <= 150 lines,
  `tools/validate-sdd-change.sh` recorded in `verification.md`.
  Documentation gate passed 2026-10-21 (DOMAIN-RULES/DATA-MODEL/
  ARCHITECTURE/TESTING updated; API-CONTRACT recorded no-change).
  Cleanup inventory done: no trailing whitespace, `gofmt` clean on
  0053 files. The two >149 test files were split by moving helpers to
  sibling doubles files — `list_offers_test.go` 103 / `list_offers_
  doubles_test.go` 72, `contract_disputes_test.go` 87 /
  `contract_disputes_doubles_test.go` 75; narrow scopes green.
  `docs/DOMAIN-RULES.md` eligibility list updated to drop the removed
  offer capacity (`available space`) in favor of live season/offer
  window/repeatability + player limit. Six domain tests added for
  `NewMissionOfferFromDefinition` identity/publication guards (frozen
  coverage inputs recorded separately; domain measured 89.77% — two
  statements short — lead re-measures after this). Evidence
  `evidence/t13-cleanup.txt`. Security investigation opened candidate
  High FIND-0053-01: stale `SaveDeparture` resurrects a departed member
  via unconditional contract Put — behavioral RED captured in
  `evidence/t13-concurrent-departure-red.txt`. Gate 1 amendment +
  scoped Gate 3 approved 2026-10-21: delta extended with
  SCN-MB-DEPART-RACE / REG-0053-03 / TEST-INF-REL-005. GREEN —
  `SaveDeparture` contract put is conditional (exists + state +
  departed-count fence); all departure/release tests and the
  leave/retake E2E pass (`evidence/t13-departure-green.txt`); the T-8
  departure test Arrange now persists the pre-leave contract. Review
  reopened the partition: count-only fence is weaker than the approved
  delta (missing `Departed` accepted with expected prior count >0;
  same-count different set passes). New RED evidence
  `evidence/t13-departure-set-red.txt`; strengthened GREEN — the fence
  now requires the exact prior departed set (`size` + `contains` per
  prior member), missing attribute tolerated only when the prior set is
  empty; all fence tests, the T-8 release test, and the leave/retake
  E2E pass (`evidence/t13-departure-set-green.txt`). FIND-0053-01
  resolved pending human security verdict; the same-set case-B test now
  asserts index-0 `ConditionalCheckFailed` (kills the
  `attribute_exists` mutation survivor,
  `evidence/t13-departure-cancel-reason.txt`). Follow-up candidate
  FIND-0053-02 recorded: unfenced non-terminal `Save` resurrects a
  departed member — RED at `evidence/t13-stale-save-red.txt`. Gate 1
  expansion approved 2026-10-21 brought `Save`/`SaveApproved` inside
  this change's canonical ownership (`mission-contract-fulfillment`);
  the fix is mapped to T-15 and stays open pending the revised Gate 2.
  Still pending approval:
  mutation, coverage verdict, security verdict, validator
  (runs `docker compose down --remove-orphans` — needs explicit
  permission or a safe equivalent).
  **Resolved current state (2026-10-26):** FIND-0053-02 landed GREEN in
  T-15 (`evidence/t15-persistence-green.txt`); mutation 10/10 killed;
  coverage lead-measured 3198/3552 = 90.03%; security approved for
  local code; validator ran in the authorized isolated environment
  (exit 0, `evidence/t17-isolated-validator.txt`). No T-13 blocker
  remains — only T-14 consolidation approval.
- [x] **T-14 — Docs + independent review + consolidate.** Deps: T-13,
  T-15. Outcome: documentation gate passed 2026-10-26; all L3 gates
  green (validator exit 0 isolated, coverage lead-measured
  3198/3552 = 90.03%, mutation 10/10, critical E2E 9/9, security
  approved local-only 2026-10-26). Verify-spec approved 2026-10-26 —
  delta consolidated into `mission-board` (REQ-MB-004, REQ-MB-008) and
  `mission-contract-fulfillment` (REQ-MCF-010, REQ-MCF-013,
  REQ-MCF-020, scenarios appended — CHG-0051 untouched); folder
  archived to `specs/changes/archive/2026/CHG-0053-fix-active-template-exclusion/`.

## Wave 7 — Canonical Expansion: Mission-Contract-Fulfillment (Gate 2 expansion approved 2026-10-21)

- [x] **T-15 — Expanded write fences (RED first, Gate 2 expansion
  required).** Deps: Gate 2 expansion approval (**granted 2026-10-21** —
  T-15 REDs proceed sequentially). Gate 1 expansion approved 2026-10-21
  added `mission-contract-fulfillment` to canonical_paths.
  `TEST-APP-LVE-002` RED captured 2026-10-21
  (`evidence/t15-application-red.txt`): `LastMemberUsesDepartureWriter`
  fails behaviorally — `departureCalls` 0, expected 1 (terminal leave
  still routed to generic `Save`). Gate 3-APPLICATION-EXPANSION
  **approved 2026-10-21** → Application GREEN/BLUE landed:
  `DepartureWriter.SaveDeparture(contract, departingPlayerID,
  previousState)`; handler captures `previousState` before `Leave` and
  routes **every** exit (incl. terminal last-member ABANDONED) through
  the focused port; ctor is `(contracts, departure)` — generic writer
  dropped. Superseded `LastMemberStillSaves` removed (its ABANDONED +
  persistence assertions are covered by the new test plus
  `prevState == ContractInProgress`). Unit doubles
  (`departureStore`, `lifecycleStore`, `disputeStore`,
  `lifecycleContractStore`) updated mechanically. Suites green:
  `./tests/unit/missions/application`, `./tests/unit/missions/interface/http`;
  `go vet ./tests/integration/http` clean
  (`evidence/t15-application-green.txt`). Boundary REDs captured
  2026-10-26: `TEST-INF-REL-007` behavioral —
  `StaleSaveApprovedDoesNotRestoreOrCredit` got nil from stale
  `SaveApproved` (departure overwritten, credits minted,
  `evidence/t15-stale-approval-red.txt`); `TEST-INF-REL-008` compile —
  3-arg `SaveDeparture(contract, id, previousState)` vs 2-arg adapter
  (`evidence/t15-last-member-red.txt`); `TEST-E2E-MB-003` compile —
  `server.go:52` old ctor arity blocks the composition package
  (`evidence/t15-solo-e2e-red.txt`). Gate 3-BOUNDARY-EXPANSION
  **approved 2026-10-26** → persistence GREEN/BLUE landed
  (`evidence/t15-persistence-green.txt`, 14/14 in scope): shared
  `departedSetMatch` helper (size + `contains` per expected member;
  missing tolerated only when empty, which also permits absent-row
  create for legacy fixtures) fences `Save` non-terminal put AND
  terminal tx put (REG-0053-04) and `SaveApproved`'s contract put plus
  `attribute_exists(PK)` (REG-0053-05; credit-range duplicate
  classification preserved). `SaveDeparture(contract, id,
  previousState)` validates a non-terminal prior state and allows
  ABANDONED only on last-member exit; the put conditions on the
  persisted prior state and departed set. Call sites updated with true
  pre-leave states (`PENDING_VALIDATION` on the stale-approval path);
  departed-row fixtures now seed the pre-leave row then persist the
  departure. Test-quality fix accepted at review:
  `SaveApprovedStaleReleaseIsNotDuplicate` was green for the wrong
  reason (index-0 `attribute_exists` rejection, never reaching the
  owner guard) — arrange now persists the pre-approval row before
  marking it completed and asserts the stored row stays
  `IN_PROGRESS`. Composition wiring landed 2026-10-26: `server.go`
  ctor is `NewLeaveContractHandler(contractStore, contractStore)`;
  `go build ./...` clean and all three composition E2Es pass —
  `SoloLeaveJourney_LastMemberRetakes` (TEST-E2E-MB-003 GREEN: solo
  leave → board re-shows `offer-x` → retake 201 under limit 1),
  `DepartedMemberRetakes`, `SameDefinitionSerializes`
  (`evidence/t15-composition-green.txt`). Note: the solo E2E needed
  `DefinitionVersion = "legacy"` on the fixture offer — the accept
  path resolves real definitions unless the sentinel is set (same
  convention as the departed-retake test).
  Planned REDs (existing RED `TEST-INF-REL-006` already captured at
  `evidence/t13-stale-save-red.txt`):
  - `TEST-APP-LVE-002` — last-member leave calls
    `SaveDeparture(contract, playerID, previousState)`; supersedes the
    T-3 `LastMemberStillSaves` guard (preserved until replaced).
  - `TEST-INF-REL-007` — stale `SaveApproved` on a pre-leave snapshot
    conflicts atomically: no contract overwrite, no credit items.
  - `TEST-INF-REL-008` — last-member `SaveDeparture` releases the
    leaver's reservation and `PLAYER_ACTIVE` slot (no orphan guard).
  - `TEST-E2E-MB-003` — public solo last-member leave → board re-shows
    the template → retake under limit 1.
  Production plan per `design.md` Expansion: `LeaveContractHandler`
  captures `previousState` and always calls the focused port
  (constructor drops `ContractWriter`); `Save`/`SaveApproved` contract
  puts fenced on the exact prior departed set (missing tolerated only
  when empty; `SaveApproved` also requires the row to exist). Commands:
  `go test -count=1 ./tests/unit/missions/application -run 'TestLeaveContract_'`,
  `go test -count=1 ./tests/integration/infrastructure/dynamodb -run 'Stale|LastMember|Departure'`,
  `go test -count=1 ./tests/integration/composition -run 'SoloLeave|DepartedMemberRetakes'`.
  All T-15 REDs reviewed and landed GREEN per the sequence above;
  remaining gates: mutation/coverage/security verdicts and validator
  (T-16 evidence captured).

- [x] **T-16 — Final verification sweep, safe validator, docs +
  security evidence.** Deps: T-15 GREEN. Outcome: standalone suites all
  green — `evidence/t16-domain.txt`, `t16-application.txt`,
  `t16-infrastructure.txt`, `t16-composition.txt`, `t16-http.txt`,
  `t16-build.txt`, `t16-vet.txt`; critical E2E `count=3` 9/9
  (`evidence/t16-critical-e2e.txt`). One fix during the sweep:
  `PersistsDeparturesAndFreesLeaverSlot` hit the departed-set fence on
  a stale pre-fence row left in its deterministic table — added
  `clearMissionsTable` (the fence behaved correctly; isolation gap in
  the test). Safe validator recipe executed against temp snapshot
  `/tmp/chg0053-validate-BKzXuF` (rsync copy, PATH without docker):
  exit 1 — only `infrastructure/s3` EvidenceStore tests fail for lack
  of an S3 emulator; every mission-scope suite, vet, build green
  (`evidence/t16-validator.txt`); original repo and coverage profiles
  untouched. Docs updated: DOMAIN-RULES (last-member release +
  stale-write rules), DATA-MODEL + ARCHITECTURE (departed-set fence),
  TESTING properties 17/18. `security-review.md`: FIND-0053-02 marked
  resolved with GREEN evidence; reviewer role + residual risks +
  human verdict approved for local code 2026-10-26 (no shared-env
  rollout). Line inventory: all CHG-0053 files
  <150 (max `contract_store_review.go` 147); gofmt clean except the
  two pre-existing non-0053 files.
  **Safety containment (post-incident):** all CHG-0053 tests now use
  fresh high-entropy `chg0053-*` tables that are never dropped;
  `DeleteTable`/`t.Cleanup` removed from both `chg0053Table` helpers;
  `contract_store_leave_test.go`, `concurrent_acceptance_e2e_test.go`,
  and `team_leave_e2e_test.go` converted off the shared
  `hbk-missions-composition`/`clearMissionsTable` convention. ~34
  pre-existing non-0053 suite files still clear shared tables — flagged
  in `verification.md`. **T-17 isolated gate:** under user
  authorization, throwaway containers (`chg0053-ddb-gate-20261026`/
  `chg0053-s3-gate-20261026`) + snapshot validator run — **exit 0**,
  all suites incl. s3 + race green, original coverage files untouched
  (`evidence/t17-isolated-validator.txt`); containers left running for
  lead cleanup.

- [x] **T-18 — Test-only coverage top-up (mission-board facet gap).**
  Deps: T-17 frozen profile + lead selection. Outcome: new
  `tests/unit/missions/interface/http/list_offer_facets_test.go` (99
  lines) — 4 characterization tests over the existing
  `ListOfferFacetsHandler` HTTP contract (distinct facets,
  INVALID status 400, NO_ACTIVE_SEASON 422, reader error 500) using
  real application handler + existing doubles; unit http package all
  green (`evidence/t18-board-facets.txt`). Follow-up: lead
  re-measurement left a 4-statement shortfall → added
  `list_offer_facets_season_test.go` (72 lines, 3 season-resolution
  cases: explicit season filter, ACTIVE_SEASON_EXISTS 422, season-reader
  500 + `failingActiveSeasons` double); all 7 facet tests green
  (`evidence/t18-board-facets-extra.txt`). Security-branch top-up:
  `accept_identity_test.go` (31 lines) — 401 UNAUTHORIZED before
  application invocation via `AcceptMissionHandler(nil)` fail-closed
  proof (`evidence/t18-accept-identity.txt`). No production/spec
  changes; lead re-measured the combined profile at
  3198/3552 = 90.03% (coverage gate passed,
  `evidence/coverage-project-combined.out`); human security approved
  for local code 2026-10-26. Next: consolidation approval only (T-14).
