# Tasks — CHG-0015 List mission offers and player contracts over HTTP

| Task | Title | Wave | Ownership | Depends on | Status |
| --- | --- | --- | --- | --- | --- |
| T-001 | RED: domain `OfferListFilter`/`ContractListFilter` parse + bucket tests | 1 | agent | Gate 2 | done |
| T-002 | GREEN: implement filter value objects and parsers | 2 | agent | T-001 RED, Gate 3-DOMAIN | done |
| T-003 | RED: `ListOffersHandler` + `ListPlayerContractsHandler` unit tests with port doubles | 2 | agent | T-002 | done |
| T-004 | GREEN: implement application handlers and ports | 3 | agent | T-003 RED, Gate 3-APPLICATION | done |
| T-005 | RED: `OfferStore.List`/`ContractStore.ListByPlayer` infra integration tests (tenant isolation) | 3 | agent | T-004 | done |
| T-006 | GREEN: persist `TenantID` and implement store list methods | 4 | agent | T-005 RED, Gate 3-INFRA | done |
| T-007 | RED: HTTP handler + `tests/unit` boundary tests (auth, filters, 200/400/500) | 4 | agent | T-006 | done |
| T-008 | GREEN: thin HTTP handlers, route wiring, `template.yaml` routes | 5 | agent | T-007 RED, Gate 3-BOUNDARY | done |
| T-009 | BLUE: refactor while green, coverage gate `>= 90%` affected scope, `< 150` lines | 5 | agent | T-008 | done |
| T-010 | Docs update + `validate-sdd-change.sh` + consolidate | 6 | agent | T-009 | in-progress |

## Outcomes

- All RED phases observed expected failures (undefined symbols for filters/handlers/store methods).
- `type`/`difficulty` filters dropped: `MissionOffer` lacks those fields (domain RED finding).
- `TenantID` now persisted on `MissionOffer`/`MissionContract`; `OfferStore.Get`/`ContractStore.ListActiveByPlayer` honor tenant.
- New: `OfferStore.List`, `ContractStore.ListByPlayer`, `ListOffersHandler`, `ListPlayerContractsHandler`, HTTP handlers, `GET /api/v1/missions/offers` + `GET /api/v1/missions/contracts` routes in `server.go` and `template.yaml`.
- Coverage gate: zero uncovered blocks in affected scope (filters, handlers, stores, HTTP). All files < 150 lines. gofmt/vet clean.
- Full unit + integration suites green against DynamoDB local.

## First RED

- `go test ./tests/unit/missions/domain/ -run OfferListFilter` → RED: `ParseOfferListFilter` undefined.
- `go test ./tests/unit/missions/application/ -run ListOffers` → RED: `ListOffersHandler` undefined.
