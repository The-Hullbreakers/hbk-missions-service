# CHG-0055 — Create local season preview scenarios: tasks

Execution: sequential, one owner (`Devin`), maximum one active task.
No parallel tracks or handoffs until the mandatory independent L3 review.
All tests are independent of other test layers and start from clean state.

| Task | Ownership / wave / dependency | Outcome, mapped test and command | Next step |
| --- | --- | --- | --- |
| T-0055-001 — Specify safe seasonal previews | Lead / 1 / Gate 1 | Record US-0055-001 and REQ-0055-001..002 in these L3 artifacts; approved Gate 1 | Request Gate 2 |
| T-0055-002 — Prove preview generation and guards RED | Tester / 2 / Gate 2 | TEST-0055-001..003: `tests/unit/seeds/test_season_scenarios.py`; `python3 -m unittest discover -s tests/unit/seeds -p 'test_season_scenarios.py'`; RED `ModuleNotFoundError: season_scenario_model` recorded in verification.md; Gate 3-TOOL approved 2026-11-23 | Implement mapped script |
| T-0055-003 — Generate safe deterministic scenarios | Coder / 3 / Gate 3-TOOL | `scripts/season_scenario_model.py`, `scripts/season_scenario_contracts.py`, `scripts/season_scenario_credits.py` created verbatim (66/33/36 lines, all <150); CLI `scripts/seed_season_scenarios.py` deferred to Gate 3 for the local-resource test; GREEN 10/10 via focused unit command; BLUE reviewed in verification.md | Boundary RED |
| T-0055-004 — Prove local adapter projection RED | Tester / 4 / tool GREEN | TEST-0055-004..005: `tests/integration/infrastructure/dynamodb/test_season_scenarios.py` plus `tests/integration/infrastructure/dynamodb/season_scenarios_readers_test.go`; `python3 -m unittest discover -s tests/integration/infrastructure/dynamodb -p "test_season_scenarios.py"` and `go test -tags=integration ./tests/integration/infrastructure/dynamodb -run TestSeasonScenarioReaders -count=1`; separate fixed persistent local preview tables, no deletion; RED recorded in verification.md; Gate 3-BOUNDARY approved 2026-11-23; GREEN via both focused commands | CLI apply path implemented |
| T-0055-005 — Populate all preview projections | Coder / 5 / Gate 3-BOUNDARY | `scripts/seed_season_scenarios.py` + `scripts/season_scenario_local.py` created verbatim (69/89 lines <150); fresh-start and mid-season previews applied to their fixed persistent tables; both focused commands GREEN; original tables verified unchanged | E2E and docs |
| T-0055-006 — Verify real user journey and safety | Lead/Tester / 6 / prior GREEN | TEST-0055-006 PASS: `go test -tags=integration ./tests/integration/http -run '^TestSeasonPreviewJourney$' -count=1` 6/6 subtests via real composition+HTTP+DynamoDB Local; OIDC session double only, no browser claim; off-season applied by lead (72 DEFINITION only), originals unchanged; no cloud/identity writes | Document |
| T-0055-007 — Document switching and validate quality | Coder / 7 / E2E | `docs/LOCAL-SEASON-PREVIEWS.md` + README bullet (`seeds/README.md` untouched >150); catalog validator PASS risk=L3 on 24 scoped files; selected mutation 5/5 mutants killed; affected Python coverage 91% (194 statements/17 missed); `code-quality-review.md` recorded; repo `tools/validate-sdd-change.sh` run once with user-approved container reset: exit 0 (gofmt/vet/unit/integration/race all ok; coverage domain 92.4%/app 91.8%/infra 61.5%; pre-existing coverage files backed up and restored); Python transport characterization 4/4, total unit 14/14 | Final review |
| T-0055-008 — Load mid-season locally and confirm | Lead / 8 / tests and explicit apply | Mid-season already applied at `http://localhost:8000` by approved Go local integration test; lead independently verified counts and original table unchanged; dry-run pending=0; no deletion | Final review |
| T-0055-009 — Consolidate approved scenario contract | Lead / 9 / final review approval | Final user review approved 2026-11-23; REQ-SCC-019/REQ-MB-018/REQ-MCF-027 merged, specs/index.md updated, entire change archived once; active duplicate absent | Done |

Traceability: REQ-0055-001 -> T-0055-002/003/004/005 ->
TEST-0055-001 (endpoint/table safety), TEST-0055-002 (idempotency
and conflict), TEST-0055-004 (no other table/identity writes) ->
the named `season_scenario_*` modules; CLI implemented only after
the local-resource Gate 3 approval.
REQ-0055-002 -> T-0055-002/003/004/005/006 ->
TEST-0055-003 (three scenario shapes), TEST-0055-005
(real reader projections), TEST-0055-006 (public journey) ->
same modules; no Go production file is unlocked.

T-0055-001..009 have recorded outcomes and verification; the
approved change is archived. Build, lint,
manual QA, and security scans cannot substitute for RED.
