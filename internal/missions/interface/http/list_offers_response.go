package http

import "github.com/hbk/hbk-missions-service/internal/missions/application"

// offersResponse preserves the lower-case list envelope while each item
// keeps the offer wire shape plus the content language actually served.
type offersResponse struct {
	Offers []offerResponse `json:"offers"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

func toOffersResponse(page application.LocalizedOfferListPage) offersResponse {
	offers := make([]offerResponse, 0, len(page.Offers))
	for _, offer := range page.Offers {
		offers = append(offers, toOfferResponse(offer))
	}
	return offersResponse{Offers: offers, Total: page.Total, Limit: page.Limit, Offset: page.Offset}
}
