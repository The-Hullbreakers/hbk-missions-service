package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-LANG-002: a published bilingual definition freezes its English
// rules and objective translations onto the offer alongside the Spanish
// scalars and economics.
func TestNewMissionOfferFromDefinition_FreezesEnglishTerms(t *testing.T) {
	// Arrange
	objectives := []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	definition, definitionErr := domain.NewMissionDefinition(
		"BH-01", "Cacería", "Descripción", "BOUNTY_HUNTING", domain.MissionModeSolo, 100, objectives, []string{"Regla uno"},
	)
	definition.ContentEn = domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"First rule"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"},
		},
	}
	publishErr := definition.Publish("v1")

	// Act
	offer, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.NoError(t, definitionErr)
	require.NoError(t, publishErr)
	require.NoError(t, err)
	assert.Equal(t, []string{"First rule"}, offer.ContentEn.Rules)
	require.Len(t, offer.ContentEn.Objectives, 1)
	assert.Equal(t, "obj-1", offer.ContentEn.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", offer.ContentEn.Objectives[0].Name)
	assert.Equal(t, "English objective", offer.ContentEn.Objectives[0].Description)
	assert.Equal(t, "Cacería", offer.Title)
	assert.Equal(t, []string{"Regla uno"}, offer.Rules)
	require.Len(t, offer.Objectives, 1)
	assert.Equal(t, "obj-1", offer.Objectives[0].ObjectiveID)
	assert.Equal(t, "Capturar", offer.Objectives[0].Name)
	assert.Equal(t, 10, offer.Objectives[0].MeritPerUnit)
	assert.Equal(t, 5, offer.Objectives[0].Bonus)
	assert.Equal(t, 3, offer.Objectives[0].UnitLimit)
}

// TEST-DOM-LANG-002: the frozen English snapshot must not alias the
// definition's slices; later edits to the source cannot rewrite a published
// offer's frozen content.
func TestNewMissionOfferFromDefinition_DoesNotAliasEnglishTerms(t *testing.T) {
	// Arrange
	objectives := []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	definition, definitionErr := domain.NewMissionDefinition(
		"BH-01", "Cacería", "Descripción", "BOUNTY_HUNTING", domain.MissionModeSolo, 100, objectives, []string{"Regla uno"},
	)
	definition.ContentEn = domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"First rule"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"},
		},
	}
	publishErr := definition.Publish("v1")

	// Act
	offer, err := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)

	// Assert
	require.NoError(t, definitionErr)
	require.NoError(t, publishErr)
	require.NoError(t, err)
	definition.ContentEn.Rules[0] = "Mutated rule"
	definition.ContentEn.Objectives[0].Name = "Mutated"
	assert.Equal(t, []string{"First rule"}, offer.ContentEn.Rules)
	assert.Equal(t, "Capture", offer.ContentEn.Objectives[0].Name)
}
