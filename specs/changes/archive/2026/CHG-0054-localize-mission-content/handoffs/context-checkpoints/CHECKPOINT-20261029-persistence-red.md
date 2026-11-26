---
handoff_id: HANDOFF-CHG-0054-CONTEXT-001
handoff_title: Continue CHG-0054 localization after persistence RED
change_id: CHG-0054
change_title: Localize mission content (definitions, offers, contracts, review queue)
capability_id: mission-board, mission-contract-fulfillment
capability_title: Mission board and mission contract fulfillment (canonical owners)
artifact_id: ART-CHG-0054-CONTEXT-HANDOFF
artifact_title: Context continuation handoff for CHG-0054-localize-mission-content
checkpoint_id: CHECKPOINT-20261029-persistence-red
checkpoint_title: Pause after T-0054-007 persistence RED for bilingual snapshots
context_used_percent: estimate ~70 (no host-observed metric)
current_task_id: T-0054-007
current_task_title: Persistence boundary RED for bilingual snapshots (first slice landed)
next_task_id: T-0054-008
next_task_title: Gate 3-PERSISTENCE decision, then minimal ContentEn DynamoDB M mapping
status: ready_for_next_context
---

# Context Handoff

This is an interim append-only continuation record. Read it before resuming the active change.

## Current Task

- Objective: localize mission definitions, offers, contracts, and review-queue titles under `lang=es|en` with atomic Spanish fallback and honest per-representation `ContentLanguage`, preserving legacy Spanish behavior, economics, frozen snapshots, authorization, response shapes, and backward compatibility.
- Done when: approved delta consolidated into canonical `mission-board` + `mission-contract-fulfillment` capabilities (NOT yet — canonical edits forbidden until consolidation).
- Current step: T-0054-007 persistence RED first slice COMPLETE; awaiting Gate 3-PERSISTENCE for definition/offer `ContentEn` round-trip only.
- Risk level: L3 (historical snapshots, data-mapping, critical accepted-contract economics).
- Owner/track: CHG-0054 agent-owned (see `parallel-tracks.md`); CHG-0051 is user-owned concurrent work — no cross-edits.

## Completed In This Context

- Tasks completed: T-0054-001/002/003/004/005 full Application+Domain inside-out; T-0054-006 partitions (definitions RED→provisional GREEN, GetOffer RED→provisional GREEN, ListOffers RED→partial GREEN); T-0054-007 first-slice RED.
- Files changed (production): `internal/missions/interface/http/mission_language_query.go`, `list_definitions.go`, `get_offer.go`, `get_offer_response.go`, `list_offers.go`, `list_offers_response.go` — all REAL (not overlay).
- Files changed (tests): `tests/integration/http/localized_*_test.go` (fixture + definitions + get-offer + list-offers + mixed); `tests/integration/infrastructure/dynamodb/localized_{items_fixture,definition,offer}_test.go` (58/141/110 lines).
- Active-change artifacts updated: `change.md` (status block), `tasks.md`, `verification.md` (gate rows + classified commands), `parallel-tracks.md` (new).
- Decisions finalized: `requestedMissionLanguage` parser (omitted→es, repeated must agree, else `INVALID_QUERY_PARAMETER`); `Content-Language` header only on homogeneous non-empty payloads; explicit `offerResponse`/`offersResponse` DTOs (fixes `LocalizedOffer` inheriting `MissionOffer.MarshalJSON` which drops `ContentLanguage`); persistence contract = `ContentEn` DynamoDB `M` attribute (Title/Description S, Rules L-of-S, Objectives L-of-M keyed ObjectiveID) — no JSON blob, no migration/backfill yet.

## Evidence

- Commands run and results (frozen — do not rerun to derive status):
  - `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListOffers'` unmodified → BUILD BLOCKER (stale handlers + CHG-0051 `NewPlacementFactPublisher` arity) — not RED; see `verification.md` for the recorded command at each partition.
  - `go test -overlay <tmp>/overlay.json ... -run '^TestLangDefinitions'` → 8/8 PASS (provisional).
  - `... -run '^TestLangGetOffer'` → 7/7 PASS (provisional).
  - `... -run '^TestLangListOffers'` → 7/7 non-mixed PASS; `MixedActualLanguages` RED: `"Hunt"/"en"/""` expected, `"Caceria"/"es"/"es"` actual — infra cause only (persistence drops `ContentEn`).
  - `go test -count=1 -tags=integration localized_items_fixture_test.go localized_definition_test.go localized_offer_test.go endpoint_test.go` (cwd `tests/integration/infrastructure/dynamodb`, real DynamoDB Local) → 3 FAIL (bilingual draft, published version, offer snapshot — `ContentEn` never written as M attribute) + 2 PASS (legacy es fallbacks).
- RED/GREEN/BLUE status: T-0054-007 first slice = RED (intended), awaiting Gate 3.
- Gates passed/still required: LAYER-GATE-DOMAIN + LAYER-GATE-APPLICATION passed (2026-10-29); Gate 3-BOUNDARY scoped approvals for definitions/GetOffer/ListOffers-HTTP-subset (conditional); Gate 3-PERSISTENCE PENDING; Gate 3-BOUNDARY for remaining partitions pending; L3 gates (mutation, critical E2E, coverage, security, cleanup, consolidation) pending.
- Overlay mechanics (reusable recipe): temp dir outside repo; JSON `{"Replace":{abs-src:abs-dst}}`; mechanical `app.Handle(..., domain.LanguageSpanish)` adapters for remaining stale files `list_contracts.go`, `contract_detail.go` (+`toContractDetail(&contract.MissionContract)`), `review_queue.go` (+`toReviewQueueEntry(contract.MissionContract)`), `create_mission_definition.go` (→`application.CreateMissionDefinitionRequest{...}`), and CHG-0051 `server.go` `outbox = application.NewPlacementFactPublisher(...)` → `outbox = nil` (dead under `DomainEvents: nil`). NEVER overlay real localized files or tests.

## Continuation

- Exact first action: ask user for scoped Gate 3-PERSISTENCE approval — definition/offer `ContentEn` round-trip via DynamoDB `M` attribute only.
- Next task and dependencies: T-0054-008 — implement minimal mapping in `definition_store.go` (`definitionItem`), `definition_store_queries.go` (`definitionFromItem`), `offer_store.go`, `offer_store_query.go`, `attribute_helpers.go`; prove the 3 RED tests GREEN + the HTTP `MixedActualLanguages` test GREEN under the same overlay; keep contracts, backfill, migration pending.
- Files/modules owned: everything listed under CHG-0054 in `parallel-tracks.md`.
- Files/modules that must not be touched: CHG-0051 files (see `parallel-tracks.md`), canonical specs/index, remaining stale HTTP handlers (`list_contracts.go`, `contract_detail.go`, `review_queue.go`, `create_mission_definition.go`) until their own Gate 3s, `MissionOffer.MarshalJSON`, `newOfferStoreForList` helper (deletes tables), any shared/prod DynamoDB endpoint.
- User decision requested: approve scoped Gate 3-PERSISTENCE (definition/offer `ContentEn` M-attribute round-trip only) — yes/no.
- Minimum verification commands for next context (frozen data, not yet run):
  - `cd tests/integration/infrastructure/dynamodb && go test -count=1 -tags=integration localized_items_fixture_test.go localized_definition_test.go localized_offer_test.go endpoint_test.go` → expect 5/5 PASS after GREEN.
  - `go test -overlay <regenerated-overlay> -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListOffers'` → expect `MixedActualLanguages` PASS after persistence GREEN.
  - `gofmt -l` on touched files → empty; all maintained files <150 lines.
- Runtime: DynamoDB Local may still run at `http://localhost:18000` — reuse and leave running; unique `hbk-missions-*` tables per test; NO DeleteTable; no external network pulls.
