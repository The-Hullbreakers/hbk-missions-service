# CHG-0056 - Design and boundary

## Model and use case

Application RED opens `GetOverallStandingsHandler` with a season
reader and a focused `OverallCreditsReader` port:
`ListOverallCredits(seasonID string) ([]domain.MeritCredit, map[string]bool, error)`.
The map reports completed contract IDs for the same credits; the
adapter fetches both facts in one focused query operation. The
Application handler takes caller identity, explicit season ID,
limit and cursor, returning a typed projection of season ID,
participant count, page and optional caller entry.
Domain computes one season/INDIVIDUAL projection per recipient from
credit facts; count distinct completed contract IDs once across all
categories and unique `(category, objective_id)` pairs. Per-category
`Standing` summaries cannot safely be summed: one contract can
contribute to more than one category. Reuse `CompareStanding` and
pagination only after testing exhausted/invalid cursors; do not
change a current leaderboard's observable contract as a side effect.
A new focused DynamoDB query adapter reads paginated season/track
credit facts and completion status without exposing evidence or
private identities. Start with the existing `PK=MERIT_CREDIT` access
pattern and measure read volume/latency; a new index or materialized
ranking needs a separately reviewed amendment if scale demands it.
No new stored leaderboard.

## Public REST contract

`GET /api/v1/missions/seasons/{season_id}/standings/overall?limit=10&cursor=...`
requires the existing authenticated session. Season ID is explicit
and bound to one season. Default limit 10, maximum 50; an invalid
limit/cursor is `400 INVALID_QUERY_PARAMETER` or `INVALID_CURSOR`.
Unknown season is `404 SEASON_NOT_FOUND`; no credits returns `200`
with `participants: 0`, `items: []`, `next_cursor: ""`, and
`my_standing: null`; missing session returns `401 UNAUTHORIZED`.
The response contains `season_id`, `track: "INDIVIDUAL"`,
`participants`, `items` (the existing public leaderboard entry shape),
`next_cursor`, and `my_standing` (same entry shape or null). The
identity used for `my_standing` comes only from the session, never
the request. No extra profile or last-login fields.

Place the new authenticated HTTP handler in
`internal/missions/interface/http/overall_standings.go`, route and
use-case wiring in composition, and a matching GET event in
`template.yaml`. The local SAM generator belongs to the separately
approved shared-infrastructure delta. Keep routes for per-category
leaderboard, per-category `/me`, statistics, and history.

## Layer gates and scope

Domain and Application are affected; first Application RED, then
demanded Domain RED/GREEN/BLUE, then Application GREEN/BLUE. Boundary
RED follows the core gate. Infrastructure gets one focused
overall-credit-facts read port and adapter; it does not extend the
existing broad store interface or alter category ranking behavior.
`template.yaml` is an existing long provider configuration file:
document its owned file-size exception with expiry instead of
truncating unrelated configuration. All new files stay below 150
physical lines. Authentication and public response privacy require
security review; locally test with an isolated DynamoDB test table.
