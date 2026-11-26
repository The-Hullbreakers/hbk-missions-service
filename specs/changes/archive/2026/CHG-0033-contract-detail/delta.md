---
change_id: CHG-0033
artifact: delta
status: consolidated
updated: 2026-10-13
---

# Delta — Contract Detail for Adjudicators

## What Is Changing

### Mission Contract Fulfillment

- Adjudicators (`admin`/`moderator`) can fetch a single contract's full
  detail — frozen terms, roster, confirmed units, and staged credits with
  their evidence `media_uri` — to inform approve/reject decisions.
- Member-facing visibility is unchanged: `GET /contracts` remains the
  roster-scoped read; the detail endpoint is reviewer-only so teammates'
  evidence references are never exposed to members.

## ADDED Requirements

### REQ-MCF-022 — Get a contract detail for adjudication

`GET /api/v1/missions/contracts/{contract_id}` returns `200` with the
contract detail: `contract_id`, `offer_id`, `definition_code`, `title`,
frozen `terms`, `state`, `roster`, `departures`, `season_id`, `team_id`,
`confirmations` (confirmed unit ids), and `pending_credits` — each staged
credit carrying `objective_id`, `unit_id`, `recipient_id`, `amount`, and
optional `media_uri`. Requires `admin` or `moderator`; any other caller
receives `403 FORBIDDEN`. Unknown ids resolve as
`422 CONTRACT_NOT_FOUND`.

#### SCN-MCF-DETAIL-REVIEWER — Reviewer inspects a pending contract

Given contract `c1` in `PENDING_VALIDATION` with two staged credits, one
carrying `media_uri`, when a moderator requests `GET /contracts/c1`, then
the response is `200` with the frozen terms, roster, confirmations, and
both staged credits including the media reference.

#### SCN-MCF-DETAIL-ANY-STATE — Detail works for any contract state

Given contract `c2` is `IN_PROGRESS`, when an admin requests
`GET /contracts/c2`, then the response is `200` — adjudicators can
inspect any state, not only `PENDING_VALIDATION`.

#### SCN-MCF-DETAIL-MEMBER-DENIED — Member cannot fetch the detail

Given `player-1` is a roster member of `c1` with no privileged role, when
they request `GET /contracts/c1`, then the response is `403 FORBIDDEN` —
member reads stay roster-scoped on `GET /contracts`.

#### SCN-MCF-DETAIL-NOT-FOUND — Unknown contract resolves cleanly

Given no contract `c9` exists, when a moderator requests
`GET /contracts/c9`, then the response is `404 CONTRACT_NOT_FOUND`.

## Unchanged

- Contract lifecycle, adjudication, credit staging — REQ-MCF-019/020.
- Member list projection and filters — REQ-MCF-006.
- Evidence read grants remain the only way to download `media_uri` — the
  detail exposes references, not bytes.

## Error Vocabulary

Reuses `FORBIDDEN`, `CONTRACT_NOT_FOUND`, `UNAUTHORIZED`,
`INTERNAL_ERROR`. No new codes.
