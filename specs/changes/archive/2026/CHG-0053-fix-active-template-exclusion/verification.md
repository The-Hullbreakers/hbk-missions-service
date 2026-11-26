---
change_id: CHG-0053
artifact: verification
status: archived
updated: 2026-10-26
---

# Verification — CHG-0053

## Gates

| Gate | Scope | Decision | Evidence | Date |
| --- | --- | --- | --- | --- |
| Gate 1 | spec writes | approved | user approval — create active change, consolidate `mission-board` | 2026-10-21 |
| Gate 1 expansion | canonical ownership | approved | user approval — extend canonical_paths to `mission-contract-fulfillment` (REQ-MCF-013/010/020: departure-write fence, stale Save/SaveApproved) | 2026-10-21 |
| Gate 2 | start RED | approved (mission-board scope) | user approval 2026-10-21 | 2026-10-21 |
| Gate 2 expansion | start MCF-scope RED | approved | user approval 2026-10-21 — T-15 REDs (`TEST-APP-LVE-002`, `TEST-INF-REL-007/008`, `TEST-E2E-MB-003`) proceed sequentially | 2026-10-21 |
| Gate 3-APPLICATION-EXPANSION | leave routing + ctor change | approved | user approval 2026-10-21 after `TEST-APP-LVE-002` RED review — landed: `leave_contract.go` previousState capture, all-exit `SaveDeparture(contract, id, previousState)` routing, ctor `(contracts, departure)` | 2026-10-21 |
| Gate 3-MCF-EXPANSION | MCF write-fence production scope | approved | satisfied by the Gate 3-BOUNDARY-EXPANSION approval 2026-10-26 — identical mapped production: `leave_contract.go` (previousState + focused-port routing, ctor drops `ContractWriter` — landed via Gate 3-APPLICATION-EXPANSION), `contract_store.go` `Save` departed-set fence, `contract_store_review.go` `SaveApproved` departed-set fence + existing-row precondition | 2026-10-26 |
| Gate 3-BOUNDARY-EXPANSION | adapter + composition MCF scope | approved (persistence partition) | user approval 2026-10-26 after RED review — landed: `contract_store_departure.go` 3-arg `SaveDeparture` + prior-state/last-member release, `contract_store.go`/`contract_store_review.go` departed-set fences; `server.go` ctor wiring deferred to composition landing | 2026-10-26 |
| Gate 3-APPLICATION | list/accept/leave use cases | approved | T-1..T-3 RED approved by user; evidence t1/t2/t3 | 2026-10-21 |
| Gate 3-DOMAIN | roster definition check | approved | T-4 RED + T-5 GREEN `evidence/t4-domain-red.txt`, `evidence/t5-domain-*.txt` | 2026-10-21 |
| Gate 3-BOUNDARY | reservation/release transaction + public-API E2E (single approval, T-7..T-9 + SaveApproved partition) | approved | T-7/T-8/T-9 RED + T-10 GREEN `evidence/t7-boundary-red.txt`, `evidence/t8-release-red.txt`, `evidence/t8-departure-red.txt`, `evidence/t9-composition-red.txt`, `evidence/t10-persistence-green.txt`; SaveApproved partition RED+GREEN `evidence/t10-saveapproved-red.txt`, `evidence/t10-saveapproved-green.txt` | 2026-10-21 |
| LAYER-GATE-DOMAIN | demanded domain complete | passed | T-5 GREEN: roster decision delegates per member | 2026-10-21 |
| LAYER-GATE-APPLICATION | core complete | passed | T-6 GREEN: board reads `ListActiveByPlayer`; leave routes `SaveDeparture`; composition wiring deferred to T-10 (known blocker, `ContractStore` lacks `SaveDeparture`) | 2026-10-21 |
| Review-test-evidence | RED before each GREEN | passed | every mapped RED captured and reviewed before its GREEN — `t1/t2/t3/t4/t7/t8/t9` application+domain+boundary REDs, `t10-saveapproved-red`, `t13-concurrent-departure-red`, `t13-departure-set-red`, `t13-stale-save-red`, `t15-application-red`, `t15-stale-approval-red`, `t15-last-member-red`, `t15-solo-e2e-red`; approvals recorded in the gate rows above | 2026-10-26 |
| Documentation gate | live docs owning CHG-0053 facts | passed | `docs/DOMAIN-RULES.md` (participation-scoped exclusion, PENDING_VALIDATION retake-block vs free slot, per-member departure, commit-time reservation, eligibility list corrected for removed offer capacity, last-member guard release + stale-write conflict rules), `docs/DATA-MODEL.md` (owned reservation vs separate active-slot count, departed-set write fence), `docs/ARCHITECTURE.md` (single acceptance transaction, member-scoped release, conditional write fence on all contract writes), `docs/TESTING.md` (properties 3, 16 + new 17/18 for last-member release and stale-write conflict). `docs/API-CONTRACT.md` — no change: `MISSION_DEFINITION_ALREADY_ACTIVE`/422 already listed and CHG-0053 changes no request/response payload (evidence: `t11-composition-green.txt` E2E uses the existing contract) | 2026-10-26 |
| Security gate | trust-boundary review | approved for local code only | User approval 2026-10-26; FIND-0053-01/02 resolved by departure/ordinary-save/approval fences (`t13-departure-set-green.txt`, `t15-persistence-green.txt`). No shared-environment rollout approved: external scanners unavailable, legacy reservation reconciliation and operational signoff still required (`security-review.md`, `rollout.md`) | 2026-10-26 |
| Coverage gate | >=90% project-wide production and affected scope | passed | Final isolated T-19 full-suite profile `/tmp/chg0053-finalreview-dScGSy/coverage-project-final.out`: 3198/3552 unique statements (90.03%, independently measured); `go tool cover -func` reports 90.0%. Standalone validator profiles: domain 90.0%, application 90.9%, DynamoDB 93.4%; the affected scope exceeds 90%. Earlier incremental block-wise OR profile `evidence/coverage-project-combined.out` independently agreed with the final full-suite result. Original root coverage files untouched | 2026-10-26 |
| Mutation gate | Go overlay mutants of changed policy and transactions | passed | Go 1.27.0, `evidence/run-mutation-gate.py` and `evidence/mutation-results.txt`: final 10/10 killed, 0 survivors/timeouts; includes departed-set fences on ordinary save/approval, terminal leave and prior-state condition. The departed-member mutant initially survived on a DynamoDB `ValidationException`; the test now asserts contract-put `ConditionalCheckFailed`. All integration mutants ran against user-authorized isolated DynamoDB `:18853`, no original-table cleanup | 2026-10-26 |
| Critical E2E | public accept + leave journeys | passed | `evidence/t16-critical-e2e.txt`: public API same-definition concurrency, team leave/retake, solo last-member leave/retake, 3 runs each (9/9), plus isolated T-17 full-suite/race pass | 2026-10-26 |
| Final verify | verify-spec + consolidation approval | passed | explicit user approval 2026-10-26 — delta consolidated into `mission-board` (REQ-MB-004, REQ-MB-008) + `mission-contract-fulfillment` (REQ-MCF-010, REQ-MCF-013, REQ-MCF-020); archived to `specs/changes/archive/2026/CHG-0053-fix-active-template-exclusion/` on `efa42234f4242e2680b92b024e72d8f6d24c68ed` + reviewed dirty worktree | 2026-10-26 |

## Commands

| Command | Scope | Result | Notes |
| --- | --- | --- | --- |
| `go test -count=1 ./tests/unit/missions/application -run '^TestListOffers_(DepartedMemberSeesTemplateAgain|ActiveRosterMemberKeepsTemplateHidden)$'` | T-1 RED | FAIL (expected) — Departed test red, Active guard green | `evidence/t1-application-red.txt` |
| `go test -count=1 ./tests/unit/missions/application -run '^TestAcceptMissionHandler_RejectsTeammate(WithSameDefinitionActive|PendingValidation)$'` | T-2 RED | FAIL (expected) — both acceptances commit with no error | `evidence/t2-application-red.txt` |
| `go test -count=1 ./tests/unit/missions/application -run '^TestLeaveContract_(ReleasesThroughDepartureWriter|LastMemberStillSaves)$'` | T-3 RED | FAIL (expected) — behavioral: `departureCalls` 0; last-member guard green | `evidence/t3-application-red.txt` |
| `go test -count=1 ./tests/unit/missions/domain -run '^TestDecideRosterAcceptance_(RejectsMemberWithSameDefinition|TeammateLimitKeepsTeammateReason)$'` | T-4 RED | FAIL (expected) — bound teammate accepted today; limit-reason guard green | `evidence/t4-domain-red.txt` |
| `go test -count=1 ./tests/unit/missions/domain -run '^TestDecideRosterAcceptance_'` | T-5 GREEN narrow | PASS | `evidence/t5-domain-narrow-green.txt` |
| `go test -count=1 ./tests/unit/missions/domain` | T-5 domain standalone | PASS | `evidence/t5-domain-standalone-green.txt` |
| `go test -count=1 -v ./tests/unit/missions/application -run '<T-1..T-3 focused>'` | T-6 GREEN focused | PASS — all six T-1/T-2/T-3 tests green | `evidence/t6-application-green.txt` |
| `go test -count=1 ./tests/unit/missions/application` | T-6 application standalone | PASS | `evidence/t6-application-green.txt` |
| `go build ./...` | whole module | FAIL (expected) — `composition/server.go:52` needs `DepartureWriter`; `ContractStore` lacks `SaveDeparture` until T-10 | boundary gate scope |
| `go test -count=1 ./tests/integration/infrastructure/dynamodb -run '^TestCommitAcceptance_(ConcurrentSameDefinitionSerializes|DifferentPlayersSameDefinitionBothCommit)$'` | T-7 RED | FAIL (expected) — both racing commits succeed; different-players guard green | `evidence/t7-boundary-red.txt` |
| `go test -count=1 ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_(TerminalSaveReleasesOnlyActiveMembers|StaleReleaseKeepsNewerReservation|PendingValidationKeepsReservation)$'` | T-8 release RED | FAIL (expected) — no reservation release; stale save silent; pending guard green | `evidence/t8-release-red.txt` |
| `go test -count=1 ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_SaveDepartureReleasesDepartedGuards$'` | T-8 departure RED | FAIL (expected) — compile: `ContractStore.SaveDeparture` undefined | `evidence/t8-departure-red.txt` |
| `go test -count=1 ./tests/integration/composition -run '^(TestConcurrentAcceptance_SameDefinitionSerializes|TestTeamLeaveJourney_DepartedMemberRetakes)$'` | T-9 RED | FAIL (expected) — compile: composition cannot wire `DepartureWriter` (missing adapter method) | `evidence/t9-composition-red.txt` |
| `go test -count=1 ./tests/integration/infrastructure/dynamodb -run 'TestCommitAcceptance_\|TestContractStore_(TerminalSaveReleasesOnlyActiveMembers\|StaleReleaseKeepsNewerReservation\|PendingValidationKeepsReservation\|SaveDepartureReleasesDepartedGuards)'` | T-10 GREEN | PASS — all boundary tests green incl `ConcurrentLimitRace` | `evidence/t10-persistence-green.txt` |
| `go test -count=1 ./tests/integration/composition -run '^(TestConcurrentAcceptance_SameDefinitionSerializes\|TestTeamLeaveJourney_DepartedMemberRetakes)$'` | T-11 E2E GREEN | PASS — both public-API journeys green post-wiring | `evidence/t11-composition-green.txt` |
| `go test -count=1 ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_SaveApprovedStaleReleaseIsNotDuplicate$'` | SaveApproved partition RED | FAIL (expected) — stale owner conflict mapped to `ErrDuplicateConfirmation` | `evidence/t10-saveapproved-red.txt` |
| `go test -count=1 ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_(SaveApprovedStaleReleaseIsNotDuplicate\|.*Review.*\|.*Approved.*)$'` | SaveApproved partition GREEN | PASS — credit-range dup mapping preserved, guard release fails closed wrapped | `evidence/t10-saveapproved-green.txt` |
| `go test -count=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_ConcurrentDeparturesDoNotRestoreFormerMember$'` | FIND-0053-01 departure partition RED | FAIL (expected) — stale `SaveDeparture` returned nil; unconditional contract Put resurrects departed member | `evidence/t13-concurrent-departure-red.txt` |
| `go test -count=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_(ConcurrentDeparturesDoNotRestoreFormerMember\|SaveDepartureReleasesDepartedGuards\|TerminalSaveReleasesOnlyActiveMembers\|StaleReleaseKeepsNewerReservation)$'` + `go test -count=1 -v ./tests/integration/composition -run '^TestTeamLeaveJourney_DepartedMemberRetakes$'` | REG-0053-03 GREEN | PASS — conditional contract put rejects stale departure; sequential leave/retake E2E green | `evidence/t13-departure-green.txt` |
| `go test -count=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_SaveDeparture(RejectsMissingDepartedWithPriorCount\|RejectsSameCountDifferentDepartedSet\|AcceptsLegacyMissingDeparted)$'` | REG-0053-03 fence-edge RED | FAIL (expected) — missing `Departed` passes with expected prior count >0; same-count different departed set passes (size equality only). Legacy-missing guard green | `evidence/t13-departure-set-red.txt` |
| `go test -count=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_(SaveDepartureRejectsMissingDepartedWithPriorCount\|SaveDepartureRejectsSameCountDifferentDepartedSet\|SaveDepartureAcceptsLegacyMissingDeparted\|ConcurrentDeparturesDoNotRestoreFormerMember\|SaveDepartureReleasesDepartedGuards)$'` + `go test -count=1 -v ./tests/integration/composition -run '^TestTeamLeaveJourney_DepartedMemberRetakes$'` | REG-0053-03 same-set GREEN | PASS — `size`+`contains` per prior member rejects divergent sets and missing-attribute-with-priors; legacy-missing zero-prior and leave/retake E2E stay green | `evidence/t13-departure-set-green.txt` |
| `go test -count=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_SaveDeparture(RejectsMissingDepartedWithPriorCount\|RejectsSameCountDifferentDepartedSet\|AcceptsLegacyMissingDeparted)$'` (with index-0 `ConditionalCheckFailed` assertion) | REG-0053-03 cancel-reason hardening | PASS — kills the `attribute_exists` mutation survivor by asserting transaction cancellation, not any error | `evidence/t13-departure-cancel-reason.txt` |
| `go test -count=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestContractStore_StaleSaveDoesNotRestoreDepartedMember$'` | FIND-0053-02 candidate RED | FAIL (expected) — unconditional non-terminal `Save` put overwrites committed departure, stale write returns nil | `evidence/t13-stale-save-red.txt` |
| `go test -count=1 ./tests/unit/missions/domain` | T-12 domain standalone | PASS | `evidence/t12-domain.txt` |
| `go test -count=1 ./tests/unit/missions/application` | T-12 application standalone | PASS | `evidence/t12-application.txt` |
| `go test -count=1 -p=1 ./tests/integration/infrastructure/dynamodb` | T-12 boundary suite | PASS | `evidence/t12-dynamodb.txt` |
| `go test -count=1 -p=1 ./tests/integration/composition` | T-12 composition E2E | PASS | `evidence/t12-composition.txt` |
| `go test -count=1 -p=1 ./tests/integration/http` | T-12 public HTTP | PASS | `evidence/t12-http.txt` |
| `go build ./...` + `go vet ./...` | T-12 whole module | PASS | `evidence/t12-build-vet.txt` |
| `git diff --check` + manual whitespace scan of CHG-0053 files | T-13 cleanup inventory | PASS — no trailing whitespace; `gofmt -l` flags only non-0053 files `tests/integration/http/season_statistics_test.go`, `tests/unit/missions/domain/merit_ledger_test.go` | line inventory below |
| line-count inventory | T-13 cleanup inventory | production all <150; test files >149: `tests/unit/missions/interface/http/list_offers_test.go` (161), `tests/integration/http/contract_disputes_test.go` (151) — pre-existing size, 0053 added ~4 lines each; resolved by helper split (below) | |
| helper split + offer-eligibility tests | T-13 cleanup | PASS — `list_offers_test.go` 103 + new `list_offers_doubles_test.go` 72; `contract_disputes_test.go` 87 + new `contract_disputes_doubles_test.go` 75; new `mission_offer_definition_test.go` 99 (6 guard tests on `NewMissionOfferFromDefinition`); DOMAIN-RULES eligibility no longer names removed offer capacity | `evidence/t13-cleanup.txt` |
| `go test -count=1 ./tests/unit/missions/domain` / `./tests/unit/missions/application` | T-16 standalone units | PASS | `evidence/t16-domain.txt`, `evidence/t16-application.txt` |
| `go test -count=1 -p=1 ./tests/integration/infrastructure/dynamodb` / `./tests/integration/composition` / `./tests/integration/http` | T-16 standalone integration suites | PASS | `evidence/t16-infrastructure.txt`, `evidence/t16-composition.txt`, `evidence/t16-http.txt` |
| `go build ./...` / `go vet ./...` | T-16 whole module | PASS | `evidence/t16-build.txt`, `evidence/t16-vet.txt` |
| `go test -count=3 -p=1 -v ./tests/integration/composition -run '^(TestConcurrentAcceptance_SameDefinitionSerializes\|TestTeamLeaveJourney_DepartedMemberRetakes\|TestSoloLeaveJourney_LastMemberRetakes)$'` | T-16 critical E2E repeated | PASS 9/9 (3 tests × 3 runs) | `evidence/t16-critical-e2e.txt` |
| `tools/validate-sdd-change.sh` (temp-snapshot safe recipe, docker absent) | full gate set | exit 1 — only `tests/integration/infrastructure/s3` EvidenceStore tests fail (no S3 emulator without docker; unrelated scope). All unit + integration suites incl. dynamodb green; vet clean; gofmt flags only pre-existing non-0053 `season_statistics_test.go`/`merit_ledger_test.go`. Snapshot left at `/tmp/chg0053-validate-BKzXuF` | `evidence/t16-validator.txt` |
| mutation / coverage / critical E2E | L3 gates | all passed — mutation 10/10 killed (line 30 above); coverage passed, lead-measured 3198/3552 on `evidence/coverage-project-combined.out` (line 29); critical E2E 9/9 `t16-critical-e2e.txt` | |

## Regression Proof

- [x] REG-0053-01 tests fail on pre-fix code and pass post-fix.
- [x] REG-0053-02 tests fail on pre-fix code and pass post-fix.
- [x] Neighboring behavior (limit counts, roster history, terminal
      release) stays green — T-12 standalone suites all pass.

## Residual Risks

- Live non-terminal contracts predating deployment hold no reservation;
  the strict invariant is not enforced on them until a separately
  approved race-safe reconciliation — see `rollout.md`.
- Known out-of-scope defect: `submit` → `PENDING_VALIDATION` does not
  release `PLAYER_ACTIVE`; recorded in `change.md` for a future change.
- The departed-set fences cover membership divergence only; other
  concurrent contract-state mutations on the same row carry no
  optimistic versioning — recorded as residual risk in `design.md`.

## T-15 expansion RED→GREEN evidence (all executed; RED preserved, GREEN verified)

- `TEST-APP-LVE-002` — `go test -count=1 ./tests/unit/missions/application -run 'TestLeaveContract_'` (last-member leave → `SaveDeparture(contract, id, previousState)`; supersedes T-3 `LastMemberStillSaves`). **RED captured** — `departureCalls` expected 1, got 0 (terminal leave routed to generic `Save`); `evidence/t15-application-red.txt`. **GREEN** — all `TestLeaveContract_*` pass incl. `prevState == ContractInProgress` assertion; `LastMemberStillSaves` removed as superseded (covered by the new test's ABANDONED + no-generic-Save + single-departure-call assertions); unit http suite green; `evidence/t15-application-green.txt`. Expected compile blocker: `composition/server.go:52` arity + infra adapter 2-arg signature — boundary Gate 3 scope.
- `TEST-INF-REL-006` — `go test -count=1 ./tests/integration/infrastructure/dynamodb -run 'StaleSave'` (stale ordinary `Save` conflicts). **GREEN** — `StaleSaveDoesNotRestoreDepartedMember` passes on the `Save` departed-set fence (RED `t13-stale-save-red.txt` preserved); `evidence/t15-persistence-green.txt`. FIND-0053-02 fix landed.
- `TEST-INF-REL-007` — `go test -count=1 ./tests/integration/infrastructure/dynamodb -run 'StaleSaveApproved'` (stale approval conflicts atomically: no overwrite, no credits). **GREEN** — `StaleSaveApprovedDoesNotRestoreOrCredit` passes on the `SaveApproved` existing-row + departed-set fence (RED `t15-stale-approval-red.txt` preserved); `evidence/t15-persistence-green.txt`.
- `TEST-INF-REL-008` — `go test -count=1 ./tests/integration/infrastructure/dynamodb -run 'LastMember'` (last-member `SaveDeparture` releases leaver's reservation + `PLAYER_ACTIVE` slot). **GREEN** — `LastMemberSaveDepartureReleasesGuards` passes on the 3-arg `SaveDeparture(contract, id, previousState)` (compile RED `t15-last-member-red.txt` preserved); `evidence/t15-persistence-green.txt`.
- `TEST-E2E-MB-003` — `go test -count=1 ./tests/integration/composition -run 'SoloLeave'` (public solo last-member leave → board re-shows → retake under limit 1). **GREEN** — `SoloLeaveJourney_LastMemberRetakes` passes end-to-end (leave 200 → board `[offer-x]` → retake 201) after `server.go` ctor wiring; `DepartedMemberRetakes` + `SameDefinitionSerializes` stay green; `go build ./...` clean (RED `t15-solo-e2e-red.txt` preserved); `evidence/t15-composition-green.txt`.
- Test-quality fix (same session): `SaveApprovedStaleReleaseIsNotDuplicate` was passing via the index-0 `attribute_exists(PK)` fence for the wrong reason — arrange now persists the pre-approval row first so the conflict exercises the foreign-owner release guard; asserts the stored row keeps `IN_PROGRESS` (not absent).

## Incident — unauthorized local DynamoDB data deletion (2026-10-26)

During T-16, two unapproved deletion paths were identified and stopped:

- `TestContractStore_PersistsDeparturesAndFreesLeaverSlot` reused the
  deterministic local table `hbk-missions-TestContractStore_PersistsDeparturesAndFreesLeaverSlot`
  and called `clearMissionsTable`, which deleted prior local test rows —
  including a pre-fence `contract-t1` row — without permission. The test now
  uses a fresh high-entropy `chg0053-*` table via `chg0053Fixture` and never
  clears any table.
- `chg0053Fixture`/`chg0053Table` auto-`DeleteTable`d the per-test temporary
  tables they created via `t.Cleanup`. The cleanup and `DeleteTable` calls
  were removed entirely: tables minted by tests remain in place until a
  separately approved cleanup pass.

The reported suite results stand for behavior correctness (the fence was
verified working), but **full validation is suspended** until a safe table
lifecycle strategy is agreed and explicitly approved.

**Containment applied (2026-10-26):** every CHG-0053-authored DynamoDB/
composition test now provisions a fresh high-entropy `chg0053-*` table via
`chg0053Fixture`/`chg0053Table` (which never deletes), including
`concurrent_acceptance_e2e_test.go` and `team_leave_e2e_test.go` — no
CHG-0053 test touches a shared deterministic table or any
`DeleteTable`/`clearMissionsTable` path.

**Still open:** ~34 pre-existing non-0053 suite files across
`tests/integration/composition` and `tests/integration/infrastructure`
call `clearMissionsTable` against shared deterministic `hbk-missions*`
tables (helpers defined in `composition/server_test.go` and
`infrastructure/dynamodb/merit_ledger_store_list_test.go`), and three more
(`readyz_test.go`, `offer_store_list_test.go`, `dynamodb_client_test.go`)
issue `DeleteTable` directly — pre-existing convention, left untouched
pending approval.

**T-17 isolated-environment gate (2026-10-26) — PASSED, exit 0.** Under
explicit user authorization, dedicated throwaway containers were
provisioned (`chg0053-ddb-gate-20261026` → `127.0.0.1:18853`,
`chg0053-s3-gate-20261026` → `127.0.0.1:14583`) and
`tools/validate-sdd-change.sh` ran from a throwaway worktree snapshot
(`/tmp/chg0053-finalgate-jJY49V`, `.git` excluded) with `docker` absent
from `PATH` (compose-down could not fire) and `DYNAMODB_TEST_PORT=18853`
+ `S3_TEST_ENDPOINT=http://127.0.0.1:14583`. Result: every unit and
integration suite green — including `infrastructure/s3` (previously
environment-blocked) and the full race-detector pass; coverage profiles
were overwritten **only in the snapshot** (original root
`coverage-*.out` untouched, Oct 21 mtimes preserved); gofmt again listed
only the two pre-existing non-0053 files (script does not fail on
gofmt). Script reported coverage lines: domain 90.0%, application
90.9%, dynamodb 93.4% — lead independently verifies thresholds.
Evidence: `evidence/t17-isolated-validator.txt`. After checking the
container names, IDs and isolated ports, the two user-authorized
`chg0053-*-gate-20261026` containers were stopped and removed on
2026-10-26. The pre-existing DynamoDB Local on `:18000` remains running
and untouched. The temporary snapshot and coverage evidence remain.

**T-18 test-only coverage work (2026-10-26):** lead measured
project-wide coverage from the frozen T-17 snapshot profile at
`/tmp/chg0053-finalgate-jJY49V/coverage-project.out` and selected the
`internal/missions/interface/http` board-facet gap (highest mission-board
relevance). Characterization tests added for the existing
`ListOfferFacetsHandler` contract in
`tests/unit/missions/interface/http/list_offer_facets_test.go` —
distinct-facet success (sorted categories/modes over the resolved active
season), `status=INVALID` → 400 `INVALID_QUERY_PARAMETER`, no active
season → 422 `NO_ACTIVE_SEASON`, offer-reader failure → 500
`INTERNAL_ERROR`; real `application.NewListOfferFacetsHandler` +
existing `fakeOfferListReader`/`fakeActiveSeasons`/
`failingOfferListReader` doubles, 99 lines, no production or spec
changes. `go test -count=1 ./tests/unit/missions/interface/http` —
all green (`evidence/t18-board-facets.txt`). No coverage percentage
asserted here; lead recomputes from a fresh profile.

**T-18 follow-up (same session):** after lead's re-measurement showed
the first four cases added 13 unique statements (still four short), three
more season-resolution characterization cases landed in sibling
`list_offer_facets_season_test.go` (72 lines; the original file stays at
99): explicit `?season_id` bypasses active-season lookup and filters to
the bound season (200, only that season's category/mode), two active
seasons + no explicit → 422 `ACTIVE_SEASON_EXISTS`, and active-season
reader error → 500 `INTERNAL_ERROR` via new `failingActiveSeasons`
double. All 7 `TestListOfferFacetsHandler_*` pass; package green
(`evidence/t18-board-facets-extra.txt`). No production/spec changes;
coverage numbers remain lead-measured inputs, not asserted here.

**T-18 security branch (same session):** lead's union measurement
(3195/3552, two statements short) selected the untested fail-closed
branch in `accept_mission.go`: `CallerIdentityFromContext` error → 401
`UNAUTHORIZED` before the application is invoked.
`tests/unit/missions/interface/http/accept_identity_test.go` (31 lines)
characterizes it with `AcceptMissionHandler(nil)` and no identity
context — the nil application proves the boundary never dereferences it
(no panic = early return taken). `TestAcceptMissionHandler_MissingCallerIdentityReturns401`
passes; package green (`evidence/t18-accept-identity.txt`).

**T-19 second isolated full validation (2026-10-26) — PASSED, exit 0.**
Fresh dedicated containers (`chg0053-ddb-final-20261026` →
`127.0.0.1:18854`, `chg0053-s3-final-20261026` → `127.0.0.1:14584`,
`-sharedDb -inMemory`, LocalStack `SERVICES=s3`) and a fresh snapshot
`/tmp/chg0053-finalreview-dScGSy` (includes all T-18 tests + converged
artifacts). `tools/validate-sdd-change.sh` from snapshot, docker absent
from `PATH`, `DYNAMODB_TEST_PORT=18854`,
`S3_TEST_ENDPOINT=http://127.0.0.1:14584` — `=== SDD validation passed ===`
(`evidence/t19-validator.txt`): all unit + integration suites and race
detector green incl. `infrastructure/s3`. The lead-authored
project-wide profile suite then ran from the same snapshot —
`go test -count=1 -p=1 -coverpkg=./internal/...
-coverprofile=coverage-project-final.out ./tests/unit/...
./tests/integration/...` — exit 0 (`evidence/t19-project-cover-suite.txt`);
profile at `/tmp/chg0053-finalreview-dScGSy/coverage-project-final.out`.
Independent block-wise measurement: 3198/3552 (90.03%) across all
production packages. Original root `coverage-*.out` remained byte-identical.
After confirming exact names, IDs and ports, the two user-authorized
T-19 containers were stopped and removed. The T-17 containers had already
been removed; the original `:18000` DynamoDB and local `:8000` server
remain untouched.

## Authorized Local Contract Cleanup

The user authorized deleting old local contracts and then confirmed the
narrower operation: remove only two `ABANDONED` contracts and the empty
`PLAYER_ACTIVE` record from `hbk-missions-local` on `localhost:8000`,
preserving its catalog and season. A fresh, strongly consistent key/state
inventory matched the original three-row inventory. One conditional
`TransactWriteItems` deleted the contracts only while `State=ABANDONED`
and the active record only while `contract_ids` remained absent. A
strongly consistent post-write scan verified 112 to 109 rows with exactly
those three keys removed: 72 `DEFINITION`, 36 `OFFER`, and 1 `SEASON`
remain. No other table was targeted; no server reset or production
migration occurred. This deletion is irreversible.

## Lead Review Before Canonical Merge

- Reviewer: lead engineer, separate from the implementation author.
  Reviewed every production/test landing, the final delta against both
  owning capabilities, and the T-19 isolated validator and complete
  project coverage profile. Source baseline:
  `efa42234f4242e2680b92b024e72d8f6d24c68ed` plus the reviewed
  uncommitted worktree; no commit or PR status is claimed.
- No open CHG-0053 code finding remains. The known
  `PENDING_VALIDATION` active-slot defect, general non-membership
  concurrency and reservation-less contracts outside local are
  explicitly separate follow-ups; no shared rollout is approved.
- Candidate merge: update only `REQ-MB-004`/`REQ-MB-008` in
  `specs/capabilities/mission-board/spec.md` and
  `REQ-MCF-010`/`REQ-MCF-013`/`REQ-MCF-020` in
  `specs/capabilities/mission-contract-fulfillment/spec.md`; preserve
  the existing MCF wording and scenarios while appending the delta's
  write-fence and last-member clauses/scenarios. Update `specs/index.md`
  and archive under
  `specs/changes/archive/2026/CHG-0053-fix-active-template-exclusion/`.
  Preserve pre-existing, uncommitted canonical edits and all CHG-0051
  work. No durable `specs/context/` convention exists, so
  `context_scope: none`.
- Final verify-spec decision and consolidation remain pending explicit
  user approval.
