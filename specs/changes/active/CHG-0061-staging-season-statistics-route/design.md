---
change_id: CHG-0061
artifact: design
status: implemented
updated: 2026-11-26
---

# Design — Staging season statistics route

## Domain / Application

- Not affected. The statistics handler (`SeasonStatisticsHandler`, canonical
  REQ-SCC-017) is already registered with `auth.Authenticate` in the
  composition mux.

## Infrastructure (template.yaml)

- Additive `MissionsSeasonStatisticsRoute` immediately before
  `MissionsFreezeSeasonRoute`, reusing `*RouteProps` for the API and Lambda
  integration — identical wiring to every neighboring mission route.
  Authentication stays in the Go route middleware, not the template.
- The `{season_id}` path part is a single segment; `statistics` is a literal
  suffix, so it cannot shadow or be shadowed by sibling routes
  (`/leaderboards/{category}`, `/standings/*`, `/statistics` are distinct
  literals at the same depth).

## Test strategy

- TEST-STG-SEASON-STATISTICS-ROUTE-001 (`tools/test-workflow-contract.sh`):
  extracts the 5-line `MissionsSeasonStatisticsRoute:` block with `sed` and
  asserts the exact `Type`, `<<: *RouteProps`, and `RouteKey` lines — static
  template contract, no AWS calls.
- Focused Go evidence: `go test ./tests/unit/missions/application -run
  '^TestGetSeasonStatistics' -count=1` keeps the existing application-level
  statistics behavior green (no code touched).
- `sam validate --template-file template.yaml --region us-east-2` for basic
  template validity (the known pre-existing `--lint` W1030 SNS warning is
  not re-litigated).

## Security / cost

- No IAM changes; existing auth reused. One additional route mapping — cost
  neutral. Rollback: delete the route block.
