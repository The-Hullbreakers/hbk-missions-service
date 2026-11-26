package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedDefinitionStore is a hand-written DefinitionReader + Writer
// double serving one draft fixture and recording saves.
type localizedDefinitionStore struct {
	draft domain.MissionDefinition
	saved domain.MissionDefinition
	saves int
}

func (s *localizedDefinitionStore) Get(definitionCode string) (domain.MissionDefinition, error) {
	if definitionCode != s.draft.DefinitionCode {
		return domain.MissionDefinition{}, domain.ErrDefinitionNotFound
	}
	return s.draft, nil
}

func (s *localizedDefinitionStore) GetVersion(_, _ string) (domain.MissionDefinition, error) {
	return s.draft, nil
}

func (s *localizedDefinitionStore) Save(definition domain.MissionDefinition) error {
	s.saves++
	s.saved = definition
	return nil
}

func publishableDraft() (domain.MissionDefinition, error) {
	return domain.NewMissionDefinition("BOUNTY-001", "Cacería", "Descripción",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 50,
		[]domain.Objective{
			{ObjectiveID: "obj-1", Category: "BOUNTY_HUNTING", Name: "Capturar", Description: "Capturar el objetivo"},
		}, []string{"Regla única"})
}

func publishableEnglish() domain.LocalizedContent {
	return domain.LocalizedContent{
		Title:       "Hunt",
		Description: "English briefing",
		Rules:       []string{"Single rule"},
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the target"},
		},
	}
}

// TEST-APP-LANG-001 (authoring partition): publishing a new draft without a
// structurally complete English snapshot is rejected before the store is
// written; historical published es-only versions stay readable elsewhere.
func TestLocalizedAuthoringPublish_RejectsMissingOrPartialEnglish(t *testing.T) {
	// Arrange
	cases := []struct {
		name string
		en   *domain.LocalizedContent
	}{
		{"spanish only", nil},
		{"missing en rules", &domain.LocalizedContent{Title: "Hunt", Description: "English briefing"}},
		{"missing en objective", &domain.LocalizedContent{
			Title: "Hunt", Description: "English briefing", Rules: []string{"Single rule"},
			Objectives: []domain.LocalizedObjective{},
		}},
		{"blank en title", &domain.LocalizedContent{
			Title: "  ", Description: "English briefing", Rules: []string{"Single rule"},
			Objectives: []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the target"}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			draft, draftErr := publishableDraft()
			if tc.en != nil {
				draft.ContentEn = *tc.en
			}
			store := &localizedDefinitionStore{draft: draft}
			handler := application.NewPublishMissionDefinitionHandler(store, store)

			// Act
			_, err := handler.Handle("BOUNTY-001", "v1")

			// Assert
			require.NoError(t, draftErr)
			assert.ErrorIs(t, err, domain.ErrDefinitionInvalid)
			assert.Equal(t, 0, store.saves)
		})
	}
}

// TEST-APP-LANG-001 (authoring partition): a new draft with a complete
// English snapshot publishes once with Spanish and English content intact.
func TestLocalizedAuthoringPublish_PublishesBilingualDraft(t *testing.T) {
	// Arrange
	english := publishableEnglish()
	draft, draftErr := publishableDraft()
	draft.ContentEn = english
	store := &localizedDefinitionStore{draft: draft}
	handler := application.NewPublishMissionDefinitionHandler(store, store)

	// Act
	published, err := handler.Handle("BOUNTY-001", "v1")

	// Assert
	require.NoError(t, draftErr)
	require.NoError(t, err)
	assert.Equal(t, 1, store.saves)
	assert.Equal(t, domain.DefinitionStatePublished, store.saved.State)
	assert.Equal(t, "v1", store.saved.Version)
	assert.Equal(t, "Cacería", store.saved.Title)
	assert.Equal(t, "Hunt", store.saved.ContentEn.Title)
	assert.Equal(t, "Hunt", published.ContentEn.Title)
}
