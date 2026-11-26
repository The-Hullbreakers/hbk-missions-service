package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

func TestActiveTemplate_ExcludesSameDefinition(t *testing.T) {
	// Arrange
	definition, _ := domain.NewMissionDefinition("bounty-template", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	definition.ContentEn = domain.LocalizedContent{Title: "Bounty", Description: "Desc"}
	_ = definition.Publish("v1")

	offer1, _ := domain.NewMissionOfferFromDefinition("offer-1", "season-1", definition)
	active := domain.NewMissionContract("contract-1", offer1, []string{"player-1"})
	startContract(t, active)

	offer2, _ := domain.NewMissionOfferFromDefinition("offer-2", "season-1", definition)
	policy := domain.NewActiveLimitPolicy(5, nil)

	// Act
	decision := domain.DecideAcceptance(offer2, domain.SeasonStateActive, []domain.ActiveContractSummary{active.Summary()}, policy)

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
	a, _ := domain.NewMissionDefinition("template-a", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	a.ContentEn = domain.LocalizedContent{Title: "A", Description: "Desc"}
	_ = a.Publish("v1")
	b, _ := domain.NewMissionDefinition("template-b", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	b.ContentEn = domain.LocalizedContent{Title: "B", Description: "Desc"}
	_ = b.Publish("v1")

	offer1, _ := domain.NewMissionOfferFromDefinition("offer-1", "season-1", a)
	active := domain.NewMissionContract("contract-1", offer1, []string{"player-1"})
	startContract(t, active)

	offer2, _ := domain.NewMissionOfferFromDefinition("offer-2", "season-1", b)
	policy := domain.NewActiveLimitPolicy(5, nil)

	// Act
	decision := domain.DecideAcceptance(offer2, domain.SeasonStateActive, []domain.ActiveContractSummary{active.Summary()}, policy)

	// Assert
	if !decision.Accepted {
		t.Fatalf("expected acceptance for a different definition, got %s", decision.Reason)
	}
}

// REG-0053-02: any roster member already bound to the offer's definition
// rejects the acceptance — teammates were only limit-checked before.
func TestDecideRosterAcceptance_RejectsMemberWithSameDefinition(t *testing.T) {
	for _, state := range []domain.ContractState{domain.ContractInProgress, domain.ContractPendingValidation} {
		t.Run(string(state), func(t *testing.T) {
			// Arrange: teammate player-2 blocks def-x while under the limit
			// (PENDING_VALIDATION keeps spare slots yet must still block).
			offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
			offer.DefinitionCode = "def-x"
			memberActive := map[string][]domain.ActiveContractSummary{
				"player-1": nil,
				"player-2": {domain.NewActiveContractSummary("offer-x", "BOUNTY_HUNTING", state, "def-x")},
			}
			policy := domain.NewActiveLimitPolicy(2, nil)

			// Act
			decision := domain.DecideRosterAcceptance(offer, domain.SeasonStateActive, memberActive, policy)

			// Assert
			require.False(t, decision.Accepted)
			require.Equal(t, domain.ReasonMissionDefinitionAlreadyActive, decision.Reason)
		})
	}
}

// A teammate at their limit keeps the teammate-specific limit reason.
func TestDecideRosterAcceptance_TeammateLimitKeepsTeammateReason(t *testing.T) {
	// Arrange: teammate player-2 is at the limit on other definitions.
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	memberActive := map[string][]domain.ActiveContractSummary{
		"player-1": nil,
		"player-2": {
			domain.NewActiveContractSummary("offer-a", "BOUNTY_HUNTING", domain.ContractInProgress, "def-a"),
			domain.NewActiveContractSummary("offer-b", "BOUNTY_HUNTING", domain.ContractInProgress, "def-b"),
		},
	}
	policy := domain.NewActiveLimitPolicy(2, nil)

	// Act
	decision := domain.DecideRosterAcceptance(offer, domain.SeasonStateActive, memberActive, policy)

	// Assert
	require.False(t, decision.Accepted)
	require.Equal(t, domain.ReasonTeammateLimitReached, decision.Reason)
}
