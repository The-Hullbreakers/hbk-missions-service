# Tasks — CHG-0045

Wave 1 opens Application RED before Domain cycles; Wave 3 (boundary) waits
for `LAYER-GATE-APPLICATION`. Unit tests use hand-written port doubles;
infrastructure covered by `tests/integration/infrastructure/dynamodb/`.

| Task | Wave | Depends on | Outcome | Test (RED) | Production unlocked |
| --- | --- | --- | --- | --- | --- |
| T-045-001 — Extend list filter value objects | 1 | — | Application tests compile against new `OfferListPage`/query fields and fail on behavior | `tests/unit/missions/application/list_offers_test.go` (RED: expects page envelope + multi-filter behavior) | `internal/missions/domain/list_filters.go` (new file `offer_list_page.go` if needed for <150 lines) |
| T-045-002 — Implement domain parsing and matching | 1 | T-045-001 RED | CSV parse, `q`, paging bounds, sort whitelist green | `tests/unit/missions/domain/list_filters_test.go` | `internal/missions/domain/list_filters.go`, `errors` sentinel |
| T-045-003 — Return paginated page from use case | 2 | T-045-002 | `Handle` returns `OfferListPage` with post-exclusion total | `tests/unit/missions/application/list_offers_test.go` GREEN | `internal/missions/application/list_offers.go` |
| T-045-004 — Add facets use case | 2 | T-045-002 | `ListOfferFacetsHandler` returns distinct sorted facets | `tests/unit/missions/application/list_offer_facets_test.go` | `internal/missions/application/list_offer_facets.go`, `CollectFacets` |
| T-045-005 — Push down filters in OfferStore | 3 | `LAYER-GATE-APPLICATION` | `FilterExpression`, direct mapping, `remaining_capacity` availability | `tests/integration/infrastructure/dynamodb/offer_store_test.go` | `internal/missions/infrastructure/dynamodb/offer_store.go` |
| T-045-006 — Wire envelope + facets endpoint | 3 | `LAYER-GATE-APPLICATION` | HTTP parses new params, returns envelope, exposes `/offers/facets` | `tests/integration/http/list_offers_test.go`, facets case | `internal/missions/interface/http/list_offers.go`, `list_offer_facets.go`, `errors.go`, `composition/server.go` |
| T-045-007 — Document contract | 4 | T-045-006 | `docs/API-CONTRACT.md` reflects envelope, params, facets | docs review | `docs/API-CONTRACT.md` |

Each task records outcome, verification (test command), and next step at
execution time. Owner: single agent. `go test ./tests/unit/...` for waves
1-2, `make test-integration` (docker-compose) for wave 3.

## Execution Record

- T-045-001/002 — RED evidenced by build failure on new API (OfferListPage,
  Categories/Modes, Page, Sort); review-test-evidence: signature RED accepted.
  GREEN: `go test ./tests/unit/missions/domain ./tests/unit/missions/application`.
- T-045-003/004 — Application GREEN: `Handle` returns `OfferListPage`;
  `ListOfferFacetsHandler` added. Verified by unit suite.
- T-045-005 — `FilterExpression` pushdown + `mapOfferItem` (N+1 removed);
  `remaining_capacity` used for availability. Verified:
  `go test -p=1 -run OfferStore ./tests/integration/infrastructure/dynamodb/`.
- T-045-006 — envelope + `/offers/facets` route; `INVALID_QUERY_PARAMETER`
  mapped. Verified: unit http suite + `go test -p=1 ./tests/integration/...`.
- T-045-007 — `docs/API-CONTRACT.md` updated; `tools/validate-sdd-change.sh`
  passed (coverage: domain 91.6%, application 91.3%, dynamodb 92.7%).
- Clean-up: split `offer_list.go`, `offer_list_page.go`, `offer_facets.go`,
  `list_query_helpers.go`, `offer_store_query.go` — all files < 150 lines.
