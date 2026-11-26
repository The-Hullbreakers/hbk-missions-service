package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type recordingDefinitionWriter struct {
	saved []domain.MissionDefinition
}

func (w *recordingDefinitionWriter) Save(definition domain.MissionDefinition) error {
	w.saved = append(w.saved, definition)
	return nil
}

type failingDefinitionWriter struct{}

func (failingDefinitionWriter) Save(domain.MissionDefinition) error {
	return errors.New("definition writer unavailable")
}

func createObjectives() []domain.Objective {
	return []domain.Objective{
		domain.NewObjectiveNamed("obj-1", "obj-1", "", "BOUNTY_HUNTING", 100, 0, 1),
	}
}

func TestCreateMissionDefinitionHandler_CreatesDraft(t *testing.T) {
	writer := &recordingDefinitionWriter{}
	handler := application.NewCreateMissionDefinitionHandler(writer)

	definition, err := handler.Handle(application.CreateMissionDefinitionRequest{
		DefinitionCode: "bounty-template", Title: "Bounty", Description: "desc",
		Category: "BOUNTY_HUNTING", Mode: domain.MissionModeSolo,
		Objectives: createObjectives(),
	})

	require.NoError(t, err)
	require.Equal(t, "bounty-template", definition.DefinitionCode)
	require.Equal(t, domain.DefinitionStateDraft, definition.State)
	require.Len(t, writer.saved, 1)
}

func TestCreateMissionDefinitionHandler_RejectsInvalidDefinition(t *testing.T) {
	handler := application.NewCreateMissionDefinitionHandler(&recordingDefinitionWriter{})
	_, err := handler.Handle(application.CreateMissionDefinitionRequest{
		Title: "Bounty", Category: "BOUNTY_HUNTING",
		Mode: domain.MissionModeSolo, Objectives: createObjectives(),
	})
	require.ErrorIs(t, err, domain.ErrDefinitionInvalid)
}

func TestCreateMissionDefinitionHandler_ReturnsWriterError(t *testing.T) {
	handler := application.NewCreateMissionDefinitionHandler(failingDefinitionWriter{})
	_, err := handler.Handle(application.CreateMissionDefinitionRequest{
		DefinitionCode: "bounty-template", Title: "Bounty", Description: "desc",
		Category: "BOUNTY_HUNTING", Mode: domain.MissionModeSolo,
		Objectives: createObjectives(),
	})
	require.Error(t, err)
}
