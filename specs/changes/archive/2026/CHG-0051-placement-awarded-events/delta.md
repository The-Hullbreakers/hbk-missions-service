---
change_id: CHG-0051
artifact: canonical-delta
status: consolidated
updated: 2026-10-21
---

# Canonical Delta — seasonal-category-competition

Revised 2026-10-21 (Gate 2 re-approved): two-phase adjudication replaces
the all-in-one finalize transaction; Gate 1 expansion adds
REQ-MCF-010/012/020 — Gate 2 (MCF scope) approved 2026-10-21, Gate 3
pending.

## ADDED

### REQ-SCC-018 — Publish adjudicated placements on official finalization

When a season with `Official: true` is adjudicated and committed to
`FINALIZED`, Missions emits one `missions.season.placement-awarded` v1.0
fact per ranked entry of every `{category, track}` bucket — at any scale.
Immutable facts are staged during adjudication, verified, and committed
as a durable MANIFEST intent with the transition; the publisher emits
each staged fact and tracks progress until drained. Each fact carries
`season_id`, `category`, `track`, `recipient_id`, `rank`, `final_merits`;
TEAM facts add `member_ids` — the union of frozen rosters of exactly the
contracts whose credits contributed to that ranked entry (uncredited
roster members and current membership are never used). Unofficial
seasons emit nothing. `occurred_at` is the adjudication cutoff minted
when the lock engaged — never the publish time; republication repeats
the same `event_id`. Which ranks produce achievements is Achievements'
decision, never Missions'.

#### SCN-SCC-AWARD-001 — Official season emits ranked facts

Given an official `FINALIZED` season where `player-1` ranks 1 in
`BOUNTY_HUNTING` `INDIVIDUAL`, when the manifest drains, then a
`placement-awarded` fact exists with `rank: 1`, merits, and cutoff `occurred_at`.

#### SCN-SCC-AWARD-002 — Unofficial season emits nothing

Given an unofficial season reaches `FINALIZED`, when finalization
completes, then no `placement-awarded` fact is emitted for any bucket.

#### SCN-SCC-AWARD-003 — Team fact carries credited frozen rosters only

Given team `team-alpha` ranks 1 in `BOUNTY_HUNTING` `TEAM` where contract
`c-1` contributed credits and `c-2` (same team, same season) none, when
the fact is emitted, then `member_ids` lists exactly `c-1`'s frozen
roster — `c-2`'s members are excluded.

#### SCN-SCC-AWARD-004 — Fact emission survives publication failure

Given a manifest publish fails mid-drain, when the publisher retries,
then published facts repeat the same `event_id` — no duplicates.

#### SCN-SCC-AWARD-005 — Large standings finalize and publish fully

Given an official `FREEZING` season with 200 ranked entries, when the
operator finalizes, then the snapshot is staged (beyond a single
transaction), the manifest commits `FINALIZED`, and all 200 facts are
eventually published.

#### SCN-SCC-AWARD-006 — Interrupted staging resumes the same snapshot

Given a crash after staging, when finalization retries, then attempt
id/cutoff are reused, identical digests replay as no-ops, the manifest
commits once under its deterministic key — nothing publishable before.

#### SCN-SCC-AWARD-007 — Divergent staged content is an integrity error

Given a staged fact row exists whose digest differs from the freshly
derived fact for the same key, when finalization retries, then the write
is refused — never overwritten; the attempt blocks for reconciliation.

## MODIFIED

### REQ-SCC-004 — Freeze and finalize

Extends: freeze unchanged — `FREEZING` still accepts eligible
in-progress work per `Season.CanAcceptMerits`. When finalization starts,
the starter conditionally acquires an internal adjudication lock and
mints `AdjudicationCutoff` + attempt id exactly once (retry keeps them;
state stays `FREEZING` throughout). While engaged, every season-bound
merit-credit write and tie-break-relevant contract mutation MUST carry
the lock check in the same write — rejected once engaged
(`SEASON_ADJUDICATING`). Finalization stages immutable facts per
REQ-SCC-018, verifies the staged count/hash, and commits one guarded
transaction: `FREEZING` -> `FINALIZED` plus the deterministic MANIFEST.
A pre-commit crash leaves the season locked with non-publishable staged
rows; orphaned locks must be observable and reconcilable by a sweep.

#### SCN-SCC-LOCK-001 — Credit racing the lock is rejected

Given a season whose adjudication lock is engaged, when a merit-credit
write or tie-break-relevant mutation targets it, then it is rejected — a
racing write either lands before the lock or is rejected, never both.

#### SCN-SCC-LOCK-002 — FREEZING still credits eligible in-progress work

Given a `FREEZING` season whose adjudication lock is NOT yet engaged,
when an eligible in-progress unit is confirmed or a pending review
approved, then the credit is recorded normally — freeze is unchanged.

#### SCN-SCC-LOCK-003 — Orphaned locked season is reconcilable

Given adjudication stopped mid-way leaving the season locked, when the
operator reconciliation runs, then it is identified and can resume the
same attempt id and cutoff.

### REQ-MCF-010 — Commit adjudication and credits atomically

Extends: pre-lock, eligible `FREEZING` credit writes proceed unchanged.
Once engaged, a rank-changing credit write MUST be rejected atomically
in the same write — contract and ledger untouched —
`422 SEASON_ADJUDICATING`, never a false 200; a post-`FINALIZED` retry
cannot add to the standing. Zero-credit writes changing no
tie-break-relevant state are NOT blocked.

#### SCN-MCF-ADJ-001 — Pre-lock approval with eligible credit succeeds

Given a `FREEZING` season whose lock is NOT yet engaged and a
`PENDING_VALIDATION` contract, when a moderator approves, then credits
commit normally — freeze holds until finalization starts.

#### SCN-MCF-ADJ-002 — Approval racing the adjudication cutoff

Given adjudication is starting, when an approval write races the lock,
then it either commits fully before the cutoff or is rejected
atomically — contract stays `PENDING_VALIDATION`, ledger unchanged.

#### SCN-MCF-ADJ-003 — Completion reward writes nothing after lock

Given the lock is engaged, when an approval would grant a completion
reward, then it is rejected with `SEASON_ADJUDICATING` — no credit or
contract transition is written.

### REQ-MCF-012 — Completion grants the frozen reward

Extends: the completion reward is a positive-credit write fenced
identically — post-lock `COMPLETION` credit is rejected and writes
nothing; pre-lock it commits per the atomic rule.

### REQ-MCF-020 — Adjudicate a pending contract

Extends: `approve` paths writing ranking-affecting credits are fenced
per REQ-MCF-010; `reject` writes zero credits and stays allowed unless
the transition mutates tie-break-relevant state. Retried approvals are
idempotent; a post-`FINALIZED` retry cannot add to the standing.
