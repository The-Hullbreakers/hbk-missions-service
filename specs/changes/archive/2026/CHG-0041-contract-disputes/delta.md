---
change_id: CHG-0041
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Contract Fulfillment

- Roster members raise disputes on rejected evidence or adjudication outcomes.
- `leave`/`abandon` on a contract with OPEN disputes returns 422 DISPUTES_PENDING.
- Moderators resolve disputes; resolution state is immutable.

## Requirements

### REQ-MCF-024 — Dispute an adjudication outcome (added)

A roster member opens a dispute on their contract; pending disputes block member exit; moderators resolve them.

## BDD Scenarios

### SCN-MCF-DISPUTE-RAISE — Member raises dispute

Given a rejected contract of player-1, when they raise a dispute with a reason, then it is OPEN and visible to reviewers.

### SCN-MCF-DISPUTE-BLOCK — Leave blocked while pending

Given a contract with an OPEN dispute, when a member leaves, then it returns 422 DISPUTES_PENDING.

### SCN-MCF-DISPUTE-RESOLVE — Moderator resolves

Given an OPEN dispute, when a moderator resolves it, then the dispute is RESOLVED and exit is unblocked.

## Error Vocabulary

- `DISPUTES_PENDING` — 422 (existing vocabulary, now enforced); `DISPUTE_STATE_INVALID` — 422 (new).
