package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// OfferListReader lists offers matching a board filter.
type OfferListReader interface {
	List(filter domain.OfferListFilter) ([]domain.MissionOffer, error)
}

// PlayerContractsReader lists the contracts where the player is in the roster.
type PlayerContractsReader interface {
	ListByPlayer(playerID string, filter domain.ContractListFilter) ([]domain.MissionContract, error)
}

// ListOffersHandler orchestrates the mission-board discovery use case.
type ListOffersHandler struct {
	offers    OfferListReader
	contracts ActiveContractsReader
	seasons   ActiveSeasonsReader
}

// NewListOffersHandler creates a handler with the required reader ports.
func NewListOffersHandler(offers OfferListReader, contracts ActiveContractsReader, seasons ActiveSeasonsReader) *ListOffersHandler {
	return &ListOffersHandler{offers: offers, contracts: contracts, seasons: seasons}
}

// Handle returns the offers visible to the player in the requested
// language. An empty SeasonID resolves to the single active season;
// offers whose definition the player already has active are hidden per
// the mission-board exclusion rule. The text search applies to the
// effective served language, so Q is never forwarded to the store's
// Spanish-only prefilter — it is applied after visibility and before
// ordering and pagination, keeping Total honest.
func (h *ListOffersHandler) Handle(playerID string, filter domain.OfferListFilter, language domain.Language) (LocalizedOfferListPage, error) {
	if filter.SeasonID == "" {
		seasonID, err := h.resolveActiveSeason()
		if err != nil {
			return LocalizedOfferListPage{}, err
		}
		filter.SeasonID = seasonID
	}

	readerFilter := filter
	readerFilter.Q = ""
	offers, err := h.offers.List(readerFilter)
	if err != nil {
		return LocalizedOfferListPage{}, err
	}

	active, err := h.contracts.ListActiveByPlayer(playerID)
	if err != nil {
		return LocalizedOfferListPage{}, err
	}
	hidden := make(map[string]bool, len(active))
	for _, c := range active {
		definitionCode := c.Offer.DefinitionCode
		if definitionCode == "" {
			definitionCode = c.Offer.OfferID
		}
		hidden[definitionCode] = true
	}

	matched := make([]domain.MissionOffer, 0, len(offers))
	for _, offer := range offers {
		definitionCode := offer.DefinitionCode
		if definitionCode == "" {
			definitionCode = offer.OfferID
		}
		if hidden[definitionCode] {
			continue
		}
		if filter.Match(LocalizeOffer(offer, language).MissionOffer) {
			matched = append(matched, offer)
		}
	}
	page := domain.PaginateOffers(domain.SortOffers(matched, filter.Sort), filter.Page)
	items := make([]LocalizedOffer, 0, len(page.Offers))
	for _, offer := range page.Offers {
		items = append(items, LocalizeOffer(offer, language))
	}
	return LocalizedOfferListPage{
		Offers: items,
		Total:  page.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}, nil
}

func (h *ListOffersHandler) resolveActiveSeason() (string, error) {
	active, err := h.seasons.ListActive()
	if err != nil {
		return "", err
	}
	if len(active) == 0 {
		return "", domain.ErrNoActiveSeason
	}
	if len(active) > 1 {
		return "", domain.ErrActiveSeasonAlreadyExists
	}
	return active[0].SeasonID, nil
}
