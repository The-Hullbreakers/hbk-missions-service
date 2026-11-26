---
change_id: CHG-0054
artifact: tasks
status: proposed
updated: 2026-11-01
---

# Tasks — CHG-0054

> Inside-out: Application RED first -> demanded Domain -> Application
> GREEN/BLUE -> Boundary RED -> adapters/interface/composition.
> Sequential execution, one agent. Gate 2 is required before any RED;
> every scoped Gate 3 is required before its mapped production.
> Expected-RED classification and mapped production files are planned;
> each Gate 3 confirms them.

## Wave 1 — Application RED

- [x] **T-0054-001 — Write Application RED for localized read use cases.**
  Owner: agent. Deps: Gate 2 (approved 2026-10-27, scoped to this
  task's first partition). Outcome: focused tests fail for the
  intended missing shape — `domain.Language`/`ParseLanguage`, effective
  language resolution, `content_language` on read results, and `lang`
  on the list/get/review-queue requests — with hand-written reader-port
  doubles serving bilingual and legacy es-only records (TEST-APP-LANG-001:
  `list_definitions`, `list_offers`, `get_offer`,
  `list_player_contracts`, `get_contract`, `list_pending_review` under
  `es`/`en`/omitted/fallback; TEST-APP-LANG-003: heterogeneous page
  reports per-item languages; `q` matches effective language).
  Tests: `tests/unit/missions/application/localized_content_test.go` +
  `localized_content_doubles_test.go`. Command:
  `go test -count=1 ./tests/unit/missions/application -run 'TestLocalized|Test.*Lang'`.
  Expected RED: build fail on undefined domain/application symbols.
  Maps to `internal/missions/application/list_definitions.go`,
  `list_offers.go`, `get_offer.go`, `list_player_contracts.go`,
  `get_contract.go`, `list_pending_review.go` + demanded domain files.
  RED evidence 2026-10-27 (GetOffer partition, first slice): design
  decision recorded — `Handle(offerID string, language domain.Language)`
  returns a projection exposing `Title`/`Description`/`ContentLanguage`;
  `go test -count=1 localized_content_test.go` (named-file, run from
  `tests/unit/missions/application`) BUILD FAIL — intended missing
  shape only: `domain.LanguageEnglish`/`domain.LanguageSpanish`,
  `domain.LocalizedContent`, `MissionOffer.ContentEn`,
  `Handle` arity, result `ContentLanguage`. Named-file command used
  because the package build is blocked by pre-existing CHG-0051 T-9/T-10
  RED (`application.PlacementManifest`, `domain.AdjudicationAttempt`,
  `application.AdjudicationSnapshot` in `manifest_publisher_doubles_test.go`
  / `season_finalizer_doubles_test.go`) — unrelated, untouched. Note:
  `get_offer_test.go` calls the old one-arg `Handle` and will need
  signature alignment during GREEN. Remaining partitions
  (list_definitions, list_offers, list_player_contracts, get_contract,
  list_pending_review, heterogeneous/search) still pending inside this
  task. Next: expand partitions as needed, then Gate 3-APPLICATION.
  Gate 2 extension approved 2026-10-27 ('si') covering remaining
  T-0054-001 read partitions + T-0054-002 authoring RED, tests only.
  Second slice executed — whole-offer language integrity for GetOffer:
  an offer may be labeled `en` only when Title/Description/Rules and
  every objective Name/Description have English content; incomplete
  locale falls back atomically to es, never mixed. Added
  `TestGetOfferLocalized_ServesEnglishRulesAndObjectives` and
  `TestGetOfferLocalized_FallsBackToSpanishWhenEnglishIncomplete` to
  `localized_content_test.go` (now 131 lines, `<150`). Demanded domain
  fixture (RED contract proposal only, no domain implementation):
  `LocalizedContent` gains `Rules []string` and
  `Objectives []LocalizedObjective`; `LocalizedObjective{ObjectiveID,
  Name, Description}` is a pure localization record keyed by stable
  ObjectiveID — economics stay on `domain.Objective` only. Actual RED
  2026-10-27 (named file): BUILD FAIL — `unknown field Rules/Objectives
  in struct literal of type domain.LocalizedContent`, `undefined:
  domain.LocalizedObjective` at localized_content_test.go:78-79;
  intended missing shape only, zero production edits. Note: runtime
  RED for the atomic fallback will emerge once the demanded domain
  shape exists (current `LocalizeOffer` labels en on title+description
  alone). Next: Gate 3-APPLICATION re-review of this partition, then
  demanded Domain RED for `LocalizedObjective`/rules translations.
  Post-domain-GREEN observation 2026-10-27 (actual, not inferred):
  `go test -count=1 -v localized_content_test.go` FAIL — runtime RED
  on the two integrity tests only; 2 legacy tests still PASS.
  New file `localized_content_integrity_test.go` (145 lines, `<150`)
  adds the defect matrix: rules count mismatch, blank translated rule,
  blank objective en description, unmatched objective ID — each
  expects atomic es fallback — plus `MatchesEnglishObjectivesByID`
  (reversed translation order renders in base objective order,
  economics preserved). `go test -count=1 -v localized_content_test.go
  localized_content_integrity_test.go`: all 4 defect subtests fail
  (mixed en title/desc + en label served despite defect); the
  ID-matching test fails on Spanish objective text only — order,
  IDs, economics already correct. Correctly classified RED: every
  failure traces to `LocalizeOffer` missing the completeness check
  and rules/objective substitution. Zero production edits; awaits
  Gate 3-APPLICATION approval before `LocalizeOffer` extension.
  ListOffers partition RED 2026-10-27 (Gate 2 extended scope): new
  `localized_list_offers_test.go` (128 lines, `<150`). Design fixed by
  lead: `Handle(playerID string, filter domain.OfferListFilter,
  language domain.Language) (application.LocalizedOfferListPage,
  error)` with `LocalizedOfferListPage{Offers []LocalizedOffer,
  Total, Limit, Offset}` mirroring the existing wire envelope.
  Covers per-item honest labels in a mixed en page, and effective-
  language `q` (en term hits bilingual; es term hits legacy fallback;
  es base term does NOT hit when en served) plus `reader.gotFilter.Q
  == ""` proving the application never forwards text search to the
  store's Spanish prefilter — `localizedOfferListReader` applies
  `filter.Match` to model the real reader. Actual RED (named file,
  `go test -count=1 localized_list_offers_test.go
  localized_list_offers_page_test.go
  list_offers_doubles_test.go`): BUILD FAIL — `Handle` 3-arg arity
  (lines 71, 121 + page test :28) + `page.Offers[i].ContentLanguage`
  undefined on `domain.MissionOffer` (lines 88, 91 + page test :37);
  intended missing shape only. Reworked per review: per-item test now
  asserts en Rules + objective Name/Description + stable ObjectiveID +
  MeritPerUnit/Bonus/UnitLimit (136 lines); new
  `localized_list_offers_page_test.go` (38 lines) proves `q` filters
  BEFORE pagination — 3 offers where the first carries the term only
  in its Spanish base (catches a Spanish prefilter), Page{Limit:1,
  Offset:1} expects Total=2 and the second matching item, with
  `gotFilter.Q == ""`. Zero production edits. Next:
  Gate 3-APPLICATION for the list partition.
  ListDefinitions partition RED 2026-10-27 (Gate 2 extended scope):
  new `localized_list_definitions_test.go` (149 lines, `<150`),
  reusing `fakeDefinitionListReader` via named-file inclusion. Design
  settled by lead: `Handle(language domain.Language)
  ([]application.LocalizedDefinition, error)` with
  `LocalizedDefinition{domain.MissionDefinition; ContentLanguage}`.
  Covers mixed en page (complete bilingual + legacy es-only published
  literal rows — literals used since publish may require en later),
  atomic es fallback on missing en rules/objectives (asserting
  objective Description on both paths after review), explicit es
  request on bilingual, and reader-error propagation. Actual RED
  (named file): BUILD FAIL — `Handle` arity have (domain.Language)
  want () at 4 sites (lines 60, 104, 128, 145) +
  `defs[i].ContentLanguage` undefined on `domain.MissionDefinition`
  at 4 sites (lines 77, 81, 114, 134); intended missing shape only.
  Zero production edits. Next: Gate 3-APPLICATION for the
  definitions partition.
  ListPlayerContracts partition RED 2026-10-27 (Gate 2 extended
  scope): new `localized_list_player_contracts_test.go` (130 lines,
  `<150`), reusing `playerContractsReaderStub` via named-file
  inclusion. Design settled by lead: `Handle(playerID, filter,
  language domain.Language) ([]application.LocalizedContract, error)`
  with `LocalizedContract{domain.MissionContract; ContentLanguage}` —
  one contract-level label covering embedded Offer + Terms; English
  only when the offer snapshot fully covers the effective
  Terms/Offer objectives. Covers mixed list (bilingual + legacy +
  outsider hidden, reader order, ids/economics), legacy persisted
  shape (Terms nonzero, Offer.Objectives nil — Terms.Objectives()
  projected en on both views, economics identical), and incomplete
  en → whole-contract es fallback incl. Terms text. Actual RED
  (named file): BUILD FAIL — `Handle` 3-arg arity at 3 sites (lines
  51, 83, 119) + `contracts[i].ContentLanguage` undefined on
  `domain.MissionContract` at 4 sites (61, 64, 100, 129); intended
  missing shape only. Expected Domain demand after Gate 3-APPLICATION:
  a Terms-level localized projection (Terms carries no locale copy of
  its own — en text resolves from the offer snapshot by ObjectiveID).
  Zero production edits. Next: Gate 3-APPLICATION for the contracts
  partition.
  GetContract partition RED 2026-10-27 (Gate 2 extended scope): new
  `localized_get_contract_test.go` (149 lines, `<150`), self-contained
  `localizedContractGetter` double + fresh fixtures. Design settled:
  `Handle(identity, contractID, language domain.Language)
  (*application.LocalizedContract, error)` — shared `LocalizeContract`
  invoked AFTER unchanged authorization (reviewer or active roster
  member). Covers member en detail (Terms nonzero + Offer.Objectives
  nil legacy shape — offer, Terms.Objectives() and Phases() all en,
  economics identical), es-only under en, outsider denied under en,
  reviewer permitted + localized, reader-error propagation. Actual
  RED (named file): BUILD FAIL — `Handle` 3-arg arity at 5 sites
  (lines 57, 91, 109, 125, 143) + `served.ContentLanguage` undefined
  on `*domain.MissionContract` at 3 sites (78, 98, 132); intended
  missing shape only. Zero production edits. Next:
  Gate 3-APPLICATION for the get-contract partition.
  ListPendingReview partition RED 2026-10-27 (Gate 2 extended scope):
  new `localized_list_pending_review_test.go` (123 lines, `<150`),
  self-contained `localizedPendingLister` double + fresh fixtures —
  both built via `NewMissionContractFromState(...,
  ContractPendingValidation, true, ...)` to model real queue
  semantics. Design settled: `Handle(identity, language
  domain.Language) ([]application.LocalizedContract, error)` reusing
  the green `LocalizeContract` — per-entry contract-level label.
  Covers mixed bilingual/legacy queue under en (en/es titles, honest
  labels, ContractID/OfferID/State/roster/order asserted unchanged),
  non-reviewer denied BEFORE the store call (`lister.calls == 0`),
  empty queue, lister error propagation. Actual RED (named file):
  BUILD FAIL — `Handle` 2-arg arity have (CallerIdentity, Language)
  want (CallerIdentity) at 4 sites (lines 56, 85, 101, 118) +
  `entries[i].ContentLanguage` undefined on `domain.MissionContract`
  at 2 sites (67, 73); intended missing shape only. Zero production
  edits. Next: Gate 3-APPLICATION for the review-queue partition.
  Gate 3-APPLICATION eighth scope approved 2026-10-27 ('si' —
  ListPendingReview GREEN only). GREEN/BLUE landed
  `list_pending_review.go` (39 lines): `Handle(identity, language)`
  keeps `isReviewer` rejection BEFORE any lister call, propagates
  lister errors unchanged, allocates a non-nil empty slice and
  projects each item through the shared `LocalizeContract`,
  preserving queue order; `PendingReviewLister` port unchanged.
  Legacy `list_pending_review_test.go` 3 call sites aligned with
  `domain.LanguageSpanish`, assertions unchanged. Focused named-file
  GREEN: all 4 localized tests PASS; full application package PASS;
  `go build ./internal/missions/application` OK. `interface/http`
  gains a sixth expected blocker (review-queue handler still calls
  the old signature); boundary untouched. BLUE: no duplicated
  localization rule — handler is a thin projection loop over the
  shared `LocalizeContract`; <150 and gofmt clean. Layer gate still
  pending: T-0054-002 bilingual authoring RED/GREEN remains.
- [x] **T-0054-002 — Write Application RED for bilingual authoring.**
  Owner: agent. Deps: Gate 2. Outcome: tests fail proving
  `CreateMissionDefinitionHandler` accepts `es`-only (backward
  compatible) and `es`+`en` atomically, and publish forwards to the
  domain completeness rule — an incomplete `en` surfaces
  `ErrDefinitionInvalid` (TEST-APP-LANG-002). Tests:
  `tests/unit/missions/application/localized_authoring_test.go` with
  port doubles. Command: `go test -count=1 ./tests/unit/missions/application -run 'Test.*Localiz|Test.*Bilingual'`.
  Expected RED: build fail on missing request/domain symbols. Maps to
  `internal/missions/application/create_mission_definition.go`,
  `publish_mission_definition.go` + demanded domain files.
  Next: Gate 3-APPLICATION -> T-0054-003.
  RED executed 2026-10-27 as TWO split files (lead-fixed API):
  `localized_authoring_create_test.go` (109 lines) — compile RED:
  `application.CreateMissionDefinitionRequest` undefined (lines
  26-27) + `Handle` arity have (unknown type) want 8-positional at
  lines 57, 79, 104. Covers es-only draft creation, bilingual
  deep-copy freeze (mutating request post-Act leaves saved snapshot
  intact), invalid es baseline with `Save` never called.
  `localized_authoring_publish_test.go` (120 lines; helper renamed
  `publishableEnglish` after review caught a package-level collision
  with `localized_content_integrity_test.go`; objective fixtures
  carry `Category: "BOUNTY_HUNTING"`) — runtime RED:
  `RejectsMissingOrPartialEnglish` 4 subtests FAIL (es-only, missing
  en rules, missing en objective, blank en title each publish
  today — `saves == 1`, expected `ErrDefinitionInvalid` + `saves
  == 0`); `PublishesBilingualDraft` already PASSes (current code
  publishes any draft). `PublishMissionDefinitionHandler.Handle`
  signature unchanged `(definitionCode, version)`. Demanded Domain
  RED expected: structural English completeness invariant on
  `MissionDefinition.Publish` (title, markdown desc, same rule
  count/order nonblank, each es objective ID exactly once with en
  Name, en Description iff es present). Published es-only
  historical versions stay readable — no migration. Package-wide
  diagnostic `go test -count=1 ./tests/unit/missions/application`
  after the collision fix: build FAIL contains ONLY the 5 intended
  create-file errors (lines 26-27, 57, 79, 104) — no redeclaration,
  no unrelated failures. Next: Gate 3-APPLICATION for the demanded
  Domain RED scope only; create and publish may need separate
  Gate 3s per inside-out.
  Gate 3-APPLICATION demanded Domain RED approved 2026-10-27
  ('si' — publish completeness invariant only). Domain RED landed
  as two standalone files: `mission_definition_locales_publish_test.go`
  (109 lines) — `RequiresCompleteEnglish` 5-subtest rejection table
  (no en / blank title / blank desc / en rule count / blank rule)
  asserting `ErrDefinitionInvalid` + State stays DRAFT + Version ""
  and `PublishesCompleteEnglish` (reversed en objective order
  succeeds, IDs/economics/Spanish text/ContentEn order preserved);
  `mission_definition_locales_edges_test.go` (126 lines) —
  `RejectsObjectiveMismatches` 5-subtest table (missing/duplicate/
  unknown ID, blank name, blank required description) +
  `AllowsOptionalEnglishDescription` + `AllowsEmptyContentShape` +
  `HistoricalSpanishOnlyStaysPublished` (literal PUBLISHED record
  cannot republish — no migration). Helpers `authoringDomain*`
  unique to the package. Actual RED: RUNTIME FAIL — all 10
  rejection subtests publish today (actual PUBLISHED/v1, want
  DRAFT/""); `PublishesCompleteEnglish`, `AllowsOptional*`,
  `AllowsEmpty*`, `Historical*` all PASS — intended missing
  completeness rule only, zero production edits. Next:
  Gate 3-DOMAIN (fourth scope) for the `Publish` English
  completeness invariant GREEN/BLUE.
  Gate 3-DOMAIN fourth scope approved 2026-10-27 ('approved' —
  new-version Publish English completeness only). GREEN/BLUE
  landed: new `mission_definition_locales.go` (46 lines) —
  `englishSnapshotComplete()` on `MissionDefinition` checks
  trimmed nonblank en Title/Description, en Rules count == es
  with no blank entries, en Objectives count == es with exactly
  one translation per Spanish ObjectiveID (dup/unknown rejected),
  nonblank en Name, en Description required iff es Description
  nonempty; `Publish` calls it after the existing draft+nonempty-
  version checks and BEFORE any state/version mutation, returning
  `ErrDefinitionInvalid` with state preserved. No content mutation,
  no migration — historical PUBLISHED es-only records never
  revalidated. Fixture alignment (en snapshot added before
  `Publish`, assertions unchanged): mission_definition_test.go
  ×2, active_template_test.go ×3, mission_rules_test.go,
  contract_approve_test.go, contract_approve_verdict_test.go,
  mission_offer_definition_test.go (publishedOfferDefinition),
  application publish_mission_fixtures_test.go +
  contract_lifecycle_test.go. Results: all 14 locale-publish
  tests PASS; full `./tests/unit/missions/domain` PASS;
  `go build ./internal/missions/domain` OK; gofmt clean,
  mission_definition.go 100 lines. Bonus GREEN: Application
  `localized_authoring_publish_test.go` named-file now PASSES
  independently — the handler propagated the domain error
  already; create RED (`CreateMissionDefinitionRequest`) still
  blocks the full application package build as intended.
  Gate 3-APPLICATION authoring-create approved 2026-10-29.
  GREEN/BLUE landed `create_mission_definition.go` (50 lines):
  `CreateMissionDefinitionRequest` with optional `ContentEn
  *domain.LocalizedContent`; `Handle(request)` builds the Spanish
  baseline via `NewMissionDefinition`, freezes `ContentEn.Clone()`
  only when non-nil, then `Save` once — writer never called on
  invalid baseline. Legacy `create_mission_definition_test.go` 3
  positional calls aligned to request literals, assertions
  unchanged; deep-copy test extended to objective-name mutation.
  `aDraftDefinition` (publish_mission_definition_test.go) aligned
  with complete en — its failure was masked while create RED
  blocked the package build. Results: create 3/3 + publish 2/2
  named-file PASS; full `./tests/unit/missions/application` PASS;
  `go build ./internal/missions/application` OK; gofmt clean,
  all files <150. Next: LAYER-GATE-APPLICATION assessment
  (T-0054-005) — all Application partitions green; then
  Boundary RED for `lang` parsing/Content-Language/response
  projection + repair of the six stale HTTP call sites.

## Wave 2 — Demanded Domain

- [x] **T-0054-003 — Write Domain RED for language, locales, and completeness.**
  Owner: agent. Deps: Gate 3-APPLICATION on T-0054-001/002 RED
  (GetOffer partition approved 2026-10-27). Outcome: tests fail for
  `Language`/`ParseLanguage` (invalid/empty/conflicting),
  effective-language resolution with `es` fallback and honest reported
  locale, per-version locale completeness at publish, stable objective
  ids/rule order across locales, and `q` matching the effective
  language (TEST-DOM-LANG-001/002/003). Tests:
  `tests/unit/missions/domain/language_test.go`,
  `localized_content_test.go`, `mission_definition_locales_test.go`.
  Command: `go test -count=1 ./tests/unit/missions/domain -run 'TestLanguage|TestLocalized|Test.*Locale'`.
  Expected RED: build fail on undefined symbols. Maps to
  `internal/missions/domain/language.go` (new, candidate),
  `mission_definition.go`, `mission_offer.go`, `objective.go`,
  `offer_list.go`, `contract_terms.go`, `errors.go`.
  RED evidence 2026-10-27 (GetOffer-demanded partition only, per
  scoped Gate 3-APPLICATION): `go test -count=1 language_test.go`
  (named-file, cwd `tests/unit/missions/domain`) BUILD FAIL —
  intended missing shape only: `domain.Language`,
  `domain.LanguageSpanish`/`LanguageEnglish`, `domain.ParseLanguage`,
  `domain.ErrLanguageInvalid`, `domain.LocalizedContent`,
  `MissionDefinition.ContentEn`, `MissionOffer.ContentEn`. Covers
  ParseLanguage table (es/en accepted, `''`/`'fr'` rejected) and the
  bilingual-definition→offer English freeze retaining Spanish scalars;
  mapped to `domain/language.go`, `domain/localized_content.go`,
  `domain/mission_offer.go`, `domain/mission_definition.go`.
  Named-file command used because the package build is blocked by
  pre-existing CHG-0051 T-10 RED (`domain.NewAdjudicationAttempt`,
  `domain.ErrAdjudicationAttemptInvalid`,
  `domain.MembersOfCreditedContracts` in `adjudication_attempt_test.go`
  / `credited_roster_test.go`) — unrelated, untouched; that baseline is
  historical, CHG-0051 domain has since advanced and is not re-claimed
  as current. Remaining partitions (publish completeness, `q`
  effective-language matching, resolution on contract/queue
  projections) still pending inside this task. Gate 3-DOMAIN approved
  2026-10-27 scoped to the GetOffer partition -> T-0054-004 partial
  GREEN.
  Gate 3-APPLICATION second scope approved 2026-10-27 ('si') for
  whole-offer integrity GetOffer — demanded Domain RED only.
  New file `tests/unit/missions/domain/localized_offer_terms_test.go`
  (87 lines, `<150`), two tests: freeze of English
  Rules/Objectives snapshot retaining Spanish scalars + economics +
  stable ObjectiveID, and defensive non-alias (source-slice mutation
  does not rewrite the frozen offer). Actual RED 2026-10-27 —
  `go test -count=1 localized_offer_terms_test.go` BUILD FAIL,
  intended missing shape only: `unknown field Rules/Objectives in
  domain.LocalizedContent`, `undefined: domain.LocalizedObjective`
  (lines 26-27, 40-44, 69-70). Zero production edits. Next:
  Gate 3-DOMAIN re-review for this demanded shape.
  Contract-terms partition RED 2026-10-27 (Gate 3-APPLICATION
  ListPlayerContracts approved 'si' — demanded Domain RED only): new
  `localized_contract_terms_test.go` (123 lines, `<150`) demanding
  `func (t ContractTerms) WithObjectiveText(translations
  []LocalizedObjective) (ContractTerms, error)` — an independent
  translated read snapshot over BOTH `Objectives()` and `Phases()`,
  matched by ObjectiveID (reversed input order proves it), preserving
  ids/economics/mandatory-vs-optional/phase order, source unchanged,
  input not aliased. Rejection table: missing/unknown/duplicate ID,
  blank Name, blank Description when source has one →
  `ErrContractTermsInvalid`; blank Description allowed when source
  blank; zero legacy terms rejected. Actual RED (named file): BUILD
  FAIL — `terms.WithObjectiveText` undefined at 3 Act sites (lines
  34, 99, 119); intended missing method only, zero production edits.
  Next: Gate 3-DOMAIN review for this demanded method, then the
  ListPlayerContracts Application GREEN under its own approval.
- [x] **T-0054-004 — Implement Domain GREEN/BLUE and pass LAYER-GATE-DOMAIN.**
  Owner: agent. Deps: T-0054-003 RED evidence + Gate 3-DOMAIN.
  Outcome: minimal domain implementation green; BLUE keeps value-object
  discipline and `< 150` physical lines per file; domain package tests
  green standalone. Command: `go test -count=1 ./tests/unit/missions/domain`.
  Second-scope GREEN/BLUE 2026-10-27 (Gate 3-DOMAIN approved 'si'):
  `localized_content.go` gained `Rules []string`,
  `Objectives []LocalizedObjective`, `LocalizedObjective{ObjectiveID,
  Name, Description}`, and `LocalizedContent.Clone()` deep-copying both
  slices (nil stays nil); `NewMissionOfferFromDefinition` now freezes
  `definition.ContentEn.Clone()` — no publish validation or read-path
  behavior added. `json:"-"` on `ContentEn` preserved on both structs.
  Evidence: `go test -count=1 -v localized_offer_terms_test.go
  language_test.go` PASS (7 tests); `go test -count=1
  ./tests/unit/missions/domain` PASS (package green incl. parallel
  CHG-0051 work); `go build ./internal/missions/domain` OK; gofmt clean.
  BLUE inventory (lines): `localized_content.go` 28,
  `mission_offer.go` 111, `localized_offer_terms_test.go` 87 — `<150`.
  Application note: `localized_content_test.go` now compiles and its
  whole-offer integrity assertions are RUNTIME RED (current
  `LocalizeOffer` labels en on title+description only) — intentional,
  awaiting separate Application GREEN approval. Whole
  LAYER-GATE-DOMAIN still not claimed — remaining partitions pending.
  Third-scope GREEN/BLUE 2026-10-27 (Gate 3-DOMAIN approved 'si',
  `WithObjectiveText` only): new focused
  `contract_terms_localized.go` (76 lines) — validates nonzero terms,
  `objectiveTranslationsByID` enforces exactly one translation per
  stored objective (count + dupes), per-objective coverage and
  trimmed non-blank Name with conditional Description (required iff
  Spanish has one); returns an independent snapshot translating
  `objectives` and every `phases[i].Mandatory/Optional` by ID —
  fresh slices, receiver and input untouched, order-independent.
  First GREEN run exposed a TEST fixture bug — shared-backing `ok`
  slice literals mutated mid-table via `append(ok[:1], …)`; fixed
  literals (not production), then PASS: all named tests + full
  `go test ./tests/unit/missions/domain` green; `go build
  ./internal/missions/domain` OK; gofmt clean; both files <150.
  BLUE: single-pass count/dup validation, fresh allocations only,
  zero → error before allocation.
  Partial GREEN/BLUE 2026-10-27 (GetOffer-demanded partition only):
  `domain/language.go` (`Language`, `LanguageSpanish`/`LanguageEnglish`,
  `ParseLanguage`, `ErrLanguageInvalid` sentinel kept in-file — no
  errors.go churn), `domain/localized_content.go` (`LocalizedContent`
  Title/Description), `ContentEn` on `MissionDefinition`/`MissionOffer`
  + freeze copy in `NewMissionOfferFromDefinition`. Evidence:
  `go test -count=1 -v language_test.go` PASS 3/3; legacy named-file
  regression `go test -count=1 mission_offer_definition_test.go
  mission_offer_test.go mission_definition_test.go mission_rules_test.go
  mission_definition_terms_test.go offer_list_filter_test.go
  offer_list_page_test.go offer_list_query_test.go` PASS.
  BLUE inventory (physical lines, all `< 150`): language.go 30,
  localized_content.go 10, mission_definition.go 95, mission_offer.go
  111, language_test.go 111. Value-object decision: `Language` =
  validated VO over es/en; `LocalizedContent` = per-locale text VO;
  provider-neutral names.
  BLUE boundary-leak fix 2026-10-27: lead review caught exported
  `ContentEn` leaking into wire JSON via `MissionOffer.MarshalJSON`
  alias and default `MissionDefinition` marshal — premature HTTP
  behavior before Boundary RED. Added wire-shape regression tests
  (`TestMissionOfferMarshalJSON_HidesEnglishContent`,
  `TestMissionDefinitionMarshalJSON_HidesEnglishContent`) — assertion
  RED proven (`ContentEn` serialized on both), then minimal fix
  `json:"-"` on both `ContentEn` fields → GREEN 5/5. Implication:
  locale content is now domain-internal only until the boundary
  projects the selected language; DynamoDB locale persistence must be
  an explicit item-attribute mapping (T-0054-007/008), never JSON
  serialization. Whole LAYER-GATE-DOMAIN NOT claimed — remaining
  partitions (completeness, `q`, contract/queue resolution) pending.
  Verification: named-file focused + legacy regression PASS recorded
  in `verification.md`; full domain package suite remains blocked by
  pre-existing CHG-0051 T-10 RED — not claimed green.
  Next: per lead — request Gate 2 extension for remaining Application
  partitions or finish approved GetOffer Application after this partial
  domain gate; do NOT implement Application until lead review of diff.

## Wave 3 — Application GREEN

- [x] **T-0054-005 — Implement Application GREEN/BLUE and pass LAYER-GATE-APPLICATION.**
  Owner: agent. Deps: T-0054-004. Outcome: all six read use cases accept
  `lang` and resolve per-item effective language with
  `content_language`; authoring use cases store/validate both locales;
  application suite green standalone with hand-written doubles only.
  Command: `go test -count=1 ./tests/unit/missions/application`.
  Partial GREEN/BLUE 2026-10-27 (GetOffer partition only, scoped
  Gate 3-APPLICATION): re-confirmed RED on named file — `Handle` arity
  + `served.ContentLanguage` only. Implemented `Handle(offerID string,
  language domain.Language) (LocalizedOffer, error)` in
  `get_offer.go`; new `localized_offer.go` holds `LocalizedOffer`
  (embeds `domain.MissionOffer`, explicit `ContentLanguage`) +
  `LocalizeOffer` resolver — separate file because the projection is
  shared across offer read use cases (ListOffers partition next).
  English served only when requested `en` AND both `ContentEn` fields
  non-empty; otherwise Spanish scalars + `ContentLanguage: es`; stored
  offer copied not mutated; reader error propagated as-is.
  `get_offer_test.go` signature-aligned (`LanguageSpanish`), assertions
  unchanged. Evidence: `go test -count=1 -v localized_content_test.go
  get_offer_test.go` PASS 4/4 (cwd `tests/unit/missions/application`).
  BLUE inventory (physical lines `<150`): get_offer.go 24,
  localized_offer.go 24, localized_content_test.go 61,
  get_offer_test.go 48.
  Recorded blocker (expected, not fixed under this gate):
  `internal/missions/interface/http/get_offer.go:16` calls the old
  one-arg `Handle` — `go build ./internal/missions/interface/http`
  fails until Boundary GREEN; domain+application build clean.
  Third-scope GREEN/BLUE 2026-10-27 (Gate 3-APPLICATION approved 'si',
  whole-offer GetOffer integrity only): RED observed — 8 integrity
  tests FAIL / 4 legacy PASS on named files. `localized_offer.go`
  (91 lines) now checks atomic English completeness —
  `englishOfferComplete` (trimmed non-blank title/description,
  positional rules count + non-blank, `englishObjectivesComplete`
  requiring equal counts, unique ID match via `objectivesByID`,
  non-blank name, description required only when the Spanish
  objective carries one) — then substitutes en rules/objective text
  in base order into copied slices; es fallback returns Spanish
  untouched; both paths copy `Rules`/`Objectives` so the projection
  never aliases the stored offer. GREEN: `go test -count=1 -v
  localized_content_test.go localized_content_integrity_test.go
  localized_content_projection_test.go get_offer_test.go` PASS 12/12.
  Whole app suite `go test -count=1 ./tests/unit/missions/application`
  now PASS (measured 2026-10-27 — parallel CHG-0051 T-11 app work is
  green; the earlier "CHG-0051 blocks app suite" note is retired,
  historical evidence preserved dated). `go build
  ./internal/missions/application` OK. BLUE: no duplication
  (`objectivesByID` shared by completeness + substitution), no
  mutation, files <150 (91/145/131/120), gofmt clean.
  Fourth-scope GREEN/BLUE 2026-10-27 (Gate 3-APPLICATION approved,
  ListOffers partition): `Handle(playerID, filter, language)` returns
  new `LocalizedOfferListPage{Offers []LocalizedOffer json:"offers",
  Total, Limit, Offset}` in `localized_offer.go` (envelope keys mirror
  the domain page). `filter.Q` is stripped before `h.offers.List` so
  the store cannot Spanish-prefilter; effective-language
  `filter.Match` runs on `LocalizeOffer(offer,language).MissionOffer`
  after visibility and before `SortOffers`/`PaginateOffers`, then the
  page items are localized — Total counts effective matches, empty
  pages stay non-nil `[]`. Visibility/season/error paths unchanged.
  Existing `list_offers_test.go`/`_page_test.go`/`_visibility_test.go`
  aligned to pass `domain.LanguageSpanish` (15 call sites), zero
  assertion changes — embedded `MissionOffer` keeps field access.
  RED→GREEN: named-file `localized_list_offers{,_page}_test.go +
  list_offers_doubles_test.go` compile-RED → PASS 5 tests incl. 3
  q-subtests + q-before-pagination. `go test -count=1
  ./tests/unit/missions/application` PASS; `go build
  ./internal/missions/application` OK. BLUE: `LocalizeOffer` reused,
  no duplication, files <150 (list_offers.go 101,
  localized_offer.go 101). `interface/http` now fails at TWO expected
  sites — `get_offer.go:16` + `list_offers.go:23` — boundary blocker,
  untouched.
  Fifth-scope GREEN/BLUE 2026-10-27 (Gate 3-APPLICATION approved 'si',
  ListDefinitions partition): pre-production RED regression added —
  `localized_list_definitions_projection_test.go` (60 lines) asserts
  served projections never alias stored es/en slices on both en and
  fallback paths; captured compile RED (2 extra arity sites — alias
  behavior unobservable until the interface exists, honestly noted).
  BLUE refactor extracted the shared projection into new
  `localized_mission_text.go` (99 lines): `localizeMissionText` +
  `englishTextComplete`/`englishObjectivesComplete`/`objectivesByID`
  now serve both `LocalizeOffer` and `LocalizeDefinition` — zero
  duplicated completeness rules, copies in both locale paths.
  `localized_offer.go` slimmed to 33 lines (types + wrapper);
  `localized_definition.go` 24 lines; `list_definitions.go` 32 lines
  with `Handle(language)` per-item projection; existing
  `list_definitions_test.go` aligned (2 sites), assertions unchanged.
  GREEN: `go test -run 'TestListDefinitions' ./tests/unit/missions/
  application` PASS 8/8 incl. both alias regressions; `go test
  -count=1 ./tests/unit/missions/application` PASS (GetOffer +
  ListOffers regressions hold through the refactor); `go build
  ./internal/missions/application` OK; gofmt clean. `interface/http`
  now fails at THREE expected sites (+ `list_definitions.go:14`).
  Sixth-scope GREEN/BLUE 2026-10-27 (Gate 3-APPLICATION approved,
  ListPlayerContracts partition): pre-production RED —
  `localized_list_player_contracts_projection_test.go` (84 lines)
  adds phased-terms translation through the list read + no-alias
  probes; captured compile RED (10 intended sites). New
  `localized_contract.go` (41 lines): `LocalizedContract{domain.
  MissionContract; ContentLanguage}` + `LocalizeContract` —
  effective objectives = `Terms.Objectives()` when nonzero else
  `Offer.Objectives`; shared `localizeMissionText` resolves the
  offer; when English + nonzero terms,
  `Terms.WithObjectiveText(ContentEn.Objectives)` produces the terms
  snapshot (phases included); any unexpected invalid projection
  falls the WHOLE contract back to es atomically rather than
  mislabeling. `list_player_contracts.go` (30 lines):
  `Handle(playerID, filter, language)` — authorization
  (`IncludesPlayer`), filtering and order unchanged;
  `list_player_contracts_test.go` aligned (2 sites), assertions
  unchanged. GREEN: `-run 'TestListPlayerContracts'` PASS 7/7 incl.
  phased projection + both alias regressions; `go test -count=1
  ./tests/unit/missions/application` PASS; `go build
  ./internal/missions/application` OK; gofmt clean. `interface/http`
  now fails at FOUR expected sites (+ `list_player_contracts`).
  Seventh-scope GREEN/BLUE 2026-10-27 (Gate 3-APPLICATION approved,
  GetContract partition): `get_contract.go` (30 lines) —
  `Handle(identity, contractID, language)` keeps the exact reader
  call/error path and `isReviewer(identity) ||
  contract.CanSubmitEvidence(identity)` gate unchanged; only the
  authorized path returns `&LocalizeContract(*contract, language)`.
  `get_contract_test.go` aligned — 5 call sites pass
  `domain.LanguageSpanish`, assertions unchanged. GREEN:
  `go test -count=1 -v localized_get_contract_test.go` PASS 5/5;
  `go test -count=1 ./tests/unit/missions/application` PASS;
  `go build ./internal/missions/application` OK; gofmt clean.
  `interface/http` now fails at FIVE expected sites (+
  `contract_detail.go:21`, `list_contracts.go:25`). BLUE:
  authorization strictly precedes projection; nil/error path
  untouched; `LocalizedContract` reused — zero new projection logic.
  Whole LAYER-GATE-APPLICATION NOT claimed — remaining read
  partitions + authoring pending.
  Verification: named-file + package PASS recorded in `verification.md`.
  Next: T-0054-006/T-0054-007 boundary RED (or Gate 2 extension for
  remaining Application partitions, per lead).

## Wave 4 — Boundary RED

Next step 2026-10-29: LAYER-GATE-APPLICATION passed — request Gate 2
approval for Boundary RED (T-0054-006/007) next; no Boundary
production is authorized by this closure.

- [x] **T-0054-006 — Write HTTP boundary RED for the language contract.**
  Owner: agent. Deps: LAYER-GATE-APPLICATION. Outcome: integration/http
  tests fail proving `?lang=` on the six endpoints (omitted→`es`,
  `en`→English, invalid/empty/conflicting-repeat→`400
  INVALID_QUERY_PARAMETER`), served-language label on every
  representation keyed by response-family casing (`ContentLanguage` on
  PascalCase projections and contract-list items, `content_language`
  on contract detail and review-queue entries), `Content-Language`
  only on homogeneous non-empty
  payloads and absent on heterogeneous/empty collections,
  contract detail keeps lowercase shape + member/reviewer projections,
  review-queue titles localize with role gates intact, facets
  unaffected (TEST-HTTP-LANG-001/002/003). Tests:
  `tests/integration/http/localized_content_test.go` (composition
  through the real server + local DynamoDB). Command:
  `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run 'TestLang|TestLocaliz'`.
  Expected RED: assertion failures on missing `lang` support /
  `content_language` / header. Maps to
  `internal/missions/interface/http/{list_definitions,list_offers,
  get_offer,list_contracts,contract_detail*,review_queue,
  create_mission_definition,errors}.go`.
  Next: Gate 3-BOUNDARY -> T-0054-008.
  Gate 2 approved 2026-10-29 ('si' — Boundary RED only, first
  partition: GET /api/v1/missions/definitions). Landed
  `localized_content_fixture_test.go` (59 lines, `//go:build
  integration`) — `langAdminSession` + `languageDefinitionsServer`
  booting real `composition.NewServer` on a unique DynamoDB table
  against local DynamoDB `http://localhost:18000` (or
  `DYNAMODB_TEST_ENDPOINT`); `localized_content_test.go` (119
  lines) — `TestLangDefinitionsRejectsInvalidQuery` 5-case table
  (fr / empty / conflicting both orders → 400 +
  `INVALID_QUERY_PARAMETER` + no Content-Language; `es&lang=es` →
  200 es item) + `TestLangDefinitionsLegacyFallbackAndEmpty`
  (es-only under en → `Rescate` + `ContentLanguage:"es"` +
  `Content-Language: es` + no exposed ContentEn; empty catalog →
  200 `[]` + header absent). Unmodified run is a BLOCKER not RED:
  `go test -count=1 -tags=integration -p=1
  ./tests/integration/http -run '^TestLangDefinitions'` → build
  FAIL at the 9 known stale call sites in 7 http files
  (contract_detail:21+35, create:67, get_offer:16,
  list_contracts:25, list_definitions:14, list_offers:23,
  review_queue:53+67) — pre-existing, not intended RED.
  Diagnostic TEMP `-overlay` (mechanical LanguageSpanish adapters
  on the 7 files + nil-outbox for CHG-0051's in-flight
  `NewPlacementFactPublisher` signature; nothing written to repo):
  SIX intended FAIL results — `unsupported_language`,
  `empty_language`, `conflicting_es_then_en`,
  `conflicting_en_then_es` return 200 instead of 400 (no lang
  parsing exists) and `spanish_only_fallback_under_en` +
  `omitted_lang_ignores_accept_language` serve the correct es item
  but lack `Content-Language: es`; TWO PASS —
  `repeated_es_accepted` and `empty_catalog_under_en` already
  correct. Counts re-verified 2026-10-29 after the seeded
  fallback was widened to a two-case table (`?lang=en` + omitted
  query with `Accept-Language: en`) and the `ContentEn`-absence
  assertion strengthened to `[]map[string]json.RawMessage` +
  `NotContains`. Overlay output is conditional diagnostic
  evidence only. Next: Gate 3-BOUNDARY for the definitions
  partition lang parsing + header production.
  Gate 3-BOUNDARY (definitions GET only, conditional) approved
  2026-10-29. Provisional GREEN/BLUE landed REAL production:
  new `mission_language_query.go` (26 lines) —
  `requestedMissionLanguage` (omitted→es, ParseLanguage, repeated
  values must agree else ErrLanguageInvalid); `list_definitions.go`
  (43 lines) — 400 `INVALID_QUERY_PARAMETER` before the app call,
  `Handle(language)`, `Content-Language` only when the non-empty
  list is homogeneous, unchanged internal-error path. Diagnostic
  overlay regenerated WITHOUT list_definitions.go (6 remaining
  stale files + CHG-0051 nil-outbox). Result: 8/8 PASS —
  PROVISIONAL, conditional on overlay; unmodified http build still
  blocked by 5 stale sites. Task NOT complete: remaining
  partitions (offers, contracts, review-queue, get_offer,
  create/publish) keep T-0054-006 open; next step is their own
  Gate 3-BOUNDARY partitions or Gate 2 extension as scoped.
  GetOffer partition RED landed 2026-10-29 under Gate 2
  extension (RED/tests only): fixture refactored —
  `languageHTTPServer(t, seed)` generic boot (79-line fixture
  file) with `languageDefinitionsServer` wrapper preserving the
  definitions tests (still PASS under overlay) plus
  `languageOfferServer` seeding one es-only offer via real
  `OfferStore.Save`; new `localized_get_offer_test.go` (118
  lines, integration tag) — `RejectsInvalidQueryBeforeLookup`
  (missing offer id + fr/empty/conflicting → 400 demanded
  BEFORE lookup), `ServesLegacySpanish` (en / omitted +
  `Accept-Language: en` / repeated en → `Rescate`,
  `ContentLanguage:"es"`, `Content-Language: es`, `ContentEn`
  key absent, Category preserved), `UnknownRemains404` (code
  `OFFER_NOT_AVAILABLE`, no header). Unmodified run: BLOCKER —
  remaining 5 stale http files + CHG-0051 composition arity;
  not RED. Overlay diagnostic (6 stale http + nil-outbox,
  `list_definitions.go` REAL): 6 FAIL — invalid queries return
  404 not 400 (lookup precedes any lang handling), legacy cases
  lack `Content-Language: es` AND the JSON loses
  `ContentLanguage` — root cause: `LocalizedOffer` inherits
  `MissionOffer.MarshalJSON` so the label is dropped from the
  payload (probe-verified); `UnknownRemains404` PASS. English-
  success e2e deferred — persistence drops ContentEn today, no
  workaround invented. Gate 3-BOUNDARY for GetOffer pending.
  Gate 3-BOUNDARY (GetOffer only, conditional) approved
  2026-10-29. Provisional GREEN/BLUE landed REAL production:
  new `get_offer_response.go` (42 lines) — explicit
  `offerResponse` DTO preserving every PascalCase wire key +
  `MeritSummary` + `ContentLanguage` (fixes the inherited
  `MissionOffer.MarshalJSON` label drop; `ContentEn` never
  serialized); `get_offer.go` (39 lines) —
  `requestedMissionLanguage` before lookup → 400
  `INVALID_QUERY_PARAMETER` precedence, `Handle(id, language)`,
  `Content-Language` on success only, error switch unchanged.
  Overlay regenerated without get_offer.go; result: GetOffer
  7/7 + definitions 8/8 PASS, gofmt clean. PROVISIONAL —
  conditional on overlay; unmodified build still blocked by 5
  stale sites. Task remains open: ListOffers, ListContracts,
  contract detail, review-queue, create/publish partitions.
  ListOffers partition RED landed 2026-10-29 under Gate 2
  extension (RED/tests only): fixture gained
  `languageMixedOffersServer` (legacy + bilingual offer; fixture
  file now 109 lines); new `localized_list_offers_test.go`
  (144 lines) — `RejectsInvalidQuery` (fr/empty/conflicting →
  400; repeated es → 200 + header), `ServesLegacySpanish` (en /
  omitted + `Accept-Language: en` → `Rescate`, `total=1`,
  `ContentLanguage:"es"`, `ContentEn` absent,
  `MeritSummary.TotalMerits=70`, `Content-Language: es`),
  `EmptyHasNoContentLanguage` (other-season filter → 200,
  `total=0`, no header); separate `localized_list_offers_mixed_test.go`
  (58 lines) — `MixedActualLanguages` demands per-item en/es
  labels + en title on the bilingual offer + no header on the
  heterogeneous page. Unmodified run: BLOCKER (5 stale http
  files + CHG-0051 composition arity), not RED. Overlay
  diagnostic (5 stale http + nil-outbox; definitions/get-offer/
  parser/DTO all REAL): 7 FAIL — invalid queries return 200 not
  400, repeated-es lacks `Content-Language`, legacy cases lack
  header + `ContentLanguage` key (MarshalJSON drop); the mixed
  test fails for TWO distinct causes — boundary label drop AND
  persistence ContentEn drop (T-0054-007 dependency, stays RED
  until infra GREEN); assert order corrected 2026-10-29 so the
  English title mismatch (`expected "Hunt", actual "Caceria"`)
  fires observably before the label requirement fails;
  `EmptyHasNoContentLanguage` PASS.
  Gate 3-BOUNDARY for ListOffers pending.
  Gate 3-BOUNDARY (ListOffers HTTP subset, conditional)
  approved 2026-10-29. Partial GREEN landed REAL production:
  `list_offers.go` (88 lines) — `requestedMissionLanguage`
  before filter parse → 400 `INVALID_QUERY_PARAMETER`,
  `Handle(callerID, filter, language)`, `Content-Language`
  only on homogeneous non-empty pages, `offersResponse`
  envelope; new `list_offers_response.go` (20 lines) —
  `offersResponse{offers,total,limit,offset}` reusing
  `toOfferResponse` (label-preserving DTO). Overlay regenerated
  excluding list_offers + response (4 stale files + nil-outbox).
  Results: 7/7 non-mixed subtests PASS; GetOffer 7/7 +
  Definitions 8/8 regression PASS; `MixedActualLanguages` STILL
  RED for infra cause only — persistence drops `ContentEn` so
  the bilingual offer truthfully serves es (title `"Caceria"`,
  label `es`, homogeneous page → header `es`); boundary logic
  verified correct given the data. T-0054-007 persistence
  ContentEn mapping is required before mixed GREEN. Task
  remains open: ListContracts, contract detail, review-queue,
  create/publish partitions.
  ContractDetail partition RED landed 2026-10-29 under Gate 2
  extension (RED/tests only, zero production edits): fixture
  refactor — `languageHTTPServer` now delegates to
  `languageSessionServer(t, session, seed)`; contracts fixture
  gained `bilingualContractFixture(roster)` +
  `legacyContractFixture(roster)` (95 lines); new
  `localized_contract_detail_fixture_test.go` (47 lines) —
  `langMemberSession` (member-1, no roles) +
  `languageContractMemberServer` seeding bilingual/legacy
  member contracts + outsider `contract-foreign`. New tests:
  `localized_contract_detail_test.go` (149 lines) —
  `RejectsInvalidQueryBeforeLookup` (fr/empty/conflicting on a
  MISSING id → 400 demanded before lookup; repeated es on real
  contract → 200), `ServesEnglish` (en label + header +
  `Rescue`/`Capture` terms + economics + no `ContentEn`),
  `SpanishDefaultIgnoresAcceptLanguage`, `LegacyFallback`;
  `localized_contract_detail_projection_test.go` (92 lines) —
  `ReviewerKeepsFullView`, `MemberKeepsMinimizedView`,
  `OutsiderDenied` (403 FORBIDDEN), `UnknownReturns404`.
  Unmodified run: BLOCKER (contract_detail:21+35,
  review_queue:53+67, create:67 + CHG-0051 composition arity),
  not RED. Overlay diagnostic (3 stale http files mechanical
  LanguageSpanish + server.go nil-outbox): 7 FAIL / 1 PASS —
  invalid queries return 500 not 400 (lookup precedes parse;
  real store maps missing id to generic error not
  `ErrContractNil`), `content_language` + `Content-Language`
  missing, en text unreachable; `UnknownReturns404` also FAILS
  — pre-existing store error-mapping gap (generic
  `contract not found` → 500 `INVALID_REQUEST`), orthogonal to
  lang; `OutsiderDenied` + `repeated_es_accepted` PASS. Task
  remains open: review-queue + create/publish partitions +
  ContractDetail Gate 3-BOUNDARY.
- [x] **T-0054-007 — Write persistence boundary RED for bilingual snapshots and backfill.**
  Owner: agent. Deps: LAYER-GATE-APPLICATION. Outcome: infrastructure
  tests fail proving definition/offer/contract items round-trip both
  locales, legacy items decode `es`-only, the offer freeze copies both
  locales, and the backfill augmentation adds `en` without touching
  `es`/economics — including rejection of mismatched translations
  (TEST-INF-LANG-001/002). Tests:
  `tests/integration/infrastructure/dynamodb/localized_items_test.go`,
  `localized_backfill_test.go` against DynamoDB Local. Command:
  `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run 'TestLocaliz|TestBackfill'`.
  Expected RED: build fail on unmapped locale attributes / missing
  backfill writer. Maps to
  `internal/missions/infrastructure/dynamodb/{definition_store*,
  offer_store*,contract_item,contract_store_mapping,objective_mapping,
  attribute_helpers}.go` (+ backfill file decided at RED).
  Gate 2 persistence RED approved 2026-10-29 (definitions+offers
  bilingual snapshots + legacy absence only). Landed three
  integration-tagged files: `localized_items_fixture_test.go` (58
  lines — `localizedStoreFixture` with localhost-guarded endpoint,
  unique table per test, NO DeleteTable; `localizedDefinitionRequest`),
  `localized_definition_test.go` (141 lines — bilingual draft
  round-trip via real `CreateMissionDefinitionRequest` + raw
  `ContentEn` M-attribute shape assertions, published version
  current+#v1 records, legacy es-only fallback + attribute
  absence), `localized_offer_test.go` (110 lines — offer frozen
  from bilingual published definition round-trips English,
  legacy offer es fallback). Named-file command + package
  `-run '^TestLocalized(Definition|Offer)'`: BOTH compile clean —
  3 FAIL (`BilingualDraftRoundTrips`,
  `PublishedVersionRoundTrips`, `BilingualSnapshotRoundTrips` —
  `ContentEn` never written as an M attribute; served text stays
  Spanish) + 2 PASS (legacy fallbacks). Intended RED maps to
  `definition_store.go` (definitionItem), `definition_store_queries.go`
  (definitionFromItem), `offer_store.go`, `offer_store_query.go`
  + `attribute_helpers.go`.
  No CHG-0051 contamination in package compile. Gate 3-PERSISTENCE
  pending — no adapter/backfill changes authorized.
  Test-quality correction 2026-10-29 (pre-Gate-3 review):
  `localizedStoreFixture` now returns `(client, table, error)` —
  no require inside helpers; strict loopback+http endpoint guard;
  `require.Len` added before objective indexing. Re-run:
  identical 3-FAIL/2-PASS, classification unchanged.
  Gate 3-PERSISTENCE approved 2026-10-29 (scoped — definition/
  offer `ContentEn` DynamoDB M only). GREEN landed REAL
  production: new `localized_content_mapping.go` (59 lines) —
  `putEnglishContent` (absent on legacy zero-English) +
  `englishContentFromItem` (zero value when attribute missing);
  `definitionItem` + `definitionFromItem` (both current and
  version records), `OfferStore.Save`, `mapOfferItem` wired;
  `attribute_helpers.go` NOT modified despite dirty state.
  Results: named-file + package `-run '^TestLocalized(Definition|
  Offer)'` both 5/5 PASS; `go build` OK; HTTP
  `MixedActualLanguages` now PASS under overlay (8/8 — bilingual
  serves `"Hunt"`/`"en"`, no header on heterogeneous page).
  `attribute_helpers.go` untouched; contracts, backfill,
  migration NOT in scope — T-0054-007 remains partial,
  T-0054-008 open for remaining persistence partitions.
  Contract snapshot RED landed 2026-10-29 under Gate 2
  extension (tests only): `localized_contract_fixture_test.go`
  (51 lines — `localizedBilingualContract`/`localizedLegacyContract`,
  error-returning, Terms via `NewContractTerms`),
  `localized_contract_test.go` (116 lines — bilingual detail
  round-trip incl. post-Save in-memory mutation isolation + raw
  `ContentEn` M shape at PK=CONTRACT/SK=id, legacy es fallback),
  `localized_contract_list_test.go` (59 lines — mixed player
  list en/es + outsider invisible). Named-file + package
  `-run '^TestLocalizedContract'`: 2 FAIL (bilingual detail,
  mixed list — `ContentEn` never persisted, served `es`/
  `Rescate`/`Capturar`) + 1 PASS (legacy). Intended RED maps to
  `contract_item.go::contractItem` +
  `contract_store_mapping.go::mapToContract`. Gate 3-PERSISTENCE
  (contract scope) pending.
  Gate 3-PERSISTENCE (contract snapshot, scoped) approved
  2026-10-29. GREEN landed REAL production: `contractItem` writes
  `ContentEn` via `putEnglishContent` (Offer snapshot only),
  `mapToContract` decodes `ContentEn` onto the offer — terms
  reconstituted unchanged, roster/state/economics untouched.
  Named-file 3/3 PASS; package `-run '^TestLocalized(Definition|
  Offer|Contract)'` 8/8 incl. prior 5/5; build OK; gofmt clean.
  Task remains partial: backfill/migration not in scope. Next:
  Gate 2 for the remaining HTTP partitions (ListContracts
  first).
  ListContracts partition RED landed 2026-10-29 under Gate 2
  extension (tests only): fixture `languageContractsServer`
  (bilingual/legacy/outsider seeds via real ContractStore;
  endpoint guard hardened to loopback+http; stale ContentEn
  comment corrected) + 3 files — `localized_list_contracts_test.go`
  (invalid-query 400s, repeated es, empty, unauthenticated 401),
  `localized_list_contracts_languages_test.go` (homogeneous en
  + Spanish-default tables), `localized_list_contracts_mixed_test.go`
  (legacy fallback + mixed filtered). Unmodified: BLOCKER (4
  stale files + CHG-0051). Overlay diagnostic (4 stale handlers
  forced-Spanish + nil-outbox): 10 FAIL / 3 PASS — no lang parse
  (200 not 400), `Content-Language` header absent, `ContentLanguage`
  key absent (LocalizedContract drops label — same DTO need as
  offers), en text unreachable under the adapter; auth/empty/
  repeated-es already pass. Gate 3-BOUNDARY (ListContracts)
  pending.
  Gate 3-BOUNDARY (ListContracts only, conditional) approved
  2026-10-29. Provisional GREEN/BLUE landed REAL production:
  `list_contracts.go` (59 lines) — `requestedMissionLanguage`
  before filter parse, `Handle(callerID, filter, language)`,
  `Content-Language` on homogeneous non-empty, `[]` empty; new
  `list_contracts_response.go` (35 lines) —
  `contractListResponse` retaining label + Terms-hydrated
  objectives + MeritSummary, `ContentEn` never exposed. Overlay
  regenerated excluding both (3 stale + nil-outbox): 13/13
  subtests PASS + `TestListContracts_OfferMeritSummaryComesFromFrozenTerms`
  merit regression PASS. PROVISIONAL — ordinary build still
  blocked by contract_detail/review_queue/create + CHG-0051.
  Task remains open: review-queue, contract-detail, create
  partitions.
  Gate 3-BOUNDARY (ContractDetail only, conditional) approved
  2026-10-29. Provisional GREEN/BLUE landed REAL production:
  `contract_detail.go` — `requestedMissionLanguage` after
  identity check / before lookup (invalid lang on missing id →
  400, auth stays before parse), `Handle(identity, id, language)`,
  `detail.ContentLanguage` set from localized contract,
  `Content-Language` on success only;
  `contract_detail_projection.go` — `content_language` json
  field added to `contractDetail`. BLUE: helpers moved to
  `localized_contract_detail_fixture_test.go` (84 lines) as
  error-returning (`detailLangGet`, `decodeDetail`,
  `detailTerm`), both detail test files updated. Overlay
  regenerated (review_queue + create + nil-outbox only):
  `^TestLangContractDetail` → 11 PASS / 1 FAIL where the single
  failure is the diagnosed orthogonal BUG-0054-001 (store
  returns generic error → 500 not 404 for unknown id; expressly
  out of scope, test preserved). Ordinary build blocked by
  review_queue + create + CHG-0051. Task remains open:
  review-queue + create partitions.
  ReviewQueue partition RED landed 2026-10-29 under Gate 2
  extension (tests only): `localized_review_queue_fixture_test.go`
  (65 lines — `languageReviewQueueServer` seeds PENDING_VALIDATION
  via `NewMissionContractFromState` from accepted fixtures + one
  accepted contract for filter proof), `localized_review_queue_
  test.go` (125 — invalid 400s, auth 401/403, empty),
  `localized_review_queue_languages_test.go` (108 — homogeneous
  en, es default+explicit, legacy), `localized_review_queue_
  mixed_test.go` (48 — mixed labels + accepted excluded).
  Unmodified: BLOCKER. Overlay diagnostic: 10 FAIL / 3 PASS —
  no parse, header + `content_language` key absent; auth/empty
  pass. Maps `review_queue.go` + future `review_queue_response.go`.
  Gate 3-BOUNDARY (ReviewQueue) pending; task open: create
  partition.
  Gate 3-BOUNDARY (ReviewQueue only, conditional) approved
  2026-10-29. Provisional GREEN/BLUE landed REAL production
  `review_queue.go` (95 lines): `content_language` field on
  `reviewQueueEntry`, `toReviewQueueEntry` now takes
  `application.LocalizedContract`, `requestedMissionLanguage`
  after identity / before app call, `Content-Language` on
  homogeneous non-empty. Overlay regenerated (create +
  nil-outbox only): all 15 localized cases PASS +
  `TestReviewQueue_*` 3/3 legacy regression. PROVISIONAL —
  ordinary build still blocked by create handler + CHG-0051.
  Create-definition partition RED landed 2026-11-01 under
  Gate 2 extension (tests only): `localized_create_definition_
  fixture_test.go` (47 — `localizedCreateFixture` captures real
  `DefinitionStore`, admin/member sessions + `postDefinitionRequest`)
  + `localized_create_definition_test.go` (114 — bilingual
  persist, Spanish-only valid, missing-Spanish 422+no record)
  + `localized_create_definition_auth_test.go` (52 —
  unauthenticated 401, non-admin 403). Unmodified: BLOCKER.
  Overlay diagnostic: `BilingualPersistsEnglish` FAIL (baseline
  drops `translations.en`), 4/5 PASS. Maps `create_mission_
  definition.go` only. Gate 3-BOUNDARY (create) pending.
  Gate 3-BOUNDARY (create, conditional) approved 2026-11-01.
  Provisional GREEN/BLUE landed REAL production:
  `create_mission_definition.go` (89 lines — `translations`
  field + `CreateMissionDefinitionRequest` struct call with
  `ContentEn`) + new `create_mission_definition_translations.go`
  (37 — `englishContent()` DTO mapping, boundary-only). Overlay
  now server.go nil-outbox only: 5/5 PASS + `^TestLangDefinitions`
  ok + `TestListDefinitionsEndpoint_*` 2/2. PROVISIONAL —
  ordinary build blocked by CHG-0051 only. GET and draft-create
  partitions are provisionally green; publish Gate 2 evidence
  landed 2026-11-01: `localized_publish_definition_fixture_test.go`
  (59 lines) + `localized_publish_definition_test.go` (104 lines)
  — 3 invalid + 1 valid = 4/4 PASS under CHG-0051-only overlay;
  unchanged publish handler; no RED, no Gate 3; ordinary tree
  compile blocker CHG-0051 persists; no full Boundary gate
  passed.

## Wave 5 — Infrastructure / Interface / Composition GREEN

- [x] **T-0054-008 — Implement persistence and HTTP GREEN/BLUE.**
  Owner: agent. Deps: Gate 3-BOUNDARY on T-0054-006/007 RED.
  Outcome: locale item mapping + legacy decode, shared `lang` parsing,
  `content_language` on all projections, `Content-Language` header
  rules, bilingual draft/publish body handling — all boundary tests
  green through the real composition root and local resources; BLUE
  keeps files `< 150` lines. Commands: the two Wave-4 commands green +
  `go test -count=1 ./tests/unit/...`. Verification: integration PASS
  recorded. Next: T-0054-009.
- [x] **T-0054-009 — Rehearse the translation backfill on local infrastructure.**
  Owner: agent. Deps: T-0054-008. Outcome: dry-run + apply rehearsal
  against DynamoDB Local with reviewed translations carrying their
  source `es` snapshot: eligible frozen records gain `en` only after
  version and exact Spanish text match; divergent snapshots remain
  untouched and enter the unresolved-record report even when version
  and objective ids match. `es`/economics remain unchanged. Gate 1
  amendment, Gate 2 RED, and scoped Gate 3 approved 2026-11-01. Command:
  `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run 'TestBackfill'`.
  Verification: rehearsal evidence + unresolved report recorded in
  `rollout.md`. Next: T-0054-010.
  Gate 3-PERSISTENCE approved 2026-11-01 for an offline adapter only.
  Local rehearsal GREEN/BLUE: 12 top-level tests + 5 drift subtests
  pass with new unique tables, prior localized persistence 8/8 and
  DynamoDB adapter build pass. `translation_backfill*.go` adds only
  conditional `ContentEn` updates; dry-run does not write, and
  version, source Spanish, existing English, and phased Terms are
  guarded. All five maintained adapter files are below 150 lines.
  This task remains open for reviewed translation inputs and a
  separately approved real execution; no live migration or final
  Boundary gate is implied.
- [x] **T-0054-010 — Confirm composition wiring and record affected scope.**
  Owner: agent. Deps: T-0054-008. Outcome: `internal/missions/composition`
  and `cmd/` wiring verified — expected `not_affected` beyond handler
  construction already covered; any needed wiring lands only under an
  existing approved RED mapping or a new RED cycle. Command:
  `go test -count=1 -tags=integration -p=1 ./tests/integration/composition`.
  Verification: composition suite PASS or explicit `not_affected` note.
  Next: T-0054-011.

## Wave 5b — Catalog Translations And Safe Redeploy (Gate 1 Amendment)

US-MB-CATALOG-002 — Review and preserve the bilingual mission
catalog: an operator needs all existing Spanish catalog missions
paired with human-reviewed English by definition version, without
losing either locale on the next deployment. Scope: 24 generated
racing and 12 explicit PvP definitions in Git. Only racing (24
staging / eight production) is automatically synchronized; this
change does not enroll PvP in that pipeline. Gate 1 approved
2026-11-01; Gate 2 approved for T-0054-013 RED and T-0054-014
draft data, then extended 2026-11-03 for T-0054-015 seed
synchronization RED only. Gate 3-SEED for scripts/loader
was approved 2026-11-03; Gate 1 for a
least-privilege OFFER GetItem amendment was approved the same day.
Gate 2/3 for its IAM policy are still pending. Run one task at a time.

- [x] **T-0054-013 — Specify source-bound translation catalog RED.**
  Owner: catalog/service owner. Deps: Gate 1 amendment + scoped Gate 2.
  Wave: 5b. Outcome: `TEST-SEED-LANG-001` asserts exactly one
  version-bound English record for each of 36 definitions, matching
  Spanish source title/description/ordered rules/objective IDs/text;
  no category, version, ID, reward, or objective economics change.
  Invalid, missing, duplicate, or incomplete translations reject
  before rendering any write. Test path:
  `tests/unit/seeds/test_mission_translations.py`. Command:
  `python3 -m unittest discover -s tests/unit/seeds -p 'test_mission_translations.py' -v`.
  Expected RED: translation manifest/loader absent. Maps to
  `seeds/translations/` + `scripts/seed_missions.py` and a small seed
  translation loader. Verification: review classified RED and
  separate Gate 3 before any script edit. Next: T-0054-014.
  Actual RED 2026-11-02: command above ran three tests, two catalog
  subcases and one production-IDs case errored on missing
  `translated_definitions`; seven invalid-manifest cases failed
  because `AttributeError` is not the expected `ValueError`. Exact
  `SourceSpanish`/English fixtures use only temporary files for
  invalid cases; no network or database. Test-only files are 60
  and 90 lines; no production edit. This is API-absence RED, not
  translation-content or behavior GREEN. Continue with approved
  T-0054-014 drafts; script code waits for later Gate 3-SEED.
- [x] **T-0054-014 — Draft English for all catalog missions.**
  Owner: catalog content author; semantic reviewer: operator. Deps:
  T-0054-013 RED review and Gate 2 data-file scope. Wave: 5b.
  Outcome: version-bound English drafts committed to repository
  artifacts for all 24 racing and 12 PvP definitions; source Spanish
  snapshots and gameplay/economic fields unchanged. Split new
  data/config files below 150 lines. Verification: independent
  structure/coverage check; human terminology approval belongs to
  T-0054-017 and remains pending. A shape check never substitutes
  for semantic signoff. No non-local persistence. Next: T-0054-015.
  Draft outcome 2026-11-03: nine version-4 JSON files hold 36
  distinct entries (24 racing, 12 PvP); each has exact source
  Spanish, version, ordered objective IDs/rules, and structurally
  complete English. Source snapshots were compared against both
  unchanged Spanish catalogs, not inferred from the missing loader.
  Files are 114-122 lines each. No seed code, CI, or database write
  occurred. These remain unreviewed proposals and must not deploy
  before T-0054-017's operator signoff. TEST-SEED-LANG-001 remains RED
  because the translated-definition loader has not been approved.
- [x] **T-0054-015 — Prove redeploy and new-version seed safety (RED).**
  Owner: seed adapter owner. Deps: T-0054-013 RED and T-0054-014
  drafts, Gate 2. Wave: 5b. Outcome: `TEST-SEED-LANG-002` reproduces
  current rejection/erasure risk when a same-version catalog is
  re-synchronized after English augmentation; proves wrong English
  or changed Spanish/economics fail closed, legacy missing `en` stays
  unchanged, and a genuinely new version cannot publish without
  complete `en`. Test paths: `tests/unit/seeds/` and
  `tests/integration/infrastructure/dynamodb/test_seed_translation*.py`
  against unique loopback DynamoDB Local tables, left in place.
  Commands: `python3 -B -m unittest discover -s tests/unit/seeds -p 'test_seed_translation*.py' -v`;
  `DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_seed_translation*.py' -v`.
  Verification: classify actual RED before scoped Gate 3-SEED; no
  shared/production query or write. Next: T-0054-016.
  Actual RED 2026-11-03: unit combined `Ran 10` —
  five failed subcases (missing/incomplete English on first/new
  publish), five errors (two same-version false rejections and
  three CLI render cases missing `ContentEn`); four unit test
  methods also pass. Integration combined `Ran 14` — 11 failures,
  one error, two passes: existing legacy/bilingual redeploy safety,
  offer Spanish/economic/English drift, dry-run, and a concurrent
  translation conflict are unprotected; CLI local bootstrap writes
  Spanish-only rows. New bilingual version advance and unchanged
  bilingual redeploy are baseline GREEN. An initial connection
  error while DynamoDB Local was down was environment-only; the
  final independent combined run reached fresh tables with no
  setup errors. No tables deleted. Only Python tests/fixtures
  added (42-144 lines), no production code changed. Gate
  3-PERSISTENCE/SEED was approved 2026-11-03 for loader,
  synchronization, and both CLIs; it does NOT include IAM or
  deployment. English semantic review remains T-0054-017.
- [x] **T-0054-018 — Specify least-privilege offer-read IAM RED.**
  Owner: deployment-policy owner. Deps: Gate 1 IAM amendment
  and Gate 2 for the IAM test (both approved 2026-11-03).
  Wave: 5b, before offer preflight GREEN/deployment. Outcome:
  TEST-SEED-IAM-001 proves the existing seed role is missing only
  `GetItem` on exact OFFER partition keys and that the proposed
  policy may not grant `Query`/`Scan` on OFFER, read CONTRACT, or
  broaden write actions/roles. Test path candidate:
  `tests/unit/seeds/test_seed_iam_policy.py`; standalone command:
  `python3 -B -m unittest discover -s tests/unit/seeds -p 'test_seed_iam_policy.py' -v`.
  Expected RED: missing constrained `GetItem` statement in
  `template.yaml`, not YAML/parser/environment failure. Next:
  review RED and request scoped Gate 3-IAM; no IAM edit yet.
  Actual RED 2026-11-03: `Ran 3 tests`, two PASS (original role,
  read/write scopes, and synthetic escalation guards), one FAIL
  solely because the exact `GetItem`/OFFER/table statement is absent.
  Test file is 112 lines; no YAML parse, import, network, fixture,
  or environment failure. `template.yaml` was unchanged during
  RED. Gate 3-IAM was subsequently approved; no policy deployment
  or live read approved.
- [x] **T-0054-019 — Grant exact-offer read without widening seed role.**
  Owner: deployment-policy owner. Deps: reviewed T-0054-018 RED
  and scoped Gate 3-IAM. Wave: 5b. Outcome: in `template.yaml`,
  the existing role gets only `dynamodb:GetItem` on the same
  MissionsTable ARN with `dynamodb:LeadingKeys: [OFFER]` as its
  own statement; other partition reads and writes remain as
  before. Verification: standalone TEST-SEED-IAM-001 GREEN/BLUE
  plus static SAM/IAM validation without any AWS call. The existing
  SAM template exceeds 150 lines; inventory it honestly and seek a
  separate controlled cleanup scope rather than splitting it as
  part of a permission hunk. No policy deployment or live read is
  authorized. Next: T-0054-016 local GREEN.
  Actual GREEN 2026-11-03: Gate 3-IAM approved, one independent
  `GetItem`/OFFER/MissionsTable statement added in `template.yaml`;
  existing role, read and write statements unchanged. Standalone
  TEST-SEED-IAM-001 passed 3/3, exit 0; static-only, no SAM
  deployment, AWS request, data change, or DynamoDB Local run.
  Template was already 311 lines; it is now 318. Existing
  over-limit template is recorded as a cleanup exception, not
  silently called BLUE-conformant. Next: T-0054-016 local GREEN.
- [x] **T-0054-016 — Preserve bilingual snapshots on seed GREEN/BLUE.**
  Owner: seed adapter owner. Deps: reviewed T-0054-015 RED, scoped
  Gate 3-SEED (approved), and T-0054-019 before deployment.
  Wave: 5b. Outcome: repeated same-version sync never
  overwrites `ContentEn` or Spanish; existing v4 without `en`
  remains legacy until separate backfill; new versions and advanced
  offers store both locales atomically without rewriting accepted
  contracts. Keep the original transaction/concurrency guards and
  split maintained over-limit Python script into files <150 lines.
  Verification: focused Python unit + unique-table integration +
  existing seed regressions; Go backfill/persistence code is
  untouched, so its independent final gate remains pending.
  Next: T-0054-017.
  Actual GREEN/BLUE 2026-11-03: six focused Python modules now own
  immutable catalog values, exact-source English validation, CLI
  transport, conditional definition/offer updates, and orchestration.
  `seed_missions.py` remains a facade; local CLI loads both
  catalogs with version-bound English and verifies preserved
  legacy snapshots. Per-key consistent `GetItem` replaces any
  OFFER partition query in the deployment seed. Old-version
  offer Spanish, English, or merit drift rejects before new
  version writes; missing-English new versions reject, and a
  concurrent English edit cancels the transaction. RED for
  corrupt prior-version offers: two tests, four failed subcases;
  RED for es-only local CLI redeploy: one failure (the old
  bilingual bootstrap case passed). Both partitions turned
  GREEN (2/2 and 2/2). Final Python evidence: seed unit 16/16,
  catalog regression 8/8, new isolated DynamoDB Local integration
  17/17; legacy CI seed integration fixture was upgraded with
  synthetic v5/v6 English and its destructive teardown removed:
  after an initial credential-only environment error and a
  classified missing-English RED, it passed 1/1 on a newly
  created table left intact. Eight maintained seed scripts are
  44-126 lines; the legacy integration test is 127 lines. Go
  code/tests and CHG-0051 were not changed. No SAM deployment,
  shared AWS query, production data write, or real migration.
  BLUE: separate catalog/text validator from AWS transport and
  conditional synchronization; unchanged Domain/Application
  ports, no new Value Object needed for CLI-only rendering.
  Full Boundary, security, coverage, mutation, E2E,
  template-size exception, and release
  checks remain open under T-0054-011/017.
  The user approved the English wording of all 36 records on
  2026-11-03, and selected a **local-only** publication target.
  Neither approval authorizes staging/production deployment.
- [x] **T-0054-017 — Confirm authoring readiness and cutover order.**
  Owner: release/operator owner. Deps: T-0054-014 human content
  review, T-0054-016 GREEN, operator authoring client supports
  `translations.en`, and ordinary Boundary gate. Wave: 5b.
  Outcome: approved staging rollout order plus go/no-go for the
  already-enforced publish rule; production stays blocked until
  operator tooling and translations are reviewed. Verification:
  reviewed content, dry-run diff, fallback and rollback evidence in
  `rollout.md`; any real AWS apply requires separate authorization.
  Next: T-0054-011.
  Content review completed by user 2026-11-03 (36/36).
  Non-local rollout remains pending operator-client readiness,
  ordinary Boundary, L3 gates, and separate environment approvals.
  User accepted that IAM `LeadingKeys: [OFFER]` restricts PK but
  cannot restrict SK; the script itself reads exact seed-owned keys.
  Read-only local preflight 2026-11-03: `applied: false` for both
  catalogs (24 racing, 12 PvP) in `local-training`; 36 current
  definitions, 36 versioned rows and 36 active offers are all
  Spanish-only; CONTRACT COUNT query returned zero. No writes.
  Ordinary seed apply cannot enrich these existing v4 rows.
  A separately gated local backfill is needed before claiming
  that the approved text is published in local data.
  Next local: T-0054-020; non-local: T-0054-011.
- [x] **T-0054-020 — Prove local-only reviewed translation invocation (RED).**
  Owner: offline infrastructure/interface owner. Deps: reviewed
  catalog, T-0054-009 backfill rehearsal, Gate 1 local amendment
  (approved 2026-11-03), and separate Gate 2 for test writing.
  Wave: 5b local-only. Outcome: TEST-SEED-LOCAL-001 proves manifest
  identity/shape and duplicate rejection before access;
  TEST-SEED-LOCAL-002 proves non-loopback rejection, read-only
  preview, unresolved conflict refusal, and English-only
  conditional apply on new private tables. Lowest independent
  tests: `tests/unit/missions/infrastructure/dynamodb/reviewed_manifest_test.go`
  via `go test -count=1 ./tests/unit/missions/infrastructure/dynamodb -run '^TestReviewedTranslationManifest' -v`;
  `tests/integration/infrastructure/dynamodb/test_local_translation_publish.py`
  via `DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_local_translation_publish.py' -v`.
  Expected RED: missing local manifest/CLI capability, not
  fixture/compiler/emulator failure. Review actual RED before
  Gate 3; no production or application-table writes. Next:
  T-0054-021.
  Gate 2 approved 2026-11-03 (tests only). TEST-SEED-LOCAL-001
  standalone RED: 137-line unit test fails to compile solely
  for missing `LoadReviewedTranslationManifests` and
  `ValidateLocalBackfillEndpoint` (3+2 references); no setup
  or fixture failure. TEST-SEED-LOCAL-002 standalone RED: five
  tests in a 128-line file each fail solely because
  `cmd/local-translation-backfill` does not exist. New private
  DynamoDB Local tables were created on port 18000 and left
  intact; no setup errors, app-table access, or production edit.
  Both RED partitions reviewed; Gate 3-LOCAL-BACKFILL is
  requested, not approved. Next: T-0054-021 only after approval.
- [x] **T-0054-021 — Implement and refactor the local backfill invoker.**
  Owner: offline infrastructure/interface owner. Deps: reviewed
  T-0054-020 RED and scoped Gate 3-LOCAL-BACKFILL.
  Wave: 5b local-only. Outcome: a manifest reader under
  `internal/missions/infrastructure/dynamodb/` and
  `cmd/local-translation-backfill/` CLI reuse the tested
  `TranslationBackfiller.Rehearse`; explicit endpoint/table,
  default no-write preview, and separate `--apply` guarded by
  a fresh unresolved-free preview. Reject remote endpoints
  before creating a client; per-row conditions add only
  `ContentEn`. BLUE keeps files below 150 lines, reruns both
  tests and existing backfill regressions on new local tables,
  and records partial-conflict recovery. No CI, IAM, or public
  API change. Next: T-0054-022.
  Gate 3 approved 2026-11-03. GREEN: four unit methods and
  five CLI integration cases PASS; existing Go backfill
  regression 12 top-level + 5 drift subtests PASS on private
  local tables. BLUE: production 99/35/95 lines, tests
  137/128 lines, gofmt clean; manifest loading, local-only
  validation and output are separate from the existing
  conditional update adapter. No new value object needed,
  Domain/Application/public HTTP not affected. The CLI
  has not read or written `hbk-missions-local`; next first
  action is a separately authorized read-only target preview,
  then specific apply approval if no unresolved records.
- [x] **T-0054-022 — Publish approved English to local history safely.**
  Owner: local data operator. Deps: T-0054-021 GREEN/BLUE,
  current read-only target preview with zero unresolved,
  and separate explicit approval for applying to
  `http://localhost:8000` / `hbk-missions-local`. Wave: 5b,
  after the code gates. Outcome: reviewed source-matching
  local records gain only `ContentEn`; verify both languages
  for current/version/offers, compare every other attribute
  to the captured pre-apply state, and report any conditional
  conflicts without pretending partial application is atomic.
  The 108 Spanish-only rows and zero contracts observed on
  2026-11-03 are not a fixed apply target or permission for
  unreviewed records. Never delete data or write outside the
  named local table. Next: T-0054-011 (non-local gates remain).
  Read-only review 2026-11-04: the CLI preview confirms 108
  eligible unique rows and zero unresolved; current definitions,
  versions and offers still have zero `ContentEn`. Local
  missions HTTP gateway returns 502 because its upstream
  on port 3006 is absent; compiling the service fails at
  CHG-0051 composition `NewPlacementFactPublisher` 3-arg
  versus 5-arg wiring. No apply, backend restart or deployment.
  Task remains open for separate exact-table write approval
  and post-apply snapshot verification.
  The user then approved apply to this exact loopback
  endpoint/table. The source-bound translation suite passed
  3/3, fresh preview returned 108 unique eligible rows and
  zero unresolved, and the local CLI reported 108 conditional
  `ContentEn` updates with no conflicts. Strongly consistent
  before/after reads confirm all other attributes, 36 reviewed
  English snapshots, season and zero contracts unchanged.
  Post-apply preview has zero eligible and both seed catalogs
  are synchronized in read-only mode. Local data is published;
  HTTP remains unavailable (gateway 502/no listener on 3006
  and CHG-0051 composition compile blocker). Do not mark
  the full Boundary or L3 gates complete. Next: restore
  separately owned local service wiring before HTTP verification.

## Wave 6 — L3 Gates And Convergence

- [x] **T-0054-011 — Run the L3 gate suite and record evidence.**
  Owner: agent. Deps: T-0054-008..010, T-0054-013..017. Outcome: security gate
  (labeling-integrity + fail-closed `lang` validation evidence in
  `security-review.md`), coverage ≥90% on production scope with no
  affected-scope regression, mutation report (`mutation-report.md`),
  critical E2E (`critical-e2e.md`: TEST-E2E-LANG-001 — member reads an
  English offer, accepts, contract keeps both frozen locales through a
  v2 translation edit), documentation gate (API contract + product
  docs), clean-up `< 150` lines, `bash tools/validate-sdd-change.sh`.
  Command: `bash tools/validate-sdd-change.sh` + focused gate commands.
  Verification: results converged in `verification.md`.
  Next: T-0054-012.
- [x] **T-0054-012 — Converge artifacts, verify, consolidate, and archive.**
  Owner: agent. Deps: T-0054-011 plus explicit final-verification
  approval via `WORKFLOW-COMMON_SDD_VERIFY_SPEC_WORKFLOW` — canonical
  merge is not automatic after T-0054-011. Outcome: verify-spec
  evidence review approved by the user; only then the approved delta
  merges into `specs/capabilities/mission-board/spec.md` and
  `specs/capabilities/mission-contract-fulfillment/spec.md`,
  `specs/index.md` is updated, and the change archives under
  `specs/changes/archive/2026/` with no active duplicate.
  Verification: final gate table all green in `verification.md`.

- CONVERGENCE + ARCHIVE (2026-11-16, user-approved): stale checkboxes
  reconciled — the implementation and test evidence for T-0054-001/003/
  004/006/007/008/009/010/017 landed across earlier sessions and was
  verified green by the full validator on 2026-11-16 (gofmt/vet/unit/
  integration/race all pass; coverage domain 92.4%, application 91.7%).
  L3 gates recorded: security-review closed with no Critical/High,
  mutation-report.md (manual analysis), critical-e2e.md
  TEST-E2E-LANG-001 PASS (English offer → accept → frozen bilingual
  contract → post-edit snapshot intact), documentation gate
  (`lang` contract in docs/API-CONTRACT.md), no production file ≥150
  lines. Delta merged into mission-board (REQ-MB-017 + extensions of
  REQ-MB-005/013/015/016) and mission-contract-fulfillment (REQ-MCF-026
  + extensions of REQ-MCF-006/021/022); index updated; change archived
  under specs/changes/archive/2026/CHG-0054-localize-mission-content.
