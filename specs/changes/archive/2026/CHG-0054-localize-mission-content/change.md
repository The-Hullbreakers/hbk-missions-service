---
change_id: CHG-0054
title: Localize mission content for Spanish and English readers
status: proposed
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/active/CHG-0054-localize-mission-content/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-GO_REST_API_WORKFLOW
context_scope: none
created: 2026-10-27
updated: 2026-11-01
---

# Localize mission content for Spanish and English readers

## Intent

Serve player-facing mission content — title, Markdown description,
ordered rules, and objective names/descriptions — in Spanish or English
so members and reviewers read the board, catalog, contracts, and review
queue in their selected language, while legacy Spanish-only records
degrade honestly instead of pretending to be translated.

## Business Value

- English-speaking members can discover, accept, and execute missions
  without translation tools; Spanish remains the default and the
  historical baseline.
- Accepted contracts keep both language snapshots frozen, so a later
  translation edit never rewrites what a player was promised.
- Every response labels the language actually served, so clients never
  render Spanish text as if it were English.

## User Stories

- **US-MB-LOCALIZE-001 — Read the board in English.** As a member, I
  want mission definitions and offers served in English when I ask for
  them, so I can evaluate rules and objectives in my language.
- **US-MCF-LOCALIZE-001 — Read my contracts in English.** As a member or
  reviewer, I want contract lists, contract detail, and the review queue
  served in my selected language, so adjudication and progress tracking
  use the same content the player saw.
- **US-MB-CATALOG-002 — Review and preserve the bilingual mission
  catalog.** As an operator, I want the existing versioned missions
  translated and saved without losing either language on redeploy, so
  English readers can trust the mission terms and rewards.

## Scope In

- `?lang=es|en` query parameter on `GET /api/v1/missions/definitions`,
  `GET /api/v1/missions/offers`, `GET /api/v1/missions/offers/{offer_id}`,
  `GET /api/v1/missions/contracts`,
  `GET /api/v1/missions/contracts/{contract_id}`, and
  `GET /api/v1/missions/review-queue` (title only). Omitted `lang`
  resolves as `es`; invalid, empty, or repeated conflicting values
  resolve as `400 INVALID_QUERY_PARAMETER`. Language is a URI query
  variant (cacheable per URI); `Accept-Language` is not negotiated.
- Explicit served-language label on every mission representation —
  `ContentLanguage` on PascalCase projections (board, catalog,
  contract-list items), `content_language` on snake_case DTOs (contract
  detail, review-queue entries); single reads label top level,
  collections label each item, and a contract item's label covers its
  embedded offer. The `Content-Language` response header is emitted
  only when a non-empty payload is homogeneous — heterogeneous and
  empty collections omit it.
- Bilingual content model: `es` and `en` are atomic per definition
  version; publishing a new version requires both locales complete;
  draft creation stays backward compatible (`es`-only POST plus an
  optional `en` translation); no machine translation.
- Frozen offer and contract snapshots inherit both locales at
  acceptance; later version updates never rewrite them.
- Legacy records without English serve the Spanish snapshot labeled
  `content_language: es` under any `lang` — never `406`, never falsely
  labeled English.
- Offer search `q` matches the selected effective language, including
  the Spanish fallback, before ordering and pagination; ordering and
  totals stay deterministic and change only through language-dependent
  matches.
- Translation backfill: reviewed translations for the current catalog
  and non-destructive augmentation of eligible frozen records by
  version; mismatched translations are rejected and unresolved records
  reported. No live migration execution is authorized by Gate 1.
- Gate 1 catalog/rollout amendment (approved 2026-11-01): prepare English
  text for all 24 generated racing and 12 explicit PvP definitions as
  version-bound, reviewable repository artifacts, paired with their
  exact Spanish source; protect repeated seed/deploy against losing
  English or rewriting same-version Spanish/economics. PvP remains
  local/manual; adding it to the deployed racing-only pipeline or
  running a real translation migration is out of scope. Human content
  review and operator authoring readiness precede any bilingual
  publish cutover; translation drafts alone are not approved data.

## Scope Out

- Locales beyond `es`/`en`, client-side rendering, and any negotiation
  of `Accept-Language`.
- New endpoints, new gameplay or economic fields, changes to
  `merit_summary`/`completion_reward`/merit values, identifiers, or
  lifecycle states (locale attributes and `content_language` labels are
  the representation concerns of this change, not new mission fields).
- `GET /api/v1/missions/offers/facets` — unaffected.
- Rewriting Spanish content, economics, or version identity of existing
  frozen offer/contract snapshots — the backfill may only enrich
  eligible records with English; it never rewrites or re-versions
  them.

## Decisions

- Language selection is a query parameter, not content negotiation:
  URLs stay self-describing and caches key on URI. Approved at Gate 1
  (2026-10-27), including the explicit decision that legacy
  Spanish-only records fall back to Spanish labeled `es`.
- Honest labeling over availability errors: a missing `en` locale
  serves `es` with `content_language: es`; `406` and falsely labeled
  English are both rejected designs. Approved 2026-10-27.
- Snapshot integrity: frozen offers/contracts capture both locales so
  a v2 translation edit cannot alter an accepted promise — same rule
  as today's economic snapshot (REQ-MB-005/REQ-MCF-009).
- Canonical scope: `mission-board` (REQ-MB-005/013/015/016 + new
  REQ-MB-017) and `mission-contract-fulfillment`
  (REQ-MCF-006/021/022 + new REQ-MCF-026). No new capability —
  localization is a read-model and authoring concern of the existing
  capabilities.
- Publish completeness is a deliberate behavioral change, not additive:
  read paths and draft POST are additive, but requiring English at
  publish breaks the existing es-only draft→publish flow. Existing
  Spanish-only published versions stay valid and servable; new-version
  publishes require `en`. Cutover is staged — updated authoring tooling
  and reviewed `en` content for the catalog land before enforcement —
  see `rollout.md`; the enforcement mechanism is decided at
  implementation with test coverage, not assumed here.

## Risk And Gates — L3

L3 per risk classification: historical snapshots and a data-migration
backfill touch accepted-contract semantics that are critical business
records. Full profile: BDD scenarios, test-first inside-out
(Application RED → demanded Domain → Application GREEN/BLUE → Boundary
RED → Infrastructure/Interface/Composition), per-layer Gate 3,
security review, coverage ≥90%, mutation, critical E2E, rollout/rollback,
and `tools/validate-sdd-change.sh` before final verification.

Initial gate status (2026-10-27): **Gate 1 approved** for the
language contract; Gate 2 and all scoped Gate 3s were pending at
change creation. Later scoped approvals and actual RED/GREEN results
are recorded in `verification.md`; this historical baseline does
not override them.

## Current Status (2026-10-29, context checkpoint)

Core Domain and Application layers passed their layer gates. HTTP
boundary is green under overlay-conditional evidence for three
partitions: definitions 8/8, GetOffer 7/7, ListOffers 8/8 — including
the mixed-content case, which turned green once persistence started
round-tripping `ContentEn`. Persistence (T-0054-007 first slice +
T-0054-008 scoped GREEN) passed Gate 3-PERSISTENCE 2026-10-29:
definitions and offers persist `ContentEn` as a DynamoDB `M`
attribute via `localized_content_mapping.go`; 5/5 bilingual
round-trip tests PASS against real DynamoDB Local and the store
builds clean. Ordinary HTTP build remains blocked by 4 stale
handlers (list_contracts, contract_detail, review_queue,
create_mission_definition) + concurrent CHG-0051 composition
changes — overlay evidence stays conditional. Pending: contract
snapshot persistence, backfill/migration, remaining HTTP partitions,
and L3 gates. Next action: Gate 2 extension for the contract
bilingual snapshot persistence RED only.

## Current Status (2026-11-01)

The GET and bilingual draft-create HTTP partitions pass provisionally
with an overlay that isolates only concurrent CHG-0051 composition;
the publish HTTP behavior passes without changing its handler.
Ordinary composition still does not compile, so the full Boundary
gate remains open. T-0054-009's reviewed-source Spanish amendment
and scoped backfill Gate 3 were approved; a local, synthetic
dry-run/apply rehearsal passes 12 tests plus five subtests against
unique DynamoDB Local tables. No live translations were supplied
or applied and no production migration is authorized. BUG-0054-001
(real unknown-contract detail 500 instead of 404) remains separately
open and excluded from this change. Coverage, mutation, security,
critical E2E, rollout sequencing, and final SDD gates remain open.

Gate 1 amendment approved 2026-11-01 for repository translation
artifacts and safe catalog synchronization only. The current racing
seed runs on each staging/production deployment; it compares full
same-version definition rows and rewrites current rows, so an
independent English backfill would otherwise break the next deploy
or discard `ContentEn`. This amendment adds test-first catalog and
redeploy work, not Gate 2 test writes or Gate 3 production changes.
Translations for the 36 source definitions will be drafted in Git
but require human semantic signoff before any non-local apply.
No AWS, shared-table, staging, or production write is approved.

## Catalog Gate 2 Progress (2026-11-02)

Gate 2 approved for T-0054-013 tests and T-0054-014 Git drafts only.
`TEST-SEED-LANG-001` is RED because the planned
`seed.translated_definitions` API does not exist: three unit tests,
with two catalog subcases, report three AttributeErrors and seven
negative-case failures (all the same missing API rather than fixture
or environment errors). `verification.md` has the exact command.
The first four PvP FPS English drafts are stored in
`seeds/translations/pvp-fps-v4.en.json`; their source Spanish,
version, objective IDs, and rule counts were checked against the
unchanged catalog (4/4; file 116 lines). These are NOT approved
translations; 32 drafts and human review remain. No seed production
code, CI, DynamoDB, or shared resource changed. Next: T-0054-014,
draft the remaining PvP and racing text, then request a separate
Gate 2 for T-0054-015 seed safety tests before any Gate 3-SEED.

## Catalog Draft Progress (2026-11-03)

T-0054-014 drafting is complete: nine version-bound JSON sidecars
cover all 36 source definitions (24 racing, 12 PvP), preserving the
original Spanish and all non-translatable economics/identifiers.
Independent read-only verification checked 36 unique codes/versions,
exact Spanish title/description/ordered rules/objective text and IDs,
complete English structure, and 114-122 lines per sidecar. English
terms were checked against the source, but human operator signoff is
still pending T-0054-017; these drafts are not approved deployment
data. No seeder, CI, AWS, DynamoDB, or published mission changed.
The unit RED remains an absent-loader RED, not a test GREEN. Next:
Gate 2 approval for T-0054-015 seed synchronization RED only.

## Catalog Seed RED (2026-11-03)

Gate 2 for T-0054-015 was approved. TEST-SEED-LANG-002 unit RED
has 10 methods, five failed subcases and five errors (missing/
incomplete `en`, false same-version immutable conflicts, and
actual CLI render omitting `ContentEn`). Against unique DynamoDB
Local tables on port 18000, the final combined integration RED
has 14 tests: 11 failures, one error, two baseline passes; no
setup failures or deleted tables. Dry-run misses offer text/merit/
English drift; apply may overwrite a concurrent English edit, and
local bootstrap stores only Spanish. `verification.md` records
commands, the initial environment-only connection failure, and
production mapping. No seed production script changed. Gate
3-PERSISTENCE/SEED was approved for scripts/loader only; all
operator English drafts remain unreviewed and any real migration
blocked.

## Exact-Offer Read Amendment (2026-11-03)

Gate 1 approved a least-privilege read of seed-owned offers.
Current deployment IAM grants Query/ConditionCheckItem only on
DEFINITION/SEASON and PutItem/UpdateItem on DEFINITION/OFFER; a
dry-run cannot detect offer Spanish, economic, or English drift
without reading the exact existing offer. The plan adds
`dynamodb:GetItem` for `dynamodb:LeadingKeys: [OFFER]` on the same
table/role, not OFFER Query/Scan or CONTRACT access. T-0054-018
will write a static policy RED only after separate Gate 2;
T-0054-019 and `template.yaml` require scoped Gate 3-IAM.
Already-approved seed-script Gate 3 does not authorize changing
IAM or running queries against real tables. Rollout remains blocked
until the policy and local seed tests pass, and English content is
reviewed.

Gate 2 for TEST-SEED-IAM-001 was approved. The standalone policy
test ran three cases on 2026-11-03: two passed baseline/negative
checks and the expected one failed solely for the missing
GetItem-on-OFFER statement; there was no environment or parser
failure. No IAM production file changed. T-0054-018 is complete;
Gate 3-IAM was then approved and only `GetItem` for OFFER on
the existing table/role was added to `template.yaml`. The
static policy test now passes 3/3; no SAM deployment or live
query was performed. T-0054-019 is complete. T-0054-016 seed
script GREEN is next, still local-only and blocked from release
until the remaining rollout gates pass.

## Local Seed GREEN/BLUE (2026-11-03)

T-0054-016 is complete for the offline Python slice under its
approved Gate 3: the seed validates version-bound English,
preserves same-version legacy/bilingual snapshots, compares
seed-owned offers by exact-key `GetItem` before dry-run/apply,
and conditionally rejects both current and previous-version
Spanish, merit, or English drift. New versions and their offers
carry both locales atomically; accepted contracts are untouched.
The main and local CLI paths both render English without adding
PvP to the deployed racing-only pipeline. The old monolithic
seed was split into focused modules of 44-126 lines.

Results: 16/16 seed unit, 8/8 catalog regression, 17/17
isolated DynamoDB Local integration, and 1/1 updated legacy CI
seed integration passed. The legacy test's initial expired
credential result was an environment blocker, not RED; its
old Spanish-only fixture was then the intended new-version RED
before synthetic English fixtures made it GREEN. Tables created
for these tests were left in place, not deleted. No deployed
table, IAM policy, staging, production, backfill, or service
data was touched. At this stage translations were unreviewed drafts;
operator tooling, full Boundary, security, coverage, mutation,
critical E2E, template size exception, and final approval
remain open.

## Local Content Approval (2026-11-03)

The user approved all 36 English texts and selected **local**
as the publication target. They accepted the PK=OFFER-only IAM
read scope, explicitly understanding that `LeadingKeys` cannot
limit SK to seed-owned rows; the CLI itself reads only those
exact keys. No local application table has yet been read or
updated under this request. Its default endpoint/table are
`http://localhost:8000` / `hbk-missions-local`; confirm the
destination before a read-only preflight. The safe seed does
not enrich same-version Spanish-only rows, so inspect the
current state before choosing whether local backfill needs
its own SDD gate. Staging/production and CI remain blocked.

The authorized local read-only preflight reached
`hbk-missions-local` on loopback port 8000. The normal seed
dry-run passed for all 24 racing and 12 PvP definitions with
season `local-training`. Explicit consistent, paginated
PK-scoped reads found 36 current definitions, 36 immutable
version rows, and 36 active-season offers: all 108 lack
`ContentEn`, and CONTRACT COUNT is zero. No write or delete
was performed. Publishing English into these existing rows
requires a separately gated local backfill; the safe seed's
`--apply` deliberately preserves them as Spanish-only.

Gate 1 for the local-only backfill invocation was approved on
2026-11-03. This amends the existing mission-board catalog
and frozen-rule delta without changing the canonical capability
until consolidation. T-0054-020 first proves the offline
invoker's manifest/loopback/preview behavior (Gate 2 pending);
T-0054-021 may implement it only after reviewed RED and
Gate 3; T-0054-022 needs separate authorization for any
`hbk-missions-local` write. The existing Go backfill adapter
remains the conditional update owner. No local table write,
new production code, CI change, or non-local access occurred
under Gate 1.

Gate 2 for T-0054-020 was approved for tests only. Standalone
TEST-SEED-LOCAL-001 unit RED fails to compile solely for the
absent manifest reader and loopback validator (137-line file).
Standalone TEST-SEED-LOCAL-002 integration RED fails five of
five cases solely because the Go invoker does not exist
(128-line test file). The latter created fresh private tables
on the dedicated port 18000 and left them intact; there were
no setup errors or app-table accesses. The actual RED was
reviewed; scoped Gate 3 for T-0054-021 is requested but
not approved. No production file changed.

Gate 3-LOCAL-BACKFILL was approved 2026-11-03 for the
local-only reader/CLI. T-0054-021 GREEN/BLUE: four unit
methods, five real private-table CLI tests, and 12 Go
backfill tests plus five drift subtests all pass.
`reviewed_translation_manifest.go` (99 lines) and the
two CLI files (35/95 lines) reuse the existing conditional
backfill. Only private test tables on port 18000 received
test writes; `hbk-missions-local` on port 8000 has not
been read by this CLI or written since the earlier
preflight. T-0054-022 begins with a fresh read-only
preview and requires separate exact-table apply approval.

The 2026-11-04 read-only local deployment review found
108 unique eligible source-matching rows, zero unresolved,
zero contracts and zero English attributes on the 72
definitions and 36 offers. The gateway's missions routes
return 502 because port 3006 has no listener. The current
Go service does not compile due to CHG-0051 composition
calling `NewPlacementFactPublisher` with three arguments
when it now takes five. Local data and HTTP are not
fully deployed. No app-table write, restart, Docker
mutation or AWS call was made.

After that review the user explicitly authorized the
conditional local-table apply. A fresh preview checked
all 108 identities against the captured Spanish-only
state and found zero unresolved. `--apply` added only
`ContentEn` to 72 DEFINITION and 36 OFFER rows.
Strongly consistent before/after reads showed the
approved English exactly, every other field unchanged,
and the season and zero contracts unchanged. The
post-apply preview reports zero eligible, and both
catalogs still pass a read-only seed dry-run. T-0054-022
is complete for local data only; the gateway continues
to return 502 until CHG-0051 composition is fixed and
the backend is started under a separate approved plan.
