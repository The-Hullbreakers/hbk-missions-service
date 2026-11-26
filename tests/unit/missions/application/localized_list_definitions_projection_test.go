package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-APP-LANG-001 (ListDefinitions partition): the served projection
// copies rules and objectives — mutating it must not rewrite the stored
// definition's Spanish or English content.
func TestListDefinitionsLocalized_ProjectionDoesNotAliasSource(t *testing.T) {
	// Arrange
	definition := bilingualListDefinition()
	handler := application.NewListDefinitionsHandler(&fakeDefinitionListReader{
		definitions: []domain.MissionDefinition{definition},
	})

	// Act
	defs, err := handler.Handle(domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, defs, 1)
	defs[0].Rules[0] = "Mutated served rule"
	defs[0].Objectives[0].Name = "Mutated served name"
	assert.Equal(t, []string{"Regla uno"}, definition.Rules)
	assert.Equal(t, "Capturar", definition.Objectives[0].Name)
	assert.Equal(t, []string{"First rule"}, definition.ContentEn.Rules)
	assert.Equal(t, "Capture", definition.ContentEn.Objectives[0].Name)
}

// TEST-APP-LANG-001 (ListDefinitions partition): the Spanish-fallback
// projection is also copied, never aliased to the stored record.
func TestListDefinitionsLocalized_FallbackDoesNotAliasSource(t *testing.T) {
	// Arrange
	definition := legacyListDefinition()
	definition.Rules = []string{"Regla legada"}
	definition.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Capturar", "Objetivo en español", "BOUNTY", 10, 5, 3),
	}
	handler := application.NewListDefinitionsHandler(&fakeDefinitionListReader{
		definitions: []domain.MissionDefinition{definition},
	})

	// Act
	defs, err := handler.Handle(domain.LanguageEnglish)

	// Assert
	require.NoError(t, err)
	require.Len(t, defs, 1)
	defs[0].Rules[0] = "Mutated served rule"
	defs[0].Objectives[0].Name = "Mutated served name"
	assert.Equal(t, []string{"Regla legada"}, definition.Rules)
	assert.Equal(t, "Capturar", definition.Objectives[0].Name)
}
