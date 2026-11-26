package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// bilingualListDefinition is a published bilingual row as the store loads it.
func bilingualListDefinition() domain.MissionDefinition {
	return domain.MissionDefinition{
		DefinitionCode:   "def-bounty",
		Title:            "Cacería",
		Description:      "Descripción",
		Category:         "BOUNTY_HUNTING",
		Mode:             domain.MissionModeSolo,
		CompletionReward: 100,
		State:            domain.DefinitionStatePublished,
		Version:          "v1",
		Rules:            []string{"Regla uno"},
		Objectives: []domain.Objective{
			domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
		},
		ContentEn: domain.LocalizedContent{
			Title:       "Hunt",
			Description: "English briefing",
			Rules:       []string{"First rule"},
			Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "English objective"}},
		},
	}
}

func legacyListDefinition() domain.MissionDefinition {
	return domain.MissionDefinition{
		DefinitionCode: "def-legacy",
		Title:          "Misión legada",
		Description:    "Contenido solo en español",
		Category:       "BOUNTY_HUNTING",
		Mode:           domain.MissionModeSolo,
		State:          domain.DefinitionStatePublished,
		Version:        "v2",
	}
}

// TEST-APP-LANG-001 (ListDefinitions partition): a mixed catalog under
// lang=en serves each item in its effective language with an honest label
// while identity, version, economics and list order stay unchanged.
func TestListDefinitionsLocalized_ServesPerItemLanguage(t *testing.T) {
	// Arrange
	handler := application.NewListDefinitionsHandler(&fakeDefinitionListReader{
		definitions: []domain.MissionDefinition{bilingualListDefinition(), legacyListDefinition()},
	})

	// Act
	defs, err := handler.Handle(domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, defs, 2)
	assert.Equal(t, "def-bounty", defs[0].DefinitionCode)
	assert.Equal(t, "v1", defs[0].Version)
	assert.Equal(t, domain.DefinitionStatePublished, defs[0].State)
	assert.Equal(t, 100, defs[0].CompletionReward)
	assert.Equal(t, "Hunt", defs[0].Title)
	assert.Equal(t, "English briefing", defs[0].Description)
	assert.Equal(t, []string{"First rule"}, defs[0].Rules)
	require.Len(t, defs[0].Objectives, 1)
	assert.Equal(t, "obj-1", defs[0].Objectives[0].ObjectiveID)
	assert.Equal(t, "Capture", defs[0].Objectives[0].Name)
	assert.Equal(t, "English objective", defs[0].Objectives[0].Description)
	assert.Equal(t, 10, defs[0].Objectives[0].MeritPerUnit)
	assert.Equal(t, domain.LanguageEnglish, defs[0].ContentLanguage)
	assert.Equal(t, "def-legacy", defs[1].DefinitionCode)
	assert.Equal(t, "v2", defs[1].Version)
	assert.Equal(t, "Misión legada", defs[1].Title)
	assert.Equal(t, domain.LanguageSpanish, defs[1].ContentLanguage)
}

// TEST-APP-LANG-001 (ListDefinitions partition): incomplete English rules
// or objective translations fall back atomically to the whole Spanish row.
func TestListDefinitionsLocalized_IncompleteEnglishFallsBack(t *testing.T) {
	tests := []struct {
		name     string
		mutateEn func(*domain.LocalizedContent)
	}{
		{"missing english rules", func(en *domain.LocalizedContent) { en.Rules = nil }},
		{"missing english objectives", func(en *domain.LocalizedContent) { en.Objectives = nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			definition := bilingualListDefinition()
			tc.mutateEn(&definition.ContentEn)
			handler := application.NewListDefinitionsHandler(&fakeDefinitionListReader{
				definitions: []domain.MissionDefinition{definition},
			})

			// Act
			defs, err := handler.Handle(domain.LanguageEnglish)

			// Assert
			require.NoError(t, err)
			require.Len(t, defs, 1)
			assert.Equal(t, "Cacería", defs[0].Title)
			assert.Equal(t, "Descripción", defs[0].Description)
			assert.Equal(t, []string{"Regla uno"}, defs[0].Rules)
			assert.Equal(t, "Capturar", defs[0].Objectives[0].Name)
			assert.Equal(t, "Objetivo en español", defs[0].Objectives[0].Description)
			assert.Equal(t, domain.LanguageSpanish, defs[0].ContentLanguage)
		})
	}
}

// TEST-APP-LANG-001 (ListDefinitions partition): an explicit Spanish
// request serves the Spanish scalars even on a bilingual record.
func TestListDefinitionsLocalized_SpanishRequestKeepsSpanish(t *testing.T) {
	// Arrange
	handler := application.NewListDefinitionsHandler(&fakeDefinitionListReader{
		definitions: []domain.MissionDefinition{bilingualListDefinition()},
	})

	// Act
	defs, err := handler.Handle(domain.LanguageSpanish)

	// Assert
	require.NoError(t, err)
	require.Len(t, defs, 1)
	assert.Equal(t, "Cacería", defs[0].Title)
	assert.Equal(t, domain.LanguageSpanish, defs[0].ContentLanguage)
}

// TEST-APP-LANG-001 (ListDefinitions partition): a reader failure is
// propagated unchanged regardless of the requested language.
func TestListDefinitionsLocalized_PropagatesReaderError(t *testing.T) {
	// Arrange
	boom := errors.New("store down")
	handler := application.NewListDefinitionsHandler(&fakeDefinitionListReader{err: boom})

	// Act
	_, err := handler.Handle(domain.LanguageEnglish)

	// Assert
	assert.ErrorIs(t, err, boom)
}
