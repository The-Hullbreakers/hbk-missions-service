# Delta — mission-board

Affected requirements: `REQ-MB-010 — List mission offers`, `REQ-MB-004 —
Discover eligible offers` (filters list), endpoint catalog.

## MODIFIED

### REQ-MB-010 — List mission offers

`GET /api/v1/missions/offers` keeps `season_id` and `status`
(`AVAILABLE` default | `ALL`) and extends discovery:

- `category` accepts a comma-separated list of categories (OR semantics,
  case-insensitive, trimmed). Unknown values keep `400 INVALID_QUERY_PARAMETER`.
- `mode` accepts a comma-separated list of mission modes (OR semantics).
  Unknown values keep `400 INVALID_QUERY_PARAMETER`.
- `q` is a case-insensitive substring match over `Title` or `Description`.
- `limit` defaults to `50`, clamps to max `200`; `offset` defaults to `0`;
  invalid numbers keep `400 INVALID_QUERY_PARAMETER`.
- `sort` accepts `completion_reward` or `category`, optional `-` prefix for
  descending; unknown fields keep `400 INVALID_QUERY_PARAMETER`. Ties and the
  default order use `OfferID` ascending.
- The response is `{ "offers": [...], "total": <int>, "limit": <int>,
  "offset": <int> }`; `total` counts matching offers before pagination.

The season resolution, `NO_ACTIVE_SEASON`/`ACTIVE_SEASON_EXISTS` handling,
and the `REQ-MB-008` active-definition exclusion remain unchanged and apply
before pagination counts.

#### SCN-MB-LIST-MULTI — Multiple categories filter with OR

Given offers in categories `BOUNTY` and `RACING`, when a member calls
`GET /offers?category=bounty,racing`, then both appear and any other
category is excluded.

#### SCN-MB-LIST-SEARCH — Substring search on title or description

Given an offer titled "Night Drift" and another "Harbor Escort", when a
member calls `GET /offers?q=drift`, then only "Night Drift" is returned.

#### SCN-MB-LIST-PAGE — Paginated envelope

Given 3 matching offers, when a member calls `GET /offers?limit=2&offset=2`,
then the response is `200` with `offers` length 1, `total` 3, `limit` 2,
`offset` 2.

#### SCN-MB-LIST-SORT — Reward descending

Given offers with rewards 100 and 500, when a member calls
`GET /offers?sort=-completion_reward`, then the 500-reward offer is first.

#### SCN-MB-LIST-INVALID — Invalid multi-value or paging input

Given an unknown mode in `mode=solo,arcade`, a negative `limit`, or an
unknown `sort=price`, when a member lists offers, then the response is
`400 INVALID_QUERY_PARAMETER`.

## ADDED

### REQ-MB-013 — List offer facets

`GET /api/v1/missions/offers/facets` returns
`{ "categories": [...], "modes": [...] }` with the distinct, sorted values
present among offers visible under the resolved season and `status`
(`AVAILABLE` default | `ALL`). Other filters do not apply. Season resolution
failures map exactly like the list endpoint.

#### SCN-MB-FACETS — Distinct categories and modes

Given offers with categories `BOUNTY`,`RACING`,`RACING` and modes
`SOLO`,`TEAM`, when a member calls `GET /offers/facets`, then
`categories` is `["BOUNTY","RACING"]` and `modes` is `["SOLO","TEAM"]`.

## REMOVED

- None. The bare-array response of `GET /offers` is superseded by the
  envelope (breaking change accepted in Gate 1).
