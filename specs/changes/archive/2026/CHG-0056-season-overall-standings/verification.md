# CHG-0056 - Verification record

Status: Gate 1 and Gate 2 approved 2026-11-24; TEST-0056-001
Application RED recorded; GREEN, BLUE, final review and
consolidation pending. No production changes are claimed.

| ID | Layer / standalone planned command | RED reason |
| --- | --- | --- |
| TEST-0056-001 | Application: `go test ./tests/unit/missions/application -run TestOverallStandings -count=1` | Overall query missing |
| TEST-0056-002 | Domain: `go test ./tests/unit/missions/domain -run TestOverallStandings -count=1` | Season-wide policy missing |
| TEST-0056-003 | HTTP: `go test -tags=integration ./tests/integration/http -run TestOverallStandings -count=1` | Route missing after core gate |
| TEST-0056-004 | Infrastructure: `go test -tags=integration ./tests/integration/infrastructure/dynamodb -run TestOverallStandings -count=1` | Focused credit reader missing after core gate |
| TEST-0056-005 | Category regression: existing unit/HTTP category, own and history tests | Existing public contract must survive |
| TEST-0056-006 | Critical journey: real composition plus owned local table; explicit overall + category checks | Not begun |

At each GREEN record the observed RED, scoped Gate 3 decision,
standalone command, source and test BLUE review, Value Object decision,
line counts, and remaining concerns. Tests must not require another
test layer's mutable fixture. Repository-wide validator
`bash tools/validate-sdd-change.sh` is not authorized here if it
resets test containers; ask before running it. The safe structural
validator may run after drafting.

Final gates: security/auth/privacy, affected and project-wide coverage,
selected mutation of aggregation/tie-break/cursor, critical HTTP/E2E,
read-volume/latency, documentation, code-quality/dead-code audit,
reviewer approval and canonical consolidation. None is passed yet.

## RED: TEST-0056-001 — Application overall standings

T-0056-002 records the Application-layer RED under approved Gate 2.
User approved Gate 3-APPLICATION on 2026-11-24 for the mapped
`internal/missions/application/get_overall_standings.go` only,
after the demanded Domain gate. Domain RED and Gate 3-DOMAIN
must occur before any production code for the overall ranking.

- Test type: unit / Application layer.
- File: `tests/unit/missions/application/overall_standings_test.go`
  (72 lines, gofmt-clean).
- Command: `go test ./tests/unit/missions/application -run
  TestOverallStandings -count=1` → exit 1, build failed.
- Failure: `undefined: application.NewGetOverallStandingsHandler`
  at lines 32 and 59 — the intended missing symbol. All other
  imports, the outgoing `OverallCreditsReader` stub
  (`ListOverallCredits`), domain values, and the existing
  `fakeLeaderboardSeasonReader` season port resolved and compiled.
- Port stub: in-memory `overallCreditsStub` returns fixed credits
  and a completion map; no DynamoDB, HTTP, or other resource.
- Standalone: no dependency on another test layer's fixtures.
- Production files changed: none. Mapped future production only
  `internal/missions/application/get_overall_standings.go`; the
  Domain RED (TEST-0056-002) is recorded below; its Gate 3
  remains pending.
- BLUE review: pending Gate 3.
- BLUE plan: after the demanded Domain gate, make the one-query
  Application port/use case pass this test; refactor with the
  isolated Application test green, no new value object without
  a validated invariant, and keep each new test/source file
  below 150 physical lines. Combined suite is not meaningful
  while this package has an unresolved missing symbol.

## RED: TEST-0056-002 — Domain overall standings

T-0056-003 records the demanded Domain-layer RED after approved
Gate 3-APPLICATION. Gate 3-DOMAIN is requested, not approved; no
production change is authorized.

- Test type: unit / Domain layer; `package domain_test` reuses the
  existing package with fresh in-memory credit slices and completion
  maps — no mutable fixture shared with another layer.
- File: `tests/unit/missions/domain/overall_standings_test.go`
  (74 lines, gofmt-clean, under 150).
- Command: `go test ./tests/unit/missions/domain -run
  TestOverallStandings -count=1` → exit 1, build failed.
- Failure: `undefined: domain.ProjectOverallStandings` at lines
  23, 52 and 70 — the three intended call sites only; all other
  imports, `domain.MeritCredit`/`Standing` fields and `require`
  resolved. No environment or tooling failure.
- Scenarios (Given/When/Then):
  - `AggregatesIndividualCreditsAcrossCategories` — Given merits in
    two categories plus TEAM and other-season credits for the same
    recipients, When projected for `season-1`, Then one INDIVIDUAL
    row per recipient sums cross-category merits (mira 200, kai
    190), counts distinct completed contracts once and distinct
    `(category, objective_id)` pairs, and orders by
    `domain.CompareStanding`.
  - `TieBreaksByCompletionsObjectivesThenRecipient` — Given four
    equal-total recipients, When ordered, Then completed-contract
    count, then unique objective pairs, then recipient ID decide
    the order `bob, zeta, carol, alice`.
  - `WithoutIndividualMeritsIsEmpty` — Given only TEAM-track merits,
    When projected, Then the result is empty (no invented rank).
- Invariants covered: season and INDIVIDUAL-track scoping, per-
  recipient aggregation across categories, deduplicated completed
  contracts and objective pairs, deterministic tie-break order,
  empty projection without individual merits.
- Production files changed: none. Mapped future production only
  `internal/missions/domain/overall_standings.go`.
- Standalone: clean-state unit test; no DynamoDB, HTTP, Application
  or other test layer's fixture.
- Gate 3-DOMAIN approved by user on 2026-11-24 for only
  `internal/missions/domain/overall_standings.go` after the
  reviewed TEST-0056-002 RED. No other production file is unlocked.
- GREEN/BLUE pending implementation: implement only the mapped
  file, keep it under 150 physical lines, reuse `Standing` and
  `CompareStanding`, and add no new Value Object unless a
  validated invariant demands one.

## GREEN/BLUE: TEST-0056-002 — Domain overall standings

Production added only `internal/missions/domain/overall_standings.go`
(43 lines, gofmt-clean); test file unchanged at 74 lines,
gofmt-clean. Both under 150 physical lines.

- Command: `go test ./tests/unit/missions/domain -run
  TestOverallStandings -count=1 -v` → PASS, 3/3 tests
  (`AggregatesIndividualCreditsAcrossCategories`,
  `TieBreaksByCompletionsObjectivesThenRecipient`,
  `WithoutIndividualMeritsIsEmpty`), exit 0.
- Independent layer check: `go test ./tests/unit/missions/domain
  -count=1` → `ok`, exit 0; no regression in the existing Domain
  suite. `git diff --check` clean on touched paths.
- Value Object decision: no new VO — existing `Standing` carries
  the projection fields and `CompareStanding` already owns the
  total/completed/objectives/recipient ordering; the new
  unexported `overallStandingFacts` is a per-recipient
  accumulator, not a domain concept.
- SOLID/CQRS/complexity: single pure function, one responsibility
  (season+track-scoped aggregation and ordering); plain-value
  input, `[]Standing` output reusing existing types; map-based
  dedup keeps counting O(credits); no I/O, ports, or framework
  dependency — Domain stays provider-neutral.
- Domain standalone gate: passed — the Domain package is green
  without Application, infrastructure, or resource dependencies.
- Scope: the Domain production change is exactly the one
  authorized file; no Application, HTTP, infrastructure, SAM,
  frontend, or test edits. Application RED (TEST-0056-001)
  remains intentionally red until its own approved GREEN.

## GREEN/BLUE: TEST-0056-001 — Application overall standings

Gate 3-APPLICATION approved by user 2026-11-24 for only
`internal/missions/application/get_overall_standings.go` after the
Domain gate. Production added only that file (57 lines,
gofmt-clean); Application test file unchanged at 72 lines,
gofmt-clean. Both under 150 physical lines.

- Command: `go test ./tests/unit/missions/application -run
  TestOverallStandings -count=1 -v` → PASS, 2/2 tests
  (`SumsCategoriesAndShowsCallerOutsideTop`,
  `NewSeasonHasNoInventedRank`), exit 0.
- Independent layer check: `go test ./tests/unit/missions/
  application -count=1` → `ok`, exit 0; no regression in the
  existing Application suite. Domain layer was already green and
  unchanged, so it was not rerun. `git diff --check` clean.
- Value Object decision: no new VO — `OverallStandingsResult`
  and `OverallCreditsReader` are Application-layer port/result
  shapes; ranking values reuse `domain.Standing`,
  `domain.Leaderboard`, and `domain.LeaderboardEntry`.
- Ports/provider-neutrality: one focused outgoing port
  `OverallCreditsReader.ListOverallCredits` per CQRS query
  operation; handler depends only on `SeasonReader` and that
  port plus `domain.CallerIdentity` — no DynamoDB, HTTP, or
  framework types in Application.
- Behavior: season existence via `SeasonReader.Get`, one credit
  read, `ProjectOverallStandings` + `BuildLeaderboard` reuse,
  caller rank derived from `identity.PlayerID()` across the full
  board (outside the requested page), `my_standing` nil when the
  caller has no merits.
- Application standalone gate: passed — the Application package
  is green with in-memory port doubles and no resource
  dependency. Scope: exactly the one authorized production file;
  no other file changed.

## RED: TEST-0056-003 — HTTP overall standings route

T-0056-005 records the Boundary HTTP RED under approved Gate 2.
Gate 3-BOUNDARY approved by user on 2026-11-24 for the mapped
overall-credit reader, HTTP handler, composition wiring and SAM
GET event. No unrelated production change is authorized.

- Test type: integration / public HTTP through real composition;
  `//go:build integration`, Testify only, one Act call per test.
- File: `tests/integration/http/overall_standings_test.go`
  (117 lines, gofmt-clean, under 150; imports regrouped to
  stdlib / testify / local module per review verdict).
- Fixture: `languageSessionServer` with `previewJourneySession`
  session fake (external OIDC replaced only at the session
  boundary) seeds a fresh test-owned table on the dedicated
  `hbk-missions-dynamodb-test` container (`:18000`); the table is
  not dropped and never touches preview/original tables.
- Command: `DYNAMODB_TEST_ENDPOINT=http://localhost:18000
  AWS_EC2_METADATA_DISABLED=true go test -tags=integration
  ./tests/integration/http -run TestOverallStandingsRoute
  -count=1 -v` → FAIL, exit 1.
- Failure: route missing — `ReturnsSeasonAndCallerRank` expected
  200 got 404; `RequiresSession` expected 401 got 404;
  `RejectsBadLimitAndCursor` subtests `zero-limit`,
  `oversized-limit`, `invalid-cursor` each expected 400 got 404.
  All five assertions fail only on the absent route; setup and
  seeding passed (no fixture or environment error); the missing
  composition route is the intended RED reason.
- Arrange exception recorded: `require.NoError` on fixture setup
  is the unavoidable existing real-resource harness pattern.
- Mapped future production: `internal/missions/interface/http/
  overall_standings.go`, wiring in
  `internal/missions/composition/server.go`, and the SAM GET
  event in `template.yaml`.
- GREEN/BLUE pending Gate 3-BOUNDARY.

## RED: TEST-0056-004 — DynamoDB overall credit facts reader

Same approved Gate 3-BOUNDARY scope as TEST-0056-003.

- Test type: integration / real DynamoDB Local adapter through
  the Application handler; `//go:build integration`, Testify
  only, one Act call.
- File: `tests/integration/infrastructure/dynamodb/
  overall_standings_test.go` (54 lines, gofmt-clean, under 150;
  imports regrouped to stdlib / AWS+testify / local module).
- Fixture: `localizedStoreFixture` creates a fresh unique
  test-owned table on `:18000`; seeds season, four credits
  (including a TEAM-track credit that must not contribute) and
  contract rows via the lead-authored direct `PutItem` fixture
  (c1 COMPLETED, c2 IN_PROGRESS); table not dropped.
- Command: `DYNAMODB_TEST_ENDPOINT=http://localhost:18000
  AWS_EC2_METADATA_DISABLED=true go test -tags=integration
  ./tests/integration/infrastructure/dynamodb -run
  TestOverallStandings -count=1 -v` → FAIL, exit 1, build
  failed before any fixture ran.
- Failure: `undefined: missiondynamo.NewOverallCreditsReader` at
  line 39 — the only error after import regrouping; the intended
  missing adapter symbol.
- Mapped future production: `internal/missions/infrastructure/
  dynamodb/overall_credit_facts.go` only.
- GREEN/BLUE pending Gate 3-BOUNDARY.

## GREEN/BLUE: TEST-0056-004 — DynamoDB overall credit facts

Gate 3-BOUNDARY approved by user 2026-11-24 (recorded above).
Production added only
`internal/missions/infrastructure/dynamodb/overall_credit_facts.go`
(98 lines, gofmt-clean, under 150); test file unchanged at 54
lines.

- Command: `DYNAMODB_TEST_ENDPOINT=http://localhost:18000
  AWS_EC2_METADATA_DISABLED=true go test -tags=integration
  ./tests/integration/infrastructure/dynamodb -run
  TestOverallStandings -count=1 -v` → PASS, exit 0
  (`DynamoDBReadsCreditsAndCompletedContracts`, 0.11s) against
  the owned `hbk-missions-dynamodb-test` container; test table
  not dropped, preview/original tables untouched.
- Adapter contract: `NewOverallCreditsReader` returns
  `*OverallCreditsStore` satisfying Application
  `OverallCreditsReader.ListOverallCredits`; paginated `Query` on
  `PK=MERIT_CREDIT` filtered to the requested season and
  INDIVIDUAL track with consistent reads; one `GetItem` per
  distinct contract ID marks `ContractCompleted`; only selected
  season/track facts returned — no evidence media, unit, or
  confirmation IDs exposed.
- Fail-closed validation: malformed `Amount` (non-N or
  unparseable), missing/empty Category, RecipientID, ContractID,
  ObjectiveID, or a season mismatch returns an error rather than
  a plausible wrong score. The query filter excludes other tracks;
  it does not treat their presence in the table as an error.
- Value Object decision: no new VO — reuses `domain.MeritCredit`,
  `domain.ContractState`, and the existing `getString` attribute
  helper; `OverallCreditsStore` is a thin adapter, not a domain
  concept.
- Read-cost caveat: the partition query scans all `MERIT_CREDIT`
  items and filters by season/track server-side; per-contract
  `GetItem` calls scale with distinct contracts. The final
  read-volume/latency gate must evaluate this before release; a
  dedicated index or projection change would need a reviewed
  amendment.
- Scope: infrastructure production change is exactly the one
  authorized file; no HTTP, composition, `template.yaml`,
  frontend, shared-infra, or test edits. HTTP GREEN is not
  claimed — the route/wiring RED remains pending its own gate.

## GREEN/BLUE: TEST-0056-003 — HTTP overall standings route

Gate 3-BOUNDARY approved by user 2026-11-24 (recorded above).
Boundary production touched exactly the authorized surfaces:
new `internal/missions/interface/http/overall_standings.go`
(76 lines, gofmt-clean), two wiring lines plus one route line in
pre-existing dirty `internal/missions/composition/server.go`
(139→142 lines, still under 150, gofmt-clean), and one
`MissionsOverallStandingsRoute` resource in the pre-existing
untracked `template.yaml` (318→324 lines; `sam validate` passes).
Every unrelated dirty change in those files and all other dirty
files preserved; `leaderboard.go` and tests untouched.

- Command: `DYNAMODB_TEST_ENDPOINT=http://localhost:18000
  AWS_EC2_METADATA_DISABLED=true go test -tags=integration
  ./tests/integration/http -run TestOverallStandingsRoute
  -count=1 -v` → PASS, exit 0: composition route returns 200
  with season/track/participants/items/next_cursor/my_standing,
  missing session 401, and `zero-limit`/`oversized-limit`/
  `invalid-cursor` each 400 — all through the real composition
  and owned `:18000` test table with the session fake only at
  the OIDC boundary.
- Old-route regression: `go test -tags=integration
  ./tests/integration/http -run 'Test(LeaderboardEndpoint|
  GetMyStanding|SeasonStatistics)' -count=1` → `ok`, exit 0;
  per-category leaderboard, own standing, and statistics
  endpoints unchanged.
- Value Object decision: no new VO — response reuses
  `leaderboardEntryResponse`/`toLeaderboardResponse` from the
  existing dirty `leaderboard.go` (read-only reuse; the one
  small personal-entry mapper is flagged for later cleanup
  review, not modified now).
- Error mapping: missing session 401, season not found 404 via
  `domain.ErrSeasonNotFound`, invalid cursor 400 via
  `domain.ErrInvalidCursor`, bad limit 400
  `INVALID_QUERY_PARAMETER`; `PublicCode` owns public codes.
- Scope: HTTP/composition/template changes are exactly the
  authorized wiring; no old endpoint, use case, or test removed
  (all still consumed). `git diff --check` clean; no E2E,
  coverage, or formatter-suite claim made.

## Local preview runtime recovery and smoke (2026-11-24)

- Context: local preview SAM on :3006 (screen
  `hbk-missions-preview`, env file
  `hbk-missions-preview-mid-season.json`, dummy AWS
  credentials, DynamoDB Local :8000) returned 502 on all
  routes after the shared-infra template regeneration. Lambda
  bootstrap panicked with `UnrecognizedClientException`
  against real AWS because the container lacked
  `DYNAMODB_ENDPOINT_URL`/`DYNAMODB_TABLE_NAME`.
- Root cause (shared infra, recorded in CHG-0001): the
  regenerated `.run/templates/hbk-missions-service.yaml` was
  produced without the renderer's `ENV_FILE` argument, so the
  template carried no `Environment.Variables` and SAM
  `--env-vars` values for undeclared names were dropped.
- Recovery: preview-specific template
  `.run/templates/hbk-missions-preview-mid-season.yaml`
  rendered with a KEY=VALUE env derived from the preview JSON,
  `sam build` refreshed `.aws-sam/build/template.yaml`, and
  only the :3006 preview screen was restarted. No source,
  backend template, or gateway changes; other services
  untouched.
- Smoke results (manual curl, unauthenticated):
  `GET :3006/healthz` → 200;
  `GET :3006/api/v1/missions/seasons/preview-mid-season/
  standings/overall?limit=5` → 401;
  same path via gateway :8080 → 401;
  `GET :3006/api/v1/missions/seasons/preview-mid-season/
  leaderboards/INDIVIDUAL` → 401. Container env verified:
  DynamoDB endpoint `http://host.docker.internal:8000`,
  table `hbk-missions-preview-mid-season`.
- Boundary: this verifies local route reachability and auth
  enforcement only; authenticated behavior evidence remains
  the earlier real-composition run with a fake OIDC session.
  No real credentials, remote calls, table writes, or E2E
  claim.

## Quality wave — T-0056-007 (2026-11-26)

- gofmt clean on all touched files; `go vet` clean
  on missions packages.
- Unit suite: all 5 missions packages pass,
  including 2 new application pins (credits
  reader error propagation, malformed cursor
  rejection) and 1 new domain pin (in-progress
  contract not counted as completion).
- Integration (scoped): http + dynamodb pass on
  :18000 owned table; new 404 season-not-found
  route pin and malformed-facts/query-failure
  reader pins pass. Pre-existing unrelated S3
  integration failure (`TestEvidenceStore_Exists
  Phantom`, no S3 emulator running) is outside
  this scope and left to its own owner.
- Coverage gate: merged unit+integration profile
  over the four new production files → 105/108
  statements = 97.2% (domain/app/http 100%,
  reader 91.4%). Uncovered: GetItem infra-fault
  guard and >1MB pagination continuation —
  documented equivalents, unreachable without
  fault injection or oversized data.
- Mutation: see mutation-report.md — 10/10
  mutants killed.
- Critical-E2E evidence: authenticated journey
  via real-composition integration test (fake
  OIDC session → 200 full payload) plus the
  owner's real authenticated browser journey
  front→gateway→backend on the mid-season
  preview (recorded in CHG-184 E2E evidence).
  Anonymous route access 401 direct and via
  gateway (curl evidence above).

## Security gate verdict — PASS (2026-11-26)

role: oauth resource server. The route is
wrapped by AuthMiddleware; identity comes from
validated context, never a caller-supplied ID.
Read-only query; no new write path, secret,
role check, or authorization branch. Anonymous
access fails closed (401 verified). No
Critical/High findings.

## Consolidation — T-0056-008 (2026-11-26, owner approved)

- REQ-SCC-020 and REQ-SCC-021 merged into
  `capabilities/seasonal-category-competition/spec.md` with canonical
  scenario IDs SCN-SCC-OVERALL-001..006, one invariant (aggregate only
  INDIVIDUAL-track season credits; each contract counts once) and one
  implementation note (handler, domain projection, reader, route).
- `specs/index.md`: capability row points to this dossier; archived
  row added for CHG-0056.
- Dossier moved to `specs/changes/archive/2026/
  CHG-0056-season-overall-standings/`.
