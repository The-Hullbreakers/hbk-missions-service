---
change_id: CHG-0051
artifact: security-review
status: pending
updated: 2026-10-20
---

# Security Review — CHG-0051

## Trust Boundaries

1. **SNS publish (outgoing)**: Lambda role gets `sns:Publish` only on the
   shared domain-events topic; message attributes `event_type`/`source`
   match the envelope.
2. **Fact content**: public competitive data only — season, category,
   track, recipient, rank, merits, frozen roster ids. No evidence,
   contract internals, or tokens.
3. **Integrity**: `event_id` deterministic + persisted intent prevents
   partial finalization states; consumers still validate.
4. **Adjudication fence (revised 2026-10-21)**: the internal lock is a
   write-integrity boundary — every ranking-input write must carry the
   same-write lock condition so a missing/bypassed condition is a
   findings-class defect, not a race detail; staged facts are immutable
   and non-publishable without the committed manifest. Coverage now
   includes MCF approve/confirm paths (REQ-MCF-010/012/020, Gate 1
   expansion approved 2026-10-21): a fenced write must surface
   `422 SEASON_ADJUDICATING` — a silent 200 with no write would be a
   High finding.

## Findings

| Severity | Finding | Status |
| --- | --- | --- |
| High | Direct `MeritLedgerStore.Save` persists a credit after the adjudication lock (TEST-INF-FENCE-001) | resolved for direct writer; same-write condition, 2/2 ledger and 10/10 combined private-table tests PASS |
| High | Stale ACTIVE acceptance commits contract/player/definition guards after FREEZING or lock (TEST-INF-FENCE-001) | resolved for acceptance writer: ACTIVE+unlocked transaction condition, 5/5 acceptance and 15/15 combined private-table cases PASS |

No Critical/High findings may remain open at consolidation.

Partial 2026-11-04 evidence: same-transaction season
lock checks now protect confirmation-stage contract
saves and positive-credit reviewer approvals on
private DynamoDB Local tables. Locked approval
returns an error without modifying contract or
ledger; pre-lock approval and duplicate-credit
classification remain intact. This does not close
the L3 security gate: finalization snapshot/commit and
the public `422 SEASON_ADJUDICATING` mapping
still need their own classified boundary evidence.

Confirmed write-integrity risk (2026-11-04):
TEST-INF-FENCE-001 demonstrates that a direct
`MeritLedgerStore.Save` call after the lock
persisted one credit instead of failing. The
separately approved Gate 3-BOUNDARY now transacts
the credit Put with the shared season-unlocked
condition; a locked call returns
`ErrSeasonAdjudicating` with zero ledger items.
Both direct-ledger tests and eight adjacent
private-table cases pass (10/10). This closes
only the named direct-writer finding: finalization
snapshot/commit and public mapping
and the full L3 security review remain pending.

The scoped staged-fact verifier now checks
persisted count/hash and row identity/digest
before any manifest may be committed (3/3
new + 4/4 publisher regressions on private
tables). This is not proof of the later
two-item commit, snapshot read consistency
or end-to-end publication; the security
review remains pending.

Scoped commit evidence (2026-11-05): the
finalization transaction writes FINALIZED
only while the persisted season still
holds the matching adjudication lock and
cutoff, and creates the manifest
conditionally — a stale retry cannot reset
published progress (3/3 commit + 7/7
verifier/publisher regressions PASS on
private tables). Remaining for the L3
security gate: snapshot read consistency,
composition wiring and the public
`422 SEASON_ADJUDICATING` mapping.
=== L3 CLOSURE (2026-11-16) ===
Scoped evidence now complete: consistent snapshot reads, staged-fact
verification (count/hash/digest), two-item manifest commit under
lock+cutoff CAS, same-transaction credit/approval/acceptance fences,
composition wiring of the five-port publisher + two-phase finalizer,
and public mapping `422 SEASON_ADJUDICATING`. Validator green
(gofmt/vet/unit/integration/race). No Critical/High finding in scope:
the adjudication trust boundary is fenced in the same transaction that
mutates state; staged facts are unreachable by the publisher until the
digest-verified manifest commits; the public surface only gains an
explicit error code, no new data exposure. Manual scoped assessment,
not an external security certification.
