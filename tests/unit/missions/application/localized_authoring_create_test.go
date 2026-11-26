package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// localizedDefinitionWriter is a hand-written DefinitionWriter double that
// records calls and the definition it was asked to persist.
type localizedDefinitionWriter struct {
	saved domain.MissionDefinition
	calls int
}

func (w *localizedDefinitionWriter) Save(definition domain.MissionDefinition) error {
	w.calls++
	w.saved = definition
	return nil
}

func bilingualDraftRequest() application.CreateMissionDefinitionRequest {
	return application.CreateMissionDefinitionRequest{
		DefinitionCode:   "BOUNTY-001",
		Title:            "Cacería",
		Description:      "Descripción",
		Category:         "BOUNTY_HUNTING",
		Mode:             domain.MissionModeSolo,
		CompletionReward: 50,
		Objectives: []domain.Objective{
			{ObjectiveID: "obj-1", Category: "BOUNTY_HUNTING", Name: "Capturar", Description: "Capturar el objetivo"},
		},
		Rules: []string{"Regla única"},
		ContentEn: &domain.LocalizedContent{
			Title:       "Hunt",
			Description: "English briefing",
			Rules:       []string{"Single rule"},
			Objectives:  []domain.LocalizedObjective{{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the target"}},
		},
	}
}

// TEST-APP-LANG-001 (authoring partition): a draft created with nil
// ContentEn is a Spanish-only draft and is persisted exactly once.
func TestLocalizedAuthoringCreate_SpanishOnlyDraft(t *testing.T) {
	// Arrange
	writer := &localizedDefinitionWriter{}
	handler := application.NewCreateMissionDefinitionHandler(writer)
	request := bilingualDraftRequest()
	request.ContentEn = nil

	// Act
	definition, err := handler.Handle(request)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, writer.calls)
	assert.Equal(t, "BOUNTY-001", writer.saved.DefinitionCode)
	assert.Equal(t, "Cacería", writer.saved.Title)
	assert.Equal(t, domain.DefinitionStateDraft, writer.saved.State)
	assert.Equal(t, "BOUNTY-001", definition.DefinitionCode)
	assert.Equal(t, "", definition.Version)
}

// TEST-APP-LANG-001 (authoring partition): a bilingual draft freezes a deep
// copy of the English snapshot — mutating the request after creation must
// not touch the persisted definition.
func TestLocalizedAuthoringCreate_FreezesEnglishSnapshot(t *testing.T) {
	// Arrange
	writer := &localizedDefinitionWriter{}
	handler := application.NewCreateMissionDefinitionHandler(writer)
	request := bilingualDraftRequest()

	// Act
	_, err := handler.Handle(request)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, writer.calls)
	request.ContentEn.Title = "Mutated"
	request.ContentEn.Rules[0] = "Mutated rule"
	request.ContentEn.Objectives[0].Name = "Mutated objective"
	assert.Equal(t, "Hunt", writer.saved.ContentEn.Title)
	assert.Equal(t, "English briefing", writer.saved.ContentEn.Description)
	assert.Equal(t, []string{"Single rule"}, writer.saved.ContentEn.Rules)
	assert.Equal(t, "Capture", writer.saved.ContentEn.Objectives[0].Name)
	assert.Equal(t, "Cacería", writer.saved.Title)
	assert.Equal(t, "Descripción", writer.saved.Description)
}

// TEST-APP-LANG-001 (authoring partition): an invalid Spanish baseline is
// rejected by the domain and the writer is never called.
func TestLocalizedAuthoringCreate_RejectsInvalidBaseline(t *testing.T) {
	// Arrange
	writer := &localizedDefinitionWriter{}
	handler := application.NewCreateMissionDefinitionHandler(writer)
	request := bilingualDraftRequest()
	request.Title = ""

	// Act
	_, err := handler.Handle(request)

	// Assert
	assert.ErrorIs(t, err, domain.ErrDefinitionInvalid)
	assert.Equal(t, 0, writer.calls)
}
