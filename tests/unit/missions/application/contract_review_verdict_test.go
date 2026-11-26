package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// pendingContractWithUnits submits a contract whose objective allows several
// units and has each listed unit staged for review.
func pendingContractWithUnits(t *testing.T, reward int, unitLimit int, units ...string) *domain.MissionContract {
	t.Helper()
	contract := lifecycleFixture(t, []domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 0, unitLimit),
	})
	contract.Offer.CompletionReward = reward
	member := lifecycleMember(t)
	require.NoError(t, contract.Start(member))
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	for _, unit := range units {
		_, err := contract.Confirm("obj-a", unit, season)
		require.NoError(t, err)
	}
	require.NoError(t, contract.Submit(member))
	return contract
}

func TestApproveContract_VerdictCommitsOnlyApprovedUnits(t *testing.T) {
	contract := pendingContractWithUnits(t, 0, 3, "u1", "u2", "u3")
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	credits, err := handler.HandleWithVerdict(reviewer(t, "moderator"), "c-1",
		[]domain.UnitRef{{ObjectiveID: "obj-a", UnitID: "u2"}})

	require.NoError(t, err)
	require.Len(t, credits, 2)
	for _, credit := range credits {
		require.NotEqual(t, "u2", credit.UnitID)
	}
	require.Equal(t, domain.ContractCompleted, store.saved.State)
}

func TestApproveContract_VerdictUnknownUnitFails(t *testing.T) {
	contract := pendingContractWithUnits(t, 0, 3, "u1")
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	_, err := handler.HandleWithVerdict(reviewer(t, "moderator"), "c-1",
		[]domain.UnitRef{{ObjectiveID: "obj-a", UnitID: "ghost"}})

	require.ErrorIs(t, err, domain.ErrUnitNotStaged)
	require.False(t, store.saved != nil || store.credits != nil)
	require.Equal(t, domain.ContractPendingValidation, contract.State)
}

func TestApproveContract_MemberCannotAdjudicateVerdict(t *testing.T) {
	contract := pendingContractWithUnits(t, 0, 3, "u1")
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	_, err := handler.HandleWithVerdict(lifecycleMember(t), "c-1",
		[]domain.UnitRef{{ObjectiveID: "obj-a", UnitID: "u1"}})

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

func TestApproveContract_FullSatisfactionGrantsBonus(t *testing.T) {
	contract := pendingContractWithUnits(t, 60, 3, "u1", "u2", "u3")
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	credits, err := handler.HandleWithVerdict(reviewer(t, "moderator"), "c-1", nil)

	require.NoError(t, err)
	var completion *domain.MeritCredit
	for i := range credits {
		if credits[i].ObjectiveID == domain.CompletionObjectiveID {
			completion = &credits[i]
		}
	}
	require.NotNil(t, completion)
	require.Equal(t, 60, completion.Amount)
}
