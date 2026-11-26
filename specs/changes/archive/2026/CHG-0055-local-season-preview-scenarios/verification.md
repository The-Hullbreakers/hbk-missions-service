# CHG-0055 — Create local season preview scenarios: verification

Status: passed. Gate 1 approved 2026-11-23. Gate 2 approved
2026-11-23. Gate 3-TOOL approved 2026-11-23. Gate 3-BOUNDARY
approved 2026-11-23 for pure local-tool and local-resource
scopes. Final review and consolidation are pending.

| Evidence | Planned independent command / acceptance |
| --- | --- |
| TEST-0055-001..003 — Unit behavior | `python3 -m unittest discover -s tests/unit/seeds -p 'test_season_scenarios.py'` succeeds after recorded RED |
| TEST-0055-004..005 — Local infrastructure | `python3 -m unittest discover -s tests/integration/infrastructure/dynamodb -p "test_season_scenarios.py"` and `go test -tags=integration ./tests/integration/infrastructure/dynamodb -run TestSeasonScenarioReaders -count=1` succeed independently on separate fixed persistent local preview tables; no deletion |
| TEST-0055-006 — Real journey | Compose the real Go HTTP server in `httptest` against the dedicated preview table; exercise current season, offers, contracts, leaderboard and missing-auth routes with simulated external OIDC |
| Existing Go regression | `go test ./tests/unit/...` and affected `go test -tags=integration ./tests/integration/infrastructure/dynamodb/...` if available |
| Source quality | `python3 -m compileall -q scripts/seed_season_scenarios.py`; focused Python unit/integration; `git diff --check` |
| Risk validation | `tools/validate-sdd-change.sh` (if repository lacks this helper, use catalog checkout's validator with recorded path) |
| Coverage, mutation and security | Record selected gate command, result, scope and any environmental limitation; never claim an unavailable gate passed |
| Independent review | Human reviewer checks safety, identity/privacy, repeatability, the canonical merge and actual commands before final validation; pending |

For each gate record the failing test name and reason (RED), reviewed
decision, minimum GREEN, BLUE changes to both source and tests,
standalone rerun, files under 150 lines, and next action. A blocked
gate remains pending, never silently waived.

## RED: TEST-0055-001..003

Command: `python3 -m unittest discover -s tests/unit/seeds -p
'test_season_scenarios.py' -v`. Expected and actual:
`ModuleNotFoundError: No module named 'season_scenario_model'` at the
test module import (line 11); a collection-time narrow missing
symbol, not an environment failure (`seed_values` resolves).

test_type: unit. test_layer: boundary-local-tool policy. No
domain/application product behavior; both layer gates
not_affected. Mapped production ONLY:
`scripts/season_scenario_model.py`, `scripts/season_scenario_contracts.py`,
`scripts/season_scenario_credits.py` (the latter two only if strictly
needed for TEST-0055-003). No CLI or AWS writes are unlocked by this
RED. Setup uses pure fixtures; owned mutable state per test; cleanup
none; depends_on_test_layer: none; manual_qa supplemental.
Arrange/Act/Assert markers use Python `#` comment syntax (canonical
`//` is unavailable in Python).

Gate 3-BOUNDARY approved 2026-11-23 (pure local-tool scope).
GREEN/BLUE recorded below.

## GREEN/BLUE: TEST-0055-001..003

Command: `python3 -m unittest discover -s tests/unit/seeds -p
'test_season_scenarios.py' -v` — 10 tests, all ok (exit 0),
2026-11-23. `python3 -m compileall -q` clean on the three modules.

Production added verbatim from lead-authored snippets (no
adaptations, no mechanical typo fixes needed):
`scripts/season_scenario_model.py` (66 lines),
`scripts/season_scenario_contracts.py` (33 lines),
`scripts/season_scenario_credits.py` (36 lines); test file 133
lines — every maintained file under 150.

BLUE review: production and test re-read together after GREEN;
each unsafe destination keeps its own Act via `subTest`; Assert
sections contain assertions only. Value Object decision: none
added — rows stay DynamoDB-JSON attribute maps produced by the
existing `seed_values` helpers (`av`/`key`/`item`/`offer_item`/
`versioned`); a domain value object would cross the
provider-neutral boundary for a local-only fixture tool with no
Domain/Application consumer, and scenario identity remains plain
`(PK, SK)` tuples inside `missing_items`.

## RED: TEST-0055-004..005

Fixture plan approved by user 2026-11-23: separate fixed persistent
local preview tables on localhost, no cleanup/deletion (see
change.md Current evidence).

Commands and actual results, 2026-11-23:

- `python3 -m unittest discover -s
  tests/integration/infrastructure/dynamodb -p
  'test_season_scenarios.py' -v`
  - Ambient environment first: ERROR in Arrange —
    `RuntimeError: AWS query failed: aws: [ERROR]: Your session has
    expired. Please reauthenticate using 'aws login'` (expired local
    AWS CLI session; environment blocker, not the intended RED).
  - With the test's own dummy credential env
    (`AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy
    AWS_EC2_METADATA_DISABLED=true AWS_DEFAULT_REGION=us-east-1`):
    `test_fresh_start_preserves_existing_table` FAIL —
    `AssertionError: 2 != 0 : python3: can't open file
    '.../scripts/seed_season_scenarios.py': [Errno 2] No such file or
    directory`; `test_remote_endpoint_is_rejected_without_contact`
    ok (vacuous at RED: exit 2 already satisfies nonzero-return and
    no-credential-leak assertions). Intended RED confirmed: missing
    `scripts/seed_season_scenarios.py`.
- `go test -tags=integration
  ./tests/integration/infrastructure/dynamodb -run
  '^TestSeasonScenarioReaders$' -count=1`
  - Compiled clean (all application/domain/dynamodb symbols
    resolved); FAIL at `require.NoError` — `exit status 2`, same
    missing `scripts/seed_season_scenarios.py`. Intended RED
    confirmed before any store or DB call.

Safety evidence: only consistent-read `partition()` queries ran
against `hbk-missions-local`; the CLI never executed, so no writes
occurred. `aws dynamodb list-tables` shows no `hbk-missions-preview-*`
table exists; original missions and identity tables unchanged.

Fixture repair (2026-11-23, test-only): the earlier ambient blocker
above was an environment issue, not the intended RED — the expired
local AWS CLI session leaked into `seed_aws`'s own subprocess queries.
`test_season_scenarios.py` now wraps each test in
`mock.patch.dict(os.environ, ENV, clear=True)` so the helper's AWS
subprocesses inherit the same dummy local credentials as the CLI
invocation and the ambient `AWS_SESSION_TOKEN` is removed. The Go
file was gofmt-formatted only. Corrected standalone reruns, no
external env overrides:

- `python3 -m unittest discover -s
  tests/integration/infrastructure/dynamodb -p
  'test_season_scenarios.py' -v` —
  `test_fresh_start_preserves_existing_table` FAIL
  `AssertionError: 2 != 0 : can't open file
  '.../scripts/seed_season_scenarios.py'`; remote-endpoint test ok
  (vacuous at RED). Intended missing-CLI RED, no blocker.
- `go test -tags=integration
  ./tests/integration/infrastructure/dynamodb -run
  '^TestSeasonScenarioReaders$' -count=1` — FAIL at
  `require.NoError`, `exit status 2`, same missing CLI file.

Both failures are the expected missing `scripts/seed_season_scenarios.py`,
not environment or build failures.

Gate 3-BOUNDARY for local-resource scope: approved by user
2026-11-23 after RED review. GREEN/BLUE recorded below.

## GREEN/BLUE: TEST-0055-004..005

Production added verbatim from lead-authored snippets (no
adaptations): `scripts/season_scenario_local.py` (89 lines) and
`scripts/seed_season_scenarios.py` (69 lines); both under 150.
CLI sets dummy local credentials itself and targets only the three
fixed `hbk-missions-preview-*` tables on HTTP loopback; `hbk-users`
and `hbk-missions-local` are read-only; writes are conditional
`attribute_not_exists` puts in batches of <=100 with pre/post
verification; no delete/truncate path exists.

Standalone commands, 2026-11-23, all exit 0:

- `python3 -m unittest discover -s tests/unit/seeds -p
  'test_season_scenarios.py' -v` — 10/10 ok (regression).
- `python3 -m unittest discover -s
  tests/integration/infrastructure/dynamodb -p
  'test_season_scenarios.py' -v` — 2/2 ok (7.27s): fresh-start
  applied to `hbk-missions-preview-fresh-start` (1 SEASON, 36
  OFFER, 0 MERIT_CREDIT), `hbk-missions-local` partitions verified
  unchanged, remote endpoint rejected without contact.
- `go test -tags=integration
  ./tests/integration/infrastructure/dynamodb -run
  '^TestSeasonScenarioReaders$' -count=1` — PASS (4.11s): mid-season
  applied to `hbk-missions-preview-mid-season`; real use cases read
  current season `preview-mid-season` (unofficial), 4
  individual-track statistics buckets, top-recipient standing, and
  non-empty player contracts in Spanish.
- `python3 -m compileall -q` clean on both new modules.

BLUE: production/test pair re-read after GREEN; both files under
150 lines; CLI stays thin orchestration over `season_scenario_*`
helpers; no scoring/query/schema changes from approved snippets;
no new Value Objects (rows remain DynamoDB-JSON maps). Persistent
preview tables remain for user review per rollout.md; nothing was
deleted.

## E2E: TEST-0055-006

Command: `go test -tags=integration ./tests/integration/http -run
'^TestSeasonPreviewJourney$' -count=1` — PASS (test 0.08s, package
0.414s), 2026-11-23; gofmt clean. All six subtests ok through real
`composition.NewServer`, real HTTP routes, and DynamoDB Local:
current season `preview-mid-season` (unofficial), offers >=1,
statistics 4 buckets, PVP_AIR INDIVIDUAL leaderboard >=1 item,
player contracts >=1, missing-auth 401.

Scope limits: `previewJourneySession` substitutes only external
OIDC (`Bearer preview` -> member roles); this is existing-behavior
characterization through real composition + HTTP + local DB — not
new product behavior and not a real login/browser E2E; no browser
screenshot claimed. Member identity was read at runtime from a
persisted CONTRACT `Roster`; no user IDs recorded. No DB writes
beyond composition table verification on the existing preview
table; no deletes. One benign AWS SDK checksum warning on a
DynamoDB Local response body; non-fatal.

Off-season applied by lead 2026-11-23:
`python3 -B scripts/seed_season_scenarios.py --scenario off-season
--endpoint-url http://localhost:8000 --apply` — verified 72
DEFINITION rows only (no SEASON/OFFER/CONTRACT/MERIT_CREDIT) and
`hbk-missions-local` unchanged. All three preview tables now
persist for user review.

Docs: `docs/LOCAL-SEASON-PREVIEWS.md` added plus README
Documentation bullet; `seeds/README.md` untouched (>150 lines).
Repo `tools/validate-sdd-change.sh` deliberately not run (Docker
compose down risk); catalog validator still fails only the
expected pending E2E/mutation/handoff evidence.

## Regression RED: BUG-0055-001 / REG-0055-001

BUG-0055-001 — Reject empty-userinfo preview endpoints.
REG-0055-001 — Empty userinfo must be rejected.

Classification: production security guard defect within approved
REQ-0055-001. `urlsplit('http://@localhost:8000').username` is the
empty string (falsy), so the userinfo guard passes and
`validate_target` returns the fixed table name instead of None —
contrary to the spec's "absence of userinfo". Expected None;
actual `hbk-missions-preview-mid-season`. Mapped production ONLY
`scripts/season_scenario_model.py::validate_target`; no canonical
intent or risk change.

Command: `python3 -m unittest discover -s tests/unit/seeds -p
'test_season_scenarios.py' -v` — intended failing subTest
`test_rejects_remote_or_unapproved_destinations`
(endpoint='http://@localhost:8000', scenario='mid-season'):
`AssertionError: 'hbk-missions-preview-mid-season' is not None`;
other 9 tests ok; exit 1. Reproduced 2026-11-23.

Scoped Gate 3 regression GREEN: approved by user 2026-11-23.
GREEN/BLUE recorded below.

## GREEN/BLUE: BUG-0055-001 / REG-0055-001

Production change (only `scripts/season_scenario_model.py::
validate_target` guard): `url.username`/`url.password` truthiness
replaced by `is not None` checks, so any userinfo component —
including the empty string from `http://@host` — is rejected while
absent userinfo (None) still passes. No dataset selection, scoring,
AWS query, or schema changes.

Commands, 2026-11-23, exit 0:

- `python3 -m unittest discover -s tests/unit/seeds -p
  'test_season_scenarios.py' -v` — 10/10 ok; the empty-userinfo
  subTest now rejects with None, all prior cases unchanged.
- `python3 -m unittest discover -s
  tests/integration/infrastructure/dynamodb -p
  'test_season_scenarios.py' -v` — 2/2 ok (4.94s); fresh-start
  re-apply idempotent (no new writes), remote endpoint still
  rejected; preview tables preserved, no deletes.

BLUE: single-expression guard tightened; `is not None` matches
urlsplit semantics (None = absent, "" = present-but-empty); both
suites rerun standalone; security-review.md notes the regression
fixed. No new Value Objects; file remains under 150 lines.

## Characterization: TEST-0055-007

Coverage gap follow-up (2026-11-23): lead measured 85% on the five
affected Python files; targeted transport branches (~64%) needed
characterization. `tests/unit/seeds/test_season_scenario_transport.py`
(70 lines) added — in-memory `LocalAWSFake` doubles only; no DB.

Command: `python3 -m unittest discover -s tests/unit/seeds -p
'test_season_scenario*.py' -v` — 14/14 ok, exit 0 (10 scenario
tests + 4 transport tests).

Characterizes already-existing production behavior, not a RED
authorization for changed production: `table_exists` pagination
cursor, `scan_rows` page loop, `create_preview_table` create ->
describe-until-ACTIVE -> `validate_schema` call order with the
readers' KEYS/ATTRIBUTES/INDEXES schema and PAY_PER_REQUEST
billing, and `insert_missing` 100-row batches of conditional
`attribute_not_exists(PK) AND attribute_not_exists(SK)` puts with
no replacement path. Production unchanged; measurement command,
config and thresholds remain lead-owned.

## Repo validator: tools/validate-sdd-change.sh

Command: `bash tools/validate-sdd-change.sh` — run once
2026-11-23 under explicit user authorization for its
`docker compose down --remove-orphans` / up / down cycle and
ephemeral test-DynamoDB :18000 row loss. Result: exit 0,
`=== SDD validation passed ===`. Stages: compose test services up
(`hbk-missions-dynamodb-test` :18000, `hbk-missions-s3-test`
:14566) then down via trap; gofmt clean; `go vet` clean; unit
tests all ok; integration tests all ok; race detector all ok;
coverage domain 92.4%, application 91.8%,
infrastructure/dynamodb 61.5% (Go package scope — separate from
the Python 91% affected-file gate). Single run, no retries; the
script exercises the repository-wide suite, not only CHG-0055.

File protection: pre-existing untracked Nov-16
`coverage-{domain,app,infra}.out` were copied to
`/tmp/hbk-chg0055-validator-backup/` and `cmp`-verified BEFORE the
run. The script rewrote app and infra profiles (domain profile
byte-identical); regenerated copies preserved at
`/tmp/hbk-chg0055-validator-generated/`; all three originals
restored via `cp -p` and `cmp`-verified after the run. Preview
DynamoDB :8000 was not part of the compose file and remained
running and untouched; no tables deleted.

## Coverage Gate (affected scope passed)

`go test -count=1 -coverpkg=./internal/missions/... -coverprofile=/tmp/hbk-chg0055-coverage.out ./tests/unit/...`
passed all five Go unit packages; `go tool cover -func=/tmp/hbk-chg0055-coverage.out`
reported 47.5% aggregate statements for unit-only Go production scope.
This unit-only figure is not project-wide unit+integration coverage.
The repository validator later ran its broader Go suites with
explicit permission to reset test resources on port 18000;
see "Repo validator" for its 92.4% Domain, 91.8% Application,
and 61.5% infrastructure results. No Go production code changed
in this change; the affected Python scope is measured separately.

Coverage.py 7.10.7 was installed in isolated
`/tmp/hbk-chg0055-venv` (no repository/global install). The lead
ran coverage with `--source=scripts` over the scenario unit suite
and each fixed-table CLI scenario under `--append`, after confirming
all three dry-runs had `pending=0`. A first coverage CLI invocation
failed on its import path; setting `PYTHONPATH=scripts` resolved that
environment issue before measuring the CLI. The first affected
Python production report was 85% (194 statements, 29 missed).
After four in-memory transport characterization tests, the
same coverage file was appended with
`coverage run --append --source=scripts -m unittest discover -s
tests/unit/seeds -p 'test_season_scenario_transport.py'`.
`coverage report --include='scripts/season_scenario_*.py,scripts/seed_season_scenarios.py' -m`
then reported **91%** (194 statements, 17 missed) across the five
changed Python production files: contracts 100%, credits 100%,
local transport 91%, model 93%, CLI 84%. The affected-scope >=90%
gate passes; no project-wide Go or Python repository coverage
percentage is claimed.

## Mutation Gate (selected scope)

`mutation-report.md` records five lead-authored source-copy mutants
against the approved endpoint, idempotency, active-user and scoring
guards. All five failed on behavioral assertions (5/5 killed; no
survivors). The full normal unit suite passed 10/10 after adding
per-unit bonus and completion-reward assertions. No native
mutation runner was installed, so no project-wide score is claimed.

## Final review gate

Scoped canonical L3 validator passed after evidence synchronization;
repo-native validator passed earlier after the last code/test edit
and before this evidence-only sync. No real browser/OIDC test is
claimed. The user approved the final human review and consolidation
on 2026-11-23; no additional external security reviewer is claimed.
Canonical owners and index were updated and CHG-0055 archived once.

## Archive check

Canonical policy validator passed pre-move (risk=L3, 25 files). The
folder moved once to `specs/changes/archive/2026/` with no active
duplicate; `specs/index.md` references the archive path from all
three capability rows and the archived-changes table; `git diff
--check` clean; canonical REQ-SCC-019/REQ-MB-018/REQ-MCF-027 present
in their owners.
