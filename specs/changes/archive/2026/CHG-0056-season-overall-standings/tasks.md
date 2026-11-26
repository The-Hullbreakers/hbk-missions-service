# CHG-0056 - Sequential tasks

One owner, no parallel tracks. Every task records actual result and
the next action before moving to the next wave.

| Task - outcome | Owner / wave / dependency | RED and verification | Next |
| --- | --- | --- | --- |
| T-0056-001 - Specify overall scoring and compatibility | Lead / 1 / Gate 1 | Approved active delta; audit consumers of old routes | Gate 2 review |
| T-0056-002 - Prove Application behavior fails without overall query | Tester / 2 / Gate 2 | TEST-0056-001 in `tests/unit/missions/application/overall_standings_test.go`; `go test ./tests/unit/missions/application -run TestOverallStandings -count=1` | Review RED / Gate 3-APPLICATION |
| T-0056-003 - Prove and implement Domain rank policy | Coder / 3 / Application RED | TEST-0056-002 in `tests/unit/missions/domain/overall_standings_test.go`; `go test ./tests/unit/missions/domain -run TestOverallStandings -count=1` RED, Gate 3-DOMAIN, GREEN/BLUE | Domain gate |
| T-0056-004 - Complete Application query | Coder / 4 / Domain gate | TEST-0056-001 GREEN/BLUE, isolated Application test | Application gate |
| T-0056-005 - Prove HTTP/auth and real credit-fact projection before outer code | Tester / 5 / Application gate | TEST-0056-003..004 in `tests/integration/http/overall_standings_test.go` and `tests/integration/infrastructure/dynamodb/overall_standings_test.go`; independent `go test -tags=integration ./tests/integration/http -run TestOverallStandings -count=1` and analogous infrastructure command, including a contract credited in two categories | Gate 3-BOUNDARY |
| T-0056-006 - Wire additive route and verify existing consumers | Coder / 6 / Boundary RED | HTTP and infrastructure GREEN/BLUE; TEST-0056-005 regression on category/own/history/statistics; no removal without approved REMOVED delta | E2E |
| T-0056-007 - Verified critical journey and quality | Tester / 7 / GREEN | TEST-0056-006 journey recorded; security PASS, coverage 97.2%, mutation 10/10 killed, gofmt/vet and scoped validator PASS | Human review |
| T-0056-008 - Consolidated approved owner | Lead / 8 / final review | REQ-SCC-020/021 merged into `seasonal-category-competition/spec.md`, index updated, dossier archived after explicit owner approval | Done |

TEST-0056-001 Application RED is recorded and awaits Gate 3;
remaining tests are planned, not executed. Each RED must fail for absent
behavior, not an environment/fixture error. Standalone tests own
their resources; no test may delete the persistent preview tables.
