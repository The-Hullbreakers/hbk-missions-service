package http

import (
	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// offerResponse preserves the existing offer wire shape and derived merits
// while exposing only the content language actually served.
type offerResponse struct {
	OfferID           string
	Title             string
	Description       string
	Category          string
	Mode              domain.MissionMode
	CompletionReward  int
	Objectives        []domain.Objective
	Rules             []string
	SeasonID          string
	DefinitionCode    string
	DefinitionVersion string
	MeritSummary      domain.MeritSummary
	ContentLanguage   domain.Language
}

func toOfferResponse(offer application.LocalizedOffer) offerResponse {
	return offerResponse{
		OfferID:           offer.OfferID,
		Title:             offer.Title,
		Description:       offer.Description,
		Category:          offer.Category,
		Mode:              offer.Mode,
		CompletionReward:  offer.CompletionReward,
		Objectives:        offer.Objectives,
		Rules:             offer.Rules,
		SeasonID:          offer.SeasonID,
		DefinitionCode:    offer.DefinitionCode,
		DefinitionVersion: offer.DefinitionVersion,
		MeritSummary:      offer.MissionOffer.MeritSummary(),
		ContentLanguage:   offer.ContentLanguage,
	}
}
