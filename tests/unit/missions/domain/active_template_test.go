package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestActiveTemplate_ExcludesSameDefinition(t *testing.T) {
	// Arrange
	definition, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = definition.Publish("v1")

	offer1, _ := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition, 0)
	active := domain.NewMissionContract("contract-1", offer1, []string{"player-1"})
	active.Start()

	offer2, _ := domain.NewMissionOfferFromDefinition("offer-2", "season-1", definition, 0)
	policy := domain.NewActiveLimitPolicy(5, nil)

	// Act
	decision := domain.DecideAcceptance(offer2, []domain.ActiveContractSummary{active.Summary()}, policy)

	// Assert
	if decision.Accepted {
		t.Fatalf("expected rejection because the definition is already active")
	}
	if decision.Reason != domain.ReasonMissionDefinitionAlreadyActive {
		t.Fatalf("expected MISSION_DEFINITION_ALREADY_ACTIVE, got %s", decision.Reason)
	}
}

func TestActiveTemplate_DoesNotExcludeDifferentDefinition(t *testing.T) {
	// Arrange
	a, _ := domain.NewMissionDefinition("template-a", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = a.Publish("v1")
	b, _ := domain.NewMissionDefinition("template-b", "BOUNTY_HUNTING", 10, domain.MissionModeSolo)
	_ = b.Publish("v1")

	offer1, _ := domain.NewMissionOfferFromDefinition("offer-1", "season-1", a, 0)
	active := domain.NewMissionContract("contract-1", offer1, []string{"player-1"})
	active.Start()

	offer2, _ := domain.NewMissionOfferFromDefinition("offer-2", "season-1", b, 0)
	policy := domain.NewActiveLimitPolicy(5, nil)

	// Act
	decision := domain.DecideAcceptance(offer2, []domain.ActiveContractSummary{active.Summary()}, policy)

	// Assert
	if !decision.Accepted {
		t.Fatalf("expected acceptance for a different definition, got %s", decision.Reason)
	}
}
