---
change_id: CHG-0048
artifact: tasks
status: implemented
updated: 2026-10-20
---

# Tasks — CHG-0048

> Read-only projection; inside-out order applies: Application RED →
> Application GREEN/BLUE → Boundary RED → Boundary GREEN → gates.
> No Domain behavior partition: new types are pure projections; every RED
> still requires evidence review before its GREEN.

## Wave 1 — Application

- [x] **T-1 — Add failing unit tests for GetSeasonStatisticsHandler.**
  Owner: agent. Deps: none. Outcome: covers grouped buckets with
  participants/total_merits/top ordering, `current` resolution, category/
  track filtering, `top_n` slicing, empty season, `SEASON_NOT_FOUND`,
  `NO_ACTIVE_SEASON`, invalid track. Test:
  `tests/unit/missions/application/get_season_statistics_test.go` with
  hand-written port doubles (seasons, buckets, standings). Verify:
  `go test ./tests/unit/missions/application -run SeasonStatistics` fails
  — handler undefined. Maps to
  `internal/missions/application/get_season_statistics.go`,
  `internal/missions/domain/season_statistics.go`. Next: evidence review
  → T-2.

- [x] **T-2 — Implement GetSeasonStatisticsHandler + projection types.**
  Owner: agent. Deps: T-1 + `common-sdd-review-test-evidence` decision.
  Outcome: minimum GREEN per design (`domain/season_statistics.go`,
  `application/get_season_statistics.go`). Verify: unit tests pass.
  Next: T-3.

## Wave 2 — Boundary

- [x] **T-3 — Add failing http integration test for the statistics
  endpoint.** Owner: agent. Deps: T-2 (LAYER-GATE-APPLICATION). Outcome:
  `GET /api/v1/missions/seasons/{id}/statistics` and `.../current/
  statistics` through the wired route with in-memory stores; asserts
  grouping, filters, `top_n`, and error codes. Test:
  `tests/integration/http/season_statistics_test.go`. Verify:
  `go test ./tests/integration/http -run SeasonStatistics` fails — 404
  route missing. Next: evidence review → T-4.

- [x] **T-4 — Implement boundary handler, route, and error mapping.**
  Owner: agent. Deps: T-3 + evidence decision. Outcome:
  `interface/http/season_statistics.go` + wiring in
  `composition/server.go` (`auth`, no role). Verify: integration test
  passes. Next: T-5.

## Wave 3 — BLUE, docs, gates

- [x] **T-5 — BLUE refactor while green.** Owner: agent. Deps: T-4.
  Outcome: Clean Code pass; maintained files < 150 lines (test files
  split into `*_errors_test.go`).
- [x] **T-6 — Update public docs.** Owner: agent. Deps: T-4. Outcome:
  `docs/API-CONTRACT.md` (Competition section) and
  `docs/API-GUIDE-FRONTEND.md` document the endpoint, params, and payload.
- [x] **T-7 — Run validate script and verify.** Owner: agent. Deps: T-5,
  T-6. Outcome: `tools/validate-sdd-change.sh` passed L2 — gofmt, unit,
  integration, race; coverage domain 90.8%, application 91.2%,
  infrastructure 92.7%. Final evidence via `common-sdd-verify-spec`.
  Next: consolidate REQ-SCC-017 into `seasonal-category-competition`
  spec and archive.

## Requirement to Test to Production Map

- **REQ-SCC-017** → `get_season_statistics_test.go` (application) +
  `season_statistics_test.go` (http) → `application/`,
  `domain/season_statistics.go`, `interface/http/`, `composition/server.go`.

## Gates

Standard per L2: evidence review after each RED; coverage gate on
affected scope; documentation before final review.
