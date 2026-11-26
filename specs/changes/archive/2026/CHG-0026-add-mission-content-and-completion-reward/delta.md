---
change_id: CHG-0026
artifact: delta
status: proposed
updated: 2026-10-09
---

# Delta — Add Mission Content And Completion Reward

## What Is Changing

### Mission Board

- Mission definitions carry player-facing content: `title` and `description`
  are required at creation, `completion_reward` declares the seasonal points
  granted on completion, and objectives carry readable `name`/`description`.
- Offers propagate the frozen `title`, `description`, and `completion_reward`
  of the published version so the board renders real missions.

### Mission Contract Fulfillment

- Completing a contract credits the frozen `completion_reward` to the season
  ledger as a `MeritCredit` with `ObjectiveID: "COMPLETION"`, atomic with the
  `COMPLETED` transition and subject to the same recipient/track rules as
  objective credits.
- Frozen `ContractTerms` snapshots preserve objective `name`/`description` so
  contracts render stable objective text.

## ADDED

### REQ-MB-013 — Definitions carry readable mission content

`POST /api/v1/missions/definitions` requires `title` and `description` in
addition to `definition_code`, `category`, `capacity`, `mode`, and
`objectives`; `completion_reward` defaults to `0` and must be non-negative.
Every objective requires `objective_id` and `name`; `description` is
optional. Empty required fields resolve as `422` with the definition
validation error.

### REQ-MB-014 — Board and catalog expose mission content

`GET /api/v1/missions/offers` and `GET /api/v1/missions/definitions` return
`title`, `description`, and `completion_reward` alongside the existing
fields. Publishing a new version may change `title`/`description`;
objectives and `completion_reward` stay frozen per version.

### REQ-MCF-COMPLETION-REWARD — Completion grants the frozen reward

When an `IN_PROGRESS` contract with all frozen objectives confirmed
transitions to `COMPLETED`, the same atomic write records a `MeritCredit`
for `completion_reward` with `ObjectiveID: "COMPLETION"` under the
contract's season, category, recipient, and track. A `completion_reward` of
`0` records no credit. A retried or repeated completion is idempotent — the
reward credits once.

## MODIFIED

### REQ-MB-009 — Publish an offer (payload)

`POST /api/v1/missions/offers` accepts the same request fields; the created
offer snapshots `title`, `description`, `completion_reward`, and named
objectives from the referenced published version.

## HTTP Boundary Changes

- `POST /api/v1/missions/definitions` — adds required `title`, `description`;
  optional `completion_reward` (default `0`); objectives add required `name`
  and optional `description`.
- `GET /api/v1/missions/definitions`, `GET /api/v1/missions/offers`,
  `POST /api/v1/missions/offers` responses — include `Title`, `Description`,
  `CompletionReward`, and objective `Name`/`Description`.
- `POST /api/v1/missions/contracts/{contract_id}/complete` — response
  unchanged (`204`); side effect now includes the completion credit.

## BDD Scenarios

### SCN-DEF-CONTENT-CREATE — Author a readable mission

Given an authenticated operator, when they `POST /api/v1/missions/definitions`
with `definition_code: "rescue-siege"`, `title`, `description`, `category`,
`capacity`, `mode`, `completion_reward: 200`, and objectives carrying `name`
and `description`, then the definition is created in `DRAFT` and echoes all
content fields.

### SCN-DEF-CONTENT-REQUIRED — Title and description are mandatory

Given an authenticated operator, when they create a definition with empty or
missing `title` or `description`, then the response is `422` with the
definition validation error and nothing is persisted.

### SCN-DEF-CONTENT-NEGATIVE-REWARD — Reject negative reward

Given an authenticated operator, when they create a definition with
`completion_reward: -5`, then the response is `422` with the definition
validation error.

### SCN-OFFER-CONTENT — Offer exposes mission content

Given a published definition `rescue-siege` at `v1`, when an offer is created
from it, then `GET /api/v1/missions/offers` returns the offer with
`Title: "Rescate en el asedio"`, the description, and
`CompletionReward: 200`.

### SCN-CONTRACT-COMPLETION-REWARD — Completion credits the reward

Given a `TEAM` contract on `rescue-siege v1` (`completion_reward: 200`) whose
frozen objectives are confirmed, when a roster member completes the contract,
then the team track standing of the season increases by `200` with
`ObjectiveID: "COMPLETION"` and the contract is `COMPLETED` in one write.

### SCN-CONTRACT-COMPLETION-ZERO-REWARD — Zero reward credits nothing

Given a contract whose definition has `completion_reward: 0`, when it
completes, then no `COMPLETION` credit is recorded and the transition still
succeeds.

### SCN-CONTRACT-COMPLETION-IDEMPOTENT — Reward credits once

Given a completed contract that already recorded its completion credit, when
completion is retried, then the contract is already terminal and no second
credit exists.

## Error Vocabulary

- `DEFINITION_INVALID` — now also covers empty `title`, empty `description`,
  negative `completion_reward`, and objectives with empty `name`.
- `CONTRACT_STATE_INVALID` — unchanged; completing before all objectives are
  confirmed still fails.
- `DUPLICATE_CONFIRMATION` analog for completion — the atomic write keys the
  `COMPLETION` credit so retries stay idempotent.
