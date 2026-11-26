//go:build integration

package http_test

import (
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// languageContractsServer seeds contracts for the admin caller's roster:
// a bilingual accepted contract, a legacy Spanish-only contract, and an
// outsider contract that must stay filtered out.
func languageContractsServer(t *testing.T, bilingual, legacy bool) (*httptest.Server, error) {
	return languageHTTPServer(t, func(client *missiondynamo.DynamoDBClient, table string) error {
		store := missiondynamo.NewContractStore(client, table)
		if bilingual {
			contract, err := bilingualContractFixture([]string{"admin-1"})
			if err != nil {
				return err
			}
			if err := store.Save(contract); err != nil {
				return err
			}
		}
		if legacy {
			legacyContract, err := legacyContractFixture([]string{"admin-1"})
			if err != nil {
				return err
			}
			if err := store.Save(legacyContract); err != nil {
				return err
			}
			outsider := domain.NewMissionContract("contract-other", legacyContract.Offer, []string{"player-2"})
			outsider.Terms = legacyContract.Terms
			if err := store.Save(outsider); err != nil {
				return err
			}
		}
		return nil
	})
}

// bilingualContractFixture builds one accepted contract (contract-rescue)
// frozen from a bilingual published definition for the given roster.
func bilingualContractFixture(roster []string) (*domain.MissionContract, error) {
	definition, err := domain.NewMissionDefinition("def-rescue", "Rescate", "Salvar al piloto",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 50,
		[]domain.Objective{
			domain.NewObjectiveNamed("obj-1", "Capturar", "Capturar al piloto", "BOUNTY_HUNTING", 10, 0, 2),
		}, []string{"No ayudas"})
	if err != nil {
		return nil, err
	}
	definition.ContentEn = domain.LocalizedContent{
		Title: "Rescue", Description: "Save the pilot", Rules: []string{"No aids"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the pilot"},
		},
	}
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
	contract := domain.NewMissionContract("contract-rescue", offer, roster)
	contract.Terms = terms
	return contract, nil
}

// legacyContractFixture builds one Spanish-only contract; errors returned.
func legacyContractFixture(roster []string) (*domain.MissionContract, error) {
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
	contract := domain.NewMissionContract("contract-old", offer, roster)
	contract.Terms = terms
	return contract, nil
}
