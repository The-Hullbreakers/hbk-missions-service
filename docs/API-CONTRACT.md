# Proposed API Contract

This surface expresses Gate 1's business needs; Routes and DTOs will be validated before Boundary RED.

## Member Board

```text
GET  /api/v1/missions/offers
GET  /api/v1/missions/offers/{offer_id}
POST /api/v1/missions/offers/{offer_id}/accept
GET  /api/v1/missions/contracts?view=active|history
GET  /api/v1/missions/contracts/{contract_id}
```

Acceptance does not receive `tenant_id`, merit recipient or points. For `FLEX` you receive the chosen mode and, when applicable, party reference.

## Mission Definitions

```text
POST /api/v1/missions/definitions
POST /api/v1/missions/definitions/{definition_id}/publish
```

Publishing an offer is now template/pool-driven:

```text
POST /api/v1/missions/offers
```

Request:

```json
{
  "tenant_id": "tenant-1",
  "offer_id": "offer-1",
  "definition_id": "bounty-template",
  "definition_version": "v1",
  "capacity": 10
}
```

`definition_id` is required and `definition_version` is optional. The published offer is bound to the active season and inherits the definition's frozen category, capacity and mode.

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
POST /api/v1/missions/contracts/{contract_id}/submit
POST /api/v1/missions/contracts/{contract_id}/abandon
```

Reviewer operations are separate from participant submission and require an explicit capability.

## Competition

```text
GET /api/v1/missions/seasons/current
GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}?track=individual|team&limit=20&cursor=<opaque>&min_merits=0
GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me
GET /api/v1/missions/seasons/{season_id}/standings/history
```

`GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}` returns a paginated leaderboard for the requested category and track.

- Query parameters:
  - `track` (required, `individual` or `team`) determines the leaderboard track.
  - `limit` (optional, default `20`, max `100`) controls page size.
  - `cursor` (optional, opaque) continues a previous page.
  - `min_merits` (optional, default `0`) excludes recipients below the threshold.
- Unknown `track` values return `400 Bad Request` with `INVALID_QUERY_PARAMETER`.
- A missing or non-existent `season_id` returns `404 Not Found` with `SEASON_NOT_FOUND`.
- Missing tenant context returns `403 Forbidden` with `TENANT_BOUNDARY_DENIED`.

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
POST /api/v1/missions/contracts/{id}/adjudications
POST /api/v1/missions/seasons/{id}/freeze
POST /api/v1/missions/seasons/{id}/finalize
```

`GET /api/v1/seasons` defaults to `status=active`. `status=inactive` returns any season whose state is not `ACTIVE`. Unknown values return `INVALID_QUERY_PARAMETER`.

`POST /api/v1/seasons` accepts a manual season with `season_id` or a scheduled season with `starts_at` and `schedule`:

```json
{
  "tenant_id": "tenant-1",
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
- Scheduled seasons auto-freeze at the end of `active_weeks` and auto-finalize at the end of `off_weeks`; the scheduler immediately creates the next `ACTIVE` season in the series.

## Error Vocabulary

`OFFER_NOT_AVAILABLE`, `NOT_ELIGIBLE`, `MISSION_DEFINITION_ALREADY_ACTIVE`, `ACTIVE_MISSION_LIMIT_REACHED`, `OFFER_CAPACITY_EXHAUSTED`, `ROSTER_INVALID`, `CONTRACT_STATE_INVALID`, `CONFIRMATION_DUPLICATE`, `EVIDENCE_REJECTED`, `SEASON_NOT_ACTIVE`, `SEASON_MISMATCH`, `NO_ACTIVE_SEASON`, `INVALID_QUERY_PARAMETER`, `SEASON_STATE_INVALID`, `SEASON_NOT_FOUND`, `DISPUTES_PENDING` y `TENANT_BOUNDARY_DENIED`.
