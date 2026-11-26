package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// GetOfferHandler serves the single-offer read; lookup failures propagate the
// domain sentinel for the boundary to map.
type GetOfferHandler struct {
	offers OfferReader
}

// NewGetOfferHandler creates a handler bound to the offer reader port.
func NewGetOfferHandler(offers OfferReader) *GetOfferHandler {
	return &GetOfferHandler{offers: offers}
}

// Handle returns the published offer for the given id projected in the
// requested language; lookup failures propagate the domain sentinel.
func (h *GetOfferHandler) Handle(offerID string, language domain.Language) (LocalizedOffer, error) {
	offer, err := h.offers.Get(offerID)
	if err != nil {
		return LocalizedOffer{}, err
	}
	return LocalizeOffer(offer, language), nil
}
