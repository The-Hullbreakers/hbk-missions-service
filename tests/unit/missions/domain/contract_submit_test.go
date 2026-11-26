package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// pendingContract builds an IN_PROGRESS contract with one confirmed unit,
// submitted for review by its roster member.
func pendingContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := rewardedContract(t, "c-1", 200)
	confirmObjective(t, contract, "obj-a")
	submitContract(t, contract)
	return contract
}

func submitContract(t *testing.T, contract *domain.MissionContract) {
	t.Helper()
	err := contract.Submit(contractMember(t, contract.Roster()[0]))
	require.NoError(t, err)
}

func TestContractSubmit_InProgressBecomesPendingValidation(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)
	confirmObjective(t, contract, "obj-a")

	err := contract.Submit(member(t))

	require.NoError(t, err)
	require.Equal(t, domain.ContractPendingValidation, contract.State)
	require.False(t, contract.State.IsActive(), "expected pending validation to release the active slot")
}

func TestContractSubmit_AllowsPartialConfirmation(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)

	err := contract.Submit(member(t))

	require.NoError(t, err)
	require.Equal(t, domain.ContractPendingValidation, contract.State)
}

func TestContractSubmit_RejectsOutsider(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)

	err := contract.Submit(outsider(t))

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Equal(t, domain.ContractInProgress, contract.State)
}

func TestContractSubmit_RejectsNonInProgress(t *testing.T) {
	contract := lifecycleContract()

	err := contract.Submit(member(t))

	require.ErrorIs(t, err, domain.ErrContractStateInvalid)
}

func TestContractSubmit_RejectsTerminal(t *testing.T) {
	contract := pendingContract(t)
	err := contract.Reject()
	require.NoError(t, err)

	err = contract.Submit(member(t))

	require.ErrorIs(t, err, domain.ErrContractTerminal)
}

func TestContractConfirm_PendingRejectsEvidence(t *testing.T) {
	contract := pendingContract(t)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	_, err := contract.Confirm("obj-a", "unit-2", season)

	require.ErrorIs(t, err, domain.ErrContractNotInProgress)
}

func TestContractState_PendingValidationIsNotActive(t *testing.T) {
	require.False(t, domain.ContractPendingValidation.IsActive(), "expected PENDING_VALIDATION to be inactive for slot accounting")
	require.False(t, domain.ContractPendingValidation.IsTerminal(), "expected PENDING_VALIDATION to remain non-terminal")
}
