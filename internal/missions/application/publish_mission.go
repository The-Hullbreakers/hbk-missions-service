package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// OfferWriter persists a published mission offer.
type OfferWriter interface {
	Save(offer domain.MissionOffer) error
}

// PublishMissionHandler publishes a new mission offer from a published definition.
type PublishMissionHandler struct {
	offers        OfferWriter
	definitions   DefinitionReader
	activeSeasons ActiveSeasonsReader
}

// NewPublishMissionHandler creates a handler with the required ports.
func NewPublishMissionHandler(offers OfferWriter, definitions DefinitionReader, activeSeasons ActiveSeasonsReader) *PublishMissionHandler {
	return &PublishMissionHandler{offers: offers, definitions: definitions, activeSeasons: activeSeasons}
}

// Handle creates and persists a mission offer bound to the active season from a published definition.
func (h *PublishMissionHandler) Handle(offerID, definitionCode, version string) (domain.MissionOffer, error) {
	active, err := h.activeSeasons.ListActive()
	if err != nil {
		return domain.MissionOffer{}, err
	}
	if len(active) == 0 {
		return domain.MissionOffer{}, domain.ErrNoActiveSeason
	}
	if len(active) > 1 {
		return domain.MissionOffer{}, domain.ErrActiveSeasonAlreadyExists
	}

	var definition domain.MissionDefinition
	if version != "" {
		definition, err = h.definitions.GetVersion(definitionCode, version)
	} else {
		definition, err = h.definitions.Get(definitionCode)
	}
	if err != nil {
		return domain.MissionOffer{}, err
	}

	offer, err := domain.NewMissionOfferFromDefinition(offerID, active[0].SeasonID, definition)
	if err != nil {
		return domain.MissionOffer{}, err
	}
	if err := h.offers.Save(offer); err != nil {
		return domain.MissionOffer{}, err
	}

	return offer, nil
}
