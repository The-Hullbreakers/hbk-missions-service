# Parallel Tracks — CHG-0054 vs CHG-0051

Two changes share this worktree concurrently. Ownership boundaries:

## CHG-0054-localize-mission-content (agent-owned)

- Files under `specs/changes/active/CHG-0054-localize-mission-content/`
- `internal/missions/domain/language.go`, `localized_content.go`,
  `contract_terms_localized.go`, `mission_definition_locales.go`
- `internal/missions/application/localized_*.go`, `list_pending_review.go`,
  `create_mission_definition.go`, `get_contract.go`, `get_offer.go`,
  `list_definitions.go`, `list_offers.go`, `list_player_contracts.go`
- `internal/missions/infrastructure/dynamodb/localized_content_mapping.go`
  (new — `putEnglishContent`/`englishContentFromItem`), plus the
  mapped locale hunks in `definition_store.go` (`putEnglishContent`
  in `definitionItem`), `definition_store_queries.go`
  (`ContentEn` in `definitionFromItem`), `offer_store.go`
  (`putEnglishContent` in `Save`), `offer_store_query.go`
  (`ContentEn` in `mapOfferItem`) — other dirty shared helper
  files remain preserved/untouched
- `internal/missions/interface/http/mission_language_query.go`,
  `list_definitions.go`, `get_offer.go`, `get_offer_response.go`,
  `list_offers.go`, `list_offers_response.go`
- Tests: `tests/unit/missions/{domain,application}/localized_*_test.go`,
  `localized_contract_terms_test.go`,
  `mission_definition_locales_*_test.go`,
  `tests/integration/http/localized_*_test.go`,
  `tests/integration/infrastructure/dynamodb/localized_*_test.go`

## CHG-0051-placement-awarded-events (user-owned, concurrent)

- `internal/missions/application/publish_placement_facts.go`,
  `placement_manifest*.go`, `manifest_*`, adjudication files
- `internal/missions/infrastructure/dynamodb/season_store_outbox.go`
  and related store refactors
- `internal/missions/composition/server.go` — in-flight: calls
  `NewPlacementFactPublisher` with a stale 3-arg signature; the new
  5-port signature has no manifest adapter yet, so the package does
  not compile. **No cross-edits**: CHG-0054 overlay nils this call
  only inside temp diagnostic files, never in the repo.
- CHG-0051 test fixtures (adjudication, manifest, placement tests)

## Cross-edit rule

Never edit the other track's files. CHG-0051's composition break is a
pre-existing blocker for ordinary builds, not a CHG-0054 defect —
boundary evidence stays overlay-conditional until CHG-0051 lands its
own composition fix or the remaining stale HTTP handlers are repaired
in their own Gate 3 scopes.
