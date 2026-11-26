---
change_id: CHG-0035
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Contract Fulfillment

- `GET /api/v1/missions/contracts/{contract_id}` is readable by roster members, not only reviewers.
- Member view exposes contract terms, state, roster, and own evidence references; reviewer-only internals stay out.

## Requirements

### REQ-MCF-022 — Read own contract detail (added)

A roster member can read the detail of a contract they belong to; non-members and outsiders get FORBIDDEN/NOT_FOUND.

## BDD Scenarios

### SCN-MCF-DETAIL-MEMBER — Member reads own contract

Given a contract with roster [player-1], when player-1 requests GET /contracts/{id}, then it returns 200 with the member projection.

### SCN-MCF-DETAIL-OUTSIDER — Non-member denied

Given a contract with roster [player-1], when player-2 requests it, then it returns 403 FORBIDDEN (or 404, per anti-enumeration decision [NEEDS CLARIFICATION]).

## Error Vocabulary

- Reuses `FORBIDDEN` — 403.
