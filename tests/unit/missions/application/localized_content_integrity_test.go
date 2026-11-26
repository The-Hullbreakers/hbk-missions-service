package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// integrityOfferFixture builds a bilingual offer whose Spanish content is
// complete; tests vary ContentEn to exercise one defect at a time.
func integrityOfferFixture() domain.MissionOffer {
	offer := domain.NewMissionOfferWithMode("offer-int", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería"
	offer.Description = "Descripción"
	offer.Rules = []string{"Regla uno"}
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	return offer
}

// completeEnglish returns a fully translated ContentEn for the fixture
// offer; defect cases modify one aspect in their Arrange step.
func completeEnglish() domain.LocalizedContent {
	return domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"First rule"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"},
		},
	}
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): English is
// complete only when rules and every objective translation exist and are
// non-blank and ID-matched; any defect falls back atomically to Spanish.
func TestGetOfferLocalized_IncompleteEnglishFallsBackAtomically(t *testing.T) {
	tests := []struct {
		name      string
		contentEn domain.LocalizedContent
	}{
		{
			name: "translated rules count differs from Spanish rules",
			contentEn: domain.LocalizedContent{
				Title:       "Hunt",
				Description: "English briefing",
				Rules:       []string{"First rule", "Unexpected extra rule"},
				Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"}},
			},
		},
		{
			name: "a translated rule is blank",
			contentEn: domain.LocalizedContent{
				Title:       "Hunt",
				Description: "English briefing",
				Rules:       []string{"   "},
				Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"}},
			},
		},
		{
			name: "an objective English description is blank while Spanish has one",
			contentEn: domain.LocalizedContent{
				Title:       "Hunt",
				Description: "English briefing",
				Rules:       []string{"First rule"},
				Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: ""}},
			},
		},
		{
			name: "an English objective ID does not match a Spanish objective",
			contentEn: domain.LocalizedContent{
				Title:       "Hunt",
				Description: "English briefing",
				Rules:       []string{"First rule"},
				Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-unknown", Name: "Capture", Description: "English objective"}},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			offer := integrityOfferFixture()
			offer.ContentEn = tc.contentEn
			handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

			// Act
			served, err := handler.Handle("offer-int", domain.LanguageEnglish)

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
		})
	}
}

// TEST-APP-LANG-001 (GetOffer partition, whole-offer integrity): English
// objective translations are matched by ObjectiveID, not position — a
// reordered translation list renders in base objective order.
func TestGetOfferLocalized_MatchesEnglishObjectivesByID(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-ord", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería"
	offer.Description = "Descripción"
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Primero", "Primer objetivo", "BOUNTY", 10, 5, 3),
		domain.NewObjectiveNamed("obj-2", "Segundo", "Segundo objetivo", "BOUNTY", 20, 0, 1),
	}
	offer.ContentEn = domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-2", Name: "Second", Description: "Second objective"},
			{ObjectiveID: "obj-1", Name: "First", Description: "First objective"},
		},
	}
	handler := application.NewGetOfferHandler(&localizedOfferReaderStub{offer: offer})

	// Act
	served, err := handler.Handle("offer-ord", domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, served.Objectives, 2)
	assert.Equal(t, "obj-1", served.Objectives[0].ObjectiveID)
	assert.Equal(t, "First", served.Objectives[0].Name)
	assert.Equal(t, "First objective", served.Objectives[0].Description)
	assert.Equal(t, 10, served.Objectives[0].MeritPerUnit)
	assert.Equal(t, "obj-2", served.Objectives[1].ObjectiveID)
	assert.Equal(t, "Second", served.Objectives[1].Name)
	assert.Equal(t, "Second objective", served.Objectives[1].Description)
	assert.Equal(t, 20, served.Objectives[1].MeritPerUnit)
	assert.Equal(t, domain.LanguageEnglish, served.ContentLanguage)
}
