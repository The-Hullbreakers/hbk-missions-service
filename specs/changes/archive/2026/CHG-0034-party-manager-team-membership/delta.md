---
change_id: CHG-0034
artifact: delta
status: proposed
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Contract Fulfillment

- Declared teammates are verified against the real team roster (Party Manager) instead of a deny-all stub.
- When the membership source is unavailable, acceptance fails closed — no contract is created.

## Requirements

### REQ-MCF-001 — Execute solo or team (modified)

Teammate validation reads the authoritative team membership source; unverifiable teammates are rejected.

## BDD Scenarios

### SCN-MCF-MEMBER-VALID — Verified teammates accepted

Given a TEAM offer and teammates that Party Manager confirms as members of the caller team, when the member accepts, then the contract is created with the full roster.

### SCN-MCF-MEMBER-INVALID — Non-member teammate rejected

Given a declared teammate is not a member of the caller team, when acceptance is attempted, then it returns 422 TEAMMATE_NOT_MEMBER and no contract is created.

### SCN-MCF-MEMBER-UNAVAILABLE — Membership source down fails closed

Given Party Manager is unreachable or returns a non-definitive response, when a member accepts a TEAM offer with teammates, then the acceptance is rejected with 503 `MEMBERSHIP_UNAVAILABLE` and no contract is created.

## Error Vocabulary

- `TEAMMATE_NOT_MEMBER` — 422 (existing): definitive negative from Party Manager.
- `MEMBERSHIP_UNAVAILABLE` — 503 (new): source unreachable, timeout, or non-definitive response; always fail closed.
