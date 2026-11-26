package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-LANG-001 (ListPlayerContracts partition): phased contract terms
// translate through the list read — every phase's mandatory and optional
// objectives carry English text while membership and order stay intact.
func TestListPlayerContractsLocalized_ProjectsPhasedTermsEnglish(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería"
	offer.Description = "Descripción"
	offer.Rules = []string{"Regla uno"}
	objectives := []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Primero", "Primer objetivo", "BOUNTY", 10, 5, 3),
		domain.NewObjectiveNamed("obj-2", "Segundo", "Segundo objetivo", "BOUNTY", 20, 0, 1),
	}
	terms, termsErr := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: objectives[:1]},
		{Mandatory: objectives[1:]},
	})
	offer.Objectives = nil
	offer.ContentEn = domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"First rule"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "First", Description: "First objective"},
			{ObjectiveID: "obj-2", Name: "Second", Description: "Second objective"},
		},
	}
	contract := *domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Terms = terms
	handler := application.NewListPlayerContractsHandler(&playerContractsReaderStub{
		contracts: []domain.MissionContract{contract},
	})

	// Act
	contracts, err := handler.Handle("player-1", domain.ContractListFilter{}, domain.LanguageEnglish)

	// Assert
	require.NoError(t, termsErr)
	require.NoError(t, err)
	require.Len(t, contracts, 1)
	phases := contracts[0].Terms.Phases()
	require.Len(t, phases, 2)
	assert.Equal(t, "obj-1", phases[0].Mandatory[0].ObjectiveID)
	assert.Equal(t, "First", phases[0].Mandatory[0].Name)
	assert.Equal(t, "obj-2", phases[1].Mandatory[0].ObjectiveID)
	assert.Equal(t, "Second", phases[1].Mandatory[0].Name)
	assert.Equal(t, domain.LanguageEnglish, contracts[0].ContentLanguage)
}

// TEST-APP-LANG-001 (ListPlayerContracts partition): the served offer
// projection copies rules and objectives — mutating it must not rewrite
// the stored contract's Spanish or English content.
func TestListPlayerContractsLocalized_ProjectionDoesNotAliasSource(t *testing.T) {
	// Arrange
	contract := *domain.NewMissionContract("contract-1", bilingualContractOffer("offer-1"), []string{"player-1"})
	handler := application.NewListPlayerContractsHandler(&playerContractsReaderStub{
		contracts: []domain.MissionContract{contract},
	})

	// Act
	contracts, err := handler.Handle("player-1", domain.ContractListFilter{}, domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, contracts, 1)
	contracts[0].Offer.Rules[0] = "Mutated served rule"
	contracts[0].Offer.Objectives[0].Name = "Mutated served name"
	assert.Equal(t, []string{"Regla uno"}, contract.Offer.Rules)
	assert.Equal(t, "Capturar", contract.Offer.Objectives[0].Name)
	assert.Equal(t, []string{"First rule"}, contract.Offer.ContentEn.Rules)
	assert.Equal(t, "Capture", contract.Offer.ContentEn.Objectives[0].Name)
}
