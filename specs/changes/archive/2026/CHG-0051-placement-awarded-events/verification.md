---
change_id: CHG-0051
artifact: verification
status: pending
updated: 2026-10-27
---

# Verification — CHG-0051

## Gates

| Gate | Scope | Decision | Evidence | Date |
| --- | --- | --- | --- | --- |
| Gate 1 | spec writes | approved | user approval 2026-10-20 | 2026-10-20 |
| Gate 1 (expansion) | canonical ownership + mission-contract-fulfillment (REQ-MCF-010/012/020) | approved | user approval 2026-10-21; `canonical_paths` updated, MCF delta added | 2026-10-21 |
| Gate 2 | start RED | approved | user approval 2026-10-20 | 2026-10-20 |
| Gate 3-DOMAIN | adjudication policy | approved | `go test -run PlacementFact ./tests/unit/missions/domain` PASS (T-3) | 2026-10-20 |
| Gate 3-APPLICATION | finalize intent + publisher | RED evidence shown | build fail: `domain.PlacementFact` + ctor arity | 2026-10-20 |
| Gate 3-BOUNDARY | outbox + SNS | approved by user | RED: `ListPending`/`MarkPublished` undefined, `infrastructure/sns` missing, `NewSeasonScheduler`/`Start`/`Members` arity | 2026-10-20 |
| Gate 2 (revised) | two-phase adjudication redesign | re-approved by user | revised design: lock+cutoff, staged facts, manifest commit | 2026-10-21 |
| Gate 2 (MCF scope) | REQ-MCF-010/012/020 lock-fence behaviors | approved | user approval 2026-10-21; allows T-9 TEST-APP-MCF-001 to start RED — Gate 3-APPLICATION/BOUNDARY still pending, no production authorized | 2026-10-21 |
| Gate 3 (revised scope) | rework tasks T-9..T-14 | pending | existing Gate 3 approvals do NOT authorize new production | 2026-10-21 |
| Gate 3-APPLICATION (revised) | T-9 Application RED → T-10 Domain RED | approved | user approval 2026-10-27; authorizes T-10 Domain RED only — no production | 2026-10-27 |
| Gate 3-DOMAIN (revised) | T-10 Domain RED → domain production | approved | user approval 2026-10-27 for TEST-DOM-ADJ-002/003; domain production only — application/boundary still gated | 2026-10-27 |
| Gate 3-APPLICATION production (revised) | T-11 application production | approved (scoped) | user instruction 2026-10-27: 'analiza e incrementa el test coverage siguiendo nuestros estandares y despues continua' — coverage step done (lead-measured 91.1% domain), 'continua' lead-interpreted as T-11 application-production approval ONLY; boundary NOT approved | 2026-10-27 |
| Gate 3-BOUNDARY (revised, partitions 1-2) | TEST-INF-ADJ-001 lock + TEST-INF-FACT-001/-002 staging | approved (scoped) | 2026-10-28: lock-adapter GREEN/BLUE + single run of two focused tests (two unique tables left intact); staging-adapter GREEN/BLUE + single run of three focused tests (three unique tables left intact); fence/snapshot/manifest/publisher partitions NOT approved | 2026-10-28 |
| Gate 2 continuation (manifest partition) | T-12 — TEST-INF-MAN-001 committed-only publisher read and progress | approved (tests only) | user approved continuing CHG-0051 T-12/T-13 on 2026-11-04; revised Gate 2 already covered T-12 RED; no Gate 3 for manifest adapters or composition, no local service startup yet | 2026-11-04 |
| Gate 3-BOUNDARY (manifest partition) | T-12 TEST-INF-MAN-001 → T-13 manifest reader/progress adapters | approved (scoped) | user approval 2026-11-04 after two-file compile RED: implement only four manifest reader/progress methods on SeasonStore; commit/fence/snapshot, composition and local service startup excluded | 2026-11-04 |
| Gate 3-DOMAIN (adjudication fence error) | TEST-APP-MCF-002 → TEST-DOM-ADJ-004 named business error | approved (scoped) | user approval 2026-11-04; only a new sentinel in `domain/errors.go`, no adapter/composition/HTTP change | 2026-11-04 |
| Gate 3-BOUNDARY (contract confirmation lock fence) | T-12 TEST-INF-FENCE-001 → T-13 credit-bearing `ContractStore.Save` | approved (scoped) | user approval after 3 private-table cases: locked confirmation FAIL with nil error/raw contract drift, pre-lock confirm and post-lock zero-credit reject PASS; only `contract_store.go` plus adjudication fence helper authorized | 2026-11-04 |
| Gate 3-BOUNDARY (reviewer approval lock fence) | T-12 TEST-INF-MCF-001 → T-13 `ContractStore.SaveApproved` | approved (scoped), GREEN/BLUE | user approval after 2 private-table cases; same-write season check added for positive credits, with shifted duplicate indexes and 3/3 approval cases + 5 adjacent private-table cases PASS; ledger/accept/commit/composition excluded | 2026-11-04 |
| Gate 3-BOUNDARY (direct merit-ledger lock fence) | T-12 TEST-INF-FENCE-001 → T-13 `MeritLedgerStore.Save` | approved (scoped), GREEN/BLUE | user approval after 2 private-table cases; same-write season check now guards ledger Put; 2 ledger plus 8 adjacent private-table tests PASS (10/10), no application-table/AWS/Docker access | 2026-11-04 |
| Gate 3-BOUNDARY (acceptance season CAS) | T-12 TEST-INF-FENCE-001 → T-13 `ContractStore.CommitAcceptance` | approved (scoped), GREEN/BLUE | user approval after 5 private-table cases: ACTIVE success and definition/limit priority PASS, stale FREEZING/lock reads RED; acceptance condition + state-specific mapping GREEN 5/5, adjacent private-table tests GREEN 10/10 | 2026-11-04 |
| Gate 3-BOUNDARY (staged fact verifier) | T-12 TEST-INF-FACT-002 → T-13 `StagedFactVerifier` adapter | approved (scoped), GREEN/BLUE | user approval after 3 integration cases compile RED only at missing constructor; 3 verifier cases + 4 manifest publisher regressions PASS on private :18000 tables, no commit/snapshot/composition | 2026-11-04 |
| Gate 3-BOUNDARY (constant-size finalization commit) | T-12 TEST-INF-MAN-002 → T-13 `FinalizationCommitter` | approved (scoped), GREEN/BLUE | user approval after 3 compile-RED cases; two-item transaction CAS FREEZING+matching attempt/cutoff plus conditional deterministic PENDING manifest; 3 commit + 3 verifier + 4 publisher cases PASS (10/10), snapshot/composition excluded | 2026-11-06 |
| Gate 3-BOUNDARY (adjudication snapshot reader) | T-12 TEST-INF-SNAP-001 → T-13 `AdjudicationSnapshotReader` adapter | approved (scoped), GREEN/BLUE | user approval after compile-RED; consistent base-table MERIT_CREDIT aggregation with deterministic bucket/entry ordering, credited contract ids and season isolation; 1 snapshot + 10 adjacent private-table cases PASS | 2026-11-06 |
| Gate 3-BOUNDARY (credited roster reader) | SCN-SCC-AWARD-003 → T-13 `CreditedRosterReader` adapter | approved (scoped), GREEN/BLUE | user approval after compile-RED; consistent per-contract reads union frozen rosters of credited ids only; 1 roster + 10 adjacent private-table cases PASS; test fixture corrected to frozen in-progress team contracts | 2026-11-06 |
| Gate 3-BOUNDARY (composition wiring) | build RED `server.go:80` → T-13 `server.go` + `season_scheduler.go` + `freeze_season.go` | approved (scoped), GREEN/BLUE | user approval after build-fail RED; five-port publisher + seven-port two-phase SeasonFinalizer wired, narrow `FinalizerRunner` interface added; `go build ./...` clean, composition tests PASS (2/2 ran), scheduler/finalizer/publisher unit tests PASS; service started locally on :3006 with healthz 200 | 2026-11-06 |
| LAYER-GATE-APPLICATION | T-11 application core | passed (unit scope) | fresh full application suite PASS; application statement coverage 91.1%; public-boundary build and project-wide gate remain pending | 2026-10-27 |

## TEST-INF-MAN-001 — Manifest Publisher Boundary RED (2026-11-04)

The user approved continuing the existing CHG-0051 T-12/T-13
scope after local API compilation found an obsolete three-port
publisher call. The use case now takes five ports, but the
real `SeasonStore` lacks four required manifest methods.
Its old `SaveFinalized` also retains the transaction-size
limit; changing only the constructor would not finish
the approved two-phase design.

Gate 2 infrastructure integration tests in
`season_manifest_publisher_test.go` (129 lines) and
`season_manifest_resume_test.go` (85 lines) use the real
`PlacementFactPublisher`, real DynamoDB adapters and
fresh unique local test tables; the event capture is a
hand-written spy without business rules. They require
uncommitted staged facts to stay unpublished, one
committed manifest to publish with frozen cutoff and
durable progress/completion, and a partially drained
manifest to resume at `NextIndex=1` without republishing.
Domain/Application are unchanged; the Application gate
already passed.

Standalone `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestManifestPublisher_' -v`
fails at the intended compile RED: missing
`ListPendingManifests`, `ListRemaining`,
`MarkFactPublished` and `MarkComplete` on
`SeasonStore`. No setup/fixture/emulator error; no test
body executed or table created. No production file was
changed before RED. Scoped Gate 3 was subsequently
approved for only the mapped manifest reader/progress
adapters, not fence/snapshot/commit or composition wiring.

GREEN/BLUE: `season_store_manifest_read.go` provides
strongly consistent, paginated committed-manifest and
attempt-scoped staged-fact reads, verifies the complete
staged set count/hash, reconstructs the frozen-cutoff
event envelope, and resumes after the stored cursor.
`season_store_manifest_progress.go` conditionally marks
one matching fact then completes only after all facts
are marked; already acknowledged marks are idempotent.
The first GREEN run exposed a test-fixture collision:
parallel tests sharing a timestamp-only table name
observed each other's manifests. A dedicated fixture
adds each test's name to the table identity; no
production logic was changed to mask the setup issue.

One further same-scope integrity RED,
`TestManifestPublisher_TamperedStagedKeyCannotPublish`,
failed at runtime because a valid digest with a
noncanonical sort key was published. The reader now
checks the canonical attempt/category/track/recipient
key before any event is exposed. Standalone manifest
suite **4/4 PASS** and adjacent staging/lock integration
regressions **5/5 PASS**, all on fresh private tables
on port 18000. Adapter files are 143/147 lines; test
files 27/129/84/90 lines; gofmt clean. No deletion,
AWS access, local application-table write, composition
edit or service startup. Remaining T-12/T-13
fence/snapshot/commit and the ordinary build are
still pending their own RED/Gate 3.

## TEST-APP-MCF-002 / TEST-DOM-ADJ-004 — Fence Error RED (2026-11-04)

The revised Gate 2 already covers the lock-fence behavior
under REQ-MCF-010. Application unit TEST-APP-MCF-002
(`tests/unit/missions/application/adjudication_fence_test.go`,
34 lines) injects a writer rejection while a FREEZING
in-progress contract is confirmed, requiring a distinct
provider-neutral `ErrSeasonAdjudicating` identity and an
unchanged persisted contract. Standalone
`GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 ./tests/unit/missions/application -run '^TestConfirmObjective_AdjudicationFence' -v`
fails at compile solely at two absent
`domain.ErrSeasonAdjudicating` references; the use case
already propagates writer failures and the doubles
contain no policy or assertions.

Demanded Domain unit TEST-DOM-ADJ-004
(`tests/unit/missions/domain/adjudication_fence_error_test.go`,
29 lines) requires that identity to remain distinguishable
under error wrapping from inactive-season and
invalid-state errors. Standalone
`GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 ./tests/unit/missions/domain -run '^TestSeasonAdjudicatingError_' -v`
also fails at compile solely on the two missing sentinel
references. Both are clean unit RED, with no network,
fixtures, or production edits. Domain Gate 3 is requested
only for one declaration in `domain/errors.go`;
Application behavior is already characterized, and
same-write DynamoDB fencing needs separate
infrastructure RED/Gate 3. BLUE: provider-neutral
sentinel, source/tests below 150 lines, standalone and
neighboring suites. No HTTP/composition change is
unlocked.

GREEN/BLUE: the approved one-line provider-neutral
`ErrSeasonAdjudicating` declaration was added to
`domain/errors.go` (65 lines), preserving all unrelated
worktree changes; gofmt clean. Focused Domain unit
test passed; the correctly anchored Application run
executed and passed both the new propagation test
and existing generic-conflict regression. An earlier
anchored command accidentally matched only the
existing test and was corrected before recording
Application GREEN. The full Domain and Application
unit packages then passed with `-count=1`. Existing
use-case code already propagates writer errors, so
no Application production change was needed.
DynamoDB same-write fencing and HTTP/DI composition
are not implemented by this gate.

## TEST-INF-FENCE-001 — Confirmation Lock Fence RED (2026-11-04)

The existing MCF Gate 2 covers the same-write
adjudication fence. Three standalone infrastructure
integration cases invoke real Application use cases
with real ContractStore, SeasonStore and AdjudicationLock
on independent named DynamoDB Local tables (port 18000).
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationFence_' -v`
ran **2 PASS, 1 FAIL**, no setup/environment error:
FREEZING before lock allows confirmation; rejecting
the contract after lock discards staged credits; but
confirming an in-progress contract after lock returns
nil instead of `ErrSeasonAdjudicating` and its raw
persisted item gains `PendingCredits`. The failure
is the missing same-write fence in `ContractStore.Save`,
not a missing handler-side read. The negative test
uses `assert.ErrorIs` to preserve the independent raw
item comparison after the first assertion fails.

Files `season_adjudication_fence_test.go` (75 lines),
`season_adjudication_fence_prelock_test.go` (53) and
`season_adjudication_fence_reject_test.go` (60) have
one-line Application actions, fresh fixtures, and
no test-double business policy. No boundary
production changed before RED; Domain and Application
unit scopes are GREEN. The scoped Gate 3-BOUNDARY
request maps only `contract_store.go` and a focused
`contract_store_adjudication_fence.go` helper for
conditional credit-bearing saves. The departed-member
CAS must remain intact; reject's zero-credit save
must not be fenced. This does not authorize
SaveApproved, merit ledger, snapshot/commit,
composition, HTTP, or local application-table
mutation/startup. If neighboring legacy fixtures
lack a season, seed the required season rather
than weaken the production fence. BLUE: maintain
both production/test files below 150 lines,
gofmt and affected-neighbor regression after GREEN.

GREEN/BLUE: `ContractStore.Save` now routes only
credit-bearing contract puts through a DynamoDB
transaction with a condition check on the existing
season item and the unchanged departed-set CAS.
The transaction rejects a locked season with
`ErrSeasonAdjudicating`, a missing season with
`ErrSeasonNotFound`, and an inactive season with
`ErrSeasonNotActive`; a zero-credit reject remains
unfenced. Helper and store are 69/138 lines and
gofmt clean. `TestAdjudicationFence_` **3/3 PASS**
on new private tables at :18000, including raw
contract non-drift after a locked confirmation.

Adjacent CHG-0053 private-table regression initially
failed during setup: its `pendingTeamFixture` staged
credit without creating any season. One test-only
Arrange insertion created the missing ACTIVE season
on that private table; the stale-approval and
stale-departure checks then **2/2 PASS**. The
fixture file is 60 lines. No production condition
was weakened. The legacy fixed-table integration
tests that invoke `clearMissionsTable` were not run
because they delete existing test-table rows; a
fresh isolated-table run or specific cleanup
authorization is required before claiming the
entire integration suite is green. No HTTP startup,
application-table write, AWS call, or Docker
mutation occurred. T-12/T-13 remain open for
SaveApproved/ledger/accept fences, snapshot,
manifest commit, and composition.

## TEST-INF-MCF-001 — Approval Lock Fence RED (2026-11-04)

Under the approved MCF Gate 2, a focused pair of real
`ApproveContractHandler`/`ContractStore` tests uses
fresh private DynamoDB Local tables on :18000. The
standalone command
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationApprovalFence_' -v`
ran **1 PASS, 1 FAIL**. FREEZING before lock completes
the contract and writes exactly one credit. After the
lock, the same approval returns nil instead of
`ErrSeasonAdjudicating`, changes the raw contract from
PENDING_VALIDATION to COMPLETED, and materializes one
MERIT_CREDIT row when zero is required. Both are real
public use-case calls with persisted, strongly read
assertions; setup, season creation, staging, and lock
acquisition passed. Two tests are 64 and 56 lines,
gofmt clean. The missing same-write fence is in
`ContractStore.SaveApproved` (`contract_store_review.go`),
which already transacts contract and ledger puts; it
is not a missing `ContractStore.Save` or handler-side
read. No production edit was made for approval.

Proposed Gate 3-BOUNDARY covers only inserting the
season lock check in that existing approval transaction,
mapping its cancellation to `ErrSeasonAdjudicating`,
and preserving departed-set CAS, exactly-once credit
condition, and `duplicateOnConditionalFail` indexes.
The 147-line approval file needs a behavior-preserving
split into a focused helper to remain below 150 lines.
MeritLedgerStore.Save, acceptance, snapshot, commit,
composition, HTTP and real application data stay
outside this gate.

GREEN/BLUE under the scoped approval: positive-credit
`ContractStore.SaveApproved` reuses the season guard
within its existing contract+credit transaction.
A season-check cancellation surfaces
`ErrSeasonAdjudicating`; the `+1` transaction
index adjustment preserves `ErrDuplicateConfirmation`
for existing credits, while the departed-set CAS
is unchanged. The approval adapter is 134 lines
after moving the existing duplicate classifier
into the 87-line fence helper; both gofmt clean.
The private-table approval matrix (pre-lock,
post-lock, existing-credit duplicate) is **3/3 PASS**.
The previously green confirmation/reject matrix
and two CHG-0053 stale-write private-table
regressions are **5/5 PASS** (7/7 in the first
combined pass, 3/3 in the added duplicate pass).
The three approval tests are 56/64/67 lines.

No full fixed-table integration run was attempted:
its legacy setup calls `clearMissionsTable` and
would delete existing test rows without approval.
The security gate remains pending because direct
merit-ledger writes, other acceptance mutations,
the snapshot/commit path, composition and HTTP
mapping still lack their own RED/Gate 3 evidence.
No application-table write, AWS call, Docker
mutation, API start or HTTP check occurred.

## TEST-INF-FENCE-001 — Direct Ledger Lock Fence RED (2026-11-04)

The approved revised Gate 2 covers the direct
season-bound merit writer. `MeritLedgerStore.Save`
is a separate public persistence path from
`ContractStore.SaveApproved`; the latter's GREEN
does not protect direct ledger calls. Two focused
tests in `season_adjudication_ledger_fence_test.go`
(83 lines, gofmt clean) use independent private
DynamoDB Local tables on :18000. Standalone
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationLedgerFence_' -v`
ran **1 PASS / 1 FAIL**: a FREEZING season
accepts the direct credit before lock; after
acquiring the lock, `Save` returns nil rather
than `ErrSeasonAdjudicating` and a strongly
read ledger row exists. No fixture/build failure
or prior production edit caused this RED. The
mapped write is `merit_ledger_save.go`; a
same-transaction season ConditionCheck is
required, not a separate read. Gate 3 is
requested only for that writer and narrowly
reusing the existing package-private season
fence helper without changing the already-green
contract write paths. Acceptance, snapshot,
commit, composition, HTTP and application-table
data remain excluded.

GREEN/BLUE: `MeritLedgerStore.Save` now writes
the credit in a two-item DynamoDB transaction
with the same season-unlocked ConditionCheck
used by the already-green contract paths.
Package-private `seasonUnlockedCheck`,
`seasonLockCheckFailed`, and
`adjudicationFenceError` were factored without
changing the contract behavior. The focused
direct-ledger pre/post-lock pair is **2/2 PASS**
and the combined private-table run includes
three approval, three contract-staging, and two
CHG-0053 stale-save regressions: **10/10 PASS**
overall (the initial run report mistakenly
said 9; the actual test-name list has 10).
`merit_ledger_save.go` is 69 lines and
`contract_store_adjudication_fence.go` is 94;
gofmt clean. All created tables remain intact.

No fixed-table integration suite was run:
its legacy `clearMissionsTable` setup deletes
test data, and tests that stage credits without
a season will require isolated fixture repair,
not a weakened production fence. Direct ledger
fencing is complete only for this scoped port;
acceptance mutations, snapshot/commit,
composition, HTTP mapping and L3 gates remain
open. No application-table write, AWS call,
Docker mutation or local API startup occurred.

## TEST-INF-FENCE-001 — Acceptance Season CAS RED (2026-11-04)

Under the approved T-12 Gate 2, five standalone
integration cases call the real atomic
`ContractStore.CommitAcceptance` port on fresh,
independent DynamoDB Local tables on :18000.
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationAcceptanceFence_' -v`
ran **3 PASS / 2 FAIL**, no fixture or build failure.
ACTIVE pre-lock acceptance writes contract, player
guard and definition reservation. Existing
definition and active-limit conflicts retain
their distinct error identities. But a stale
ACTIVE decision after FREEZING without lock
returns nil rather than `ErrSeasonNotActive`,
and after acquiring the lock returns nil
rather than `ErrSeasonAdjudicating`; both
wrongly persist all three guard rows. The
fresh-read Domain decision already rejects
FREEZING, so this is a same-write boundary
race, not a new Application/Domain rule.
Five test files are 50/50/56/57/48 lines,
gofmt clean; the Action in each is a single
call to the public writer port.

Gate 3-BOUNDARY is requested only for
`contract_store_accept.go` plus a narrowly
shared package-private season condition.
The acceptance transaction must check its
persisted contract season is **ACTIVE and
unlocked** (unlike merit credits, which may
write in FREEZING before the lock), map
conditional failures to the matching domain
errors, and preserve the existing
definition-reservation-over-active-limit
conflict priority when the condition
prepends an item. No production changes
for this partition yet. Snapshot/commit,
composition/HTTP and application data
remain separately gated.

GREEN/BLUE under scoped Gate 3: `CommitAcceptance`
prepends a ConditionCheck requiring the stored
contract season be **ACTIVE with no lock** in
the same transaction as the contract put and
player/definition guards. Reservation and limit
cancellation indexes remain derived from
`len(items)`, preserving their precedence.
The first GREEN run was **14/15**: FREEZING
was rejected atomically, but the shared
error mapper returned a generic invalid-state
error. `adjudicationFenceError` now
distinguishes acceptance (`allowFreezing=false`,
`ErrSeasonNotActive`) from credit writes
(`allowFreezing=true`, FREEZING pre-lock
allowed), while an acquired lock maps to
`ErrSeasonAdjudicating`. A fresh private-table
run is **15/15 PASS**: five acceptance,
two ledger, three reviewer approval,
three confirmation/reject, and two
CHG-0053 stale-save cases. Adapter/helper
files 90/109 lines, gofmt clean.

The ordinary fixed-table integration suite
was not run because legacy setup calls
`clearMissionsTable`; no existing test data
was deleted. This scoped GREEN does not
authorize snapshot, two-item manifest
commit, composition or HTTP startup.
Application table :8000 and gateway :8080
were untouched; no AWS/Docker mutation.

## TEST-INF-FACT-002 — Staged Fact Verifier RED (2026-11-04)

Three real-adapter cases on private DynamoDB
Local tables require that a complete staged
set verifies without creating a manifest,
a missing row fails with
`ErrStagedFactSetMismatch`, and a fact whose
merits changed without updating its digest
cannot pass verification. Two test files
`season_staged_fact_verifier_test.go` (94 lines)
and `_integrity_test.go` (50) are gofmt clean.
The standalone
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestStagedFactVerifier_' -v`
fails at compile solely at three absent
`missiondynamo.NewStagedFactVerifier` references.
No fixture/table is opened in that run and
no production or data was changed. Gate 3
is requested only for `season_store_fact_verify.go`
reusing the package's existing attempt-scoped
consistent staged-fact validation; if a
mechanical extraction is needed, preserve
the manifest reader's current behavior.
Manifest commit, snapshot, composition and
public HTTP remain separately gated.

GREEN/BLUE: `season_store_fact_verify.go` (42
lines, gofmt clean) validates the attempt ID
and expected `StagedFactSet`, then reuses
the strongly consistent attempt-scoped
`SeasonStore.ListRemaining` count/hash,
identity, canonical key and digest checks
against an in-memory expected manifest.
It discards the derived intents and neither
writes nor publishes a manifest. Standalone
`TestStagedFactVerifier_` **3/3 PASS**;
neighbor `TestManifestPublisher_` **4/4
PASS** on fresh private :18000 tables
(7/7 total). The tests remain 94/50
lines. Only the verifier adapter was added;
snapshot, two-item FINALIZED+MANIFEST commit,
composition and HTTP remain open. The
legacy fixed-table suite was not run because
its setup deletes test data.

## TEST-INF-MAN-002 — Constant-Size Commit RED (2026-11-04)

Under the previously approved T-12 Gate 2,
three private-table infrastructure cases demand
a guarded `FinalizationCommitter`: 200 staged
and verified facts commit a FINALIZED season
with exactly one deterministic PENDING
manifest (count/hash/cursor; no per-fact
outbox rows), while an unlocked FREEZING
season cannot commit even an empty manifest;
a retried stale commit must leave a previously
PUBLISHED manifest and its cursor unchanged.
`season_manifest_commit_test.go` (87 lines),
`season_manifest_commit_lock_test.go`
(52 lines) and `season_manifest_commit_retry_test.go`
(85 lines) are gofmt clean. Standalone
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestSeasonManifestCommit_' -v`
fails at compile solely at the three absent
`missiondynamo.NewFinalizationCommitter`
references; fixtures never ran. No production
or data was changed. Gate 3-BOUNDARY was
approved for the commit writer only: the
reviewed RED covers size, lock CAS and
published-manifest preservation.

GREEN/BLUE: `season_store_manifest_commit.go`
(83 lines, gofmt clean) performs a two-item
transaction: Update sets State=FINALIZED only
while the persisted season is FREEZING and
holds the matching AdjudicationLock/cutoff,
and a conditional Put creates the one
deterministic `MANIFEST#<season>#<attempt>`
PENDING row with FactCount/FactHash/NextIndex=0.
Any cancellation maps to
`ErrInvalidSeasonStateTransition`; existing
manifests are never overwritten. Standalone
run `TestSeasonManifestCommit_` 3/3 PASS,
`TestStagedFactVerifier_` 3/3 and
`TestManifestPublisher_` 4/4 regressions —
10/10 on private :18000 tables. Snapshot
reader, composition wiring, HTTP startup and
app-table data remain outside this partition.

## TEST-INF-SNAP-001 — Snapshot Reader RED (2026-11-06)

One private-table integration case
(`season_adjudication_snapshot_test.go`, 82
lines, gofmt clean) seeds five MERIT_CREDIT
rows — two for one recipient across two
contracts, one for another recipient, one
TEAM bucket, and one foreign-season row —
then demands
`AdjudicationSnapshotReader.Read(attempt)`
return two buckets where INDIVIDUAL totals
aggregate 100 (contracts c-1,c-2) and 10,
TEAM carries c-4, and the foreign season is
excluded. Standalone
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestAdjudicationSnapshot_' -v`
fails at compile solely at the missing
`NewAdjudicationSnapshotReader` constructor.
Gate 3-BOUNDARY was approved only for the
snapshot adapter.

GREEN/BLUE: `season_store_snapshot.go` (130
lines, gofmt clean) issues paginated
strongly-consistent Queries on the
MERIT_CREDIT partition filtered by the
attempt's SeasonID — never the standings
GSI — and aggregates totals plus distinct
ContractIDs per recipient. Buckets and
entries sort deterministically (category,
track, then total desc/recipient). Snapshot
1/1 PASS plus 10 commit/verifier/publisher
adjacent cases — 11/11 on private :18000
tables. Composition and HTTP remain
separate partitions.

## TEST-INF-ROSTER-001 — Credited Roster RED (2026-11-06)

One private-table case
(`credited_roster_reader_test.go`, 40 lines)
seeds two TEAM contracts of the same team and
season — c-1 roster {p-1,p-2} and uncredited
c-2 {p-3} — then demands
`Members(season, category, recipient, [c-1])`
return exactly {p-1,p-2}. Standalone
`-run '^TestCreditedRosterReader_'` fails at
compile solely at `NewCreditedRosterReader`.
Gate 3 requested only for a roster adapter
union over credited contract ids; existing
`Members(seasonID, teamID)` behavior is
unchanged. Composition/HTTP separate.

## Composition Wiring RED (2026-11-06)

`go test ./tests/integration/composition`
fails at build solely at
`server.go:80`: the composition root still
constructs `NewPlacementFactPublisher` with
the obsolete three-port signature. The
five-port publisher (manifests, staged
reader, bus, progress, completion) plus
the two-phase `SeasonFinalizer` (lock,
snapshot, credited rosters, stager,
verifier, committer) now all exist. Gate 3 was approved for composition.

GREEN/BLUE: `server.go` wires
`NewPlacementFactPublisher(seasonStore,
seasonStore, cfg.DomainEvents,
seasonStore, seasonStore)` and the
two-phase `NewSeasonFinalizer` with the
real lock, snapshot reader, credited
roster reader, stager, verifier and
committer. `season_scheduler.go` gained
a narrow `FinalizerRunner` interface so
either finalizer satisfies the scheduler,
and `freeze_season.go`'s HTTP adapter
accepts the same narrow signature —
concrete types preserved, no behavior
change. `go build ./...` clean;
scheduler/finalizer/publisher unit tests
PASS; composition CreateSeason +
FreezeAndFinalize PASS. Service started
locally `PORT=3006 DYNAMODB_ENDPOINT=
http://localhost:8000 DYNAMODB_TABLE=
hbk-missions-local` — healthz 200;
authenticated GETs return 401 without
OIDC (expected — local verification of
`lang` proceeds via the real gateway
login flow, not bare curl).

## Commands

| Command | Scope | Result | Notes |
| --- | --- | --- | --- |
| `go test -count=1 ./tests/unit/...` | domain+application+http+infra units | PASS (all 5 packages) | includes scheduler-drain, `Start` onError, unsorted-ranking tests |
| `go test -count=1 -p=1 ./tests/integration/infrastructure/dynamodb/` | outbox + finalize + roster store | PASS | DynamoDB Local :18000; CAS refinalize, missing-season, oversized-batch, PUBLISHED-prefix pagination, frozen-roster union |
| `go test -count=1 -v ./tests/integration/infrastructure/sns/` | SNS publisher | PASS (2 tests) | Query-protocol form body: Action/TopicArn/Message + `event_type`/`source` attributes; 4xx propagates error |
| `go test -count=1 -p=1 ./tests/integration/composition/` | server wiring | PASS | `DomainEvents` optional; nil disables drain |
| `go test -count=1 ./tests/unit/missions/application -run 'TestSeasonFinalizer_|Test.*Adjudication|TestPlacementFactPublisher_'` | T-9 Application RED | RED (build fail) | 2026-10-27: intended missing shape — `AdjudicationAttempt`/`AdjudicationSnapshot`/`PlacementManifest` undefined + `NewSeasonFinalizer`/publisher arity; TEST-APP-FIN-003/004 (8 `season_finalizer*`) + TEST-APP-PUB-001 (6 `manifest_publisher*`: order/progress, failure, retry, committed-only, empty + incomplete-manifest guard); Gate 3-APPLICATION approved for T-10 Domain RED; legacy finalize/publisher tests superseded |
| `go test adjudication_mcf_test.go adjudication_mcf_doubles_test.go contract_review_helpers_test.go contract_lifecycle_test.go -run 'Test.*Adjudication'` | TEST-APP-MCF-001 characterization | PASS (5 tests) | 2026-10-27: named-file compile — approve/confirm conflict propagates with persisted contract untouched (clone-per-Get proof), FREEZING pre-lock credit, zero-credit reject; package stays RED via finalizer/manifest types |
| `go test -count=1 ./tests/unit/missions/domain -run 'TestAdjudicationAttempt_|TestStagedFactSet_|TestCreditedRoster_|TestPlacementFact_Digest'` | T-10 Domain RED (TEST-DOM-ADJ-002/003) | RED (build fail) | 2026-10-27: intended missing shape — `NewAdjudicationAttempt`/`ErrAdjudicationAttemptInvalid`/`MembersOfCreditedContracts` undefined first, `NewStagedFactSet`/`Digest`/`ErrStagedFactSet*`/`ErrCreditedRosterInvalid` capped by `too many errors`; 25 test funcs in 6 files, syntax-clean; was RED pending — Gate 3-DOMAIN subsequently approved, GREEN in next row |
| named-file `go test` (6 CHG-0051 domain files + `placement_fact_test.go`) then full `go test -count=1 ./tests/unit/missions/domain` | T-10 Domain GREEN | PASS | 2026-10-27: 28 test funcs green (named-file 0.262s); full package PASS 0.306s — concurrent CHG-0054 language work did not block; `go build`/`go vet` clean |
| named-file `go test` (7 CHG-0051 files incl. `adjudication_regressions_test.go` + `placement_fact_test.go`); full `go test -count=1 ./tests/unit/missions/domain` | T-10 regression additions | PASS | 2026-10-27: +5 funcs (+4 blank-variant subtests) — empty-EventID & same-ID-divergent-payload staging rejection, blank credited-ID/member rejection, `MemberIDs` non-alias; named-file `ok 0.541s`, full package `ok 0.266s`; coverage percentages lead-measured, not recomputed here |
| `go test -count=1 -coverpkg=./internal/missions/domain -coverprofile=/tmp/hbk-chg0051-domain-coverage-after-20261027.out ./tests/unit/missions/domain` | domain production statement coverage, Go 1.27.0 | PASS: 91.1% | 2026-10-27: fresh pre-regression profile `/tmp/hbk-chg0051-domain-coverage-20261027.out` 90.8%; persisted 2026-10-21 `coverage-domain.out` 89.6% by `go tool cover -func` (older validator reported 89.7%). After tests: `NewStagedFactSet` and `MembersOfCreditedContracts` each 100%; no exclusions. Standalone domain threshold passed, not the project-wide gate. |
| `go test -count=1 -coverpkg=./internal/missions/application -coverprofile=/tmp/hbk-chg0051-app-coverage-20261027.out ./tests/unit/missions/application` | application coverage diagnostic, Go 1.27.0 | BLOCKED (build fail) | T-9 RED still lacks `application.PlacementManifest` and `application.AdjudicationSnapshot`; project-wide/current application coverage cannot yet be measured. No application production change authorized. |
| named-file `go test` (10 app files: `season_finalizer*`/`manifest_publisher*`/`placement_fact_publisher` + `finalize_season_placement`/`freeze_season` helpers) | T-11 Application GREEN | PASS | 2026-10-27: 27 test funcs `ok 0.260s` — finalizer pipeline, manifest drain, migrated legacy publisher (`IdempotentRepublish` reworked to single-Act via duplicate committed-manifest delivery); legacy finalize/freeze still green in parallel; `go vet ./internal/missions/application` clean |
| `go test -count=1 -coverpkg=./internal/missions/application -coverprofile=/tmp/hbk-chg0051-app-coverage-current-20261027.out ./tests/unit/missions/application` | current application statement coverage, Go 1.27.0 | PASS: 91.1% | 2026-10-27: CHG-0054 GetOffer partition advanced independently; full application unit package now GREEN, prior missing-shape blocker historical; no exclusions. |
| `go test -count=1 -coverpkg=./internal/missions/domain -coverprofile=/tmp/hbk-chg0051-domain-coverage-current-20261027.out ./tests/unit/missions/domain` | current domain statement coverage, Go 1.27.0 | PASS: 91.2% | 2026-10-27: fresh full domain package GREEN after concurrent localization changes; previous 91.1% profile remains historical. |
| `go test -count=1 -coverpkg=./internal/... -coverprofile=/tmp/hbk-chg0051-project-unit-coverage-20261027.out ./tests/unit/...` | project unit production coverage diagnostic | BLOCKED (build fail) | 2026-10-27: `internal/missions/interface/http/get_offer.go:16` still calls `Handle(string)` rather than `Handle(string, domain.Language)`; partial per-package numbers are NOT a project percentage. Composition also retains CHG-0051's obsolete 3-arg publisher constructor. |
| `go build ./internal/missions/composition` | composition wiring | BLOCKED (expected) | 2026-10-27: fails first on CHG-0054 `http/get_offer.go` `Handle` arity; own `server.go` 3-arg publisher ctor call also stale — both deferred to boundary GREEN + CHG-0054 completion |
| `go test -count=1 -tags=integration ./tests/integration/infrastructure/dynamodb -run '^TestSeasonAdjudicationLock_'` | T-12 Boundary RED (partition 1: adjudication lock) | RED (build fail) | 2026-10-28: solely `undefined: missiondynamo.NewAdjudicationLock`; TEST-INF-ADJ-001 drives real `SeasonFinalizer` with capture-only doubles over a per-run unique test-owned table, ConsistentRead GetItem asserts `AdjudicationLock`/`AdjudicationCutoff` attrs + still-FREEZING `State`; :18000 never contacted; Gate 3-BOUNDARY was pending, subsequently approved for this partition; retry test preseeded `ADJ#`+known cutoff and proved no re-mint — same missing symbol |
| `go test -count=1 -tags=integration -v ./tests/integration/infrastructure/dynamodb -run '^TestSeasonAdjudicationLock_'` | T-12 partition-1 GREEN | PASS (2/2, one approved run) | 2026-10-28: `season_store_lock.go` `AdjudicationLock.Acquire` — conditional lock claim + consistent-read resume; retry `hbk-chg0051-adj-retry-1793222105026163000` 0.14s, acquire `hbk-chg0051-adj-1793222105164110000` 0.01s; both tables left intact; SDK checksum WARN is a Local quirk; scoped build+vet clean |
| `go test -count=1 -tags=integration [-v] ./tests/integration/infrastructure/dynamodb -run '^TestSeasonAdjudicationStaging_'` | T-12 Boundary RED→GREEN (partition 2: fact staging) | RED then PASS (3/3, one approved run) | 2026-10-28 RED: solely `undefined: missiondynamo.NewPlacementFactStager`/`ErrStagedFactConflict`; TEST-INF-FACT-001 stages 200 INDIVIDUAL facts (no tx cap), ConsistentRead `Query` on `FACT#<season>` = 200 rows + p-001 `Digest` equality; identical-replay no-op (`ReplayMarker` untouched); divergent `ErrStagedFactConflict` halts pipeline, no overwrite. GREEN: `season_store_facts.go` sequential conditional `PutItem` + consistent-read replay; -v run — conflict `hbk-chg0051-conflict-1793222872198063000` 0.05s, replay `hbk-chg0051-replay-1793222872244527000` 0.01s, 200-fact `hbk-chg0051-stage-1793222872255176000` 0.21s; all left intact; scoped build+vet clean |
| `go build` + `go vet ./internal/... ./cmd/...` | compile+vet | PASS (pre-T-11 codebase) | superseded: composition expected-broken until boundary + CHG-0054 |
| `bash tools/validate-sdd-change.sh` | structural SDD gate | PASS | 2026-10-21: unit+integration+race green; coverage domain 89.7%, application 90.7%, infrastructure/dynamodb 92.5% (domain sits below the 90% L3 bar — flagged, not waived). Script flagged gofmt on two files NOT changed by this slice: `tests/integration/http/season_statistics_test.go` (untracked user file) and `tests/unit/missions/domain/merit_ledger_test.go` (user modification) — left untouched. |
| `git diff --check` | whitespace | PASS | no output; only relevant findings were the two pre-existing user files above |
| coverage / mutation / critical E2E | L3 gates | pending | T-7 |

## Boundary corrections applied after first landing

- `SaveFinalized` compare-and-swaps `State = FREEZING` on the season item:
  a repeated/concurrent finalization is rejected and can no longer reset
  already-PUBLISHED outbox rows (`RefinalizeIsRejectedAndKeepsState`).
  Missing season maps to `ErrSeasonNotFound`, state conflict to
  `ErrInvalidSeasonStateTransition` via a verifying `Get`.
- `ListPending` paginates past PUBLISHED intents at the head of the
  partition (`ListPendingPagesPastPublishedPrefix`, limit=1).
- `SeasonScheduler.Start(ctx, interval, onError)` surfaces tick errors —
  including SNS outages — to a callback; composition logs them. Publish
  failures are observable end-to-end (logged), not silently dropped.
- TEAM `member_ids` now resolve from `ContractStore.Members(seasonID,
  teamID)`: the union of frozen contract rosters for that team in that
  season (design's frozen-roster source). `TeamRosterReader` gained the
  `seasonID` param; `DenyAllReader` no longer serves the roster port.

## T-10 Mandatory BLUE Review — 2026-10-27 (production + tests)

Reviewed after GREEN; no refactor required — code already clean.

- **Line-count inventory (<150 all):** `adjudication_errors.go` 12,
  `adjudication_attempt.go` 30, `hash_fields.go` 21,
  `staged_fact_set.go` 47, `credited_roster.go` 40,
  `placement_fact.go` 83; tests 26–113 per file.
- **Provider-neutral / domain purity:** imports only `crypto/sha256`,
  `encoding/hex`, `errors`, `slices`, `sort`, `strconv`, `strings`,
  `time` — no AWS SDK, JSON, IO, or `time.Now` (cutoff caller-supplied).
- **Source immutability:** `StagedFactSet` keeps only private
  `count`/`hash` captured at construction; `PlacementFact` clones
  `MemberIDs`; `MembersOfCreditedContracts` returns a fresh sorted slice.
- **CQRS / architecture:** pure domain value objects + package function;
  no ports, orchestration, or persistence concerns; per-layer
  independence holds (application/boundary unaffected).
- **Post-BLUE commands:** named-file run `ok 0.262s` (28 funcs); full
  `go test -count=1 ./tests/unit/missions/domain` `ok 0.306s`;
  `go build`/`go vet` clean; `git diff --check` clean.
- **Gate distinction:** revised domain standalone test and coverage
  threshold PASSED (fresh 91.1% after regressions); project-wide
  coverage gate remains PENDING while Application RED blocks its
  current measurement; mutation/critical-E2E pending. No
  overall L3 completion claimed.

## T-11 Mandatory BLUE Review — 2026-10-27 (production + tests)

Reviewed after GREEN; one correction applied while green, no remaining
smells.

- **Line-count inventory (<150 all):** `adjudication_ports.go` 65,
  `season_finalizer.go` 105, `placement_manifest.go` 46,
  `publish_placement_facts.go` 86; tests 58–126 per file.
- **Provider-neutral / purity:** ports are context-scoped domain
  interfaces only; no SDK, transport, or provider names in the layer.
- **CQRS / boundaries:** 11 focused single-operation ports (7 finalizer
  + 4 manifest); existing `SeasonReader` reused, no broadened interface;
  legacy `finalize_season.go`/`finalize_season_placement_test.go`
  untouched — parallel models coexist until boundary migration.
- **Post-GREEN correction:** `Official` guard on fact derivation was
  initially dropped vs legacy semantics; restored + new
  `PracticeSeasonStagesNoFacts` test — caught by BLUE, not tests.
- **Gate update:** full application unit package now GREEN (91.1%
  coverage); CHG-0054's former compile RED is historical. The HTTP
  GetOffer call remains stale, and CHG-0051 composition still calls the
  3-arg publisher ctor; both boundary builds remain blocked. Neither
  the project-wide coverage nor the boundary gate is approved.
- **T-12 lock/staging BLUE:** focused single-operation adapters (94/110 lines), tests (87/98 and 109/91/92 lines) use owned tables and the real Application entry; conditional writes plus consistent-read retries preserve identity and immutable digests, sequential staging avoids the transaction cap; GREEN 2/2 and 3/3, scoped build/vet/gofmt/diff-check clean; no full boundary gate claimed.

## Residual Risks

- **High — ADDRESSED BY REVISED DESIGN (pending implementation T-9..T-13):**
  seasons with more than 99 ranked placement facts could not finalize
  under the all-in-one transaction. The 2026-10-21 revised design stages
  immutable facts via conditional `PutItem` writes (digest-verified
  replay) and commits a constant-size MANIFEST — the 100-item cap no
  longer bounds ranked-entry count.
  Remains a live risk in deployed code until T-13 lands.
- **TEAM roster not constrained to credited contributors — ADDRESSED BY
  REVISED DESIGN (pending implementation T-9/T-10/T-13):** the revised
  `CreditedRosterReader` resolves `member_ids` only from contract ids
  credited in the exact season/category/TEAM bucket/recipient. Live risk
  until implemented.
- `occurred_at` — REVISED: now the adjudication cutoff minted when the
  lock engages (never publish time); `correlation_id` remains a
  deterministic hash of the season id, stable across republication of
  the same finalized batch.
- `SharedEventsTopicArn` defaults to empty in `template.yaml`; when unset,
  `SNS_TOPIC_ARN` is empty, composition wires no `DomainEvents`, and the
  scheduler does not drain the outbox — the capability is inert until the
  shared topic ARN is provided at deploy time (same convention as
  requisition-service). The `sns:Publish` statement targets the param
  ref directly. The Achievements side of the pipeline is likewise
  pending: Achievements Lambda/outbox/SAM/shared SQS subscription remain
  unimplemented (lead-reviewed, unresolved).

=== 2026-11-16 L3 convergence ===
- Validator PASS end-to-end (gofmt/vet/unit/integration/race) after the
  fixture-repair wave aligning DynamoDB/composition fixtures with the
  season fence and English-snapshot invariants.
- Public contract: `422 SEASON_ADJUDICATING` mapped; contract detail
  404s now surface `CONTRACT_NOT_FOUND` (`ContractStore.Get` returns
  `domain.ErrContractNil`).
- Critical E2E TEST-E2E-AWD-002 executed: 200 ranked entries → one
  pending manifest → 200 unique publishable intents.
- T-14 orphan sweep: scheduler reports locks >24h via log reporter;
  listing is consistent and integrity-checked; no auto-unlock.
- Coverage: domain 92.4%, application 91.7%. The 62.2% dynamodb figure
  is whole-package including unrelated adapters; all CHG-0051-touched
  adapters are covered by dedicated integration suites above.
