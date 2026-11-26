package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type recordingDefinitionStore struct {
	definitions map[string]domain.MissionDefinition
}

func (s *recordingDefinitionStore) Get(definitionCode string) (domain.MissionDefinition, error) {
	def, ok := s.definitions[definitionCode]
	if !ok {
		return domain.MissionDefinition{}, domain.ErrDefinitionNotFound
	}
	return def, nil
}

func (s *recordingDefinitionStore) GetVersion(_, _ string) (domain.MissionDefinition, error) {
	return domain.MissionDefinition{}, errors.New("not implemented")
}

func (s *recordingDefinitionStore) Save(definition domain.MissionDefinition) error {
	s.definitions[definition.DefinitionCode] = definition
	return nil
}

func aDraftDefinition() domain.MissionDefinition {
	def, _ := domain.NewMissionDefinition("bounty-template", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	def.ContentEn = domain.LocalizedContent{Title: "Bounty", Description: "Desc"}
	return def
}

func TestPublishMissionDefinitionHandler_PublishesVersion(t *testing.T) {
	store := &recordingDefinitionStore{definitions: map[string]domain.MissionDefinition{"bounty-template": aDraftDefinition()}}
	handler := application.NewPublishMissionDefinitionHandler(store, store)

	definition, err := handler.Handle("bounty-template", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if definition.State != domain.DefinitionStatePublished || definition.Version != "v1" {
		t.Fatalf("unexpected published definition: %+v", definition)
	}
}

func TestPublishMissionDefinitionHandler_RejectsUnknownDefinition(t *testing.T) {
	store := &recordingDefinitionStore{definitions: map[string]domain.MissionDefinition{}}
	handler := application.NewPublishMissionDefinitionHandler(store, store)

	_, err := handler.Handle("bounty-template", "v1")
	if !errors.Is(err, domain.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestPublishMissionDefinitionHandler_RejectsEmptyVersion(t *testing.T) {
	store := &recordingDefinitionStore{definitions: map[string]domain.MissionDefinition{"bounty-template": aDraftDefinition()}}
	handler := application.NewPublishMissionDefinitionHandler(store, store)

	_, err := handler.Handle("bounty-template", "")
	if !errors.Is(err, domain.ErrDefinitionInvalid) {
		t.Fatalf("expected ErrDefinitionInvalid, got %v", err)
	}
}
