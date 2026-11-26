//go:build integration

package dynamodb_test

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedBilingualContract freezes English terms from a published
// bilingual definition onto a contract the roster member can read.
func localizedBilingualContract() (*domain.MissionContract, error) {
	req := localizedDefinitionRequest()
	definition, err := domain.NewMissionDefinition(req.DefinitionCode, req.Title, req.Description, req.Category, req.Mode, req.CompletionReward, req.Objectives, req.Rules)
	if err != nil {
		return nil, err
	}
	definition.ContentEn = req.ContentEn.Clone()
	if err := definition.Publish("v1"); err != nil {
		return nil, err
	}
	offer, err := domain.NewMissionOfferFromDefinition("offer-rescue", "season-1", definition)
	if err != nil {
		return nil, err
	}
	terms, err := domain.NewContractTerms(offer.Objectives)
	if err != nil {
		return nil, err
	}
	contract := domain.NewMissionContract("contract-rescue", offer, []string{"player-1"})
	contract.Terms = terms
	return contract, nil
}

// localizedLegacyContract models a Spanish-only persisted contract: no
// English snapshot anywhere.
func localizedLegacyContract() (*domain.MissionContract, error) {
	offer := domain.NewMissionOffer("offer-old", "BOUNTY_HUNTING", "season-1")
	offer.Title = "Rescate antiguo"
	offer.Description = "Salvar al piloto"
	offer.Rules = []string{"No ayudas"}
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Capturar al piloto", "BOUNTY_HUNTING", 10, 0, 2),
	}
	terms, err := domain.NewContractTerms(offer.Objectives)
	if err != nil {
		return nil, err
	}
	contract := domain.NewMissionContract("contract-old", offer, []string{"player-1"})
	contract.Terms = terms
	return contract, nil
}
