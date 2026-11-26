# Design — CHG-0045

## Domain (`internal/missions/domain/list_filters.go`)

- `OfferListQuery`: `Category`, `Mode` stay `string` (raw CSV); add `Q`,
  `Limit`, `Offset`, `Sort` as raw strings.
- `OfferListFilter`: `Categories []string`, `Modes []MissionMode`,
  `Q string`, `Page OfferPage` (`Limit`, `Offset`), `Sort OfferSort`
  (`Field`, `Descending`).
- `ParseOfferListFilter` splits CSV lists (upper-case, trim, reject empties),
  validates modes via `MissionMode.IsValid`, parses `limit`/`offset` with
  `strconv.Atoi` (limit default 50, clamp `[1,200]`; offset `>= 0`), and
  whitelists `sort` in `{completion_reward, category}` with `-` prefix.
  All rejections reuse existing sentinels (`ErrInvalidOfferModeFilter`,
  `ErrInvalidOfferStatusFilter`) plus `ErrInvalidOfferListFilter` → all map
  to `INVALID_QUERY_PARAMETER`.
- `Match` checks OR-membership for categories/modes and case-insensitive
  `strings.Contains` on `Title`+`Description` for `q`.
- `SortOffers(offers, sort)` orders by field then `OfferID` asc (stable).
- `PaginateOffers(offers, page)` returns `OfferListPage{Offers, Total,
  Limit, Offset}` slicing safely.
- `OfferFacets{Categories, Modes}` built by `CollectFacets(offers)`
  (distinct + sorted).

## Application

- `OfferListReader.List(filter)` unchanged signature; the store applies
  structural filters + returns all matches; sorting and paging happen in
  the handler after the `REQ-MB-008` exclusion so `total` is caller-visible.
- `ListOffersHandler.Handle` returns `domain.OfferListPage`.
- New `ListOfferFacetsHandler` (ports: `OfferListReader`,
  `ActiveSeasonsReader`) resolves the season like list, lists with
  `status` only, and returns `OfferFacets`.

## Infrastructure (`infrastructure/dynamodb/offer_store.go`)

- `List` keeps `Query` on `PK = OFFER` and builds a `FilterExpression`
  for `SeasonID`, `Category IN (...)`, `Mode IN (...)`; `q` stays
  post-filter in memory (substring over two attributes is cheap to keep
  in `Match`). Items map directly via a shared `mapOfferItem` — no
  per-item `GetItem` (removes N+1).
- `remaining_capacity > 0` is pushed into the `FilterExpression` for
  `AVAILABLE` (was post-filter on `Capacity`; note: store previously used
  `Capacity`, the seeded `remaining_capacity` is the truthful field).
- `ListFacets` reuses `List` with the facet filter and calls
  `CollectFacets`.

## Boundary (`interface/http/list_offers.go`)

- Handler reads the new query params, encodes `OfferListPage`.
- `GET /api/v1/missions/offers/facets` handler; route registered in the
  mux and composition `server.go`. Route order: `facets` is a literal
  segment so it does not conflict with `/offers/{offer_id}`.
- `errors.go`/`PublicCode`: new sentinel maps to `INVALID_QUERY_PARAMETER`.

## Risks / Rollback

- Envelope is breaking: rollback = revert commit; `hbk-web-app` ships the
  matching client change.
- `remaining_capacity` vs `Capacity` discrepancy documented; AVAILABLE
  semantics now use `remaining_capacity` (correct atomic counter).
