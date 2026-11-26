---
change_id: CHG-0046
artifact: tasks
status: completed
updated: 2026-10-19
---

# Tasks — CHG-0046

## Wave 1 — Application RED

- [x] **T-1 — Add failing list-offers assertions for objectives and
  merit_summary.** Owner: agent. Deps: none. Outcome: `ListOffersHandler`
  unit test expects structured objectives and derived totals from the
  offer. Verify: `go test ./tests/unit/missions/application` fails for the
  right reason. Next: T-2.

## Wave 2 — Domain RED/GREEN/BLUE

- [x] **T-2 — Add failing offer freeze + summary tests.** Owner: agent. Deps:
  T-1. Outcome: `NewMissionOfferFromDefinition` must copy objectives;
  `MeritSummary` must compute `Σ (mpu + bonus) × limit + reward` and handle
  empty objectives. Verify: domain tests RED. Next: T-3.
- [x] **T-3 — Implement objectives freeze and MeritSummary.** Owner: agent.
  Deps: T-2 + evidence review. Outcome: minimum GREEN. Verify: domain tests
  pass. Next: T-4.
- [x] **T-4 — BLUE domain while green.** Owner: agent. Deps: T-3. Outcome:
  naming/size cleanup, files <150 lines. Next: LAYER-GATE-APPLICATION.

## Wave 3 — Application GREEN/BLUE + core gate

- [x] **T-5 — Green application tests via domain changes.** Verify: unit suite
  passes. Next: boundary.

## Wave 4 — Boundary RED/GREEN/BLUE

- [x] **T-6 — Add failing HTTP offer-payload tests.** Outcome: `GET /offers`
  and `GET /offers/{id}` projections include `objectives` and
  `merit_summary`. Verify: interface tests RED. Next: T-7.
- [x] **T-7 — Implement offer payload projection.** Verify: interface tests
  pass; http integration suite green. Next: T-8.
- [x] **T-8 — Extend `offer_item` seeds to keep `Objectives`.** Verify:
  `tests/seeds` pass. Next: T-9.

## Wave 5 — Infrastructure

- [x] **T-9 — Persist and decode offer objectives in DynamoDB.** Outcome:
  `offer_store` writes the attribute; `mapOfferItem` tolerates legacy items
  without it. Verify: `tests/integration/infrastructure/dynamodb` green.

## Wave 6 — Gates and consolidation

- [x] **T-10 — Run validate script, coverage, and clean-up; verify spec.**
  Verify: `tools/validate-sdd-change.sh` clean, suites green from clean
  state. Next: consolidate into `mission-board` spec and archive.
