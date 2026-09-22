package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestMissionDefinition_NewDraftDefinition(t *testing.T) {
	def, err := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if def.DefinitionID != "bounty-template" || def.State != domain.DefinitionStateDraft {
		t.Fatalf("unexpected draft definition: %+v", def)
	}
}

func TestMissionDefinition_RejectsEmptyDefinitionID(t *testing.T) {
	_, err := domain.NewMissionDefinition("", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	if !errors.Is(err, domain.ErrDefinitionInvalid) {
		t.Fatalf("expected ErrDefinitionInvalid, got %v", err)
	}
}

func TestMissionDefinition_Publish(t *testing.T) {
	def, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	if err := def.Publish("v1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if def.State != domain.DefinitionStatePublished || def.Version != "v1" {
		t.Fatalf("expected published v1, got %+v", def)
	}
}

func TestMissionDefinition_PublishRejectsEmptyVersion(t *testing.T) {
	def, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	err := def.Publish("")
	if !errors.Is(err, domain.ErrDefinitionInvalid) {
		t.Fatalf("expected ErrDefinitionInvalid, got %v", err)
	}
}

func TestMissionDefinition_Retire(t *testing.T) {
	def, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = def.Publish("v1")
	if err := def.Retire(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if def.State != domain.DefinitionStateRetired {
		t.Fatalf("expected retired, got %s", def.State)
	}
}
