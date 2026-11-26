---
change_id: CHG-0032
artifact: delta
status: consolidated
updated: 2026-10-12
---

# Delta — Review Queue for Pending Contracts

## What Is Changing

### Mission Contract Fulfillment

- Adjudicators (`admin`/`moderator`) can list every contract currently in
  `PENDING_VALIDATION` through a dedicated read endpoint — the discovery
  gap left by CHG-0031.
- Member-facing listings are unchanged: `GET /contracts` stays
  roster-scoped; the queue never returns contracts outside the pending
  state.

## ADDED

### REQ-MCF-021 — List the pending review queue

`GET /api/v1/missions/review-queue` returns `200` with the contracts in
`PENDING_VALIDATION`, each carrying `contract_id`, `offer_id`,
`definition_code`, `title`, `state`, `roster`, `season_id`, `team_id`,
the confirmed-unit count, and `pending_credits` (staged credit count).
Requires `admin` or `moderator` — any other caller receives
`403 FORBIDDEN`. An empty queue returns `200` with `[]`.

## MODIFIED

None — member contract listing and adjudication semantics unchanged.

## BDD Scenarios

### SCN-MCF-QUEUE-REVIEWER — Reviewer sees the pending queue

Given contracts `c1` (PENDING_VALIDATION) and `c2` (IN_PROGRESS) exist,
when a moderator requests the review queue, then the response is `200`
with `c1` only, including its staged-credit count.

### SCN-MCF-QUEUE-MEMBER-DENIED — Member cannot list the queue

Given `player-1` holds no privileged role, when they request the review
queue — even while owning a pending contract — then the response is
`403 FORBIDDEN`.

### SCN-MCF-QUEUE-EMPTY — Empty queue returns an empty list

Given no `PENDING_VALIDATION` contracts exist, when a moderator requests
the review queue, then the response is `200` with `[]`.

### SCN-MCF-QUEUE-DRAINS — Queue reflects adjudication

Given a pending contract, when a moderator approves or rejects it, then
a subsequent queue request no longer contains it.

## Error Vocabulary

- `FORBIDDEN` — `403` (caller without `admin`/`moderator` role).
- `UNAUTHORIZED` — `401` (missing or invalid token; existing middleware).
