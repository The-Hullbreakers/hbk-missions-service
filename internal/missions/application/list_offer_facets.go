package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// ListOfferFacetsHandler returns the distinct category and mode values present
// for the resolved season and status, to populate board filter pickers.
type ListOfferFacetsHandler struct {
	offers  OfferListReader
	seasons ActiveSeasonsReader
}

// NewListOfferFacetsHandler creates a handler with the required reader ports.
func NewListOfferFacetsHandler(offers OfferListReader, seasons ActiveSeasonsReader) *ListOfferFacetsHandler {
	return &ListOfferFacetsHandler{offers: offers, seasons: seasons}
}

// Handle resolves the season like the offers list and collects facets over the
// matching offers. Search, paging, and sort do not apply.
func (h *ListOfferFacetsHandler) Handle(filter domain.OfferListFilter) (domain.OfferFacets, error) {
	if filter.SeasonID == "" {
		active, err := h.seasons.ListActive()
		if err != nil {
			return domain.OfferFacets{}, err
		}
		if len(active) == 0 {
			return domain.OfferFacets{}, domain.ErrNoActiveSeason
		}
		if len(active) > 1 {
			return domain.OfferFacets{}, domain.ErrActiveSeasonAlreadyExists
		}
		filter.SeasonID = active[0].SeasonID
	}
	filter.Q = ""
	filter.Page = domain.OfferPage{}
	filter.Sort = domain.OfferSort{}

	offers, err := h.offers.List(filter)
	if err != nil {
		return domain.OfferFacets{}, err
	}
	return domain.CollectFacets(offers), nil
}
