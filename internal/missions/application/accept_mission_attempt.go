package application

import (
	"fmt"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// nextAttempt mints the attempt-unique contract id and freezes the accepted
// definition version's objectives into ContractTerms. Offers bound to a
// definition freeze its objectives; definition-less offers keep empty terms.
func (h *AcceptMissionHandler) nextAttempt(offer domain.MissionOffer, playerID string) (string, domain.ContractTerms, error) {
	prior, err := h.history.ListByPlayer(playerID, domain.ContractListFilter{Status: domain.ContractStatusAll})
	if err != nil {
		return "", domain.ContractTerms{}, err
	}
	attempt := 1
	for _, c := range prior {
		if c.Offer.OfferID == offer.OfferID {
			attempt++
		}
	}
	contractID := fmt.Sprintf("%s:%s:%d", offer.OfferID, playerID, attempt)

	if offer.DefinitionCode == "" || offer.DefinitionVersion == "legacy" {
		return contractID, domain.ContractTerms{}, nil
	}
	definition, err := h.definitions.GetVersion(offer.DefinitionCode, offer.DefinitionVersion)
	if err != nil {
		return "", domain.ContractTerms{}, err
	}
	terms, err := domain.NewContractTerms(definition.Objectives)
	if err != nil {
		return "", domain.ContractTerms{}, err
	}
	return contractID, terms, nil
}
