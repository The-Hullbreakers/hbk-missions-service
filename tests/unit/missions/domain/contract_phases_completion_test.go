package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// phasedRewardFixture builds a two-phase contract carrying a completion
// reward: phase 1 mandatory obj-a + optional obj-x; phase 2 mandatory obj-b.
func phasedRewardFixture(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	offer.CompletionReward = 200
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})

	objA := domain.NewObjectiveNamed("obj-a", "mandatory a", "", "BOUNTY", 10, 0, 0)
	objX := domain.NewObjectiveNamed("obj-x", "optional x", "", "BOUNTY", 5, 0, 0)
	objB := domain.NewObjectiveNamed("obj-b", "mandatory b", "", "BOUNTY", 20, 0, 0)
	terms, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{objA}, Optional: []domain.Objective{objX}},
		{Mandatory: []domain.Objective{objB}},
	})
	require.NoError(t, err)
	contract.Terms = terms

	startContract(t, contract)
	return contract
}

// Completing every mandatory objective — with the optional one unmet — still
// emits the completion credit on approval.
func TestContractApprove_PhasedMandatoryCompletionGrantsReward(t *testing.T) {
	// Arrange
	contract := phasedRewardFixture(t)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	for _, objectiveID := range []string{"obj-a", "obj-b"} {
		_, err := contract.Confirm(objectiveID, "u1", season)
		require.NoError(t, err)
	}
	member := contractMember(t, "player-1")
	err := contract.Submit(member)
	require.NoError(t, err)

	// Act
	credits, err := contract.Approve()

	// Assert
	require.NoError(t, err)
	var completion *domain.MeritCredit
	for i := range credits {
		if credits[i].ObjectiveID == "COMPLETION" {
			completion = &credits[i]
		}
	}
	require.NotNil(t, completion)
	require.Equal(t, 200, completion.Amount)
}
