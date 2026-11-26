package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// authoringDomainDraft builds a fresh draft definition per test with the
// given English snapshot; construction errors are returned for Assert.
func authoringDomainDraft(en domain.LocalizedContent) (domain.MissionDefinition, error) {
	definition, err := domain.NewMissionDefinition("BOUNTY-001", "Cacería", "Descripción",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 50,
		[]domain.Objective{
			{ObjectiveID: "obj-1", Category: "BOUNTY_HUNTING", Name: "Capturar",
				Description: "Capturar el objetivo", MeritPerUnit: 5, Bonus: 2, UnitLimit: 1},
			{ObjectiveID: "obj-2", Category: "BOUNTY_HUNTING", Name: "Rastrear"},
		}, []string{"Regla única"})
	if err != nil {
		return domain.MissionDefinition{}, err
	}
	definition.ContentEn = en
	return definition, nil
}

func authoringDomainEnglish() domain.LocalizedContent {
	return domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"Single rule"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the target"},
			{ObjectiveID: "obj-2", Name: "Track"},
		},
	}
}

// TEST-DOM-LANG-001 (authoring partition): Publish refuses a draft whose
// English snapshot is absent or structurally incomplete and leaves it a
// draft with no version.
func TestMissionDefinitionPublish_RequiresCompleteEnglish(t *testing.T) {
	// Arrange
	english := authoringDomainEnglish()
	cases := []struct {
		name string
		en   domain.LocalizedContent
	}{
		{"no english", domain.LocalizedContent{}},
		{"blank en title", func() domain.LocalizedContent { c := english; c.Title = " "; return c }()},
		{"blank en description", func() domain.LocalizedContent { c := english; c.Description = ""; return c }()},
		{"en rule count differs", func() domain.LocalizedContent { c := english; c.Rules = nil; return c }()},
		{"blank en rule", func() domain.LocalizedContent {
			c := english
			c.Rules = []string{" "}
			return c
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			draft, draftErr := authoringDomainDraft(tc.en)

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

// TEST-DOM-LANG-001 (authoring partition): a draft whose English snapshot
// covers every Spanish field publishes and keeps ids, economics, state and
// version — objective translations match by ID, not position.
func TestMissionDefinitionPublish_PublishesCompleteEnglish(t *testing.T) {
	// Arrange
	english := authoringDomainEnglish()
	english.Objectives = []domain.LocalizedObjective{
		{ObjectiveID: "obj-2", Name: "Track"},
		{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the target"},
	}
	draft, draftErr := authoringDomainDraft(english)

	// Act
	err := draft.Publish("v1")

	// Assert
	require.NoError(t, draftErr)
	require.NoError(t, err)
	assert.Equal(t, domain.DefinitionStatePublished, draft.State)
	assert.Equal(t, "v1", draft.Version)
	assert.Equal(t, "obj-1", draft.Objectives[0].ObjectiveID)
	assert.Equal(t, 5, draft.Objectives[0].MeritPerUnit)
	assert.Equal(t, 2, draft.Objectives[0].Bonus)
	assert.Equal(t, 1, draft.Objectives[0].UnitLimit)
	assert.Equal(t, "obj-2", draft.Objectives[1].ObjectiveID)
	assert.Equal(t, "Capturar", draft.Objectives[0].Name)
	assert.Equal(t, "obj-2", draft.ContentEn.Objectives[0].ObjectiveID)
	assert.Equal(t, "Track", draft.ContentEn.Objectives[0].Name)
	assert.Equal(t, "Capture", draft.ContentEn.Objectives[1].Name)
}
