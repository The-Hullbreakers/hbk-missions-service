# Proposed API Contract

This surface expresses Gate 1's business needs; Routes and DTOs will be validated before Boundary RED.

## Member Board

```text
GET  /api/v1/missions/offers
GET  /api/v1/missions/offers/facets
GET  /api/v1/missions/offers/{offer_id}
POST /api/v1/missions/offers/{offer_id}/accept
GET  /api/v1/missions/contracts?view=active|history
GET  /api/v1/missions/contracts/{contract_id}
```

`GET /api/v1/missions/offers` accepts:

- `season_id` — defaults to the single active season.
- `lang` — `es` (default when omitted) or `en`; selects the served language
  of player-facing content (titles, descriptions, rules, objective names).
  A record storing only Spanish serves Spanish under any `lang`, including
  `lang=en`, and reports `ContentLanguage: "es"`. An unsupported, empty, or
  conflicting repeated `lang` returns `400 INVALID_QUERY_PARAMETER`;
  `Accept-Language` is never negotiated. When every item in a non-empty
  payload resolves the same language, the response also carries a
  `Content-Language` header; mixed-language payloads omit it. Each item
  declares its served language via `ContentLanguage` (PascalCase board
  projections) or `content_language` (contract detail, review queue).
- `status` — `AVAILABLE` (default, offers open for acceptance) or `ALL`.
- `category` — comma-separated list, OR semantics (e.g. `category=BOUNTY,RACING`).
- `mode` — comma-separated list of `SOLO|TEAM|FLEX`, OR semantics.
- `q` — case-insensitive substring over `Title` and `Description`.
- `limit` — page size, default `50`, max `200`.
- `offset` — items to skip, default `0`.
- `sort` — `completion_reward` or `category`; prefix `-` for descending.
  Default order is stable by `OfferID`.

Unknown `status`, `mode`, `sort`, or malformed `limit`/`offset` return
`400 INVALID_QUERY_PARAMETER`.

The response is a paginated envelope; `total` counts the visible offers
before pagination:

```json
{
  "offers": [{ "OfferID": "offer-1", "...": "..." }],
  "total": 37,
  "limit": 50,
  "offset": 0
}
```

Each offer exposes the frozen objectives and the maximum earnable merits:

```json
{
  "Objectives": [
    { "ObjectiveID": "pvp-kills", "Name": "Derribos verificados",
      "Category": "PVP_AIR", "MeritPerUnit": 270, "Bonus": 0, "UnitLimit": 1 }
  ],
  "CompletionReward": 0,
  "MeritSummary": { "ObjectiveMerits": 300, "CompletionReward": 0, "TotalMerits": 300 }
}
```

`ObjectiveMerits` is `Σ (MeritPerUnit + Bonus) × UnitLimit` over the frozen
objectives; `TotalMerits` adds `CompletionReward`. Offers stored before this
field existed return an empty `Objectives` list.

`GET /api/v1/missions/contracts` returns the caller's contracts filtered by
`season_id` and `status` (`ALL` default, `ACTIVE` for non-terminal, or an
exact state). Each item's embedded `Offer` exposes `Objectives` and
`MeritSummary` derived from the contract's **frozen `Terms`** — the same
formula as offers — so the earnable total always reflects the promise
captured at acceptance, never live offer data. Contracts without terms
report `ObjectiveMerits: 0` and `TotalMerits` equal to `CompletionReward`.

`GET /api/v1/missions/offers/facets` returns the distinct categories and
modes present under the resolved season and `status`, for filter pickers:

```json
{ "categories": ["BOUNTY", "RACING"], "modes": ["SOLO", "TEAM"] }
```

Acceptance does not receive identity claims, merit recipient or points. For `FLEX` you receive the chosen mode and, when applicable, party reference.

The optional body carries `team_id` and `teammates`. `team_id` is meaningful
only on `TEAM`/`FLEX` offers — a `team_id` on a `SOLO` offer returns `422
TEAM_NOT_ALLOWED`, and `TEAM` offers without it return `422 TEAM_ID_REQUIRED`.
Declared teammates are validated against team membership (`422
TEAMMATE_NOT_MEMBER`) and each one consumes a slot of their own active-contract
limit (`422 TEAMMATE_LIMIT_REACHED`). The committed roster is the requester
plus every validated teammate; each confirmed member consumes one active slot
and team credits land on the team track only.

## Mission Definitions

```text
POST /api/v1/missions/definitions
POST /api/v1/missions/definitions/{definition_code}/publish
```

Request:

```json
{
  "definition_code": "bounty-template",
  "title": "Elimina 100 enemigos",
  "description": "Derriba 100 hostiles en el sector asignado",
  "category": "BOUNTY_HUNTING",
  "mode": "SOLO",
  "completion_reward": 200,
  "objectives": [
    {"objective_id": "obj-kills", "name": "Elimina atacantes", "merit_per_unit": 10, "bonus": 5, "unit_limit": 100}
  ],
  "rules": ["Solo PU/LIVE", "Video continuo sin cortes"]
}
```

`title`, `description` and each objective's `name` are required (`INVALID_MISSION_DEFINITION` on missing/empty). `description` is Markdown-formatted content that clients render verbatim. `rules` is an optional ordered list of non-blank requirement strings the client renders as bullets; an empty or blank item resolves as `INVALID_MISSION_DEFINITION`. `completion_reward` defaults to `0` and must be non-negative. Approving a submitted contract credits the frozen `completion_reward` once to the season ledger when every objective is confirmed; the credit is atomic with the `COMPLETED` transition.

Publishing an offer is now template/pool-driven:

```text
POST /api/v1/missions/offers
```

Request:

```json
{
  "offer_id": "offer-1",
  "definition_code": "bounty-template",
  "definition_version": "v1"
}
```

`definition_code` is required and `definition_version` is optional. The published offer is bound to the active season and inherits the definition's frozen category and mode.

The list excludes definitions already active for the user on the server. Acceptance verifies it again even if the offer is queried by ID.

Limit rejection response:

```json
{
  "code": "ACTIVE_MISSION_LIMIT_REACHED",
  "activeCount": 3,
  "effectiveLimit": 3,
  "category": "BOUNTY_HUNTING"
}
```

## Contract Progress

```text
POST /api/v1/missions/contracts/{contract_id}/start
POST /api/v1/missions/contracts/{contract_id}/evidence
POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url
POST /api/v1/missions/contracts/{contract_id}/evidence/media-url
POST /api/v1/missions/contracts/{contract_id}/submit
POST /api/v1/missions/contracts/{contract_id}/leave
POST /api/v1/missions/contracts/{contract_id}/abandon
```

Evidence media flow: the client requests an upload grant
(`{content_type, content_length}` → `201 {upload_url, media_uri}`), PUTs
the object to `upload_url`, then submits evidence with the optional
`media_uri` field — the reference must live under
`evidence/{contract_id}/` and exist in storage (`422
EVIDENCE_MEDIA_INVALID` otherwise). Disallowed content types and
declared oversize return `400 INVALID_REQUEST`. Grants require roster
membership or `admin`/`moderator`; read grants
(`{media_uri}` → `200 {download_url}`) are restricted to
`admin`/`moderator` reviewers.

Confirmed merit is staged on the contract — it does not pay until review.
`POST .../submit` moves an `IN_PROGRESS` contract to `PENDING_VALIDATION`
(roster member only, any confirmation level; `422
CONTRACT_STATE_INVALID` otherwise). Pending contracts close evidence,
upload and read grants; `leave`/`abandon` remain available.

Reviewer operations are separate from participant submission and require an explicit capability.

## Competition

```text
GET /api/v1/missions/seasons/current
GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}?track=individual|team&limit=20&cursor=<opaque>&min_merits=0
GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me
GET /api/v1/missions/seasons/{season_id}/standings/history
GET /api/v1/missions/seasons/{season_id}/statistics?category=&track=&top_n=3
```

`GET /api/v1/missions/seasons/{season_id}/statistics` returns a
dashboard-ready season summary in one call, grouped by `{category, track}`
bucket. `season_id` accepts `current` to resolve the single `ACTIVE`
season. Unlike `standings/history` it serves any season state — `ACTIVE`
standings are live.

- Query parameters:
  - `category` (optional) narrows the buckets to one category.
  - `track` (optional, `individual` or `team`) narrows the buckets to one track.
  - `top_n` (optional, default `3`, max `50`) caps the ranked entries per bucket.
- `400 INVALID_QUERY_PARAMETER` for an invalid `track` or a malformed/out-of-range `top_n`.
- `404 NO_ACTIVE_SEASON` when `current` is requested with no active season; `404 SEASON_NOT_FOUND` for a missing season; `422 ACTIVE_SEASON_EXISTS` signals a broken single-active invariant.

Response:

```json
{
  "SeasonID": "season-1",
  "State": "ACTIVE",
  "Official": true,
  "Buckets": [
    {
      "Bucket": { "Category": "BOUNTY_HUNTING", "Track": "INDIVIDUAL" },
      "Participants": 12,
      "TotalMerits": 2500,
      "Top": [
        { "Rank": 1, "RecipientID": "player-1", "Track": "INDIVIDUAL",
          "Total": 150, "CompletedMissions": 3, "UniqueObjectives": 4 }
      ]
    }
  ]
}
```

`Participants` counts recipients with credits in the bucket and
`TotalMerits` sums them; `Top` carries the first `top_n` ranked entries
using the same public leaderboard projection (rank, total, tie-break
facts) — never evidence, contract ids, or unit ids. A season with no
credits returns `200` with `Buckets: []`.

`GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}` returns a paginated leaderboard for the requested category and track.

- Query parameters:
  - `track` (required, `individual` or `team`) determines the leaderboard track.
  - `limit` (optional, default `20`, max `100`) controls page size.
  - `cursor` (optional, opaque) continues a previous page.
  - `min_merits` (optional, default `0`) excludes recipients below the threshold.
- Unknown `track` values return `400 Bad Request` with `INVALID_QUERY_PARAMETER`.
- A missing or non-existent `season_id` returns `404 Not Found` with `SEASON_NOT_FOUND`.
- Missing or invalid credentials return `401 Unauthorized`.

Response:

```json
{
  "items": [
    {
      "rank": 1,
      "recipient_id": "player-1",
      "track": "INDIVIDUAL",
      "total": 150,
      "completed_missions": 3,
      "unique_objectives": 4
    }
  ],
  "next_cursor": ""
}
```

The response includes rank, merits and public tie-break facts; it never includes private evidence, contract identifiers or unit identifiers.

## Administration

```text
POST /api/v1/missions/definitions
POST /api/v1/missions/definitions/{id}/publish
POST /api/v1/missions/offers
GET  /api/v1/seasons{?status=active|inactive|all}
POST /api/v1/seasons
PUT  /api/v1/missions/policies/active-contract-limit
GET  /api/v1/missions/review-queue
GET  /api/v1/missions/contracts/{contract_id}
POST /api/v1/missions/contracts/{contract_id}/approve
POST /api/v1/missions/contracts/{contract_id}/reject
POST /api/v1/missions/seasons/{id}/freeze
POST /api/v1/missions/seasons/{id}/finalize
```

`GET /api/v1/seasons` defaults to `status=active`. `status=inactive` returns any season whose state is not `ACTIVE`. Unknown values return `INVALID_QUERY_PARAMETER`.

`GET /api/v1/missions/review-queue` lists every `PENDING_VALIDATION`
contract for `admin`/`moderator` reviewers (`403 FORBIDDEN` otherwise).
Each entry carries `contract_id`, `offer_id`, `definition_code`, `title`,
`state`, `roster`, `season_id`, `team_id`, `confirmed_count`, and
`pending_credits` (the staged credit count) — enough to size and select
adjudication work. An empty queue returns `200` with `[]`.

`GET /api/v1/missions/contracts/{contract_id}` returns the reviewer-facing
contract detail — `admin`/`moderator` only (`403 FORBIDDEN` otherwise,
`404 CONTRACT_NOT_FOUND` for unknown ids). The projection carries the
frozen `terms`, `roster`, `departures`, `confirmations`,
`pending_credits` with each staged credit's `media_uri`, and
`rejected_units` — the units a verdict excluded, each as
`{objective_id, unit_id, recipient_id}`; members see only their own.
Media bytes still require the reviewer read grant.

`POST .../approve` and `POST .../reject` adjudicate a `PENDING_VALIDATION`
contract (`admin`/`moderator` only, `422
CONTRACT_NOT_PENDING_VALIDATION` otherwise). Approval accepts an optional
verdict body — `{"rejected_units": [{"objective_id", "unit_id"}]}` — that
excludes the listed staged units; every unit not listed is approved. An
unknown unit reference fails the whole call with `422 UNIT_NOT_STAGED`.
Approval materializes every surviving staged credit — plus the completion
reward when every mandatory objective reached its `unit_limit` in
approved units — in one atomic write and returns the committed credits;
replayed approvals are idempotent. An empty body approves all units.
Rejection writes `DISQUALIFIED` with zero credits.

`POST /api/v1/seasons` accepts a manual season with `season_id` or a scheduled season with `starts_at` and `schedule`:

```json
{
  "official_statistics": true,
  "starts_at": "2026-09-02T00:00:00Z",
  "schedule": {
    "active_weeks": 4,
    "off_weeks": 1,
    "series": {
      "prefix": "season-2026",
      "iteration": 1
    }
  }
}
```

For scheduled seasons the `season_id` is derived as `{prefix}-{iteration}`. When `schedule` is omitted the request behaves as a manual season.

Administration notes:

- `POST /api/v1/missions/seasons/{id}/freeze` requires the season to be `ACTIVE`; it returns the season with `State: FREEZING`.
- `POST /api/v1/missions/seasons/{id}/finalize` requires the season to be `FREEZING`; it returns the season with `State: FINALIZED`.
- A non-existent season returns `404 Not Found` with `SEASON_NOT_FOUND`.
- An invalid state transition returns `422 Unprocessable Entity` with `SEASON_STATE_INVALID`.
- While `FREEZING`, only contracts already `IN_PROGRESS` can still earn merits; a finalized season rejects all confirmations.
- Creating a `season_id` that already exists returns `422 SEASON_ALREADY_EXISTS`; the stored season is never overwritten, so its state and `Official` flag are immutable through creation.
- Scheduled seasons auto-freeze at the end of `active_weeks` and auto-finalize at the end of `off_weeks`; the scheduler then creates the next `ACTIVE` season in the series. If that creation fails, the season records pending next-season creation and the scheduler retries it on every tick until it succeeds.

## Error Vocabulary

`OFFER_NOT_AVAILABLE`, `NOT_ELIGIBLE`, `MISSION_DEFINITION_ALREADY_ACTIVE`, `ACTIVE_MISSION_LIMIT_REACHED`, `ROSTER_INVALID`, `CONTRACT_STATE_INVALID`, `CONFIRMATION_DUPLICATE`, `EVIDENCE_REJECTED`, `EVIDENCE_MEDIA_INVALID`, `SEASON_NOT_ACTIVE`, `SEASON_MISMATCH`, `NO_ACTIVE_SEASON`, `SEASON_ALREADY_EXISTS`, `ACTIVE_SEASON_EXISTS`, `TEAM_ID_REQUIRED`, `TEAM_NOT_ALLOWED`, `NOT_TEAM_MEMBER`, `TEAMMATE_NOT_MEMBER`, `TEAMMATE_LIMIT_REACHED`, `INVALID_QUERY_PARAMETER`, `SEASON_STATE_INVALID`, `SEASON_NOT_FOUND`, `DISPUTES_PENDING`.
