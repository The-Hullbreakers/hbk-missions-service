---
change_id: CHG-0049
artifact: tasks
status: active
updated: 2026-10-20
---

# Tasks — CHG-0049

## Wave 1 — Application RED

- [x] **T-1 — Add failing create-definition assertions for `rules`
  pass-through.** Owner: agent. Deps: none. Outcome:
  `CreateMissionDefinitionHandler` unit test expects `rules` to reach the
  definition. Verify: `go test ./tests/unit/missions/application` fails for
  the right reason. Next: T-2.

## Wave 2 — Domain RED/GREEN/BLUE

- [x] **T-2 — Add failing definition/offer `Rules` tests.** Owner: agent.
  Deps: T-1. Outcome: `NewMissionDefinition` stores `rules` and rejects a
  blank item; `NewMissionOfferFromDefinition` freezes a copy; markdown
  description stored verbatim. Verify: domain tests RED. Next: T-3.
- [x] **T-3 — Implement `Rules` on definition and offer.** Owner: agent.
  Deps: T-2 + evidence review. Outcome: minimum GREEN. Verify: domain tests
  pass. Next: T-4.
- [x] **T-4 — BLUE domain while green.** Owner: agent. Deps: T-3. Outcome:
  naming/size cleanup, files <150 lines. Next: T-5.

## Wave 3 — Application GREEN/BLUE + core gate

- [x] **T-5 — Green application tests via domain changes; pass `rules`
  through the use case.** Verify: unit suite passes →
  LAYER-GATE-APPLICATION. Next: boundary.

## Wave 4 — Boundary RED/GREEN/BLUE

- [x] **T-6 — Add failing HTTP tests.** Outcome: `POST /definitions`
  accepts `rules` and echoes `Rules`; `GET /offers`/`GET /offers/{id}`
  include `Rules` and verbatim markdown `Description`; blank `rules` item
  → `422 INVALID_MISSION_DEFINITION`. Verify: interface/http tests RED.
  Next: T-7.
- [x] **T-7 — Implement request/projection changes.** Verify: interface
  and `tests/integration/http` suites green. Next: T-8.

## Wave 5 — Infrastructure

- [x] **T-8 — Persist and decode `Rules` in DynamoDB with legacy
  tolerance.** Verify: `tests/integration/infrastructure/dynamodb` green,
  including a legacy item without the attribute. Next: T-9.

## Wave 6 — Seeds and documentation

- [x] **T-9 — Emit markdown `Description` + `Rules` in seeds.** Outcome:
  `seed_missions.py` and both catalogs updated; `tests/seeds` green. Next:
  T-10.
- [x] **T-10 — Update API-CONTRACT.md and API-GUIDE-FRONTEND.md.** Outcome:
  `description` documented as markdown; `rules` documented on request and
  responses. Next: T-11.

## Wave 7 — Gates and consolidation

- [x] **T-11 — Run validate script, clean-up gate, verify spec.** Verify:
  `tools/validate-sdd-change.sh` clean; suites green from clean state.
  Next: consolidate into `mission-board` spec and archive.
