package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// authoringDomainEdgeDraft builds a single-objective draft with the given
// English objective list; construction errors are returned for Assert.
func authoringDomainEdgeDraft(esDescription string, enObjectives []domain.LocalizedObjective) (domain.MissionDefinition, error) {
	definition, err := domain.NewMissionDefinition("BOUNTY-002", "Cacería", "Descripción",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 50,
		[]domain.Objective{
			{ObjectiveID: "obj-1", Category: "BOUNTY_HUNTING", Name: "Capturar", Description: esDescription},
		}, []string{"Regla única"})
	if err != nil {
		return domain.MissionDefinition{}, err
	}
	definition.ContentEn = domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"Single rule"},
		Objectives:  enObjectives,
	}
	return definition, nil
}

// TEST-DOM-LANG-001 (authoring partition): Publish refuses English
// objective sets that cannot cover the Spanish objectives exactly once.
func TestMissionDefinitionPublish_RejectsObjectiveMismatches(t *testing.T) {
	// Arrange
	cases := []struct {
		name string
		en   []domain.LocalizedObjective
	}{
		{"missing objective", []domain.LocalizedObjective{}},
		{"duplicate id", []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the target"},
			{ObjectiveID: "obj-1", Name: "Again", Description: "Capture the target"},
		}},
		{"unknown id", []domain.LocalizedObjective{
			{ObjectiveID: "obj-9", Name: "Capture", Description: "Capture the target"},
		}},
		{"blank name", []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: " ", Description: "Capture the target"},
		}},
		{"blank required description", []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			draft, draftErr := authoringDomainEdgeDraft("Capturar el objetivo", tc.en)

			// Act
			err := draft.Publish("v1")

			// Assert
			require.NoError(t, draftErr)
			assert.ErrorIs(t, err, domain.ErrDefinitionInvalid)
			assert.Equal(t, domain.DefinitionStateDraft, draft.State)
			assert.Equal(t, "", draft.Version)
		})
	}
}

// TEST-DOM-LANG-001 (authoring partition): a blank English objective
// description is allowed when the Spanish objective carries none.
func TestMissionDefinitionPublish_AllowsOptionalEnglishDescription(t *testing.T) {
	// Arrange
	draft, draftErr := authoringDomainEdgeDraft("", []domain.LocalizedObjective{
		{ObjectiveID: "obj-1", Name: "Capture"},
	})

	// Act
	err := draft.Publish("v1")

	// Assert
	require.NoError(t, draftErr)
	require.NoError(t, err)
	assert.Equal(t, domain.DefinitionStatePublished, draft.State)
	assert.Equal(t, "v1", draft.Version)
}

// TEST-DOM-LANG-001 (authoring partition): a draft with no rules and no
// objectives publishes when its English snapshot covers the empty sets.
func TestMissionDefinitionPublish_AllowsEmptyContentShape(t *testing.T) {
	// Arrange
	draft, draftErr := domain.NewMissionDefinition("BOUNTY-003", "Cacería", "Descripción",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 50, nil, nil)
	draft.ContentEn = domain.LocalizedContent{Title: "Hunt", Description: "English briefing"}

	// Act
	err := draft.Publish("v1")

	// Assert
	require.NoError(t, draftErr)
	require.NoError(t, err)
	assert.Equal(t, domain.DefinitionStatePublished, draft.State)
}

// TEST-DOM-LANG-001 (authoring partition): historical already-published
// Spanish-only records stay readable and cannot be republished through the
// new completeness rule — no migration is required.
func TestMissionDefinitionPublish_HistoricalSpanishOnlyStaysPublished(t *testing.T) {
	// Arrange
	historical := domain.MissionDefinition{
		DefinitionCode: "BOUNTY-LEGACY", Title: "Misión legada", Description: "Vieja",
		Category: "BOUNTY_HUNTING", Mode: domain.MissionModeSolo,
		State: domain.DefinitionStatePublished, Version: "v-legacy",
	}

	// Act
	err := historical.Publish("v2")

	// Assert
	assert.ErrorIs(t, err, domain.ErrDefinitionInvalid)
	assert.Equal(t, "Misión legada", historical.Title)
	assert.Equal(t, domain.DefinitionStatePublished, historical.State)
	assert.Equal(t, "v-legacy", historical.Version)
}
