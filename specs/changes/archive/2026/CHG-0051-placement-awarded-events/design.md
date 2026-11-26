---
change_id: CHG-0051
artifact: design
status: proposed
updated: 2026-10-21
---

# Design — CHG-0051

Revised 2026-10-21: two-phase adjudication replaces the all-in-one
finalize transaction (Gate 2 re-approved; Gate 3 for this model pending —
earlier GREEN evidence does not authorize new production).

## Domain Additions

- `PlacementFact{SeasonID, Category, Track, RecipientID, Rank,
  FinalMerits, MemberIDs[]}` — immutable staged fact; unchanged shape.
- `AdjudicationAttempt{ID, SeasonID, Cutoff}` — attempt id stable per
  season (derived `ADJ#<season_id>`); `Cutoff` minted once when the lock
  engages and reused by every retry.
- `StagedFactSet`/`Verify(count, hash)` — domain check that the staged
  rows equal the adjudication snapshot before the commit.
- Season item gains internal `AdjudicationLock` + `AdjudicationCutoff`
  fields; public `State` stays `FREEZING` — no new enum.

## Application

- `SeasonFreezer` — UNCHANGED: `ACTIVE` -> `FREEZING` writes no lock.
  Eligible in-progress confirmations/approvals keep crediting per
  `Season.CanAcceptMerits` throughout FREEZING.
- `SeasonFinalizer` (reworks `FinalizeSeasonHandler`) — first step:
  conditionally acquire the adjudication lock on the season item
  (`attribute_not_exists` on `AdjudicationLock`), minting
  `AdjudicationCutoff` + `attempt_id` once; a retried/concurrent start
  reads back the existing values instead of overwriting. Public `State`
  stays `FREEZING` for the whole adjudication window.
- Then: read a strongly-consistent `AdjudicationSnapshot` (standings +
  credited contract ids per bucket/recipient) via
  `AdjudicationSnapshotReader`; derive `PlacementFact`s;
  `PlacementFactStager` writes them as conditional `PutItem`s
  (`attribute_not_exists`, canonical digest on each row; identical digest
  -> replay no-op, divergent digest -> `ErrStagedFactConflict`) under
  `attempt_id` with bounded parallelism; `StagedFactVerifier` confirms
  count/hash; `FinalizationCommitter` runs one guarded transaction — CAS
  `FREEZING`->`FINALIZED` + deterministic conditional MANIFEST intent.
  Zero ranked entries still commit an empty manifest.
- `PlacementFactPublisher` (manifest-driven): `ManifestIntentReader`
  lists PENDING manifests; `StagedFactReader` loads that attempt's facts;
  each published via `DomainEventPublisher` with stable `event_id`;
  `ManifestProgressWriter` marks per-manifest progress so a mid-drain
  failure resumes where it stopped, then marks the manifest PUBLISHED.
- Credit path: season-bound credit writes and tie-break-relevant contract
  mutations pass `ExpectUnlocked(seasonID)`; the write adapter fences it
  in the same write and returns `ErrSeasonAdjudicating` when locked.
  Roster resolution port becomes `CreditedRosterReader`
  (`member_ids` = frozen rosters of the credited contract ids only).

## Envelope v1.0 (snake_case) — unchanged

`event_id` (deterministic: season+attempt+bucket+recipient+rank),
`event_type`, `event_version: "1.0"`, `source: "hbk-missions-service"`,
`occurred_at` = `AdjudicationCutoff` (never publish time),
`correlation_id`, `payload{season_id, category, track, recipient_id,
rank, final_merits, member_ids?}`.

## Staging + Manifest Model (DynamoDB)

- Staged fact item: `PK=FACT#<season_id>`,
  `SK=<attempt_id>#<category>#<track>#<recipient>`; written by CONDITIONAL
  `PutItem` (`attribute_not_exists(PK)`) — one conditional write per fact,
  no all-fact transaction, bounded parallelism to avoid throttling.
  Each staged row carries a canonical `digest` of its fact content.
  Replay: an existing row whose stored digest equals the freshly derived
  digest is a successful no-op; a row with a DIFFERENT digest is a hard
  integrity error (`ErrStagedFactConflict`) — never overwritten, blocks
  the attempt for operator reconciliation. Staged rows are never
  publishable until the manifest commits.
- Manifest intent: `PK=OUTBOX`,
  `SK=MANIFEST#<season_id>#<attempt_id>` — deterministic from the stable
  attempt/cutoff, written conditionally inside the commit transaction, so
  a retried start/commit can never create two publishable manifests for
  one season. Content: `{season_id, attempt_id, cutoff, fact_count,
  fact_hash, state: PENDING}` — constant size.
- Commit transaction is always 2 items: season CAS + manifest — the
  100-item cap no longer bounds ranked-entry count.
- Publisher only honors COMMITTED manifests; staged rows under a lock
  without a manifest are never publishable.

## Fenced Writes

Every write that can alter ranking input — merit-credit appends
(`merit_ledger_save.go` transaction), contract credit materialization /
review approval (`contract_store_review.go`), and other
tie-break-relevant contract mutations — adds a condition asserting the
season's lock is not engaged. Before adjudication starts these writes
proceed normally (freeze changes nothing); once engaged the same-write
condition rejects with `ErrSeasonAdjudicating` -> `422 SEASON_ADJUDICATING`,
never a false 200, and a post-`FINALIZED` retry cannot add to the
standing. The fence constrains only positive-credit/rank-affecting
writes — `reject` paths writing zero credits stay allowed unless the
transition mutates tie-break-relevant state. These conditions land in
mission-contract-fulfillment write adapters owned by
REQ-MCF-010/012/020; Gate 1 canonical ownership expansion was approved
2026-10-21; Gate 2 (MCF scope) approved 2026-10-21, Gate 3 pending.

## Consistency + Recovery

- Snapshot reads are strongly consistent base-table reads (never the
  standings GSI) so post-freeze writes cannot be missed.
- Crash before commit: lock + staged rows remain, nothing published;
  retry reuses attempt id/cutoff, re-stages as digest-verified no-ops,
  commits once; a divergent staged digest blocks with
  `ErrStagedFactConflict` instead of overwriting.
- Orphaned locks: internal scheduler sweep lists `AdjudicationLock`
  seasons past a threshold and surfaces them for operator reconcile —
  no new external web route in scope.

## Error Policy

- `ErrSeasonAdjudicating` on fenced writes; `ErrInvalidSeasonStateTransition`
  on commit conflicts; `ErrStagedFactConflict` on a staged row with a
  divergent digest (hard integrity error — blocks the attempt, never
  overwrites); other staging/verification failures leave the lock engaged
  for retry; SNS failures leave manifest progress pending — next tick
  resumes with identical `event_id`s.
