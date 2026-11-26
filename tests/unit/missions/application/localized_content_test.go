package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedOfferReaderStub is a hand-written OfferReader double for the
// language-selection partition: it returns the stored offer untouched.
type localizedOfferReaderStub struct {
	offer domain.MissionOffer
	err   error
}

func (s *localizedOfferReaderStub) Get(offerID string) (domain.MissionOffer, error) {
	return s.offer, s.err
}

// TEST-APP-LANG-001 (GetOffer partition): a bilingual offer serves its
// English content declared as English when the member chooses English.
func TestGetOfferLocalized_ServesEnglishWhenRequested(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería de recompensas"
	offer.Description = "Descripción en **Markdown**"
	offer.ContentEn = domain.LocalizedContent{Title: "Bounty hunt", Description: "English **Markdown** briefing"}
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-1", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Bounty hunt", served.Title)
	assert.Equal(t, "English **Markdown** briefing", served.Description)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetOffer partition): a Spanish-only record serves its
// Spanish content honestly declared as Spanish under an English request.
func TestGetOfferLocalized_LegacySpanishDeclaredUnderEnglishRequest(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-2", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Misión legada"
	offer.Description = "Contenido solo en español"
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-2", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Misión legada", served.Title)
	assert.Equal(t, "Contenido solo en español", served.Description)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): a fully
// translated offer serves English rules and objective text while preserving
// objective identity and economics, declared as English.
func TestGetOfferLocalized_ServesEnglishRulesAndObjectives(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-3", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
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
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"},
		},
	}
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-3", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Hunt", served.Title)
	assert.Equal(t, "English briefing", served.Description)
	assert.Equal(t, []string{"First rule"}, served.Rules)
	require.Len(t, served.Objectives, 1)
	assert.Equal(t, "obj-1", served.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", served.Objectives[0].Name)
	assert.Equal(t, "English objective", served.Objectives[0].Description)
	assert.Equal(t, 10, served.Objectives[0].MeritPerUnit)
	assert.Equal(t, 5, served.Objectives[0].Bonus)
	assert.Equal(t, 3, served.Objectives[0].UnitLimit)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): an offer
// whose English content covers title and description but not rules or
// objectives falls back as a whole to Spanish, never serving a mixed locale.
func TestGetOfferLocalized_FallsBackToSpanishWhenEnglishIncomplete(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-4", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería"
	offer.Description = "Descripción"
	offer.Rules = []string{"Regla uno"}
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	offer.ContentEn = domain.LocalizedContent{Title: "Hunt", Description: "English briefing"}
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-4", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Cacería", served.Title)
	assert.Equal(t, "Descripción", served.Description)
	assert.Equal(t, []string{"Regla uno"}, served.Rules)
	require.Len(t, served.Objectives, 1)
	assert.Equal(t, "obj-1", served.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capturar", served.Objectives[0].Name)
	assert.Equal(t, "Objetivo en español", served.Objectives[0].Description)
	assert.Equal(t, domain.LanguageSpanish, served.ContentLanguage)
}
