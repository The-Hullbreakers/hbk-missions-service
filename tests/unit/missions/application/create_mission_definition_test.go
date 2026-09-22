package application_test

import (
	"errors"
	"testing"

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

func TestCreateMissionDefinitionHandler_CreatesDraft(t *testing.T) {
	writer := &recordingDefinitionWriter{}
	handler := application.NewCreateMissionDefinitionHandler(writer)

	definition, err := handler.Handle("tenant-1", "bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if definition.DefinitionID != "bounty-template" || definition.State != domain.DefinitionStateDraft {
		t.Fatalf("unexpected definition: %+v", definition)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("expected definition to be saved")
	}
}

func TestCreateMissionDefinitionHandler_RejectsInvalidDefinition(t *testing.T) {
	handler := application.NewCreateMissionDefinitionHandler(&recordingDefinitionWriter{})
	_, err := handler.Handle("tenant-1", "", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	if !errors.Is(err, domain.ErrDefinitionInvalid) {
		t.Fatalf("expected ErrDefinitionInvalid, got %v", err)
	}
}

func TestCreateMissionDefinitionHandler_ReturnsWriterError(t *testing.T) {
	handler := application.NewCreateMissionDefinitionHandler(failingDefinitionWriter{})
	_, err := handler.Handle("tenant-1", "bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	if err == nil {
		t.Fatalf("expected writer error")
	}
}
