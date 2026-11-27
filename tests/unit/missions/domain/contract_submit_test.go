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

// TEST-0063-DOM-001: submit requires at least one staged credit carrying a
// media reference; partial progress with evidence still submits.
func TestContractSubmit_RejectsInProgressWithoutEvidence(t *testing.T) {
	// Arrange
	contract := rewardedContract(t, "c-1", 200)

	// Act
	err := contract.Submit(member(t))

	// Assert
	require.ErrorIs(t, err, domain.ErrContractEvidenceRequired)
	require.Equal(t, domain.ContractInProgress, contract.State)
}

// TEST-0063-DOM-001: a staged legacy credit without a media reference does
// not satisfy the evidence requirement.
func TestContractSubmit_RejectsLegacyOnlyStagedCredit(t *testing.T) {
	// Arrange
	base := rewardedContract(t, "c-1", 200)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	legacy, err := base.Confirm("obj-a", "unit-1", season)
	require.NoError(t, err)
	contract := domain.NewMissionContractFromState(
		base.ContractID, base.Offer, base.Roster(), domain.ContractInProgress,
		base.RosterFrozen(), "", base.SeasonID,
		base.Confirmations(), nil, []domain.MeritCredit{legacy},
	)
	contract.Terms = base.Terms

	// Act
	err = contract.Submit(member(t))

	// Assert
	require.ErrorIs(t, err, domain.ErrContractEvidenceRequired)
	require.Equal(t, domain.ContractInProgress, contract.State)
}

// TEST-0063-DOM-001: one media-backed staged credit submits even while
// objectives remain unconfirmed.
func TestContractSubmit_AllowsPartialProgressWithEvidence(t *testing.T) {
	// Arrange
	contract := rewardedContract(t, "c-1", 200)
	extra := domain.NewObjectiveNamed("obj-b", "Segundo objetivo", "", "BOUNTY", 20, 0, 0)
	terms, err := domain.NewContractTerms(append(contract.Terms.Objectives(), extra))
	require.NoError(t, err)
	contract.Terms = terms
	confirmObjective(t, contract, "obj-a")

	// Act
	err = contract.Submit(member(t))

	// Assert
	require.NoError(t, err)
	require.Equal(t, domain.ContractPendingValidation, contract.State)
	require.Less(t, len(contract.Confirmations()), len(contract.Terms.Objectives()),
		"objectives remain unconfirmed — partial progress still submits")
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
