package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-LANG-002 (contract terms partition): WithObjectiveText produces
// an independent translated snapshot — Objectives() and Phases() carry the
// English text in original order and phase membership while ids, economics
// and the source terms stay untouched.
func TestContractTermsWithObjectiveText_TranslatesAcrossPhases(t *testing.T) {
	// Arrange
	terms, termsErr := domain.NewContractTermsPhased([]domain.Phase{
		{
			Mandatory: []domain.Objective{domain.NewObjectiveNamed("obj-1", "Primero", "Primer objetivo", "BOUNTY", 10, 5, 3)},
			Optional:  []domain.Objective{domain.NewObjectiveNamed("obj-2", "Segundo", "Segundo objetivo", "BOUNTY", 20, 0, 1)},
		},
		{
			Mandatory: []domain.Objective{domain.NewObjectiveNamed("obj-3", "Tercero", "Tercer objetivo", "BOUNTY", 30, 7, 2)},
		},
	})
	translations := []domain.LocalizedObjective{
		{ObjectiveID: "obj-3", Name: "Third", Description: "Third objective"},
		{ObjectiveID: "obj-2", Name: "Second", Description: "Second objective"},
		{ObjectiveID: "obj-1", Name: "First", Description: "First objective"},
	}

	// Act
	translated, err := terms.WithObjectiveText(translations)

	// Assert
	require.NoError(t, termsErr)
	require.NoError(t, err)
	objectives := translated.Objectives()
	require.Len(t, objectives, 3)
	assert.Equal(t, "obj-1", objectives[0].ObjectiveID)
	assert.Equal(t, "First", objectives[0].Name)
	assert.Equal(t, "First objective", objectives[0].Description)
	assert.Equal(t, 10, objectives[0].MeritPerUnit)
	assert.Equal(t, "obj-3", objectives[2].ObjectiveID)
	assert.Equal(t, "Third", objectives[2].Name)
	phases := translated.Phases()
	require.Len(t, phases, 2)
	require.Len(t, phases[0].Mandatory, 1)
	require.Len(t, phases[0].Optional, 1)
	require.Len(t, phases[1].Mandatory, 1)
	assert.Equal(t, "First", phases[0].Mandatory[0].Name)
	assert.Equal(t, "obj-2", phases[0].Optional[0].ObjectiveID)
	assert.Equal(t, "Second", phases[0].Optional[0].Name)
	assert.Equal(t, 0, phases[0].Optional[0].Bonus)
	assert.Equal(t, 1, phases[0].Optional[0].UnitLimit)
	assert.Equal(t, "Third", phases[1].Mandatory[0].Name)
	assert.Equal(t, 30, phases[1].Mandatory[0].MeritPerUnit)
	assert.Equal(t, 7, phases[1].Mandatory[0].Bonus)
	assert.Equal(t, 2, phases[1].Mandatory[0].UnitLimit)
	translations[0].Name = "Mutated"
	assert.Equal(t, "Third", translated.Objectives()[2].Name)
	sourceObjectives := terms.Objectives()
	assert.Equal(t, "Primero", sourceObjectives[0].Name)
	assert.Equal(t, "Segundo", sourceObjectives[1].Name)
	assert.Equal(t, "Tercero", sourceObjectives[2].Name)
}

// TEST-DOM-LANG-002 (contract terms partition): translations must cover
// every objective exactly once by ID with usable text, or the snapshot is
// rejected; a blank description is fine only when the source has none.
func TestContractTermsWithObjectiveText_RejectsInvalidTranslations(t *testing.T) {
	terms, termsErr := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-1", "Primero", "Primer objetivo", "BOUNTY", 10, 5, 3),
		domain.NewObjectiveNamed("obj-2", "Segundo", "", "BOUNTY", 20, 0, 1),
	})
	okOne := domain.LocalizedObjective{ObjectiveID: "obj-1", Name: "First", Description: "First objective"}
	okTwo := domain.LocalizedObjective{ObjectiveID: "obj-2", Name: "Second"}
	tests := []struct {
		name         string
		translations []domain.LocalizedObjective
		wantErr      bool
	}{
		{"missing objective translation", []domain.LocalizedObjective{okOne}, true},
		{"unknown objective id", []domain.LocalizedObjective{okOne, {ObjectiveID: "obj-x", Name: "X"}}, true},
		{"duplicate objective id", []domain.LocalizedObjective{okOne, {ObjectiveID: "obj-1", Name: "Again"}}, true},
		{"blank translated name", []domain.LocalizedObjective{{ObjectiveID: "obj-1"}, okTwo}, true},
		{"blank description when source has one", []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "First"}, okTwo}, true},
		{"blank description allowed when source has none", []domain.LocalizedObjective{okOne, okTwo}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			translations := tc.translations

			// Act
			_, err := terms.WithObjectiveText(translations)

			// Assert
			require.NoError(t, termsErr)
			if tc.wantErr {
				assert.ErrorIs(t, err, domain.ErrContractTermsInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TEST-DOM-LANG-002 (contract terms partition): zero (legacy) terms have no
// objectives to translate and are rejected rather than silently accepted.
func TestContractTermsWithObjectiveText_RejectsZeroTerms(t *testing.T) {
	// Arrange
	var terms domain.ContractTerms

	// Act
	_, err := terms.WithObjectiveText(nil)

	// Assert
	assert.ErrorIs(t, err, domain.ErrContractTermsInvalid)
}
