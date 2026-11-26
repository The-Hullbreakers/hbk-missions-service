package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-LANG-001: Language parses only the supported codes.
func TestParseLanguage_SupportedCodes(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want domain.Language
	}{
		{"spanish", "es", domain.LanguageSpanish},
		{"english", "en", domain.LanguageEnglish},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			raw := tc.raw

			// Act
			got, err := domain.ParseLanguage(raw)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TEST-DOM-LANG-001: empty and unsupported codes are rejected.
func TestParseLanguage_RejectsUnsupported(t *testing.T) {
	for _, raw := range []string{"", "fr"} {
		t.Run("input "+raw, func(t *testing.T) {
			// Arrange
			input := raw

			// Act
			_, err := domain.ParseLanguage(input)

			// Assert
			require.ErrorIs(t, err, domain.ErrLanguageInvalid)
		})
	}
}

// TEST-DOM-LANG-002 (GetOffer partition): an offer frozen from a bilingual
// definition retains the English text alongside the Spanish scalars.
func TestNewMissionOfferFromDefinition_FreezesEnglishContent(t *testing.T) {
	// Arrange
	definition, definitionErr := domain.NewMissionDefinition("def-1", "Cacería", "Descripción **Markdown**", "BOUNTY_HUNTING", domain.MissionModeSolo, 100, nil, nil)
	definition.ContentEn = domain.LocalizedContent{Title: "Bounty hunt", Description: "English **Markdown** briefing"}
	publishErr := definition.Publish("v1")

	// Act
	offer, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.NoError(t, definitionErr)
	require.NoError(t, publishErr)
	require.NoError(t, err)
	assert.Equal(t, "Bounty hunt", offer.ContentEn.Title)
	assert.Equal(t, "English **Markdown** briefing", offer.ContentEn.Description)
	assert.Equal(t, "Cacería", offer.Title)
	assert.Equal(t, "Descripción **Markdown**", offer.Description)
}

// TEST-DOM-LANG-002 (wire-shape regression): the legacy offer projection
// must not expose the English snapshot before Boundary work — locale
// selection owns what the wire serves.
func TestMissionOfferMarshalJSON_HidesEnglishContent(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOfferWithMode("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	offer.Title = "Cacería de recompensas"
	offer.Description = "Descripción en **Markdown**"
	offer.ContentEn = domain.LocalizedContent{Title: "Bounty hunt", Description: "English **Markdown** briefing"}

	// Act
	raw, err := json.Marshal(offer)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, string(raw), "Cacería de recompensas")
	assert.NotContains(t, string(raw), "Bounty hunt")
	assert.NotContains(t, string(raw), "ContentEn")
}

// TEST-DOM-LANG-002 (wire-shape regression): the definition projection
// must not expose the English snapshot before Boundary work either.
func TestMissionDefinitionMarshalJSON_HidesEnglishContent(t *testing.T) {
	// Arrange
	definition, definitionErr := domain.NewMissionDefinition("def-1", "Cacería", "Descripción en **Markdown**", "BOUNTY_HUNTING", domain.MissionModeSolo, 100, nil, nil)
	definition.ContentEn = domain.LocalizedContent{Title: "Bounty hunt", Description: "English **Markdown** briefing"}

	// Act
	raw, err := json.Marshal(definition)

	// Assert
	require.NoError(t, definitionErr)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "Cacería")
	assert.NotContains(t, string(raw), "Bounty hunt")
	assert.NotContains(t, string(raw), "ContentEn")
}
