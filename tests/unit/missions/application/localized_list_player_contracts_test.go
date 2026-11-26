package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func bilingualContractOffer(offerID string) domain.MissionOffer {
	offer := domain.NewMissionOfferWithMode(offerID, "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería"
	offer.Description = "Descripción"
	offer.Rules = []string{"Regla uno"}
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	offer.ContentEn = domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"First rule"},
		Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"}},
	}
	return offer
}

func legacyContractOffer(offerID string) domain.MissionOffer {
	offer := domain.NewMissionOfferWithMode(offerID, "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Misión legada"
	offer.Description = "Contenido solo en español"
	return offer
}

// TEST-APP-LANG-001 (ListPlayerContracts partition): a player's mixed
// contract list serves each contract's embedded offer in its effective
// language with one honest contract-level label, in reader order, with
// ids and economics unchanged — and the outsider's contract stays hidden.
func TestListPlayerContractsLocalized_ServesPerItemLanguage(t *testing.T) {
	// Arrange
	reader := &playerContractsReaderStub{contracts: []domain.MissionContract{
		*domain.NewMissionContract("contract-1", bilingualContractOffer("offer-1"), []string{"player-1"}),
		*domain.NewMissionContract("contract-2", legacyContractOffer("offer-2"), []string{"player-1"}),
		*domain.NewMissionContract("contract-3", legacyContractOffer("offer-3"), []string{"player-9"}),
	}}
	handler := application.NewListPlayerContractsHandler(reader)

	// Act
	contracts, err := handler.Handle("player-1", domain.ContractListFilter{}, domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "player-1", reader.gotPlayer)
	require.Len(t, contracts, 2)
	assert.Equal(t, "contract-1", contracts[0].ContractID)
	assert.Equal(t, "Hunt", contracts[0].Offer.Title)
	assert.Equal(t, "English briefing", contracts[0].Offer.Description)
	assert.Equal(t, 10, contracts[0].Offer.Objectives[0].MeritPerUnit)
	assert.Equal(t, domain.LanguageEnglish, contracts[0].ContentLanguage)
	assert.Equal(t, "contract-2", contracts[1].ContractID)
	assert.Equal(t, "Misión legada", contracts[1].Offer.Title)
	assert.Equal(t, domain.LanguageSpanish, contracts[1].ContentLanguage)
}

// TEST-APP-LANG-001 (ListPlayerContracts partition): a persisted contract
// whose offer carries no objectives still serves the frozen Terms
// objectives — projected in the effective language on both the offer view
// and Terms.Objectives(), economics identical.
func TestListPlayerContractsLocalized_ProjectsEnglishTermsObjectives(t *testing.T) {
	// Arrange
	offer := bilingualContractOffer("offer-1")
	terms, termsErr := domain.NewContractTerms(offer.Objectives)
	offer.Objectives = nil
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
	require.Len(t, contracts[0].Offer.Objectives, 1)
	assert.Equal(t, "obj-1", contracts[0].Offer.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", contracts[0].Offer.Objectives[0].Name)
	assert.Equal(t, "English objective", contracts[0].Offer.Objectives[0].Description)
	termsObjectives := contracts[0].Terms.Objectives()
	require.Len(t, termsObjectives, 1)
	assert.Equal(t, "Capture", termsObjectives[0].Name)
	assert.Equal(t, "English objective", termsObjectives[0].Description)
	assert.Equal(t, 10, termsObjectives[0].MeritPerUnit)
	assert.Equal(t, 5, termsObjectives[0].Bonus)
	assert.Equal(t, 3, termsObjectives[0].UnitLimit)
	assert.Equal(t, domain.LanguageEnglish, contracts[0].ContentLanguage)
}

// TEST-APP-LANG-001 (ListPlayerContracts partition): an offer whose English
// snapshot does not cover the frozen Terms objectives falls the whole
// contract back to Spanish — offer text, terms text and label agree.
func TestListPlayerContractsLocalized_IncompleteEnglishFallsBack(t *testing.T) {
	// Arrange
	offer := bilingualContractOffer("contract-1")
	terms, termsErr := domain.NewContractTerms(offer.Objectives)
	offer.Objectives = nil
	offer.ContentEn.Objectives = nil
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
	assert.Equal(t, "Cacería", contracts[0].Offer.Title)
	termsObjectives := contracts[0].Terms.Objectives()
	require.Len(t, termsObjectives, 1)
	assert.Equal(t, "Capturar", termsObjectives[0].Name)
	assert.Equal(t, domain.LanguageSpanish, contracts[0].ContentLanguage)
}
