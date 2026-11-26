---
change_id: CHG-0051
title: Emit placement-awarded facts on official season finalization
status: proposed
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/CHG-0051-placement-awarded-events/
execution_mode: parallel_tracks
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_AWS_SNS_PUBLISH_WORKFLOW
context_scope: existing
created: 2026-10-20
updated: 2026-10-21
---

# Emit placement-awarded facts on official season finalization

## Intent

When an official season finalizes, Missions publishes one durable
`missions.season.placement-awarded` fact per ranked entry so Achievements
can materialize permanent unlocks. Missions adjudicates the standings —
which rank each entry earned; it never names achievement definitions.

## Business Value

- Competitive results become permanent player recognition.
- Finalization is a two-phase adjudication: a fenced lock freezes ranking
  input, immutable facts are staged at any scale, and a constant-size
  manifest commits atomically with `FINALIZED`.

## User Stories

- **US-SCC-AWARD**: As a season operator, I want finalized standings to
  emit adjudicated placement facts, so player achievements are provable
  and repeatable.

## Scope In

- One fact per ranked entry of each official bucket — no tier/policy
  vocabulary; the rank is the only adjudicated value.
- Adjudication lock + stable cutoff minted when adjudication STARTS
  (public state stays `FREEZING`; freeze itself is unchanged — eligible
  in-progress work keeps crediting until then). From that point every
  season-bound credit and tie-break-relevant contract mutation is fenced
  against the lock in the same write.
- Immutable ranked facts staged by conditional `PutItem` writes
  (digest-verified replay), verified by count/hash, then one guarded
  transaction commits `FINALIZED` plus a constant-size outbox MANIFEST.
- Reliable manifest-driven publisher -> shared SNS topic (past-tense
  facts, envelope v1.0); `occurred_at` is the adjudication cutoff.
- TEAM facts carry `member_ids` from the frozen rosters of the contracts
  credited in that exact season/category/TEAM bucket for that recipient.

## Scope Out

- Achievement selection/unlocks (Achievements), notifications, corrections
  of already-finalized placements (compensating events are future work).

## Decisions

- No tier vocabulary: every ranked entry of an official bucket emits a fact;
  which ranks grant achievements is decided by Achievements definitions
  (`criteria.ranks`). Agreed 2026-10-20.
- Revised finalization model approved 2026-10-21 (Gate 2 re-approved):
  public state vocabulary stays `FREEZING`/`FINALIZED` (no new enum for
  the web app); an internal adjudication lock + cutoff replaces the
  all-in-one transaction, removing the 100-item ceiling; fact emission is
  driven by a committed manifest, not per-fact transaction items.
- Canonical scope covers `seasonal-category-competition` (REQ-SCC-004/018)
  and, per user-approved Gate 1 expansion on 2026-10-21,
  `mission-contract-fulfillment` (REQ-MCF-010/012/020): the lock fence
  constrains approval/credit writes in `merit_ledger_save.go`,
  `contract_store_review.go`, `contract_store_accept.go`, and the
  review/confirm use cases. Gate 2 (MCF scope) was approved by the user
  on 2026-10-21 — T-9 TEST-APP-MCF-001 may start RED. Gate
  3-APPLICATION (revised) was approved 2026-10-27 (T-10 Domain RED
  only); Gate 3-DOMAIN (revised) was approved 2026-10-27 for
  TEST-DOM-ADJ-002/003, authorizing T-10 domain production. Gate
  3-APPLICATION production was approved 2026-10-27 via scoped user
  instruction ('...despues continua' post-coverage) authorizing T-11
  only. Gate 3-BOUNDARY (partition 1, TEST-INF-ADJ-001) was approved
  2026-10-28 scoped to the lock adapter + one local run. Gate
  3-BOUNDARY (partition 2, TEST-INF-FACT-001/-002) was approved
  2026-10-28 scoped to the staging adapter + one local run of three
  tests; fence/snapshot/manifest/publisher partitions remain pending.

## Risk And Gates — L3

Per constitution: competitive adjudication + cross-service publication.
BDD, per-layer RED/GREEN/BLUE, security, coverage, mutation, critical E2E,
rollout/rollback, independent review.

Continuation: `handoffs/context-checkpoints/CHECKPOINT-20261026-1029.md`
records T-9's approved RED scope and the pending revised Gate 3.

## Local Composition Blocker — Manifest Boundary RED (2026-11-04)

The user approved continuing existing T-12/T-13 in CHG-0051's
own scope after the local missions service failed to build:
`internal/missions/composition/server.go:80` calls the new
five-port placement publisher with three arguments. The
`SeasonStore` also lacks all four new manifest read/progress
ports, and the old finalization store still writes the
limited outbox transaction. Changing only the constructor
would risk losing durable publication.

Under the already approved revised Gate 2,
TEST-INF-MAN-001 adds two infrastructure integration test
files (129 and 85 lines) for committed-only publication,
durable progress and resumed drain. The standalone RED
fails solely at compile on the four missing `SeasonStore`
methods. Neither production nor local application data
was touched. Gate 3-BOUNDARY for manifest reader/progress
adapters is requested; fence/snapshot/commit, composition,
full local API startup and the remaining L3 gates are
still separate work.

The scoped manifest reader/progress adapters are
GREEN/BLUE on private DynamoDB Local tables (four new
cases plus five adjacent lock/staging regressions).
A supplementary RED proved that a staged row with
a valid digest but wrong sort key could be published;
the adapter now rejects that identity mismatch. A
parallel test fixture initially reused the same
timestamp-only table and was isolated with the test
name before rerunning. Both adapter files and all
four test files are below 150 lines. No change was
made to the legacy finalization commit, write fence,
snapshot reader or composition, and local HTTP
remains unavailable pending those separate scopes.

The next fence partition is RED at Application and
Domain under the previously approved MCF Gate 2.
`TEST-APP-MCF-002` requires a named
`ErrSeasonAdjudicating` when confirmation's writer
rejects the staged credit, without mutating its
persisted contract. `TEST-DOM-ADJ-004` requires
the business error identity to survive wrapping and
remain distinct from season-state errors. Both
focused unit commands fail only because the Domain
sentinel is undefined; no production code changed.
Gate 3-DOMAIN is requested for `domain/errors.go` alone;
real same-write DynamoDB fencing and API mapping
remain gated.

Scoped Gate 3-DOMAIN was approved 2026-11-04.
`ErrSeasonAdjudicating` is now a distinct,
provider-neutral sentinel in `domain/errors.go`;
the focused Domain and Application tests plus
both full unit packages pass. No existing
worktree changes were removed and no Application,
DynamoDB, HTTP, or composition production was
edited. Same-write locking is the next separate
Boundary RED/Gate 3 partition.

TEST-INF-FENCE-001 now runs through the real
confirmation use case and stores on independent
DynamoDB Local tables: FREEZING before lock and
post-lock zero-credit rejection PASS, while a
post-lock confirmation FAILS with nil error and
persisted PendingCredits drift. No boundary
production changed. A scoped Gate 3-BOUNDARY
is requested only for the same-write contract
save fence; approval/ledger, snapshot/commit,
composition and HTTP remain separately gated.

Scoped Gate 3-BOUNDARY approved 2026-11-04:
credit-bearing `ContractStore.Save` now checks
the persisted season lock in the same DynamoDB
transaction as the contract put, retaining the
departure CAS. Three isolated fence scenarios
and two private CHG-0053 regressions pass. A
CHG-0053 test fixture gained its missing ACTIVE
season row; production was not weakened to
accommodate the old fixture. The broader
fixed-table integration suite is unverified
because it deletes existing test data in setup.
SaveApproved/ledger/accept, snapshot/commit,
composition and local HTTP are still pending.

Operational pause: the resumable checkpoint is
`handoffs/context-checkpoints/CHECKPOINT-20261104-123038.md`.
Next T-12 partition is approval/ledger/accept fence RED;
it does not inherit the scoped `ContractStore.Save` Gate 3.

The next T-12 approval partition is now RED under the
approved MCF Gate 2. Real reviewer approval before
the lock passes; after lock it incorrectly completes
the contract and writes one merit credit. The
missing fence belongs to `ContractStore.SaveApproved`
in `contract_store_review.go`. A separate Gate
3-BOUNDARY is requested only for that transaction
and a focused helper split; no approval production
code or application-table data was changed.

The scoped approval Gate 3-BOUNDARY was approved
and is GREEN/BLUE: positive-credit `SaveApproved`
checks the season lock in the same transaction
as contract and credit puts. Existing-credit
duplicate detection was adjusted for the new
condition-check position and proved by a new
private-table regression. Three approval cases
and five neighboring private-table cases pass.
Direct ledger/accept mutations, finalization
commit, composition, local API and full L3
verification remain separate and pending.

The direct `MeritLedgerStore.Save` path is a
separate T-12 RED: pre-lock credit is accepted,
but after acquiring the season lock the writer
returns nil and persists a credit. Both tests
use fresh private DynamoDB Local tables; no
ledger production code was changed. Gate
3-BOUNDARY is requested only for an atomic
season-lock check on that writer, with
acceptance/commit/composition still outside.

Scoped Gate 3-BOUNDARY for direct
`MeritLedgerStore.Save` is GREEN/BLUE: its
credit Put shares the season-unlocked
transaction condition with contract staging
and approval. Ten private-table tests pass
across the three writer paths plus two
CHG-0053 stale-write regressions. No
fixed-table test cleanup, app-table write,
AWS access, Docker mutation, API startup
or public HTTP check occurred. Acceptance
and finalization remain separately gated.

Acceptance stale-read RED is now classified:
`CommitAcceptance` writes contract, player
slot and definition reservation after either
FREEZING or the adjudication lock, despite
an earlier ACTIVE decision. Three existing
behavior cases pass, two stale-read cases
fail on private DynamoDB Local tables.
Only the acceptance transaction's
ACTIVE-and-unlocked same-write condition
is proposed for its next Gate 3-BOUNDARY;
no production or application data changed.

The scoped acceptance Gate 3-BOUNDARY is
GREEN/BLUE: stale ACTIVE decisions cannot
commit after FREEZING or the adjudication
lock. The same transaction checks the
persisted season before writing contract,
player slot and definition reservation.
Acceptance maps unlocked FREEZING to
`ErrSeasonNotActive`, and a locked season
to `ErrSeasonAdjudicating`; five acceptance
and ten adjacent private-table tests pass.
The local API is still unavailable until
snapshot and composition converge.

TEST-INF-FACT-002 staged-set verifier is RED:
three private-table integration scenarios fail
only at the missing `NewStagedFactVerifier`
constructor. They require exact persisted
count/hash and rejection of altered staged
merits, without creating a publishable
manifest. Gate 3-BOUNDARY is requested only
for the verifier adapter; no verifier
production code or local data was changed.

Scoped Gate 3-BOUNDARY for the staged-fact
verifier is GREEN/BLUE. The 42-line adapter
checks the exact persisted set under a
strongly consistent attempt-scoped read
before the later manifest commit; it
does not create or publish that manifest.
Three focused cases and four publisher
regressions pass in private DynamoDB Local
tables. Snapshot, commit, composition
and public API verification are still
unimplemented or unverified in this change.

The T-12 commit partition is GREEN/BLUE
under its approved Gate 3-BOUNDARY:
`season_store_manifest_commit.go` performs
the constant-size two-item transaction —
FREEZING season to FINALIZED only while
holding the matching attempt lock/cutoff,
plus the conditional deterministic PENDING
manifest. Stale retries cannot overwrite a
PUBLISHED manifest. Ten private-table cases
pass (3 commit + 3 verifier + 4 publisher).
The snapshot reader is GREEN/BLUE under
its approved Gate 3-BOUNDARY: consistent
base-table credit aggregation feeds the
finalizer. The credited-roster adapter is
GREEN/BLUE under its approved Gate 3:
consistent per-contract reads union the
frozen rosters of credited contracts
only; the test fixture was corrected to
in-progress frozen team contracts.
Composition is GREEN/BLUE under its
approved Gate 3: the five-port manifest
publisher and two-phase finalizer are
wired, `go build` is clean, and the
service now starts locally (:3006,
healthz 200). Remaining: HTTP `lang`
verification via the real gateway login
and the L3 gates; no overall pass
claimed.
