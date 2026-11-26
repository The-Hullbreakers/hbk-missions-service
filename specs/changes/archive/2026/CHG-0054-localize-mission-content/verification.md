---
change_id: CHG-0054
artifact: verification
status: pending
updated: 2026-11-01
---

# Verification — CHG-0054

## Gates

| Gate | Scope | Decision | Evidence | Date |
| --- | --- | --- | --- | --- |
| Gate 1 | spec writes (L3, both canonical specs, labeled es fallback) | approved | user approval 2026-10-27 | 2026-10-27 |
| Gate 2 | start RED — scoped to T-0054-001 first partition only | approved | user approval 2026-10-27; does not authorize production or further partitions | 2026-10-27 |
| Gate 3-APPLICATION | T-0054-001 GetOffer partition only | approved (scoped) | user approval 2026-10-27; authorizes the demanded Domain RED below only — no production, no other partitions | 2026-10-27 |
| Gate 3-APPLICATION (second scope) | T-0054-001 whole-offer integrity GetOffer partition | approved (scoped) | user approval 2026-10-27; authorizes demanded Domain RED for English rules/objectives snapshot only — no production | 2026-10-27 |
| Gate 3-APPLICATION (third scope) | T-0054-005 GetOffer whole-offer integrity GREEN/BLUE | approved (scoped) | user approval 2026-10-27; authorizes `LocalizeOffer` atomic-completeness + ID-matched substitution in `localized_offer.go` only — no other use cases, no boundary | 2026-10-27 |
| Gate 3-APPLICATION (fourth scope) | T-0054-005 ListOffers GREEN/BLUE | approved (scoped) | user approval 2026-10-27; authorizes `LocalizedOfferListPage`, 3-arg `Handle`, `filter.Q` strip + effective-language Match before sort/pagination, and signature alignment of existing list tests — no other use cases, no boundary | 2026-10-27 |
| Gate 3-APPLICATION (fifth scope) | T-0054-005 ListDefinitions GREEN/BLUE | approved (scoped) | user approval 2026-10-27; authorizes `LocalizedDefinition`, `Handle(language)`, shared `localizeMissionText` extraction refactor, and existing test alignment — no other use cases, no boundary | 2026-10-27 |
| Gate 3-DOMAIN | T-0054-003 GetOffer-demanded partition | approved (scoped) | user approval 2026-10-27 after re-review of restructured `language_test.go`; authorizes GREEN for the mapped symbols only | 2026-10-27 |
| Gate 3-DOMAIN (second scope) | T-0054-003 English terms freeze/non-alias partition | approved (scoped) | user approval 2026-10-27; authorizes GREEN/BLUE for `LocalizedObjective` + `LocalizedContent.Rules/Objectives` + deep-copy freeze only | 2026-10-27 |
| Gate 3-DOMAIN (third scope) | T-0054-003 `ContractTerms.WithObjectiveText` partition | approved (scoped) | user approval 2026-10-27; authorizes the method in `contract_terms_localized.go` only | 2026-10-27 |
| Gate 3-APPLICATION (sixth scope) | T-0054-005 ListPlayerContracts GREEN/BLUE | approved (scoped) | user approval 2026-10-27; authorizes `LocalizedContract`, `LocalizeContract` (Terms-effective objectives + `WithObjectiveText` snapshot), 3-arg `Handle`, and existing test alignment — no boundary | 2026-10-27 |
| Gate 3-APPLICATION (seventh scope) | T-0054-005 GetContract GREEN/BLUE | approved (scoped) | user approval 2026-10-27; authorizes 3-arg `Handle` + `LocalizeContract` wrap post-authorization and existing test alignment — no boundary | 2026-10-27 |
| Gate 3-APPLICATION (authoring create) | T-0054-002 bilingual draft creation GREEN/BLUE | approved (scoped) | user approval 2026-10-29; create request with optional deep-copied ContentEn, legacy tests aligned; no HTTP/persistence | 2026-10-29 |
| LAYER-GATE-DOMAIN | domain complete for CHG-0054 core | passed | focused language/terms/publish 14/14 and full Domain package PASS 2026-10-27; no HTTP/persistence; final coverage/mutation remain pending | 2026-10-29 |
| LAYER-GATE-APPLICATION | core complete for CHG-0054 | passed | GetOffer, ListOffers, ListDefinitions, ListPlayerContracts, GetContract, ListPendingReview + bilingual authoring focused GREEN/BLUE; full Application package PASS 2026-10-29; HTTP and persistence compile/integration deferred to Boundary | 2026-10-29 |
| Gate 2 (boundary RED) | T-0054-006 first partition — definitions lang | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 2 (boundary RED extension) | T-0054-006 GetOffer partition lang | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 2 (boundary RED extension) | T-0054-006 ListOffers partition lang | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 2 (persistence RED) | T-0054-007 bilingual definition+offer snapshots, legacy absence | approved | user approval 2026-10-29; RED/tests only, no production — no adapter/backfill authorized | 2026-10-29 |
| Gate 3-PERSISTENCE | T-0054-008 definition/offer ContentEn DynamoDB M round-trip only | approved (scoped) | user approval 2026-10-29; authorizes `localized_content_mapping.go` + 4 store files only — contracts/backfill/migration excluded | 2026-10-29 |
| Gate 2 (persistence RED extension) | T-0054-007 contract bilingual snapshot | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 3-PERSISTENCE (contract snapshot) | T-0054-008 contract ContentEn M round-trip only | approved (scoped) | user approval 2026-10-29; authorizes `contractItem` + `mapToContract` hunk only — backfill/migration excluded | 2026-10-29 |
| Gate 2 (boundary RED extension) | T-0054-006 ListContracts partition lang | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 3-BOUNDARY (ListContracts only) | T-0054-006 list-contracts partition provisional GREEN/BLUE | approved (scoped, conditional) | user approval 2026-10-29; authorizes `list_contracts.go` + new `list_contracts_response.go` only — evidence overlay-conditional | 2026-10-29 |
| Gate 3-BOUNDARY (ContractDetail only) | T-0054-006 contract-detail partition provisional GREEN/BLUE | approved (scoped, conditional) | user approval 2026-10-29; authorizes `contract_detail.go` + `contract_detail_projection.go` only — evidence overlay-conditional | 2026-10-29 |
| Gate 2 (boundary RED extension) | T-0054-006 ReviewQueue partition lang | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 3-BOUNDARY (ReviewQueue only) | T-0054-006 review-queue partition provisional GREEN/BLUE | approved (scoped, conditional) | user approval 2026-10-29; authorizes `review_queue.go` only — evidence overlay-conditional | 2026-10-29 |
| Gate 2 (boundary RED extension) | T-0054-006 Create definition bilingual draft partition | approved | user approval 2026-11-01; RED/tests only, no production | 2026-11-01 |
| Gate 3-BOUNDARY (Create definition only) | T-0054-006 create bilingual draft provisional GREEN/BLUE | approved (scoped, conditional) | user approval 2026-11-01; authorizes `create_mission_definition.go` + new `create_mission_definition_translations.go` only — evidence overlay-conditional | 2026-11-01 |
| Gate 2 (boundary RED extension) | T-0054-006 Publish definition HTTP behavior | approved | user approval 2026-11-01; tests only — behavior may already be green | 2026-11-01 |
| Gate 1 amendment + Gate 2 (backfill RED extension) | T-0054-009 reviewed source Spanish snapshot match | approved (tests only) | user approvals 2026-11-01; version rows can be overwritten by `DefinitionStore.Save`; compare source es text even for same-version frozen records; new DynamoDB Local tables only, no production or live migration | 2026-11-01 |
| Gate 3-PERSISTENCE (backfill rehearsal only) | T-0054-009 offline adapter for reviewed translation | approved (scoped) | user approval 2026-11-01 after TEST-INF-LANG-002 compile RED; authorizes only a new backfill adapter with dry-run, unresolved report, conditional ContentEn-only updates, and GREEN on isolated DynamoDB Local tables; no CLI, live migration, or CHG-0051 edits | 2026-11-01 |
| Gate 1 amendment (catalog + safe redeploy) | T-0054-013..017 source-bound drafts and release order | approved (artifacts only) | user approval 2026-11-01; 24 racing + 12 PvP versioned English drafts in Git, preserve same-version Spanish/economics/ContentEn on repeated seed and require English on new versions; no Gate 2/3 for tests/code, no human content signoff or non-local data operations implied | 2026-11-01 |
| Gate 2 (catalog RED and draft data) | T-0054-013..014 source-bound manifest tests and English drafts | approved (scoped) | user approval 2026-11-01; write/review TEST-SEED-LANG-001 RED first, then draft 36 English records under `seeds/translations/` in Git; no seed-script production edit, CLI/CI change, shared-table query, or data migration | 2026-11-01 |
| Gate 2 (seed synchronization RED) | T-0054-015 — TEST-SEED-LANG-002 safe redeploy and new versions | approved (tests only) | user approval 2026-11-03; Python unit tests and integration tests against new loopback DynamoDB Local tables only; no seeder production edits, real deployment, live migration, or deletion of existing tables | 2026-11-03 |
| Gate 3-PERSISTENCE/SEED | T-0054-013/015 RED → T-0054-016 script GREEN | approved (scoped, IAM excluded) | user approval 2026-11-03; `scripts/seed_missions.py`, `scripts/seed_local_missions.py`, and a focused translation loader only; local verification, no deploy/live data/IAM/CI change | 2026-11-03 |
| Gate 1 amendment (exact OFFER read) | T-0054-018/019 least-privilege IAM plan | approved (artifacts only) | user approval 2026-11-03; plan read-only `dynamodb:GetItem` conditioned to `dynamodb:LeadingKeys: [OFFER]` on existing table/role; separate Gate 2 test RED and Gate 3-IAM needed before editing `template.yaml` | 2026-11-03 |
| Gate 2 (IAM policy RED) | T-0054-018 — TEST-SEED-IAM-001 static policy scope | approved (tests only) | user approval 2026-11-03; static read of `template.yaml` under `tests/unit/seeds/test_seed_iam_policy.py` only; no IAM/template edit, SAM deploy, or AWS request | 2026-11-03 |
| Gate 3-IAM | T-0054-018 RED → T-0054-019 exact OFFER GetItem statement | approved (scoped) | user approval 2026-11-03; one `GetItem` statement in `template.yaml` for OFFER keys only, static GREEN 3/3; no SAM deploy, live read, broader grant or other policy edit | 2026-11-03 |
| Content signoff + IAM-scope decision | T-0054-017 — approve English catalog wording for local use | approved (local content only) | user confirmed all 36 English translations and chose local as publication target 2026-11-03; accepted that the role can `GetItem` any row with PK=OFFER (IAM cannot limit SK) while the script reads only seed-owned SK; no local table query/apply, non-local release, or backfill executed by this decision | 2026-11-03 |
| Local read-only preflight | T-0054-017 — inspect `hbk-missions-local` before local publication | approved (reads only) | user confirmed `http://localhost:8000` / `hbk-missions-local` on 2026-11-03 for consultation/dry-run; no apply/delete/backfill approved by this confirmation | 2026-11-03 |
| Gate 1 (local reviewed backfill invocation) | T-0054-020..022 — publish approved English locally | approved (artifacts only) | user approval 2026-11-03 after read-only preflight: reuse Go offline backfiller via loopback-only invoker and version/source-bound sidecars; Gate 2 test RED, Gate 3 invoker GREEN/BLUE, and separate exact-table apply approval still required; no local table write by this gate | 2026-11-03 |
| Gate 2 (local invoker RED) | T-0054-020 — TEST-SEED-LOCAL-001/002 | approved (tests only) | user approval 2026-11-03; unit manifest and private unique-table CLI RED only; no production edit, application-table apply/delete, or remote endpoint access | 2026-11-03 |
| Gate 3-LOCAL-BACKFILL | T-0054-020 RED → T-0054-021 manifest reader + local invoker | approved (scoped) | user approval 2026-11-03; manifest reader/loopback validator and CLI only, GREEN/BLUE on isolated tables; no `hbk-missions-local` write, CI/deploy or remote access | 2026-11-03 |
| `python3 -m unittest discover -s tests/unit/seeds -p 'test_mission_translations.py' -v` | T-0054-013 — TEST-SEED-LANG-001 catalog translation unit RED | FAIL (intended API absence) | 2026-11-02: 3 top-level tests; two catalog subcases and production racing case ERROR `AttributeError: seed_missions has no attribute translated_definitions`; seven invalid-manifest cases FAIL because the same missing function returns `AttributeError` rather than expected `ValueError`. No fixture/import/environment error; no DB touched. Tests assert 24 racing + 12 PvP with exact Spanish/economics unchanged, eight production racing IDs, stale/missing/duplicate/invalid translations rejected. Test-only files `tests/unit/seeds/translation_fixture.py` (60 lines) and `test_mission_translations.py` (90 lines). This RED proves the planned catalog translation seam is absent, not that the actual English drafts are complete; human translation review remains separate. No seed production file changed; Gate 3-SEED not requested. |
| Gate 3-BOUNDARY (definitions GET only) | T-0054-006 definitions partition provisional GREEN/BLUE | approved (scoped, conditional) | user approval 2026-10-29; authorizes `list_definitions.go` + new `mission_language_query.go` only — evidence remains overlay-conditional until the other five stale handlers are repaired in later partitions | 2026-10-29 |
| Gate 3-BOUNDARY (GetOffer only) | T-0054-006 get-offer partition provisional GREEN/BLUE | approved (scoped, conditional) | user approval 2026-10-29; authorizes `get_offer.go` + new `get_offer_response.go` DTO only — evidence remains overlay-conditional until remaining stale handlers are repaired | 2026-10-29 |
| Gate 3-BOUNDARY (ListOffers HTTP subset) | T-0054-006 list-offers partition partial GREEN | approved (scoped, conditional) | user approval 2026-10-29; authorizes `list_offers.go` + new `list_offers_response.go` only — mixed-content case remains RED pending T-0054-007 persistence ContentEn mapping | 2026-10-29 |
| Gate 2 (boundary RED extension) | T-0054-006 ContractDetail partition lang | approved | user approval 2026-10-29; RED/tests only, no production | 2026-10-29 |
| Gate 3-BOUNDARY | T-0054-006/007 boundary RED | pending | | |

## Catalog Translation Draft Check (2026-11-03)

T-0054-014's nine files under `seeds/translations/` contain exactly
36 distinct version-4 records (24 racing, 12 PvP). A read-only
`python3 -B -` check used `seed.definitions` to compare every
`SourceSpanish` title, description, ordered rules, and ordered
objective id/name/description against its unchanged source row,
including each version; required English title/description/rules
and objective names/descriptions are nonblank with matching
cardinality and IDs. All JSON parsed, no duplicate records or
non-translatable economics fields were present, and each JSON file
has 114-122 lines (<150). English titles/descriptions for all 36
and every distinct Spanish-to-English rule/objective pair were
read for consistency; no conflicting translations for the same
Spanish text were found. This is independent draft-data
inspection, NOT `TEST-SEED-LANG-001` GREEN: the production
`translated_definitions` loader still does not exist. English
semantic approval is pending the operator, and neither a seed nor
an AWS/DynamoDB operation was run.

## TEST-SEED-LANG-002 — Safe Redeploy RED (2026-11-03)

Gate 2 approved for tests only, risk L3, scope: offline seed adapter
and its two operator CLIs; Go Domain/Application behavior and HTTP
boundary are `not_affected` by this slice (their existing CHG-0054
layer gates remain independent). SCN-MB-CATALOG-REDEPLOY and
SCN-MB-CATALOG-NEW-VERSION map through REQ-MB-005/013 and
US-MB-CATALOG-002 to T-0054-015, then to the candidate
`scripts/seed_missions.py`, `scripts/seed_local_missions.py`, and a
small translation-loader module. The seed CLI is an offline
DynamoDB adapter, not an Application use case or an HTTP resource;
the Python unit test is the lowest stable logic seam, and the
infrastructure integration tests call the actual seed/CLI with a
real, owned DynamoDB Local table. No new test suite or live query.

- `python3 -B -m unittest discover -s tests/unit/seeds -p 'test_seed_translation*.py' -v`
  exits 1: **10 test methods**, five failed subcases (two missing
  English first/new-version, three incomplete English), five errors
  (two unexpected immutable-version rejections; two racing CLI
  subcases and explicit PvP CLI missing `ContentEn`). Four methods
  have passing outcomes; subtest failures/errors make the numeric
  totals larger than the number of failing methods. The render
  tests call the CLI without AWS credentials or network access.
- `DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_seed_translation*.py' -v`
  exits 1: **14 tests, 11 FAIL, one ERROR, two PASS**. Intended RED
  partitions: translated same-version legacy/new English and
  old-catalog redeploy reject harmless no-ops; same-version offer
  Spanish, merit, or English drift is accepted or overwritten;
  missing/incomplete English new versions are written; a
  translation authored between read and transaction is overwritten;
  racing/PvP local CLI bootstrap saves Spanish only. The two PASS
  cases characterize bilingual v5 offer advancement (accepted
  contracts unchanged) and idempotent bilingual redeploy.
- Each integration test creates a distinct uniquely named table
  on the **dedicated loopback port 18000**, sets dummy credentials,
  and leaves the table in place; tests make no shared/AWS call and
  perform no deletes. The dedicated emulator was initially down:
  that first connection failure was an environment/setup blocker,
  **not RED**. After it was restarted, the two combined commands
  above reached all tests with no setup errors; no other service
  container on port 8000 was touched.
- New test-only files: `tests/unit/seeds/test_seed_translation_actions.py`
  (132 lines), `test_seed_translation_cli.py` (62 lines),
  `tests/integration/infrastructure/dynamodb/seed_translation_fixture.py`
  (88 lines), `test_seed_translation.py` (144 lines),
  `test_seed_translation_dry.py` (79 lines),
  `test_seed_translation_race.py` (45 lines), and
  `test_seed_translation_local_cli.py` (42 lines). All use fresh
  fixtures, one-line Act calls and assertions in Assert. The race
  test injects only one concurrent write to its own local offer
  before the seed transaction; no third-party API is mocked.
  Actual old `seed_missions.py` and `seed_local_missions.py` were
  **not edited**. `TEST-SEED-LANG-001` remains independently RED
  on the absent loader; this is not a full suite or final L3 gate.

Gate 3-PERSISTENCE/SEED **approved only for scripts/loader**. Minimum
GREEN: load and validate version/source-bound English before any
write, preserve existing same-version `ContentEn`, reject mismatched
Spanish/economics/English in definitions and offers during dry-run
and conditional apply, atomically advance offers with complete
bilingual version changes, and route both actual seed CLIs through
the loader. BLUE: keep seed module(s)/tests <150 physical lines
and rerun the same focused commands; no Go domain value object,
new AWS/IAM permission, CI workflow change, real migration, or
deployed table action. Human content approval remains pending.

## TEST-SEED-IAM-001 — Exact OFFER Read RED (2026-11-03)

Gate 2 approved a static Python unit test only.
`python3 -B -m unittest discover -s tests/unit/seeds -p 'test_seed_iam_policy.py' -v`
ran **3 tests; 2 PASS, 1 FAIL**, exit 1. The failing
`test_get_item_only_for_offers_on_existing_table` compares all
statements in `MissionCatalogDeploymentPolicy` and sees exactly
one missing statement: `dynamodb:GetItem` for the existing
MissionsTable ARN with `ForAllValues:StringEquals` and
`dynamodb:LeadingKeys: [OFFER]`. The other two tests verify that
the current role's read and write scopes match the approved
baseline and that synthetic Query/Scan, contract reads, expanded
leading keys, and extra writes do not equal the approved policy.
No parser, fixture, or environment errors; the test uses only
Python's standard library, is 112 physical lines, and makes no
AWS/localhost request. `template.yaml` was not modified during
RED. Gate 3-IAM was approved afterwards, and exactly one
`Allow dynamodb:GetItem` statement was added for the existing
MissionsTable ARN with `dynamodb:LeadingKeys: [OFFER]`; the old
statements and role were unchanged. Re-running the same static
command passed **3/3**, exit 0; no AWS/network/deployment or
DynamoDB Local run. The pre-existing 311-line template is now
318 lines; its size remains a documented BLUE exception pending
a separate controlled cleanup decision. T-0054-016 script GREEN
is still pending.

## TEST-SEED-LANG-001/002 — Local GREEN/BLUE (2026-11-03)

The previously recorded catalog and seed REDs turned GREEN after
the scoped Gate 3-SEED. The seed CLI and local bootstrap now load
version/source-bound English from `seeds/translations/`; new
versions require complete English. Same-version current,
immutable-version, and offer rows keep their existing English
or intentional Spanish-only legacy state. Each seed-owned offer
is read via strongly consistent `GetItem` with its exact PK/SK;
dry-run rejects Spanish, English, and merit drift before writes,
including prior-version offer drift before advancing a version.
`UpdateItem` guards the entire observed offer (including absence
of `ContentEn`), so an intervening authored English edit aborts
the transaction. Contracts are not written.

Additional REDs reviewed during BLUE: two prior-version offer
tests produced four intended failures (Spanish title, English
title, merit, and apply); a new local CLI legacy redeploy case
failed on an incorrect post-write expectation. After corrections,
each focused partition passed 2/2 against unique DynamoDB Local
tables. Final focused independent results:

- `python3 -B -m unittest discover -s tests/unit/seeds -p 'test_*.py' -v`
  **16/16 PASS**, including catalog, seed, and static IAM policy.
- `python3 -B -m unittest discover -s tests/seeds -p 'test_seed_missions.py' -v`
  **8/8 PASS**; Spanish catalog identity/economics unchanged.
- `DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_seed_translation*.py' -v`
  **17/17 PASS** (dedicated loopback DynamoDB Local, fresh UUID
  tables left in place, no shared AWS).
- The legacy CI seed integration test originally had an expired
  credential error (**environment, not RED**). With dummy local
  credentials it reached the expected RED: old Spanish-only v4
  fixture could not create a new version. It was aligned with
  version-bound v4 English and synthetic v5/v6 English and its
  `delete-table` teardown removed. Standalone
  `DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/seeds -p 'test_seed_integration.py' -v`
  then **1/1 PASS** (new table retained).

BLUE inventory: all eight maintained seed scripts are 44-126
physical lines, old integration test 127; `seed_catalog` keeps
Spanish/economics, `seed_translations` validates source/English,
`seed_actions` owns conditional semantics, and `seed_aws` owns
the AWS CLI adapter. No new Domain/Application Value Object or
service-level port is needed for the offline seed. No Go production
file changed; separate Go/Boundary and L3 gates are not claimed
from these Python runs. `template.yaml` remains an existing
318-line cleanup exception. No SAM deploy, real seed, backfill,
human content approval, or AWS/staging/production read/write
occurred. T-0054-016 is complete for the offline Python slice;
rollout and final gates remain pending.

## Local Application Preflight — Read Only (2026-11-03)

The user confirmed consultation/dry-run only against
`http://localhost:8000` / `hbk-missions-local`. With dummy local
credentials, `python3 -B scripts/seed_local_missions.py --endpoint-url http://localhost:8000 --table hbk-missions-local`
returned `applied: false`, season `local-training`, and both
catalogs synchronized in dry-run (24 racing and 12 PvP;
no conflicts). Subsequent strongly consistent, paginated
`PK = :pk` queries against DEFINITION and OFFER classified the
36 selected codes: all **36 current definitions**, **36 immutable
version rows**, and **36 active-season offers** are Spanish-only;
none is bilingual, missing, or on another version. A paginated
`PK = :pk` COUNT query for CONTRACT returned **0**. The table
has 72 DEFINITION rows and 36 OFFER rows in total.

This is a point-in-time observation, not an authorization to
write. Running the ordinary seed's `--apply` would preserve
all 108 legacy Spanish-only rows and **would not publish the
approved English**. A local-only, source-checked, conditional
ContentEn backfill needs a scoped implementation and explicit
apply approval; no table write/delete, AWS request, or
staging/production operation was made.

## TEST-SEED-LOCAL-001 — Manifest/Loopback Unit RED (2026-11-03)

Under Gate 2, only `tests/unit/missions/infrastructure/dynamodb/reviewed_manifest_test.go`
(137 lines) was added. Standalone
`go test -count=1 ./tests/unit/missions/infrastructure/dynamodb -run '^TestReviewedTranslationManifest' -v`
stopped at the intended compile RED: `LoadReviewedTranslationManifests`
(three call sites) and `ValidateLocalBackfillEndpoint` (two call
sites) are undefined. No Go import, fixture, or environment error
was reported. Tests demand source/version-bound content, duplicate
and invalid/economic-field rejection, and a pure loopback URL
guard. At that point TEST-SEED-LOCAL-002 integration RED was
still pending. This is **not** Gate 3 or permission to modify production; the local
application table was not queried or changed in this test.

## TEST-SEED-LOCAL-002 — Local CLI Integration RED (2026-11-03)

Standalone
`DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_local_translation_publish.py' -v`
ran five tests, **5 FAIL**, exit 1. All five fail for the same
intended missing boundary: `stat cmd/local-translation-backfill:
directory not found`. The 128-line test verifies preview has
three eligible source-matching rows and zero writes, apply on a
fresh table adds only `ContentEn`, a divergent offer blocks
every pre-apply update, rerun is idempotent, and the CLI rejects
a non-HTTP endpoint. Each scenario created its own UUID table
on dedicated loopback DynamoDB Local port 18000 and left it
in place; no environment/fixture failure, user-app access,
production code change, AWS call, or deletion occurred.

Reviewed Gate 3-LOCAL-BACKFILL evidence: TEST-SEED-LOCAL-001
unit compile RED is limited to two absent Go API functions;
TEST-SEED-LOCAL-002 boundary integration RED is limited to
the missing CLI. Tests are standalone with independent state,
use an existing real offline infrastructure adapter for
publication, and do not change Domain/Application/public
HTTP behavior (`not_affected`). No manual QA substitutes for
these tests. GREEN maps only to a focused manifest reader
and `cmd/local-translation-backfill/`; BLUE keeps files
below 150 lines and reruns both partitions plus existing
backfill regressions. This review **does not approve** a
production edit or an apply to the user's local table.

## TEST-SEED-LOCAL-001/002 — Local Invoker GREEN/BLUE (2026-11-03)

Gate 3 approved only `reviewed_translation_manifest.go` and
`cmd/local-translation-backfill/`, not an app-table apply.
The reader rejects empty/duplicate/malformed/economic-field
records; `ValidateLocalBackfillEndpoint` rejects remote or
non-HTTP targets before creating the client. The CLI loads all
36 reviewed records before connecting, defaults to no-write
preview, and on `--apply` repeats a full preview, refuses
unresolved records, and reuses the existing per-item
`TranslationBackfiller.Rehearse` conditional `ContentEn` update.
Concurrent conflicts report partial success without claiming
global transactionality.

Standalone `go test -count=1 ./tests/unit/missions/infrastructure/dynamodb -run '^TestReviewedTranslationManifest' -v`
passed all four methods and their cases. Standalone
`DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_local_translation_publish.py' -v`
passed **5/5**: no-write preview, English-only apply on three
owned rows, unresolved pre-apply refusal, idempotent rerun,
and invalid-endpoint refusal. Existing
`DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestBackfill' -v`
passed **12 top-level + 5 drift subtests** with no build or
environment failure. Every integration test used fresh unique
loopback tables and left them in place.

BLUE: manifest validation (99 lines), CLI input/output (35
lines), CLI orchestration (95 lines), Go unit test (137 lines),
and Python integration test (128 lines) are each below 150;
gofmt applied. The existing adapter remains the sole
conditional update owner; no Domain/Application or public HTTP
change and no new value object is warranted for an offline
invoker. The L3 final suite remains pending. The repository's
`tools/validate-sdd-change.sh` was not run: inspection shows it
calls `docker compose down --remove-orphans`, which would
destroy existing local containers without specific approval.
No app-table apply, deletion, AWS call, or deployment occurred.

## Local Deployment Review — Read Only (2026-11-04)

The local-only CLI ran against the previously confirmed
`http://localhost:8000` / `hbk-missions-local` with `--dry-run`:
`applied: false`, **108 eligible unique rows** (36 current
definitions, 36 version rows, 36 offers), **zero unresolved**
and zero contracts. A separate partition-key-scoped inspection
found one active `local-training` season, 72 DEFINITION and 36
OFFER rows at version `4`, zero CONTRACT rows, and **zero
`ContentEn`** fields on definitions or offers. The reviewed
English is still not published to this local application table.

`docker ps` showed the gateway and both DynamoDB Local containers
running but no missions-service container. The gateway's local
configuration points `/api/v1/missions/` at
`host.docker.internal:3006`; no process is listening on port
3006. Read-only GETs to `/api/v1/missions/healthz`,
`/api/v1/missions/readyz`, `/api/v1/missions/offers?lang=en`,
and `/api/v1/missions/definitions?lang=en` via port 8080 all
returned **502**. No authenticated locale response could be
verified. An offline, network-disabled compile-only check
`GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -run '^$' ./cmd/hbk-missions-service`
failed at `internal/missions/composition/server.go:80`:
`NewPlacementFactPublisher` is called with 3 arguments but
currently requires 5. This is the independently tracked
CHG-0051 composition blocker, not evidence of a translation
behavior failure. The service cannot build from this tree;
the missing listener explains the gateway's 502 response.

Conclusion: **local missions are not fully deployed**.
Catalog data exists in Spanish only, and the HTTP missions
backend is unavailable. The CLI was used in read-only mode;
no `--apply`, restart, deployment, data mutation, table
deletion, or AWS request was performed. T-0054-022 needs
explicit exact-table apply approval; restoring the backend
also needs the separate CHG-0051 composition owner and a
reviewed local startup plan.

## Local English Publication — Approved Apply (2026-11-04)

The user explicitly approved `--apply` only for
`http://localhost:8000` / `hbk-missions-local`, with a fresh
preview and before/after verification. The source-bound
catalog unit suite passed **3/3** immediately before apply.
Strongly consistent, paginated partition reads captured
72 DEFINITION, 36 OFFER, zero CONTRACT, and one active
`local-training` SEASON row. All 108 mission rows were
Spanish-only. The fresh CLI `--dry-run` returned
`applied: false`, **108 unique eligible identities**
(36 current definitions, 36 version rows, 36 offers),
zero updated and zero unresolved; its identities matched
the captured pre-apply rows exactly.

The approved local CLI `--apply` exited 0 and reported
**108 updated unique IDs** (72 DEFINITION, 36 OFFER),
zero unresolved. A strongly consistent post-apply read
confirmed the same 108 keys, and a per-item comparison
showed **only `ContentEn` changed**. Every new English
attribute exactly matches the corresponding one of the
36 reviewed version-bound translations; all Spanish,
IDs, merit/economic fields, objectives, version fields,
season, and zero contract records are unchanged.
There was no partial failure or automatic retry.

An independent read-only CLI preview afterwards reported
zero eligible/updated/unresolved and `applied: false`.
The ordinary seed dry-run also reported both catalogs
`synchronized` (`24/24` racing and `12/12` PvP,
`applied: false`) in `local-training`. No table was
deleted and no AWS/shared-table call, CI deploy, or
staging/production write was made. The HTTP backend
remains unavailable: gateway missions health still
returns 502 and direct port 3006 has no listener.
**Local data publication is complete; full local API
deployment and `?lang=en` HTTP verification remain blocked
by the separately owned CHG-0051 composition issue.**

## Commands

| Command | Scope | Result | Notes |
| --- | --- | --- | --- |
| `go test -count=1 localized_content_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — GetOffer partition (TEST-APP-LANG-001 subset) | RED (build fail) | 2026-10-27: intended missing shape only — `domain.LanguageEnglish`/`LanguageSpanish`, `domain.LocalizedContent`, `MissionOffer.ContentEn`, `Handle(offerID, language)` arity, result `ContentLanguage`; production files unchanged |
| `go test -count=1 -v localized_content_test.go get_offer_test.go` (cwd `tests/unit/missions/application`) | T-0054-005 Application GREEN — GetOffer partition | RED→GREEN | 2026-10-27 RED re-confirm: `Handle` arity + `served.ContentLanguage` only; GREEN after `localized_offer.go`/`get_offer.go` (PASS 4/4); `get_offer_test.go` signature-aligned, assertions unchanged |
| `go test -count=1 localized_content_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — whole-offer integrity partition (Gate 2 extension 2026-10-27) | FAIL (build, intended) | 2026-10-27: `unknown field Rules/Objectives in domain.LocalizedContent`, `undefined: domain.LocalizedObjective` (lines 78-79) — demanded atomic-fallback shape absent; zero production edits |
| `go test -count=1 localized_offer_terms_test.go` (cwd `tests/unit/missions/domain`) | T-0054-003 Domain RED — English terms freeze/non-alias partition (Gate 3-APPLICATION second scope 2026-10-27) | FAIL (build, intended) | 2026-10-27: `unknown field Rules/Objectives in domain.LocalizedContent`, `undefined: domain.LocalizedObjective` (lines 26-27, 40-44, 69-70) — zero production edits |
| `go test -count=1 -v localized_offer_terms_test.go language_test.go` (cwd `tests/unit/missions/domain`) | T-0054-004 Domain GREEN — English terms freeze/non-alias partition (second Gate 3-DOMAIN) | RED→GREEN | 2026-10-27: PASS 7/7 after `LocalizedObjective` + `LocalizedContent.Rules/Objectives` + `Clone()`; deep-copy freeze via `definition.ContentEn.Clone()` |
| `go test -count=1 ./tests/unit/missions/domain` | T-0054-004 Domain GREEN — whole package check | PASS | 2026-10-27: package green including parallel CHG-0051 domain work; `go build ./internal/missions/domain` OK; gofmt clean |
| `go test -count=1 -v localized_content_test.go` (cwd `tests/unit/missions/application`) | Application whole-offer integrity — post-domain-GREEN observation | FAIL (runtime, intended) | 2026-10-27 observed: 2 legacy tests PASS; `ServesEnglishRulesAndObjectives` fails — rules "Regla uno"≠"First rule", objective "Capturar"≠"Capture" (en label + title/description already correct); `FallsBackToSpanishWhenEnglishIncomplete` fails — mixed locale served ("Hunt"/"English briefing", label en≠es) |
| `go test -count=1 -v localized_content_test.go localized_content_integrity_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — integrity matrix partition (Gate 2 extended) | FAIL (runtime, intended) | 2026-10-27 observed: all 4 `IncompleteEnglishFallsBackAtomically` subtests fail — every defect (rules count, blank rule, blank objective description, unmatched objective ID) still serves mixed en title/description + label en; `MatchesEnglishObjectivesByID` fails — objectives rendered Spanish, order/IDs/economics correct; correctly classified: all failures trace to `LocalizeOffer` completeness check + rules/objective substitution; zero production edits |
| `go test -count=1 -v localized_content_test.go localized_content_integrity_test.go localized_content_projection_test.go get_offer_test.go` (cwd `tests/unit/missions/application`) | T-0054-005 Application GREEN — GetOffer whole-offer integrity (Gate 3-APPLICATION third scope) | RED→GREEN | 2026-10-27: RED observed 8 FAIL/4 PASS incl. new missing-name, duplicate-ID, whitespace, no-alias tests; GREEN after `LocalizeOffer` atomic completeness + ID-matched substitution + copied slices — PASS 12/12 |
| `go test -count=1 localized_list_offers_test.go localized_list_offers_page_test.go list_offers_doubles_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — ListOffers partition (Gate 2 extended) | FAIL (build, intended) | 2026-10-27: `Handle` 3-arg arity (3 sites) + `page.Offers[i].ContentLanguage` undefined on `domain.MissionOffer` (3 sites); `LocalizedOfferListPage` + `filter.Q` strip + q-before-pagination demanded; reworked asserts en rules/objective text + economics and adds `localized_list_offers_page_test.go` (Total=2 under Limit1/Offset1); zero production edits |
| `go test -count=1 localized_list_offers_test.go localized_list_offers_page_test.go list_offers_doubles_test.go` (cwd `tests/unit/missions/application`) | T-0054-005 Application GREEN — ListOffers partition (Gate 3-APPLICATION approved) | RED→GREEN | 2026-10-27: PASS all — per-item mixed page, 3 effective-language q subtests, q-before-pagination (Total=2 under Limit1/Offset1); `LocalizedOfferListPage` + Q-strip + post-visibility effective Match implemented; existing list tests signature-aligned (15 sites), assertions unchanged |
| `go test -count=1 ./tests/unit/missions/application` | T-0054-005 application package check | PASS | 2026-10-27 measured: suite green incl. parallel CHG-0051 T-11 work — prior "CHG-0051 blocks app suite" note retired as stale; `go build ./internal/missions/application` OK; `interface/http` still fails at `get_offer.go:16` (deliberate boundary blocker) |
| `go test -count=1 localized_list_definitions_test.go list_definitions_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — ListDefinitions partition (Gate 2 extended) | FAIL (build, intended) | 2026-10-27: `Handle` arity have (domain.Language) want () ×4 (lines 60, 104, 128, 145) + `defs[i].ContentLanguage` undefined on `domain.MissionDefinition` ×4 (lines 77, 81, 114, 134); `LocalizedDefinition` demanded; reworked per review to assert objective Description on en + es-fallback paths (still 149 lines); zero production edits |
| `go test -count=1 ./tests/unit/missions/application` (post-ListOffers) | T-0054-005 application package re-check | PASS | 2026-10-27: suite still green after ListOffers signature change; `interface/http` now fails at `get_offer.go:16` AND `list_offers.go:23` — expected boundary blockers |
| `go test -count=1 localized_list_definitions_test.go localized_list_definitions_projection_test.go list_definitions_test.go` + `go test -run 'TestListDefinitions' ./tests/unit/missions/application` | T-0054-005 Application GREEN — ListDefinitions partition (Gate 3-APPLICATION approved) | RED→GREEN | 2026-10-27: compile RED incl. 2 new no-alias arity sites → PASS 8/8 (mixed page, fallback table, es request, reader error, both alias regressions); `localizeMissionText` extracted and shared by LocalizeOffer+LocalizeDefinition |
| `go test -count=1 localized_contract_terms_test.go` (cwd `tests/unit/missions/domain`) | T-0054-003 Domain RED — contract Terms text projection (demanded by ListPlayerContracts Gate 3-APPLICATION) | FAIL (build, intended) | 2026-10-27: `terms.WithObjectiveText` undefined at 3 Act sites (lines 34, 99, 119); translation snapshot over Objectives()+Phases() with ID-match, validation, and non-alias demanded; zero production edits |
| `go test -count=1 ./tests/unit/missions/domain -run 'TestContractTermsWithObjectiveText'` + `go test -count=1 ./tests/unit/missions/domain` | T-0054-004 Domain GREEN — WithObjectiveText (third Gate 3-DOMAIN) | RED→GREEN | 2026-10-27: initial run exposed a TEST fixture bug (shared-backing slice literal mutated `ok` mid-table — "blank description allowed" corrupted); fixed literals, then PASS — all 9 tests + full domain package green; `contract_terms_localized.go` 76 lines; build OK |
| `go test -count=1 localized_list_player_contracts_test.go list_offers_doubles_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — ListPlayerContracts partition (Gate 2 extended) | FAIL (build, intended) | 2026-10-27: `Handle` 3-arg arity ×3 (lines 51, 83, 119) + `contracts[i].ContentLanguage` undefined on `domain.MissionContract` ×4 (61, 64, 100, 129); `LocalizedContract` contract-level label + Terms projection demanded; zero production edits |
| `go test -count=1 ./tests/unit/missions/application` (post-ListDefinitions) | T-0054-005 application package re-check | PASS | 2026-10-27: suite green after definitions change + shared-projection refactor (GetOffer/ListOffers regressions hold); `interface/http` now fails at three sites: get_offer.go:16, list_offers.go:23, list_definitions.go:14 — expected boundary blockers |
| `go test -count=1 localized_list_player_contracts_test.go localized_list_player_contracts_projection_test.go list_offers_doubles_test.go` + `go test -run 'TestListPlayerContracts' ./tests/unit/missions/application` | T-0054-005 Application GREEN — ListPlayerContracts partition (Gate 3-APPLICATION sixth scope) | RED→GREEN | 2026-10-27: compile RED (10 sites: 3-arg arity + `ContentLanguage`) → PASS 7/7 incl. phased-terms projection + both alias regressions; `LocalizeContract` uses Terms-effective objectives + `WithObjectiveText` snapshot with atomic whole-contract es fallback |
| `go test -count=1 localized_get_contract_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — GetContract partition (Gate 2 extended) | FAIL (build, intended) | 2026-10-27: `Handle` 3-arg arity ×5 (lines 57, 91, 109, 125, 143) + `served.ContentLanguage` undefined on `*domain.MissionContract` ×3 (78, 98, 132); shared `LocalizeContract` after unchanged authorization demanded; zero production edits |
| `go test -count=1 ./tests/unit/missions/application` (post-ListPlayerContracts) | T-0054-005 application package re-check | PASS | 2026-10-27: suite green; `interface/http` now fails at four sites: get_offer.go:16, list_offers.go:23, list_definitions.go:14, list_player_contracts — expected boundary blockers |
| `go test -count=1 -v localized_get_contract_test.go` (cwd `tests/unit/missions/application`) | T-0054-005 Application GREEN — GetContract partition (Gate 3-APPLICATION seventh scope) | RED→GREEN | 2026-10-27: compile RED → PASS 5/5 (member en detail incl. Terms/Phases, es fallback, outsider denied, reviewer en, read error); `LocalizeContract` invoked only after unchanged authorization |
| `go test -count=1 localized_list_pending_review_test.go` (cwd `tests/unit/missions/application`) | T-0054-001 Application RED — ListPendingReview partition (Gate 2 extended) | FAIL (build, intended) | 2026-10-27: `Handle` 2-arg arity ×4 (lines 56, 85, 101, 118) + `entries[i].ContentLanguage` undefined on `domain.MissionContract` ×2 (67, 73); `[]LocalizedContract` projection + auth-before-store demanded; fixtures model `ContractPendingValidation`; zero production edits |
| `go test -count=1 -v localized_list_pending_review_test.go` + `./tests/unit/missions/application` + `go build ./internal/missions/application` | T-0054-005 Application GREEN/BLUE — ListPendingReview | PASS | 2026-10-27: all 4 localized tests + full package green; auth-before-store preserved, non-nil empty result, per-item `LocalizeContract`, order kept; http gains sixth expected blocker (review queue) |
| `go test -count=1 localized_authoring_create_test.go` (cwd `tests/unit/missions/application`) | T-0054-002 Application RED — bilingual create (Gate 2 extended) | FAIL (build, intended) | 2026-10-27: `application.CreateMissionDefinitionRequest` undefined (lines 26-27) + `Handle` arity have (unknown type) want 8-positional (lines 57, 79, 104); self-contained `localizedDefinitionWriter`; zero production edits |
| `go test -count=1 -v localized_authoring_publish_test.go` (cwd `tests/unit/missions/application`) | T-0054-002 Application RED — bilingual publish (Gate 2 extended) | FAIL (runtime, intended) | 2026-10-27: `RejectsMissingOrPartialEnglish` 4 subtests FAIL — es-only/missing-rules/missing-objective/blank-title drafts publish today (`saves 1`, want `ErrDefinitionInvalid` + `saves 0`); `PublishesBilingualDraft` already green; demands domain completeness gate on `Publish`; helper renamed `publishableEnglish` (was `completeEnglish`, collided with `localized_content_integrity_test.go`); zero production edits |
| `go test -count=1 ./tests/unit/missions/application` (post-RED package diagnostic) | T-0054-002 RED classification cross-check | FAIL (build, intended only) | 2026-10-27: package errors limited to the 5 intended `CreateMissionDefinitionRequest`/arity sites (lines 26-27, 57, 79, 104); no redeclaration after `publishableEnglish` rename; publish file compiles, its runtime RED is masked by the build failure as expected |
| `go test -count=1 -v mission_definition_locales_publish_test.go mission_definition_locales_edges_test.go` (cwd `tests/unit/missions/domain`) | T-0054-003 Domain RED — Publish English completeness (Gate 3-APPLICATION demanded, approved as Domain RED only) | FAIL (runtime, intended) | 2026-10-27: 10 rejection subtests FAIL — drafts publish today without the completeness rule (actual PUBLISHED/v1, want DRAFT/""); 4 positive tests PASS (complete en incl. reversed IDs, optional desc, empty shape, historical es-only published untouched); Gate 3-DOMAIN fourth scope pending |
| `go test -count=1 -v mission_definition_locales_publish_test.go mission_definition_locales_edges_test.go` + `./tests/unit/missions/domain` + `go build ./internal/missions/domain` | T-0054-004 Domain GREEN/BLUE — Publish English completeness (Gate 3-DOMAIN fourth scope approved) | PASS | 2026-10-27: all 14 locale-publish tests + full domain package green; 9 fixture sites aligned with complete en before `Publish` (assertions unchanged); `englishSnapshotComplete` in new `mission_definition_locales.go` (46 lines), mission_definition.go 100 lines; application `localized_authoring_publish_test.go` named-file PASSes independently via domain rule — create RED still blocks app package as intended |
| `go test -count=1 -v localized_authoring_create_test.go` + `-v localized_authoring_publish_test.go` + `./tests/unit/missions/application` + `go build ./internal/missions/application` | T-0054-002 Application GREEN/BLUE — bilingual create (Gate 3-APPLICATION authoring-create scope approved) | PASS | 2026-10-29: `CreateMissionDefinitionRequest` + request-shaped `Handle` with `ContentEn.Clone()` freeze landed; create 3/3 + publish 2/2 named-file PASS, full application package PASS; legacy `create_mission_definition_test.go` 3 calls aligned to request literal; deep-copy case extended to objective-name mutation; one additional es-only draft fixture (`aDraftDefinition`) aligned with complete en — masked earlier by the create compile RED; gofmt clean, all files <150 |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangDefinitions'` (unmodified tree) | T-0054-006 Boundary RED — definitions partition, unmodified run | FAIL (build — BLOCKER, not RED) | 2026-10-29: build FAIL at 9 pre-existing stale call sites in 7 http files (contract_detail:21+35, create_mission_definition:67, get_offer:16, list_contracts:25, list_definitions:14, list_offers:23, review_queue:53+67) — pre-existing breakage, classified as blocker; no production edits |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangDefinitions'` (diagnostic only) | T-0054-006 Boundary RED — definitions partition, overlay diagnostic | FAIL (runtime, intended — conditional) | 2026-10-29: temp overlay mechanically adapts the 7 stale files with `LanguageSpanish` + nils CHG-0051 in-flight `NewPlacementFactPublisher` call; against real `composition.NewServer` + local DynamoDB, re-run after widening to two-case fallback + ContentEn key-absence assertion: **6 FAIL** — `unsupported_language`/`empty_language`/`conflicting_es_then_en`/`conflicting_en_then_es` return 200 not 400 (no lang parsing), `spanish_only_fallback_under_en` + `omitted_lang_ignores_accept_language` serve es item but lack `Content-Language: es`; **2 PASS** — `repeated_es_accepted`, `empty_catalog_under_en`; conditional on overlay — not an ordinary production-run result |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangDefinitions'` (PROVISIONAL — overlay adapts remaining 6 stale http files + CHG-0051 nil-outbox only; `list_definitions.go` + `mission_language_query.go` run as REAL production source) | T-0054-006 Boundary provisional GREEN — definitions partition | PASS 8/8 | 2026-10-29: all 5 query cases (400 INVALID_QUERY_PARAMETER on fr/empty/conflicting×2; 200 on repeated es) + 3 fallback cases (en fallback, omitted-lang + Accept-Language: en ignored, empty catalog) PASS against real `composition.NewServer` + local DynamoDB; `Content-Language: es` on homogeneous non-empty payload, absent on empty; `ContentEn` key absent; gofmt clean, list_definitions.go 43 + mission_language_query.go 26 lines; unmodified `interface/http` build still blocked by the remaining 5 stale call sites — not a full-boundary or ordinary-run claim |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangGetOffer'` (unmodified tree) | T-0054-006 Boundary RED — GetOffer partition, unmodified run | FAIL (build — BLOCKER, not RED) | 2026-10-29: build FAIL at remaining stale call sites (contract_detail:21+35, create:67, get_offer:16, list_contracts:25, list_offers:23, review_queue:53+67) + CHG-0051 `NewPlacementFactPublisher` arity in composition — pre-existing, classified as blocker; no production edits |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangGetOffer'` (diagnostic only; overlay = 6 stale http files + CHG-0051 nil-outbox; `list_definitions.go` runs REAL) | T-0054-006 Boundary RED — GetOffer partition, overlay diagnostic | FAIL (runtime, intended — conditional) | 2026-10-29: `RejectsInvalidQueryBeforeLookup` ×3 FAIL — 404 not 400 (lookup precedes any lang parse); `ServesLegacySpanish` ×3 FAIL — `Content-Language: es` absent AND `ContentLanguage` key dropped from JSON because `LocalizedOffer` inherits `MissionOffer.MarshalJSON` (probe-verified); `UnknownRemains404` PASS; conditional on overlay — not an ordinary production-run result |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangGetOffer'` + `-run '^TestLangDefinitions'` (PROVISIONAL — overlay adapts only remaining 5 stale handlers + CHG-0051 nil-outbox; `get_offer.go`, `get_offer_response.go`, `list_definitions.go`, `mission_language_query.go` run REAL) | T-0054-006 Boundary provisional GREEN — GetOffer partition | PASS 7/7 + 8/8 | 2026-10-29: all GetOffer subtests PASS — 400 precedence before lookup, es fallback/omitted/repeated-en with `Content-Language: es`, `ContentLanguage` retained by explicit `offerResponse` DTO (fixes MarshalJSON drop), MeritSummary 20/50/70 + objective economics wire-verified, `ContentEn` absent, 404 contract unchanged; definitions partition still 8/8; gofmt clean, 39+42 lines; unmodified build still blocked by remaining stale sites — not a full-boundary or ordinary-run claim |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListOffers'` (unmodified tree) | T-0054-006 Boundary RED — ListOffers partition, unmodified run | FAIL (build — BLOCKER, not RED) | 2026-10-29: build FAIL at remaining stale call sites (list_offers:23, list_contracts:25, contract_detail:21+35, review_queue:53+67, create:67) + CHG-0051 composition arity — pre-existing, classified as blocker; no production edits |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListOffers'` (diagnostic only; overlay = 5 stale http files + CHG-0051 nil-outbox; `list_definitions.go`, `get_offer.go`, `get_offer_response.go`, `mission_language_query.go` REAL) | T-0054-006 Boundary RED — ListOffers partition, overlay diagnostic | FAIL (runtime, intended — conditional) | 2026-10-29: `RejectsInvalidQuery` ×4 FAIL — 200 not 400 (no lang parse) + `repeated_es_accepted` missing `Content-Language`; `ServesLegacySpanish` ×2 FAIL — header + `ContentLanguage` key absent (MarshalJSON drop); `MixedActualLanguages` FAIL — English title now observably fired (`expected "Hunt", actual "Caceria"` — persistence drops ContentEn, T-0054-007 dependency) followed by boundary `ContentLanguage` absent in the same subtest (assert order corrected 2026-10-29 so the infra cause executes first); `EmptyHasNoContentLanguage` PASS; conditional on overlay — not an ordinary production-run result |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListOffers(RejectsInvalidQuery\|ServesLegacySpanish\|EmptyHasNoContentLanguage)$'` + `-run '^TestLangListOffers'` + `-run '^TestLang(GetOffer\|Definitions)'` (PROVISIONAL — overlay adapts only 4 remaining stale handlers + CHG-0051 nil-outbox) | T-0054-006 Boundary partial GREEN — ListOffers partition | PASS 7/7 + mixed still RED (infra-only cause) | 2026-10-29: all query-validation/legacy/empty cases PASS (`requestedMissionLanguage` + `offersResponse` DTO reusing `toOfferResponse`; `Content-Language` on homogeneous non-empty only); `MixedActualLanguages` remains RED for infra cause only — `expected "Hunt"/"en"/"", actual "Caceria"/"es"/"es"` because persistence drops `ContentEn` making the page truthfully homogeneous-es; boundary logic correct given data; GetOffer 7/7 + Definitions 8/8 still PASS; gofmt clean, 88+20 lines; no full-boundary claim |
| `go test -count=1 -tags=integration localized_items_fixture_test.go localized_definition_test.go localized_offer_test.go endpoint_test.go` (cwd `tests/integration/infrastructure/dynamodb`, real local DynamoDB) | T-0054-007 Persistence RED — bilingual snapshots | FAIL (runtime, intended) | 2026-10-29: `BilingualDraftRoundTrips`, `PublishedVersionRoundTrips`, `BilingualSnapshotRoundTrips` FAIL — `ContentEn` never written (no M attribute; served text stays Spanish: `"Capture the pilot"` → `"Capturar al piloto"`, `Rescue` → `Rescate`); both `LegacySpanishOnlyFallsBack` PASS — absent attribute decodes to zero English + honest `es` label; maps to `definition_store.go`, `definition_store_queries.go`, `offer_store.go`, `offer_store_query.go`, `attribute_helpers.go`; package run `^TestLocalized(Definition\|Offer)` compiles clean (no CHG-0051 contamination) with identical 3-FAIL/2-PASS; test-quality correction 2026-10-29 — fixture converted to error-returning (no require in helper) with strict loopback+http guard and `require.Len` guard before objective indexing; re-run after correction: identical 3-FAIL/2-PASS |
| `go test -count=1 -tags=integration localized_items_fixture_test.go localized_definition_test.go localized_offer_test.go endpoint_test.go` + `-p=1 ./tests/integration/infrastructure/dynamodb -run '^TestLocalized(Definition\|Offer)'` + `go build ./internal/missions/infrastructure/dynamodb` | T-0054-008 Persistence GREEN — ContentEn M-attribute round-trip | PASS 5/5 | 2026-10-29: all bilingual round-trips green — `ContentEn` persists as DynamoDB M (Title/Description S, Rules L, Objectives L-of-M), legacy items keep attribute absent + es fallback; new `localized_content_mapping.go` (59 lines) reuses `stringSliceToAttribute`/`getString`/`getStringSlice`; `definitionItem`, `definitionFromItem`, `OfferStore.Save`, `mapOfferItem` wired; `attribute_helpers.go` untouched; gofmt clean, all <150 |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListOffers'` (PROVISIONAL — overlay = 4 stale handlers + CHG-0051 nil-outbox) | T-0054-008 HTTP mixed regression — persistence now serves English | PASS 8/8 | 2026-10-29: `MixedActualLanguages` previously RED (`"Caceria"/"es"/"es"`) now PASS — bilingual offer serves `"Hunt"`/`"en"` per item, heterogeneous page has no header; all 8 ListOffers cases green; conditional on overlay — ordinary tree still blocked by 4 stale handlers + CHG-0051 |
| `go test -count=1 -tags=integration localized_items_fixture_test.go localized_contract_fixture_test.go localized_contract_test.go localized_contract_list_test.go endpoint_test.go` + `-p=1 ./tests/integration/infrastructure/dynamodb -run '^TestLocalizedContract'` | T-0054-007 Persistence RED — contract bilingual snapshot | FAIL (runtime, intended) | 2026-10-29: `BilingualSnapshotRoundTrips` + `MixedSnapshotsRoundTrip` FAIL — contract `ContentEn` never written (no M attribute; served Offer/Terms stay Spanish: `Rescue`→`Rescate`, `Capture`→`Capturar`, label `en`→`es`, post-Save in-memory mutation irrelevant); `LegacySpanishOnlyFallsBack` PASS — absent attribute + es; maps to `contract_item.go::contractItem`, `contract_store_mapping.go::mapToContract`; 2-FAIL/1-PASS both commands |
| `go test -count=1 -tags=integration localized_items_fixture_test.go localized_contract_fixture_test.go localized_contract_test.go localized_contract_list_test.go endpoint_test.go` + `-p=1 ./tests/integration/infrastructure/dynamodb -run '^TestLocalized(Definition\|Offer\|Contract)'` + `go build ./internal/missions/infrastructure/dynamodb` | T-0054-008 Persistence GREEN — contract ContentEn M round-trip | PASS 3/3 + 8/8 regression | 2026-10-29: `putEnglishContent(item, contract.Offer.ContentEn)` in `contractItem` + `ContentEn: englishContentFromItem(item)` in `mapToContract` — bilingual contract serves en Offer + en Terms objectives, legacy absent-attribute stays es, post-Save in-memory mutation isolated; 8/8 incl. prior definition/offer 5/5; gofmt clean, 64+58 lines; backfill/migration not in scope |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListContracts'` (unmodified) | T-0054-006 Boundary RED — ListContracts, unmodified run | FAIL (build — BLOCKER, not RED) | 2026-10-29: build FAIL at remaining stale sites (list_contracts:25, contract_detail:21+35, review_queue:53+67, create:67) + CHG-0051 composition arity — pre-existing blocker; no production edits |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListContracts'` (diagnostic; overlay = 4 stale handlers mechanical `LanguageSpanish` + CHG-0051 nil-outbox; real localized files + tests untouched) | T-0054-006 Boundary RED — ListContracts, overlay diagnostic | FAIL (runtime, intended — conditional) | 2026-10-29: 10 FAIL / 3 PASS — invalid queries return 200 not 400 (no parse); `HomogeneousEnglish`/`SpanishDefault`/`ExplicitSpanish`/`LegacyFallback`/`MixedAndFiltered` fail on missing `Content-Language` header AND missing `ContentLanguage` key (observed `expected "en"/"es", actual ""` — `LocalizedContract` drops the label like `LocalizedOffer` did); en text unreachable under forced-Spanish adapter; `Empty`, `Unauthorized` (401 UNAUTHORIZED before parse), `repeated_es_accepted` PASS; conditional on overlay — the create handler itself is mechanically overlaid (request-struct call omitting `ContentEn`) to compile, so this diagnostic cannot count as unmodified-tree RED |
| `go test -overlay /tmp/chg0054-overlay-create/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangCreateDefinition'` + `-run '^TestLangDefinitions'` + `-run '^TestListDefinitionsEndpoint_'` (PROVISIONAL — overlay = CHG-0051 nil-outbox ONLY; create_mission_definition.go + translations DTO run REAL) | T-0054-006 Boundary provisional GREEN — Create definition partition | PASS 5/5 + regressions | 2026-11-01: `BilingualPersistsEnglish` now green — `translations.en` maps to `ContentEn` via new `createMissionDefinitionTranslations` DTO (`englishContent()`), persisted M snapshot verified via real `DefinitionStore`; Spanish-only valid, missing-Spanish 422, 401/403 unchanged; response exposes no `ContentEn`/`translations`; definitions regression `^TestLangDefinitions` ok + `TestListDefinitionsEndpoint_*` 2/2; gofmt clean, 89+37 lines; ordinary tree still blocked by CHG-0051 composition only — not full-boundary |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangListContracts'` + `-count=1 ./tests/unit/missions/interface/http -run '^TestListContracts_OfferMeritSummaryComesFromFrozenTerms$'` (PROVISIONAL — overlay = 3 stale handlers + CHG-0051 nil-outbox; list_contracts.go + DTO run REAL) | T-0054-006 Boundary provisional GREEN — ListContracts partition | PASS 13/13 + merit regression | 2026-10-29: all subtests green — 400 before filter, homogeneous `Content-Language`, `contractListResponse` DTO retains label + Terms-hydrated objectives + MeritSummary economics, mixed page no header, `[]` empty, `ContentEn` absent, auth/404 contracts unchanged; gofmt clean, 59+35 lines; ordinary build still blocked by 3 stale handlers + CHG-0051 — not a full-boundary claim |
| `go test -overlay /tmp/chg0054-overlay/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangContractDetail'` (PROVISIONAL — overlay = review_queue + create + CHG-0051 nil-outbox; contract_detail.go + projection run REAL) | T-0054-006 Boundary provisional GREEN — ContractDetail partition | PASS 11/12 + 1 known orthogonal FAIL | 2026-10-29: all lang/auth cases PASS — 400 before lookup, en detail + `content_language` + `Content-Language`, es default ignores Accept-Language, legacy es fallback, reviewer full / member minimized / outsider 403 projections preserved; `UnknownReturns404` FAILS with 500 `INVALID_REQUEST` — diagnosed pre-existing store defect (generic `errors.New("contract not found")` ≠ `domain.ErrContractNil`), excluded from localization scope; gofmt clean; ordinary build blocked by 2 stale handlers + CHG-0051 |
| `go test -overlay /tmp/chg0054-overlay/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangContractDetail(RejectsInvalidQueryBeforeLookup\|ServesEnglish\|SpanishDefaultIgnoresAcceptLanguage\|LegacyFallback\|ReviewerKeepsFullView\|MemberKeepsMinimizedView\|OutsiderDenied)$'` + `-run '^TestContractDetail_'` (same overlay) | T-0054-006 Boundary provisional GREEN — selected locale cases + legacy contract-detail regression | PASS 11/11 + 5/5 | 2026-10-29: selected-language cases all green incl. reviewer/member/outsider projections; existing `TestContractDetail_{ReviewerGetsDetail,MemberGetsMinimizedDetail,OutsiderDenied,UnknownContract,ExposesPhases}` all pass — `UnknownContract` uses the unit double returning `domain.ErrContractNil` so it stays green; BUG-0054-001 only affects the real-Dynamo path, full `^TestLangContractDetail` remains 11P/1F, not fixed |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangReviewQueue'` (unmodified) | T-0054-006 Boundary RED — ReviewQueue, unmodified | FAIL (build — BLOCKER, not RED) | 2026-10-29: build FAIL at `review_queue.go` (app.Handle arity + `toReviewQueueEntry` type), `create_mission_definition.go`, CHG-0051 composition — pre-existing blocker |
| `go test -overlay /tmp/chg0054-overlay/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangReviewQueue'` (diagnostic; overlay = review_queue forced-Spanish mechanical + create + nil-outbox; real localized files + tests untouched) | T-0054-006 Boundary RED — ReviewQueue, overlay diagnostic | FAIL (runtime, intended — conditional) | 2026-10-29: 10 FAIL / 3 PASS — invalid queries return 200 not 400 (no parse); `HomogeneousEnglish`/`SpanishDefaultAndExplicit`/`LegacyFallback`/`MixedAndFiltered`/`repeated_en_accepted` fail on missing `Content-Language` header + missing `content_language` key + es-served text under forced adapter; `Unauthenticated` (401), `NonReviewerForbidden` (403), `Empty` PASS already; test-quality correction 2026-10-29: `reviewQueueRequest` error-returning (no assertions in helper), `AuthBoundaries` split into two single-Act tests, `mustMarshal` dependence removed — re-run produced same 10 intended fails; conditional on overlay; maps to `review_queue.go` + future `review_queue_response.go` (TBD at Gate 3) |
| `go test -overlay /tmp/chg0054-overlay/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangReviewQueue'` + `-run '^TestReviewQueue_'` (PROVISIONAL — overlay = create_mission_definition + CHG-0051 nil-outbox only; review_queue.go runs REAL) | T-0054-006 Boundary provisional GREEN — ReviewQueue partition | PASS all + 3/3 regression | 2026-10-29: all 8 tests + 7 subtests green — 400 before lookup, `content_language` per entry via updated `toReviewQueueEntry(application.LocalizedContract)`, homogeneous `Content-Language`, mixed no header, `[]` empty, auth 401/403 unchanged; `TestReviewQueue_{ReviewerListsPendingContracts,MemberDenied,EmptyQueueReturnsEmptyList}` legacy regression PASS; no extra DTO file needed (snake_case projection retained label in-file); gofmt clean, 95 lines; ordinary build blocked by create + CHG-0051 — not full-boundary |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangCreateDefinition'` (unmodified) | T-0054-006 Boundary RED — Create definition, unmodified | FAIL (build — BLOCKER, not RED) | 2026-11-01: build FAIL at `create_mission_definition.go` positional `app.Handle` arity + CHG-0051 composition — pre-existing blocker |
| `go test -overlay /tmp/chg0054-overlay/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangCreateDefinition'` (diagnostic; overlay = baseline create copy omitting `ContentEn` intentionally + CHG-0051 nil-outbox; real localized files + tests untouched) | T-0054-006 Boundary RED — Create definition, overlay diagnostic (TEST-HTTP-LANG-003 / SCN-MB-CONTENT-DRAFT-EN / REQ-MB-013) | FAIL (runtime, intended — conditional) | 2026-11-01: `BilingualPersistsEnglish` FAIL — baseline handler ignores `translations.en`, persisted `ContentEn` zero (`expected "Rescue"/"Save the pilot"/["Capture the pilot"]` absent); `SpanishOnlyValid` (201 + zero English), `MissingSpanishRejected` (422 + no record), `Unauthenticated` (401 + no record), `NonAdminForbidden` (403 + no record) PASS; maps to `create_mission_definition.go` only (translations DTO + `ContentEn` on request); BLUE plan: explicit translations DTO + parser; Gate 3-BOUNDARY (create) pending; conditional on overlay — the create handler itself is mechanically overlaid (request-struct call omitting `ContentEn`) to compile, so this diagnostic cannot count as unmodified-tree RED |
| `go test -overlay /tmp/chg0054-overlay-create/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangCreateDefinition'` + `-run '^TestLangDefinitions'` + `-run '^TestListDefinitionsEndpoint_'` (PROVISIONAL — overlay = CHG-0051 nil-outbox ONLY; create_mission_definition.go + translations DTO run REAL) | T-0054-006 Boundary provisional GREEN — Create definition partition | PASS 5/5 + regressions | 2026-11-01: `BilingualPersistsEnglish` now green — `translations.en` maps to `ContentEn` via new `createMissionDefinitionTranslations` DTO (`englishContent()`), persisted M snapshot verified via real `DefinitionStore`; Spanish-only valid, missing-Spanish 422, 401/403 unchanged; response exposes no `ContentEn`/`translations`; definitions regression `^TestLangDefinitions` ok + `TestListDefinitionsEndpoint_*` 2/2; gofmt clean, 89+37 lines; ordinary tree still blocked by CHG-0051 composition only — not full-boundary |
| `go test -overlay /tmp/chg0054-overlay-create/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangPublishDefinition'` (overlay = CHG-0051 nil-outbox only; unmodified run blocked by composition arity) | T-0054-006 Boundary — Publish definition (TEST-HTTP-LANG-003 / SCN-MB-CONTENT-INCOMPLETE / REQ-MB-013) | PASS 4/4 — existing-behavior GREEN characterization | 2026-11-01: incomplete-English table (missing/incomplete rule/incomplete objective) → 422 `INVALID_MISSION_DEFINITION`, draft untouched, no version row (`ErrDefinitionNotFound`); complete bilingual → 200 PUBLISHED v1, `GetVersion` carries frozen en snapshot + Spanish economics, response has no `ContentEn`; **no Gate 3 required** — endpoint already delegates to domain completeness check, persistence GREEN round-trips the snapshot |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangPublishDefinition'` (unmodified) | T-0054-006 Boundary — Publish definition, unmodified run | FAIL (build — BLOCKER, not behavior RED) | 2026-11-01: build FAIL on CHG-0051 `NewPlacementFactPublisher` arity in `composition/server.go`; no full ordinary package pass claimed |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestBackfillOffer'` | T-0054-009 Backfill partition 1 — offer rehearsal RED (TEST-INF-LANG-002) | FAIL (compile — intended RED) | 2026-11-01: `undefined: missiondynamo.ReviewedMissionTranslation`, `NewTranslationBackfiller`, `BackfillRecord` — exactly the settled API; tests demand `Rehearse(ctx, reviewed, apply)` report with `Eligible`/`Updated`/`Unresolved` + `SPANISH_SNAPSHOT_MISMATCH`/`VERSION_MISMATCH` reasons; dry-run no-mutation + apply-only-compatible + wrong-version-skip cases written against real unique tables; maps to `internal/missions/infrastructure/dynamodb` new backfiller; Gate 3-PERSISTENCE (backfill) pending |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestBackfill'` | T-0054-009 Backfill partition 1 — expanded RED (TEST-INF-LANG-002) | FAIL (compile — intended RED) | 2026-11-01: same three undefined symbols only; new cases — 5-case text-drift table (description/rule/objective name+desc/ID → `SPANISH_SNAPSHOT_MISMATCH`), prior-authored-English candidate left byte-identical on apply, contract dry/apply (terms-name drift unresolved, match en title + localized `Capture` term, roster/state/lineage preserved), published definition both rows (`DEFINITION`/`def-rescue` + `#v1` in `Updated`, `GetVersion` serves en); phased-contract coverage deferred to future partition |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestBackfill'` | T-0054-009 Backfill partition 1 — review-corrected RED (TEST-INF-LANG-002) | FAIL (compile — intended RED) | 2026-11-01: same three undefined symbols only, zero non-target errors; corrections — contract fixture objective-slice copy (terms-only drift now real), ordered-rule drift case (`localized_backfill_rules_test.go`), phased-contract apply+drift (`localized_backfill_contract_phases_test.go` — 2 phases, en `Capture`/`Track` localized mandatory terms, phase order + merits preserved); duplicate partition-1 gate row removed, lead's detailed backfill-RED-extension row retained |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestBackfill'` | T-0054-009 Backfill partition 1 — fail-closed input coverage (TEST-INF-LANG-002) | FAIL (compile — intended RED) | 2026-11-01: same three undefined symbols only, zero unrelated errors; correction — rule-order fixture now sets `English.Rules=["No aids","No weapons"]` so reviewed translation is structurally valid and isolates order drift; new `localized_backfill_invalid_input_test.go` asserts `ErrorIs(domain.ErrDefinitionInvalid)` on incomplete reviewed English with byte-identical item and no `ContentEn` — input completeness must be verified before any enrichment |
| `go test -count=1 -tags=integration -p=1 -v ./tests/integration/infrastructure/dynamodb -run '^TestBackfill'` (lead-checked) + focused `^TestLocalized(Definition\|Offer\|Contract)` + `go build ./internal/missions/infrastructure/dynamodb` | T-0054-009 local rehearsal GREEN/BLUE, TEST-INF-LANG-002 | PASS 12 top-level tests + 5 drift subtests; 8/8 localized persistence regression; adapter build PASS | 2026-11-01: isolated DynamoDB Local tables only; source-es/version match and incomplete-input fail closed, dry-run leaves raw items unchanged, apply adds only `ContentEn` to eligible definition/offer/contract snapshots, existing English and divergent terms/phases remain untouched, unresolved reasons recorded. Five new adapter files (59+75+70+60+44 lines) use paginated, consistent partition queries and conditional `UpdateItem` (no whole-item overwrite); gofmt and file-cap checked. BLUE: separated scanning, validation, matching, and conditional write; existing domain content types reused, no new Value Object or runtime wiring. No reviewed live catalog translations, live migration, or full Boundary/rollout gate claimed. |
| **BUG-0054-001 — Unknown contract detail maps to 500 rather than 404** | separate unresolved finding (no production fix authorized) | open | `ContractStore.Get` returns generic `errors.New("contract not found")` instead of `domain.ErrContractNil`; handler's `ErrContractNil` case unreachable → 500 `INVALID_REQUEST` where 404 is expected; user expressly excluded from CHG-0054 scope; `TestLangContractDetailUnknownReturns404` kept as regression evidence |
| `go test -count=1 ./tests/unit/missions/application` (post-GetContract) | T-0054-005 application package re-check | PASS | 2026-10-27: suite green; `interface/http` now fails at five sites: get_offer.go:16, list_offers.go:23, list_definitions.go:14, list_contracts.go:25, contract_detail.go:21 — expected boundary blockers |
| `go build ./internal/missions/interface/http` | boundary compile check (diagnostic) | FAIL (expected blocker) | 2026-10-27: `get_offer.go:16` calls old one-arg `Handle` — deferred to Boundary GREEN per approved gate scope; `go build ./internal/missions/application ./internal/missions/domain` clean |
| `go test -count=1 ./tests/unit/missions/application` | package baseline diagnostic | FAIL (pre-existing) | 2026-10-27: CHG-0051 T-9/T-10 RED — `application.PlacementManifest`, `domain.AdjudicationAttempt`, `application.AdjudicationSnapshot` undefined in `manifest_publisher_doubles_test.go`/`season_finalizer_doubles_test.go`; unrelated, untouched; named-file command used for classified RED |
| `go test -count=1 ./tests/unit/missions/application -run 'TestLocalized|Test.*Lang'` | T-0054-001 Application RED (remaining partitions) | pending | blocked by pre-existing CHG-0051 compile failure; use named-file commands per partition until baseline clears |
| `go test -count=1 ./tests/unit/missions/application -run 'Test.*Localiz|Test.*Bilingual'` | T-0054-002 Application RED (bilingual authoring) | pending | expected RED: missing request/domain symbols |
| `go test -count=1 language_test.go` (cwd `tests/unit/missions/domain`) | T-0054-003 Domain RED — GetOffer-demanded partition (TEST-DOM-LANG-001/002 subset) | RED→GREEN | 2026-10-27 RED (build fail): intended missing shape only — `domain.Language`/`LanguageSpanish`/`LanguageEnglish`, `ParseLanguage`, `ErrLanguageInvalid`, `LocalizedContent`, `MissionDefinition.ContentEn`, `MissionOffer.ContentEn`. GREEN same day: `go test -count=1 -v language_test.go` PASS 3/3 after minimal `language.go`/`localized_content.go`/`ContentEn` implementation |
| `go test -count=1 -v language_test.go` | BLUE wire-leak regression (TEST-DOM-LANG-002 extension) | RED(assertion)→GREEN | 2026-10-27: `TestMissionOfferMarshalJSON_HidesEnglishContent` + `TestMissionDefinitionMarshalJSON_HidesEnglishContent` FAILED on assertion — `ContentEn` serialized on both wire shapes (premature HTTP exposure before Boundary RED); fixed with `json:"-"` on both fields → PASS 5/5 |
| `go test -count=1 mission_offer_definition_test.go mission_offer_test.go mission_definition_test.go mission_rules_test.go mission_definition_terms_test.go offer_list_filter_test.go offer_list_page_test.go offer_list_query_test.go` | legacy regression on touched structs (named-file, self-contained) | PASS | 2026-10-27: offer/definition/rules/filter suites green against new fields; `objective_test.go` excluded — needs `contract_fixtures_test.go` helpers (pre-existing cross-file dep, unrelated) |
| `go test -count=1 ./tests/unit/missions/domain` | package baseline diagnostic | FAIL (pre-existing) | 2026-10-27: CHG-0051 T-10 RED — `NewAdjudicationAttempt`/`ErrAdjudicationAttemptInvalid`/`MembersOfCreditedContracts` undefined in `adjudication_attempt_test.go`/`credited_roster_test.go`; unrelated, untouched |
| `go test -count=1 ./tests/unit/missions/domain -run 'TestLanguage|TestLocalized|Test.*Locale'` | T-0054-003 Domain RED (remaining partitions) | pending | blocked by pre-existing CHG-0051 compile failure; use named-file commands per partition until baseline clears |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangContractDetail'` (unmodified) | T-0054-006 Boundary RED — ContractDetail partition, unmodified run | FAIL (build — BLOCKER, not RED) | 2026-10-29: build FAIL at remaining stale sites (contract_detail.go:21+35, review_queue.go:53+67, create_mission_definition.go:67); CHG-0051 `NewPlacementFactPublisher` 5-arg arity also masks `composition/server.go:80` — pre-existing blockers; zero production edits |
| `go test -overlay <tmp>/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangContractDetail'` (diagnostic; overlay = contract_detail/review_queue/create mechanical `LanguageSpanish` adapters + composition/server.go nil-outbox for CHG-0051 arity; real localized files + tests untouched) | T-0054-006 Boundary RED — ContractDetail partition, overlay diagnostic | FAIL (runtime, intended — conditional) | 2026-10-29: 7 FAIL / 1 PASS top-level (12 subtests: 8 FAIL / 4 PASS) — invalid queries return 500 not 400 (lookup precedes parse AND real `ContractStore.Get` maps missing id to generic error → `INVALID_REQUEST`, not `ErrContractNil`); `ServesEnglish`/`SpanishDefault`/`LegacyFallback` lack `content_language` key + `Content-Language` header and en text is unreachable under forced-Spanish adapter (Spanish text assertions themselves pass); `ReviewerKeepsFullView`/`MemberKeepsMinimizedView` lose label/header/translation but keep wire shape; `UnknownReturns404` FAILS for a pre-existing cause — real store never returns `ErrContractNil` so unknown id is 500 `INVALID_REQUEST` today (store error-mapping gap, orthogonal to lang); `OutsiderDenied` + `repeated_es_accepted` PASS; conditional on overlay — the create handler itself is mechanically overlaid (request-struct call omitting `ContentEn`) to compile, so this diagnostic cannot count as unmodified-tree RED |
| `go test -overlay /tmp/chg0054-overlay-create/overlay.json -count=1 -tags=integration -p=1 ./tests/integration/http -run '^TestLangCreateDefinition'` + `-run '^TestLangDefinitions'` + `-run '^TestListDefinitionsEndpoint_'` (PROVISIONAL — overlay = CHG-0051 nil-outbox ONLY; create_mission_definition.go + translations DTO run REAL) | T-0054-006 Boundary provisional GREEN — Create definition partition | PASS 5/5 + regressions | 2026-11-01: `BilingualPersistsEnglish` now green — `translations.en` maps to `ContentEn` via new `createMissionDefinitionTranslations` DTO (`englishContent()`), persisted M snapshot verified via real `DefinitionStore`; Spanish-only valid, missing-Spanish 422, 401/403 unchanged; response exposes no `ContentEn`/`translations`; definitions regression `^TestLangDefinitions` ok + `TestListDefinitionsEndpoint_*` 2/2; gofmt clean, 89+37 lines; ordinary tree still blocked by CHG-0051 composition only — not full-boundary |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/http -run 'TestLang|TestLocaliz'` | T-0054-006 HTTP boundary RED | pending | expected RED: missing lang/content_language/header support |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run 'TestLocaliz|TestBackfill'` | T-0054-007/009 persistence RED + backfill rehearsal | pending | expected RED: unmapped locale attributes / missing backfill writer |
| `go test -count=1 ./tests/unit/...` | unit scope | pending | |
| `go test -count=1 -tags=integration -p=1 ./tests/integration/composition` | composition | pending | expected `not_affected` beyond existing wiring |
| `bash tools/validate-sdd-change.sh` | structural SDD gate | pending | before final verification |
| coverage / mutation / critical E2E / docs / clean-up | L3 gates | pending | T-0054-011 |

## Test-First Constraints

- Gate 2 must approve these artifacts before any test file is created;
  Gate 1 approval covers spec writes only.
- Each scoped Gate 3 reviews classified RED evidence before the mapped
  production edit; manual QA/build/lint/coverage never unlocks code.
- Domain work is demanded-only: the Application RED of T-0054-001/002
  decides which domain cycles open in T-0054-003.
- Boundary RED (T-0054-006/007) waits for LAYER-GATE-APPLICATION;
  infrastructure/interface/composition production waits for
  Gate 3-BOUNDARY.

## Residual Risks

- Backfill correctness on historical frozen records is mitigated by the
  rehearsal gate (T-0054-009): en-only augmentation, mismatch
  rejection, unresolved-record report. No live migration runs without
  a separate approval.
- Legacy items decoding as es-only must not regress today's
  "missing `Rules` decodes empty" tolerance — covered by
  TEST-INF-LANG-001.
- `content_language`/`ContentLanguage` labels and the
  `Content-Language` header must stay consistent; a header claiming
  `en` over mixed items is a findings-class labeling defect (see
  `security-review.md`).
- The publish completeness rule is a deliberate behavioral change to
  `POST /definitions/{code}/publish`: existing es-only published
  versions stay valid, but new publishes require `en`. The staged
  cutover in `rollout.md` (authoring tooling + reviewed translations
  first) is the mitigation; an enforced-early publish rejection of a
  still-Spanish workflow is a deployment-sequencing risk, not a code
  defect.

=== 2026-11-16 FINAL GATE TABLE ===
| Gate | Result |
| --- | --- |
| validate-sdd-change.sh | PASS (gofmt/vet/unit/integration/race) |
| Coverage | domain 92.4%, application 91.7% |
| Security | closed, no Critical/High |
| Mutation | manual analysis recorded |
| Critical E2E | TEST-E2E-LANG-001 PASS |
| Docs | `lang` contract in API-CONTRACT.md |
| File size | no production file ≥150 lines |
