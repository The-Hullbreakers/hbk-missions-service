package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestContractApprove_FullConfirmationCreditsEverything(t *testing.T) {
	contract := pendingContract(t)

	credits, err := contract.Approve()

	require.NoError(t, err)
	require.Equal(t, domain.ContractCompleted, contract.State)
	require.Len(t, credits, 2)
	var unitCredit, completionCredit *domain.MeritCredit
	for i := range credits {
		switch credits[i].ObjectiveID {
		case domain.CompletionObjectiveID:
			completionCredit = &credits[i]
		case "obj-a":
			unitCredit = &credits[i]
		}
	}
	require.NotNil(t, unitCredit)
	require.False(t, unitCredit.Amount != 50 || unitCredit.RecipientID != "player-1")
	require.NotNil(t, completionCredit)
	require.Equal(t, 200, completionCredit.Amount)
}

func TestContractApprove_PartialCreditsOnlyConfirmedUnits(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)
	contract.Terms, _ = domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 50, 0, 0),
		domain.NewObjectiveNamed("obj-b", "obj-b", "", "BOUNTY", 20, 0, 0),
	})
	confirmObjective(t, contract, "obj-a")
	submitContract(t, contract)

	credits, err := contract.Approve()

	require.NoError(t, err)
	require.Equal(t, domain.ContractCompleted, contract.State)
	require.Len(t, credits, 1)
	require.NotEqual(t, credits[0].ObjectiveID, domain.CompletionObjectiveID)
}

func TestContractReject_DisqualifiesWithoutCredits(t *testing.T) {
	contract := pendingContract(t)

	err := contract.Reject()

	require.NoError(t, err)
	require.Equal(t, domain.ContractDisqualified, contract.State)
	require.True(t, contract.State.IsTerminal(), "expected DISQUALIFIED to be terminal")
}

func TestContractReject_NoStagedCreditsRemain(t *testing.T) {
	contract := pendingContract(t)
	err := contract.Reject()
	require.NoError(t, err)
	// Reconstruction must not resurrect discarded credits.
	rebuilt := domain.NewMissionContractFromState(
		contract.ContractID, contract.Offer, contract.Roster(),
		contract.State, contract.RosterFrozen(), contract.TeamID,
		contract.SeasonID, contract.Confirmations(), contract.Departures(),
		contract.PendingCredits(),
	)
	require.Equal(t, 0, len(rebuilt.PendingCredits()))
}

func TestContractApprove_RejectsNonPendingState(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)

	_, err := contract.Approve()

	require.ErrorIs(t, err, domain.ErrContractNotPendingValidation)
}

func TestContractReject_RejectsNonPendingState(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)

	err := contract.Reject()

	require.ErrorIs(t, err, domain.ErrContractNotPendingValidation)
}
