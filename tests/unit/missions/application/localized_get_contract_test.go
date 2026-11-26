package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedContractGetter is a hand-written ContractReader double.
type localizedContractGetter struct {
	contract *domain.MissionContract
	err      error
	calls    int
}

func (g *localizedContractGetter) Get(string) (*domain.MissionContract, error) {
	g.calls++
	return g.contract, g.err
}

func bilingualDetailOffer() domain.MissionOffer {
	offer := domain.NewMissionOfferWithMode("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
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

// TEST-APP-LANG-001 (GetContract partition): a roster member reads the
// bilingual contract detail in English — offer text, Terms objectives and
// phases all translated, one honest contract-level label.
func TestGetContractLocalized_MemberGetsEnglishDetail(t *testing.T) {
	// Arrange
	offer := bilingualDetailOffer()
	terms, termsErr := domain.NewContractTerms(offer.Objectives)
	offer.Objectives = nil
	contract := *domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Terms = terms
	member, memberErr := domain.NewCallerIdentity("player-1", nil, nil)
	handler := application.NewGetContractHandler(&localizedContractGetter{contract: &contract})

	// Act
	served, err := handler.Handle(member, "contract-1", domain.LanguageEnglish)

	// Assert
	require.NoError(t, termsErr)
	require.NoError(t, memberErr)
	require.NoError(t, err)
	require.NotNil(t, served)
	assert.Equal(t, "Hunt", served.Offer.Title)
	assert.Equal(t, []string{"First rule"}, served.Offer.Rules)
	require.Len(t, served.Offer.Objectives, 1)
	assert.Equal(t, "obj-1", served.Offer.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", served.Offer.Objectives[0].Name)
	termsObjectives := served.Terms.Objectives()
	require.Len(t, termsObjectives, 1)
	assert.Equal(t, "Capture", termsObjectives[0].Name)
	assert.Equal(t, "English objective", termsObjectives[0].Description)
	assert.Equal(t, 10, termsObjectives[0].MeritPerUnit)
	phases := served.Terms.Phases()
	require.Len(t, phases, 1)
	require.Len(t, phases[0].Mandatory, 1)
	assert.Equal(t, "Capture", phases[0].Mandatory[0].Name)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetContract): an es-only contract stays Spanish.
func TestGetContractLocalized_LegacySpanishUnderEnglish(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-2", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Misión legada"
	contract := *domain.NewMissionContract("contract-2", offer, []string{"player-1"})
	member, memberErr := domain.NewCallerIdentity("player-1", nil, nil)
	handler := application.NewGetContractHandler(&localizedContractGetter{contract: &contract})

	// Act
	served, err := handler.Handle(member, "contract-2", domain.LanguageEnglish)

	// Assert
	require.NoError(t, memberErr)
	require.NoError(t, err)
	require.NotNil(t, served)
	assert.Equal(t, "Misión legada", served.Offer.Title)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetContract): an outsider is denied even with lang=en.
func TestGetContractLocalized_OutsiderDenied(t *testing.T) {
	// Arrange
	contract := *domain.NewMissionContract("contract-1", bilingualDetailOffer(), []string{"player-1"})
	outsider, outsiderErr := domain.NewCallerIdentity("player-9", nil, nil)
	handler := application.NewGetContractHandler(&localizedContractGetter{contract: &contract})

	// Act
	_, err := handler.Handle(outsider, "contract-1", domain.LanguageEnglish)

	// Assert
	require.NoError(t, outsiderErr)
	assert.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

// TEST-APP-LANG-001 (GetContract partition): a reviewer outside the roster
// still reads the translated detail.
func TestGetContractLocalized_ReviewerGetsEnglishDetail(t *testing.T) {
	// Arrange
	contract := *domain.NewMissionContract("contract-1", bilingualDetailOffer(), []string{"player-1"})
	reviewer, reviewerErr := domain.NewCallerIdentity("mod-1", []string{"moderator"}, nil)
	handler := application.NewGetContractHandler(&localizedContractGetter{contract: &contract})

	// Act
	served, err := handler.Handle(reviewer, "contract-1", domain.LanguageEnglish)

	// Assert
	require.NoError(t, reviewerErr)
	require.NoError(t, err)
	require.NotNil(t, served)
	assert.Equal(t, "Hunt", served.Offer.Title)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetContract partition): a reader failure propagates.
func TestGetContractLocalized_PropagatesReadError(t *testing.T) {
	// Arrange
	boom := errors.New("store down")
	member, memberErr := domain.NewCallerIdentity("player-1", nil, nil)
	handler := application.NewGetContractHandler(&localizedContractGetter{err: boom})

	// Act
	served, err := handler.Handle(member, "contract-1", domain.LanguageEnglish)

	// Assert
	require.NoError(t, memberErr)
	assert.Nil(t, served)
	assert.ErrorIs(t, err, boom)
}
