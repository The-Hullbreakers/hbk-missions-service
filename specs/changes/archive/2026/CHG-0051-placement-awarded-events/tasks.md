---
change_id: CHG-0051
artifact: tasks
status: proposed
updated: 2026-10-20
---

# Tasks — CHG-0051

> Inside-out: Application RED first -> demanded Domain -> Application
> GREEN/BLUE -> Boundary RED -> adapters/composition.

## Wave 1 — Application RED
- [x] **T-1 — Application RED: finalize intends placement facts.** Owner:
  agent. Deps: none. Outcome: `FinalizeSeasonHandler` tests fail — official
  season intends one fact per policy-ranked entry, unofficial intends none,
  intents passed to the same writer call (port doubles assert). Test:
  `tests/unit/missions/application/finalize_season_test.go`. Maps to
  `internal/missions/application/finalize_season.go`, demanded domain files.
  RED evidence 2026-10-20: build fails — `undefined: domain.PlacementFact`
  + constructor arity. Next: Gate 3-APPLICATION -> T-2.

## Wave 2 — Demanded Domain
- [x] **T-2 — Domain RED: adjudication policy + placement fact.** Deps:
  T-1 RED. Outcome: `NewPlacementFact` tests fail — deterministic
  `event_id`, TEAM requires non-empty `member_ids`, rank >= 1. Test:
  `tests/unit/missions/domain/placement_fact_test.go`. RED evidence
  2026-10-20: `undefined: domain.NewPlacementFact`. Gate 3-DOMAIN approved.
  Next: T-3.
- [x] **T-3 — Domain GREEN/BLUE + LAYER-GATE-DOMAIN.** Deps: T-2 evidence.
  Outcome: `PlacementFact` + `NewPlacementFact` in
  `internal/missions/domain/placement_fact.go`; `go test -run PlacementFact
  ./tests/unit/missions/domain` PASS 2026-10-20 (3 tests). Next: T-4.

## Wave 3 — Application GREEN
- [x] **T-4 — Application GREEN/BLUE + publisher use case.** Deps: T-3 +
  T-1 evidence. Outcome: `FinalizeSeasonHandler` rewired
  (buckets/standings/rosters/atomic `FinalizedSeasonPersister`);
  `PlacementFactPublisher` with `PendingIntentReader` +
  `DomainEventPublisher` + `IntentStateWriter` ports; deterministic
  `event_id`; failed publish leaves intent pending. `SeasonStore.
  SaveFinalized` (TransactWriteItems season+outbox) and
  `DenyAllReader.Members` fail-closed landed early because composition
  must compile — their integration evidence stays in T-5/T-6. GREEN
  evidence 2026-10-20: `go test ./tests/unit/missions/application` PASS
  standalone after user refactor converged (11 finalize/publisher tests).
  Next: T-5.

## Wave 4 — Boundary RED
- [x] **T-5 — Boundary RED: outbox persistence + SNS publish.** Deps: T-4.
  Outcome: `tests/integration/infrastructure/dynamodb/outbox_intents_test.go`
  + `season_store_finalize_test.go` + `season_store_finalize_guard_test.go`
  + `contract_store_roster_test.go` — finalize persists intents atomically
  (CAS on `State = FREEZING`, `ErrOutboxBatchTooLarge` past the 100-item
  transaction ceiling, `occurred_at`/`correlation_id` in the envelope),
  `ListPending` honours the bound and pages past a PUBLISHED prefix,
  `MarkPublished` is conditional/idempotent with no phantom row, and
  `ContractStore.Members` returns the frozen season-scoped roster union;
  `tests/integration/infrastructure/sns/publisher_test.go` — publisher
  emits the exact envelope with `event_type` AND `source` attributes and
  propagates SNS failures. RED evidence 2026-10-20: build fail —
  `ListPending`/`MarkPublished` undefined, `infrastructure/sns` package
  missing, `NewSeasonScheduler` arity. Gate 3-BOUNDARY approved by user.
  Next: T-6.

## Wave 5 — Infrastructure/Composition
- [x] **T-6 — DynamoDB outbox + SNS adapter + wiring.** Deps: T-5 RED.
  Outcome: `SeasonStore.SaveFinalized` transacts season+intents with a
  `State = FREEZING` compare-and-swap (missing season ->
  `ErrSeasonNotFound`, conflict -> `ErrInvalidSeasonStateTransition`,
  PUBLISHED intents never reset); `season_store_outbox.go` implements
  `ListPending` (paginates past PUBLISHED prefix) + conditional
  `MarkPublished`; `internal/missions/infrastructure/sns` publishes via
  cached `aws-sdk-go-v2/service/sns` v1.39.17 with `event_type`/`source`
  attributes; TEAM `member_ids` resolve from `ContractStore.Members`
  (frozen contract rosters, union per season); `SeasonScheduler.Tick`
  drains the `PendingIntentDrainer` port and `Start` reports tick errors
  to an `onError` callback (composition logs them); composition wires the
  publisher only when `DomainEvents` is configured
  (`SNS_TOPIC_ARN`/`SNS_ENDPOINT_URL`); `template.yaml` grants
  `sns:Publish` on `SharedEventsTopicArn`. GREEN evidence 2026-10-20/21:
  `go test -count=1` PASS for `./tests/unit/...`,
  `./tests/integration/infrastructure/{dynamodb,sns}`,
  `./tests/integration/composition` against DynamoDB Local :18000.
  Residual: >99 ranked facts cannot finalize (DynamoDB 100-item
  transaction cap; needs an approved transaction-model change) — see
  verification.md. Next: T-7 L3 gates.

## Wave 6 — Revised Finalization Rework (Gate 2 re-approved 2026-10-21)
> Two-phase adjudication replaces the all-in-one transaction. Existing
> Gate 3 approvals do NOT authorize the new production below.
- [x] **T-9 — Application RED: two-phase finalization rework.** Owner:
  agent. Deps: T-4/T-6 evidence preserved. RED 2026-10-27: focused
  BUILD FAIL — intended missing shape (evidence in verification.md);
  TEST-APP-FIN-003/004 + TEST-APP-PUB-001 + TEST-APP-MCF-001 (MCF: 5
  characterization tests PASS named-file, clone-per-Get proof). Gate
  3-APPLICATION approved 2026-10-27 (T-10 Domain RED only); legacy
  model tests superseded. Next: T-10.
- [x] **T-10 — Demanded Domain RED→GREEN: attempt identity + staged-set
  verification + credited-roster rule.** Deps: T-9 RED. Gate 3-DOMAIN
  approved 2026-10-27. RED->GREEN 2026-10-27: missing-shape BUILD FAIL
  then named-file + full `go test ./tests/unit/missions/domain` PASS.
  `adjudication_attempt.go` (`ADJ#`, lock-supplied cutoff),
  `staged_fact_set.go` (count/hash snapshot, Verify),
  `credited_roster.go` (dedup sorted union), `hash_fields.go` framing,
  `adjudication_errors.go` sentinels, `placement_fact.go` Digest +
  MemberIDs clone; BLUE clean +5 focused regression tests 2026-10-27;
  coverage/mutation/E2E pending. Next: T-11.
- [x] **T-11 — Application GREEN/BLUE (approved 2026-10-27).** Deps:
  T-10. 7-port `SeasonFinalizer` + manifest publisher (5-port) +
  `adjudication_ports.go`/`placement_manifest.go` landed; legacy tests
  migrated; named-file PASS 27 funcs. Full application suite GREEN
  2026-10-27 (91.1% coverage); revised LAYER-GATE-APPLICATION passed.
  Composition/HTTP still blocked until their boundary cycles. Next: T-12.
- [x] **T-12 — Boundary RED: staged facts, fence, manifest, publisher.**
  Deps: T-11. Outcome: conditional `PutItem` staging
  (`attribute_not_exists`, digest per row; identical -> no-op,
  divergent -> `ErrStagedFactConflict`; bounded parallelism); same-write
  lock fence on `merit_ledger_save.go`/`contract_store_review.go`/
  `contract_store_accept.go` mutations (`ErrSeasonAdjudicating`;
  raced approval all-or-nothing, post-lock contract+ledger unchanged;
  zero-credit reject allowed); strong-consistent snapshot reads;
  2-item commit tx (CAS + deterministic MANIFEST — no duplicate
  publishable manifest); progress marking. Tests:
  `tests/integration/infrastructure/dynamodb/` — TEST-INF-FACT-001/-002,
  TEST-INF-FENCE-001, TEST-INF-MCF-001, TEST-INF-SNAP-001. PARTIAL:
  TEST-INF-ADJ-001 GREEN 2026-10-28 (scoped approval, 2 tables intact);
  TEST-INF-FACT-001/-002 GREEN 2026-10-28 (scoped approval,
  `season_store_facts.go`; 3/3 single run, 3 tables intact).
  Fence/snapshot/manifest partitions pending Gate 3-BOUNDARY.
  User approved continuing T-12/T-13 in their own CHG-0051
  scope on 2026-11-04 after local API compilation exposed
  the obsolete 3-port publisher call. First remaining
  partition: TEST-INF-MAN-001 — drain only a committed
  manifest with real `SeasonStore`, never publish orphaned
  staged facts, and persist progress/completion. Integration
  scope `tests/integration/infrastructure/dynamodb/season_manifest_publisher_test.go`;
  retry/progress case in `season_manifest_resume_test.go`;
  standalone RED:
  `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestManifestPublisher_' -v`.
  Expected RED: four missing manifest reader/progress methods,
  not fixture/emulator errors. `application: not_affected`,
  LAYER-GATE-APPLICATION passed. Mapped GREEN files after
  a separate scoped Gate 3-BOUNDARY:
  `internal/missions/infrastructure/dynamodb/season_store_manifest_read.go`
  and `season_store_manifest_progress.go`; composition wiring
  needs a subsequent RED/Gate 3 and cannot be altered yet.
  Actual RED 2026-11-04: two test files (129/85 lines)
  fail solely because `SeasonStore` lacks all four new
  manifest reader/progress methods. The three cases
  require committed-only publication, frozen-cutoff
  envelope/progress/completion, and retry from
  `NextIndex=1` without republishing the first fact.
  No test fixture was executed and no production changed.
  Gate 3-BOUNDARY for only these four SeasonStore
  methods approved 2026-11-04; fence/snapshot/commit
  and composition remain gated.
  Manifest-reader/progress GREEN/BLUE 2026-11-04:
  three original scenarios plus an integrity case
  PASS 4/4. The additional case was RED at runtime
  on a noncanonical staged SK with a valid digest;
  canonical key verification closed it. Five existing
  staging/lock regressions pass. A timestamp-only
  parallel-table fixture collision was fixed in
  a private 27-line manifest fixture, not production.
  Files 143/147/27/129/84/90 lines, gofmt clean.
  T-12/T-13 remain open for fence/snapshot/commit.
  After the fence partition, under the already approved revised
  Gate 2: TEST-INF-MAN-002 — real `SeasonFinalizer`
  stages 200 facts, verifies the persisted count/hash,
  then commits exactly one deterministic PENDING manifest
  and the `FINALIZED` season via a two-item transaction;
  a missing staged fact leaves the locked FREEZING season
  without a publishable manifest. Tests:
  `tests/integration/infrastructure/dynamodb/season_manifest_commit_test.go`
  and `season_manifest_verify_test.go`, integration_scope:
  infrastructure, standalone command
  `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestSeasonManifestCommit_' -v`.
  Application gate passed; Domain/Application not affected.
  Expected RED: missing `NewStagedFactVerifier` and
  `NewFinalizationCommitter`, not setup or 100-item runtime
  limit. Mapped production after separate Gate 3:
  `season_store_fact_verify.go` and
  `season_store_manifest_commit.go`. Fence/snapshot
  and composition remain separate pending partitions.
  Fence partition starts with TEST-APP-MCF-002, an
  Application RED demanding the provider-neutral
  `ErrSeasonAdjudicating` signal on a locked-season
  confirmation whose writer rejects the save. The
  existing `ConfirmObjectiveHandler` must propagate the
  rejection without committing the staged credit.
  A demanded Domain RED/Gate 3 for the missing sentinel
  precedes any boundary implementation; afterward
  TEST-INF-FENCE-001 must prove same-write fencing
  against real DynamoDB Local. Source test:
  `tests/unit/missions/application/adjudication_fence_test.go`;
  standalone command:
  `go test -count=1 ./tests/unit/missions/application -run '^TestConfirmObjective_AdjudicationFence' -v`.
  This is a new Application partition, not permission
  to change Domain, Application, DynamoDB, or HTTP yet.
  TEST-DOM-ADJ-004 then distinguishes the new
  adjudication-lock error from an inactive season or
  invalid season-state transition. Test path
  `tests/unit/missions/domain/adjudication_fence_error_test.go`;
  standalone `go test -count=1 ./tests/unit/missions/domain -run '^TestSeasonAdjudicatingError_' -v`.
  The Domain RED maps only to
  `internal/missions/domain/errors.go` after its own
  Gate 3-DOMAIN; no Application production edit is
  required if the use case already propagates the
  writer's named error unchanged.
  Application standalone RED 2026-11-04: two absent
  `domain.ErrSeasonAdjudicating` references in a 34-line
  test. Demanded Domain standalone RED: the same
  missing sentinel at two sites in a 29-line test.
  No production edit; scoped Gate 3-DOMAIN is requested
  before adding the declaration to `domain/errors.go`.
  Outer same-write fence needs its own RED and Gate 3.
  Gate 3-DOMAIN was approved 2026-11-04. One
  `ErrSeasonAdjudicating` sentinel (errors.go 65
  lines, gofmt clean) made Domain and the
  correctly anchored Application test GREEN; both
  full affected unit packages PASS. The use case
  already forwards writer errors, so no Application
  production file changed. Next: TEST-INF-FENCE-001
  boundary RED; same-write DynamoDB and HTTP
  mapping remain pending Gate 3-BOUNDARY.
  TEST-INF-FENCE-001 boundary RED reviewed: the real
  `ConfirmObjectiveHandler` and stores on a fresh
  private table FAIL after lock (nil error and raw
  contract gains PendingCredits), while FREEZING
  pre-lock confirmation and zero-credit post-lock
  reviewer rejection both PASS. Test files
  `season_adjudication_fence_test.go` (75 lines),
  `season_adjudication_fence_prelock_test.go` (53),
  `season_adjudication_fence_reject_test.go` (60).
  Standalone `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationFence_' -v`.
  Gate 3-BOUNDARY is requested only for
  `ContractStore.Save` in `contract_store.go` and
  `contract_store_adjudication_fence.go`: same-write
  season-lock condition for credit-bearing contract
  saves, preserving departure checks and zero-credit
  rejection. Missing-season legacy fixtures must
  seed a season, not weaken the business fence.
  Approval-ledger, snapshot, commit, and composition
  remain separately gated.
  Gate 3-BOUNDARY approved 2026-11-04 for this
  `ContractStore.Save` partition. The atomic
  season ConditionCheck + unchanged departure
  CAS yields TEST-INF-FENCE-001 **3/3 PASS**
  on private :18000 tables. Production files
  `contract_store.go` 138 lines and
  `contract_store_adjudication_fence.go` 69
  lines, gofmt clean. One CHG-0053 private
  stale-approval fixture required an ACTIVE
  season row before staging credit; after
  this test-only setup fix its stale-approval
  and stale-departure regressions PASS 2/2.
  Legacy fixed-table tests that clear rows
  were deliberately not run. T-12/T-13
  still open for remaining fences, snapshot
  and commit before any composition/API work.
  Checkpoint for the next T-12 RED:
  `handoffs/context-checkpoints/CHECKPOINT-20261104-123038.md`.
  Continued T-12 TEST-INF-MCF-001 approval partition:
  `season_adjudication_fence_approval_test.go` (64
  lines) and `season_adjudication_fence_approval_prelock_test.go`
  (56) use real reviewer use cases + stores on fresh
  tables. Standalone `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationApprovalFence_' -v`
  yields 1 PASS/1 FAIL. Pre-lock FREEZING approval
  completes and credits once. After lock approval
  returns nil, persists COMPLETED, and writes one
  credit: exact RED for absent same-transaction
  guard in `ContractStore.SaveApproved`. Gate
  3-BOUNDARY is requested only for that approval
  transaction in `contract_store_review.go` plus a
  focused helper split for the <150-line rule.
  Preserve departed-set CAS and correct duplicate
  cancellation indexes when adding the condition.
  MeritLedgerStore.Save, acceptance, snapshot,
  commit, composition and HTTP are separate.
  Gate 3-BOUNDARY approved 2026-11-04 for this
  `SaveApproved` partition only. Same-write
  season ConditionCheck now guards positive-credit
  approval; shifted credit cancellation indexes
  still return `ErrDuplicateConfirmation`. Existing
  departed-member CAS is unchanged. Approval
  matrix 3/3 PASS including a duplicate-credit
  regression; five adjacent private-table fence
  and stale-departure cases PASS in the combined
  run. `contract_store_review.go` 134 lines,
  `contract_store_adjudication_fence.go` 87;
  tests 64/56/67 lines; gofmt clean. Broader
  fixed-table tests would clear existing rows,
  so are not counted as GREEN. T-12/T-13 remain
  open for ledger/accept, snapshot/commit,
  composition and HTTP mapping.
  Direct-ledger TEST-INF-FENCE-001 RED reviewed:
  `season_adjudication_ledger_fence_test.go` (83
  lines) on fresh private :18000 tables proves
  FREEZING pre-lock `MeritLedgerStore.Save`
  PASS and post-lock direct Save FAIL with nil
  error and a written credit. Standalone
  `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationLedgerFence_' -v`.
  Gate 3-BOUNDARY requested only for
  `merit_ledger_save.go` plus a focused reuse
  of the existing package-private season-lock
  helper. No production code changed for this
  partition. Acceptance, snapshot, commit and
  composition remain separately gated.
  Scoped Gate 3-BOUNDARY approved 2026-11-04
  for `MeritLedgerStore.Save` only. Its credit
  Put now shares a transaction with the
  factored season-unlocked check; direct
  ledger tests PASS 2/2 and the combined
  private-table matrix PASS 10/10 (two
  ledger, three approval, three staging,
  two CHG-0053 stale-write). Helper/store
  files 94/69 lines, gofmt clean. Existing
  fixed-table tests with destructive setup
  were not run. T-12/T-13 continue with
  acceptance, snapshot and commit after
  their own RED/Gate 3. No API startup.
  Acceptance TEST-INF-FENCE-001 RED is
  classified: five private-table cases
  `TestAdjudicationAcceptanceFence_`
  (50/50/56/57/48-line files) yield 3
  PASS / 2 FAIL. ACTIVE pre-lock commit
  and existing definition/limit error
  priority pass; a stale ACTIVE read
  followed by FREEZING without lock or
  by the acquired lock erroneously
  commits contract/player/definition
  guard rows. Gate 3-BOUNDARY requested
  only for `contract_store_accept.go`
  and a focused package-private
  ACTIVE+unlocked season condition in
  the same acceptance transaction.
  Credit writes retain their distinct
  ACTIVE-or-FREEZING unlocked rule.
  Snapshot, commit and composition
  remain separately gated.
  Scoped Gate 3-BOUNDARY approved 2026-11-04.
  `CommitAcceptance` now checks ACTIVE and
  unlocked in the same transaction as its
  contract/player/definition writes. A
  first GREEN run was 14/15: FREEZING
  failed safely but mapped to a generic
  error; state-specific mapping fixed it.
  Final private-table run PASS 15/15
  (5 acceptance + 10 neighboring cases).
  Reservation-before-limit error priority
  still passes. `contract_store_accept.go`
  90 lines and shared fence helper 109,
  gofmt clean. Full legacy tests with
  destructive fixture setup not run.
  T-12/T-13 remain open for snapshot,
  manifest commit, composition and HTTP.
  Staged-verifier TEST-INF-FACT-002 RED reviewed:
  `season_staged_fact_verifier_test.go` (94 lines)
  and `_integrity_test.go` (50) require complete
  set verification without a manifest, missing
  row mismatch, and rejection of tampered merits
  with an unchanged digest. Standalone
  `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestStagedFactVerifier_' -v`
  fails solely on three references to the absent
  `NewStagedFactVerifier`; no emulator setup ran.
  Gate 3-BOUNDARY requested only for
  `season_store_fact_verify.go` and reuse of
  attempt-scoped consistent staged validation.
  Commit, snapshot and composition remain gated.
  Scoped Gate 3-BOUNDARY approved 2026-11-04:
  `season_store_fact_verify.go` (42 lines)
  verifies expected count/hash, canonical
  staged identities/digests via the existing
  strongly consistent attempt reader and
  creates no manifest. Three verifier and
  four manifest-publisher cases PASS 7/7
  on private :18000 tables; tests 94/50
  lines, gofmt clean. Commit/snapshot
  and composition remain separate RED/Gate 3
  partitions; no API startup or app-table
  data mutation occurred.
  TEST-INF-MAN-002 commit RED reviewed:
  `season_manifest_commit_test.go` (87 lines)
  and `season_manifest_commit_lock_test.go`
  (52), plus `season_manifest_commit_retry_test.go`
  (85), compiled RED only because
  `NewFinalizationCommitter` was missing.
  Scoped Gate 3-BOUNDARY approved 2026-11-05:
  `season_store_manifest_commit.go` (83
  lines) commits FINALIZED + one
  deterministic PENDING manifest in a
  two-item transaction guarded by the
  FREEZING+lock/cutoff CAS and
  attribute_not_exists manifest Put.
  Tests PASS 10/10 (3 commit + 3
  verifier + 4 publisher regressions)
  on private :18000 tables; gofmt clean.
  Remaining T-12/T-13 partitions:
  snapshot reader, composition/HTTP —
  each its own RED/Gate 3.
  TEST-INF-SNAP-001 snapshot GREEN/BLUE:
  scoped Gate 3 approved 2026-11-06.
  `season_store_snapshot.go` (130 lines)
  aggregates MERIT_CREDIT via consistent
  base-table reads (never the GSI) into
  deterministic buckets with credited
  contract ids; 1 snapshot + 10 adjacent
  cases PASS 11/11 on private :18000.
  Credited roster GREEN/BLUE (scoped
  Gate 3, 2026-11-06): adapter reads
  each credited contract consistently
  and unions frozen rosters; 1 case +
  10 adjacent PASS 11/11; fixture fixed
  to frozen in-progress team contracts.
  Composition GREEN/BLUE (scoped Gate 3,
  2026-11-06): five-port publisher and
  two-phase SeasonFinalizer wired;
  `FinalizerRunner` interface keeps both
  finalizers valid; `go build` clean,
  composition+scheduler tests PASS;
  local service up on :3006 (healthz
  200, authenticated GETs 401 as OIDC
  is unconfigured — expected). Next:
  HTTP `lang` verification via real
  gateway login, then L3 gates. No
  commit, snapshot or composition
  production edit is authorized yet.
- [x] **T-13 — Infrastructure GREEN/BLUE + integration evidence.** Deps:
  T-12 RED. Maps to `season_store_lock.go`, `season_store_facts.go`,
  `season_store_finalize.go` (rework), `season_store_outbox.go`,
  `contract_store_credited.go` (new), `merit_ledger_save.go` +
  `contract_store_review.go` + `contract_store_accept.go` (fence
  conditions). Next: T-14.
- [x] **T-14 — Orphaned-lock sweep + operator reconcile.** Deps: T-13.
  Outcome: scheduler sweep lists locked seasons past threshold (no new
  web route). Maps to `season_scheduler.go` + lock listing query.

## Wave 7 — Gates And Convergence
- [x] **T-7 — L3 gates.** Deps: T-1..T-6 + T-9..T-14. Outcome: security
  review (fenced-write boundary), coverage >=90% incl. rework scope,
  mutation (`mutation-report.md`), critical E2E (`critical-e2e.md` —
  TEST-E2E-AWD-002: 200 entries finalize -> manifest -> 200 facts),
  clean-up <=150 lines, verification.md.
- [x] **T-8 — Docs + independent review + consolidate.** Deps: T-7.
  Outcome: EVENT-CATALOG snake_case alignment; handoff review resolved;
  verify; merge delta; archive.
- TEST-HTTP-ADJ-001 public mapping GREEN/BLUE (scoped Gate 3,
  2026-11-16): RED `TestPublicCode_MapsDomainErrors` expected
  `SEASON_ADJUDICATING`, got `INVALID_REQUEST`; GREEN one `case` in
  `errors.go` → `422 SEASON_ADJUDICATING` via existing
  `writeSeasonError`. Suite PASS, `go build` clean, no BLUE needed.
  Fixture fix (unrelated pre-existing RED): publish definition HTTP
  test fake now supplies `ContentEn` required since CHG-0054.
- L3 GATE PROGRESS (2026-11-16): TEST-HTTP-ADJ-001 GREEN — `422
  SEASON_ADJUDICATING` public mapping. Fixture repair wave: stale
  DynamoDB/composition fixtures aligned with CHG-0051 season fence +
  CHG-0054 English-snapshot invariants (new shared `englishize` +
  `seedActiveSeason` test helpers; `ContractStore.Get` now returns
  `domain.ErrContractNil` so detail 404 maps `CONTRACT_NOT_FOUND`).
  `tools/validate-sdd-change.sh` PASS: gofmt/vet clean, unit +
  integration + race all green; coverage domain 92.4%, application
  91.7%, dynamodb package 62.2% (whole-package figure incl. unrelated
  adapters — affected-scope record kept in verification.md).
  TEST-E2E-AWD-002 critical E2E PASS (200 entries → manifest → 200
  unique intents). Mutation report completed via manual analysis
  (repo convention, tools unavailable). Security review closed.
  Remaining: T-14 orphan-lock sweep, T-8 docs + consolidation +
  archive (needs explicit approval).
- T-14 orphan-lock sweep GREEN/BLUE (scoped Gate 3, 2026-11-16):
  RED `TEST-APP-SWP-001`/`TEST-INF-SWP-001` compile-RED → GREEN.
  Application: `LockedSeason` + `AdjudicationLockLister` +
  `OrphanedLockReporter` ports (`adjudication_ports.go`, 90 lines),
  `SeasonScheduler.WithOrphanSweep` + `sweepOrphanedLocks`
  (`season_scheduler_sweep.go`, 40 lines) — locks past 24h threshold
  are reported, never unlocked automatically; lister failure fails the
  tick. Infrastructure: `AdjudicationLock.ListLocked` consistent-query
  on `SEASON` partition filtered by `attribute_exists(AdjudicationLock)`;
  reminted-attempt check rejects corrupt rows. Composition wires
  `orphanLogReporter` (service log, no new web route per design).
  Unit 2/2 + infra 1/1 + composition regression green; all files <150.

- ARCHIVED 2026-11-16 (user-approved consolidation): delta merged into
  `specs/capabilities/seasonal-category-competition` (REQ-SCC-004
  extended + LOCK scenarios + new REQ-SCC-018 with AWARD scenarios +
  two invariants) and `mission-contract-fulfillment` (REQ-MCF-010
  extended + ADJ scenarios, REQ-MCF-012 and REQ-MCF-020 fence
  extensions). `specs/index.md` updated; change moved to
  `specs/changes/archive/2026/CHG-0051-placement-awarded-events`.
