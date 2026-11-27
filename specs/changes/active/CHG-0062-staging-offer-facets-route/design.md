---
change_id: CHG-0062
artifact: design
status: implemented
updated: 2026-11-26
---

# Design — Staging offer facets route

## Domain / Application

- Not affected. The facets handler is already registered with
  `auth.Authenticate` in the composition mux alongside the other offer
  routes.

## Infrastructure (template.yaml)

- Additive `MissionsOfferFacetsRoute` immediately after
  `MissionsListOffersRoute` / before `MissionsListContractsRoute`, reusing
  `*RouteProps` for the API and Lambda integration — identical wiring to
  every neighboring route. Authentication stays in the Go route middleware.
- `facets` is a literal segment; API Gateway prefers literal matches over the
  parameterized `/offers/{offer_id}` route, so there is no shadowing.

## Test strategy

- TEST-STG-OFFER-FACETS-ROUTE-001 (`tools/test-workflow-contract.sh`):
  extracts the 5-line `MissionsOfferFacetsRoute:` block with `sed` and
  asserts the exact `Type`, `<<: *RouteProps`, and `RouteKey` lines — static
  template contract, no AWS calls.
- Focused Go evidence: `go test ./tests/unit/missions/application
  ./tests/unit/missions/interface/http -run '^TestListOfferFacets' -count=1`
  keeps the existing application and HTTP facets behavior green (no code
  touched).
- `sam validate --template-file template.yaml --region us-east-2` for basic
  template validity (known pre-existing `--lint` W1030 SNS warning not
  re-litigated).

## Security / cost

- No IAM changes; existing auth reused. One additional route mapping — cost
  neutral. Rollback: delete the route block.
