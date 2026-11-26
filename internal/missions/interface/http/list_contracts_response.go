package http

import (
	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// contractListResponse retains the existing contract fields and nested offer
// merit breakdown while adding the language served for the entire contract.
type contractListResponse struct {
	ContractID      string
	Offer           domain.MissionOffer
	Terms           domain.ContractTerms
	State           domain.ContractState
	SeasonID        string
	TeamID          string
	ContentLanguage domain.Language
}

func toContractListResponse(contract application.LocalizedContract) contractListResponse {
	offer := contract.Offer
	if len(offer.Objectives) == 0 {
		// Legacy contract JSON projects objectives from frozen terms.
		offer.Objectives = contract.Terms.Objectives()
	}
	return contractListResponse{
		ContractID:      contract.ContractID,
		Offer:           offer,
		Terms:           contract.Terms,
		State:           contract.State,
		SeasonID:        contract.SeasonID,
		TeamID:          contract.TeamID,
		ContentLanguage: contract.ContentLanguage,
	}
}
