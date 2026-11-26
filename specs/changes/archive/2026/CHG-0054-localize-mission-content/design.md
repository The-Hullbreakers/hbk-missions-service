---
change_id: CHG-0054
artifact: design
status: proposed
updated: 2026-11-01
---

# Design — CHG-0054

Settled at Gate 1 (2026-10-27): the language contract, fallback and
labeling rules, atomic bilingual versions, frozen bilingual snapshots,
and the backfill constraints below. Remaining choices are
implementation details derived at RED, not unsettled Gate 2 questions.

## Layer Split — Settled

- **HTTP (transport)**: parses raw `?lang=` syntax on the six
  endpoints — absent → the request carries no explicit language;
  repeated values must agree, and empty or conflicting values resolve
  `400 INVALID_QUERY_PARAMETER` before any use case runs. Facets keeps
  its current signature. Handlers stay thin: parse, call one use case,
  map the response.
- **Domain**: `Language` is a value object over `es`/`en`;
  `ParseLanguage` validates supported codes only — it never sees
  transport repetition syntax. The localized content set exposes the
  stored locales for lookup; it does not decide which one a request
  sees.
- **Application**: each read use case takes a single
  `domain.Language` request value (default `es` when the transport
  supplies none) and resolves the effective language per returned
  representation — the requested locale when present, `es` fallback
  otherwise, always reporting the locale actually served. Default and
  fallback resolution live here, not in the transport or the domain
  parser. Authoring use cases forward both locales to the domain and
  rely on the domain completeness rule.

## Domain

- `Language` — `es`/`en` codes only; `es` is the default.
- `LocalizedContent` (candidate name) — one locale's player-facing
  text: title, Markdown description (verbatim), ordered rules, and
  per-objective `name`/`description`. The mission content set holds
  `es` (required) plus optional `en`; objective ids and rule order
  exist once, shared by both locales. The concrete Go shape
  (locale-keyed map vs. paired fields) is derived at Domain RED.
- Publish completeness — structural only: every `es` `objective_id`
  present in `en` with required text, `en` rule count equals `es`
  count; no semantic translation checking. Violation resolves as
  `ErrDefinitionInvalid`. `es`-only draft creation stays valid.
- Search — `OfferListFilter.Match`/`matchesQuery` compares `q` against
  the offer's effective-language title+description (Spanish fallback
  included) before sort/pagination.

## Application

- Read handlers gaining a `Language` request:
  `ListDefinitionsHandler`, `ListOffersHandler`, `GetOfferHandler`,
  `ListPlayerContractsHandler`, `GetContractHandler`,
  `ListPendingReviewHandler`. Ports stay unchanged query operations.
- `CreateMissionDefinitionHandler` accepts the optional `en`
  translation and forwards both locales atomically;
  `PublishMissionDefinitionHandler` relies on the domain completeness
  rule — no application-level revalidation.
- Whether `lang` rides inside `OfferListFilter` or a sibling request
  field is an implementation detail fixed at Application RED.

## HTTP Boundary — Settled Contract

- `?lang=es|en`; absent → `es`; unsupported/empty/conflicting repeats
  → `400 {code: INVALID_QUERY_PARAMETER}`; agreeing repeats resolve as
  that value.
- Served-language label on every representation, keyed by the response
  family's existing casing: `ContentLanguage` on the PascalCase
  board/catalog projections and contract-list items,
  `content_language` on snake_case contract-detail and review-queue
  DTOs. Single reads label at top level; collections label each item;
  a contract item's label sits at contract level and covers its
  embedded `Offer`.
- `Content-Language` header: emitted if and only if every
  representation in a non-empty payload resolved the same language —
  never on heterogeneous or empty collections.
- `POST /api/v1/missions/definitions` keeps its existing snake_case
  body as Spanish and accepts optional `translations.en`
  `{title, description, rules, objectives: [{objective_id, name,
  description}]}`; `rules` may be empty; objective `description`
  optional as `es` allows; `name` required.
- Serialization site (existing `MarshalJSON`/projection DTOs vs. a
  transport decorator) is an implementation detail fixed at Boundary
  RED; the settled contract is the fields above.

## Persistence Model

- Definition, offer, and contract items store locale attributes (exact
  attribute names/structure fixed at Boundary RED). Legacy items
  without locale attributes decode as `es`-only content from today's
  scalar fields — extending the existing "missing `Rules` decodes
  empty" tolerance. `ContentEn` fields carry `json:"-"`: locale
  snapshots are domain-internal until the boundary projects the
  selected language, so locale persistence is an explicit
  item-attribute mapping — never JSON serialization.
- Frozen snapshots: offer items persist both locales at creation;
  contract items persist the bilingual `ContractTerms`/`Offer`
  snapshot at acceptance. Reads never rewrite them.
- Backfill is an offline, rehearsal-first operation — no live
  migration is authorized. Tooling shape is an implementation detail
  decided at T-0054-009 within these constraints: `en` attributes only,
  per definition version; `es` content, economics, ids, and states
  untouched; records whose stored identity (version, objective ids,
  rule count) mismatches the translation are rejected and reported
  unresolved. Amendment approved 2026-11-01: reviewed input also
  carries the Spanish source snapshot for that version. Compare its
  title, description, ordered rules, and objective ids, names, and
  descriptions against each candidate's frozen Spanish content
  (including effective contract terms and phases). A differing
  snapshot is unresolved even when code, version, ids, and rule count
  match: `DefinitionStore.Save` can overwrite a version row. Dry-run
  never writes; apply may add `ContentEn` only after this match and
  must not replace any intervening Spanish or existing English data.

## Versioned Catalog And Safe Seed (Gate 1 Amendment 2026-11-01)

The source catalogs are `seeds/racing-missions.json` (24 generated
definitions, eight production-enabled) and `seeds/pvp-missions.json`
(12 explicit definitions; local/manual seeding, not CI). Keep their
Spanish text and economics unchanged. Reviewable English artifacts
under `seeds/translations/` group records by definition code and
version; each record contains the exact `SourceSpanish` snapshot
(`Title`, `Description`, ordered `Rules`, `Objectives` keyed by stable
`ObjectiveID`) and a corresponding `English` snapshot. Split new
data/config files below 150 lines. Only these text fields translate;
IDs, categories, mode, version, and merits remain shared.
Each sidecar JSON file has a `translations` array of records with
`DefinitionCode`, `DefinitionVersion`, `SourceSpanish`, and `English`.
The last two objects use `Title`, `Description`, ordered `Rules`, and
`Objectives` entries containing only `ObjectiveID`, `Name`, and
`Description`. The seed's `translated_definitions(catalog,
environment, translations_dir)` yields the existing AWS attribute
rows plus the mapped `ContentEn`, without changing Spanish fields.

The seed adapter checks every selected definition against exactly one
translation record and rejects missing, duplicate, stale-source,
incomplete, or wrong-version English before proposing any write.
It encodes `ContentEn` with the same DynamoDB map shape as
`localized_content_mapping.go`. The repository drafts are not a
content-review signoff; an operator must approve the English text
before applying it to any non-local table.

Compatibility matrix for `scripts/seed_missions.py` (which CI runs
after every racing deploy):

- Existing same-version rows: compare ALL Spanish, structural,
  economic, and identity fields exactly. Existing `ContentEn` must
  match the reviewed English or synchronization fails closed. Do not
  rewrite the immutable version row, erase English from current or
  offer rows, or silently add English to a historical row missing it.
  The offline conditional backfill owns historical augmentation.
- New definition version: require complete English and atomically
  write Spanish plus `ContentEn` in the current, versioned, and new
  offer snapshots. When an existing offer moves to a new version,
  update its English in the same operation as its Spanish and
  `DefinitionVersion`; accepted contracts stay frozen.
- No active season: only definitions are eligible; do not create an
  offer. Existing concurrent checks, transaction limit, and
  no-downgrade semantics remain. Fail closed on version/text conflict
  or missing authoring readiness rather than relaxing invariants.

Gate 1 offer-read amendment (approved 2026-11-03): the deploy
policy in `template.yaml` currently allows `Query` and
`ConditionCheckItem` only for DEFINITION/SEASON, and
`PutItem`/`UpdateItem` for DEFINITION/OFFER. An offer dry-run
cannot detect existing Spanish, merit, or `ContentEn` drift without
reading its stored item first. Use a strongly consistent `GetItem`
for each exact `PK=OFFER, SK=seed-<code>-<season hash>` key already
derived by `offer_item`; do not query the whole OFFER partition.
Keep the read limited to the existing MissionsTable ARN and one
independent policy statement with `dynamodb:GetItem` and
`dynamodb:LeadingKeys: [OFFER]` only. The policy grants no
`Query`/`Scan`, no contract access, no new write action, and no
new role. This IAM condition restricts PK to OFFER but cannot
limit SK; the script computes exact seed-owned SKs. Gate 3-IAM
approved 2026-11-03 and its static policy
test passed 3/3; the source permission is not deployed. Reuse
the existing conditional `UpdateItem` action for
an offer with a full-old-item condition so a concurrent authored
English edit aborts the transaction. TEST-SEED-IAM-001 first
proved the missing statement RED; scoped Gate 3-IAM then allowed
the one `template.yaml` edit. Gate 3-SEED remains separate;
neither the policy nor the script may deploy until the remaining
rollout gates are complete.

Test strategy: first isolated Python seed unit RED under
`tests/unit/seeds/` for catalog completeness, source alignment,
legacy versus new-version actions, and no changed economics; then
integration RED under `tests/integration/infrastructure/dynamodb/`
against a unique DynamoDB Local table for repeat seed after
backfill, conflicting English, and version advancement. Do not
delete any new test table; only loopback endpoints. Existing
`tests/seeds/` tests are legacy regression evidence, not a new suite.
After scoped Gate 3, split touched over-limit Python seed code into
focused modules below 150 lines as behavior-preserving BLUE.

Rollout mechanism: deploy operator authoring support (including
`translations.en`) and have translations reviewed before releasing
the service's already-enforced new-version publish rule. No new
runtime toggle or public endpoint is authorized by this amendment.
If an operator client cannot be ready first, stop and seek another
design approval rather than deploying a breaking publish path.

## Local Reviewed-Translation Publication (Gate 1 2026-11-03)

The user approved all 36 English texts and chose the local application
only. Its read-only preflight at `http://localhost:8000` /
`hbk-missions-local` found one active season, 36 current definitions,
36 versioned definitions, and 36 offers: all 108 are same-version
Spanish-only; CONTRACT count is zero. An ordinary seed apply must
continue preserving these legacy rows, not silently enrich them.

Reuse the existing `TranslationBackfiller.Rehearse` adapter rather
than adding a parallel migration algorithm. A new offline local-only
invoker loads the nine reviewed sidecars into
`ReviewedMissionTranslation` records and checks identity,
duplicates, source Spanish, and complete English before connecting.
It rejects non-loopback endpoints and never falls back to a cloud
endpoint or ambient AWS credentials. The default mode first reports
eligible and unresolved records without writes;
an explicitly requested apply repeats the preview and refuses to
start if it has unresolved records. Each eligible update changes
only `ContentEn` under the adapter's source/version/absence
conditions. An intervening conflict is reported, not overwritten.
The 108-row preflight is an observation, not a baked-in threshold
or permission to update unexpected rows. Apply is per-record, not
globally atomic: on a partial failure, stop, report affected IDs,
and re-preview before any idempotent retry; never delete or revert
Spanish, economics, or accepted terms.

Gate 2 tests use new uniquely named tables on the dedicated
loopback port 18000, left in place. They prove invalid manifests
and non-local endpoints fail before access, preview is read-only,
and apply on owned test data adds only English while conflicts
remain unchanged. Gate 3 may add only the local invoker and focused
manifest reader after actual RED; Go domain/application and the
HTTP composition are not affected. The source files remain under
150 lines each. No staging/production/AWS access, CI change,
local application write, or real backfill is authorized by Gate 1.
After GREEN/BLUE and an actual preview of the target table,
request explicit approval for the local table apply separately.

## Compatibility

- Read paths and the draft POST are additive: omitted `lang`
  reproduces today's behavior plus the new label field; old readers
  ignore unknown fields/headers.
- The publish completeness rule is a deliberate behavioral change:
  existing published es-only versions stay valid, but new-version
  publishes require `en` — the staged cutover lives in `rollout.md`,
  and the enforcement mechanism is decided at implementation with test
  coverage of the chosen approach.
- One new `400` reach on `lang` misuse via the existing
  `INVALID_QUERY_PARAMETER` code; one optional response header.

## Implementation Details (Derived At RED — Not Gate 2 Questions)

1. Concrete locale-set type and DynamoDB attribute layout — Domain and
   Boundary RED output; the atomic-per-version and `es`-required
   invariants are settled.
2. Serialization mechanism for the label — Boundary RED output; the
   placement and casing rules above are settled.
3. Backfill tooling form (script vs. operator procedure) — T-0054-009
   output; rehearsal-first and no-live-migration are settled.
4. `lang` request plumbing inside each handler — Application RED
   output; one `domain.Language` per request is settled.
