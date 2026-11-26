package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// startedContractFixture builds an IN_PROGRESS solo contract.
func startedContractFixture(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	freezeTerms(t, contract, domain.NewObjectiveNamed("obj-a", "a", "", "BOUNTY", 10, 0, 0))
	startContract(t, contract)
	return contract
}

// A roster member raises an OPEN dispute targeting a rejected credit (objective).
func TestRaiseDispute_MemberOpensDispute(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")

	// Act
	dispute, err := contract.RaiseDispute(member, "obj-a", "credit was wrongly rejected")

	// Assert
	require.NoError(t, err)
	require.Equal(t, domain.DisputeOpen, dispute.State)
	require.False(t, dispute.RaisedBy != "player-1" || dispute.ObjectiveID != "obj-a")
}

// Outsiders cannot raise disputes.
func TestRaiseDispute_OutsiderRejected(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	outsider, _ := domain.NewCallerIdentity("player-9", nil, nil)

	// Act
	_, err := contract.RaiseDispute(outsider, "obj-a", "no rights")

	// Assert
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

// An OPEN dispute blocks leave.
func TestLeave_OpenDisputeBlocksExit(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")
	_, err := contract.RaiseDispute(member, "obj-a", "contested")
	require.NoError(t, err)

	// Act
	err = contract.Leave(member)

	// Assert
	require.ErrorIs(t, err, domain.ErrDisputesPending)
}

// An OPEN dispute blocks abandon.
func TestAbandon_OpenDisputeBlocksExit(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")
	_, err := contract.RaiseDispute(member, "obj-a", "contested")
	require.NoError(t, err)

	// Act
	err = contract.Abandon(member)

	// Assert
	require.ErrorIs(t, err, domain.ErrDisputesPending)
}

// Resolving the dispute unblocks exit; a second resolve is rejected.
func TestResolveDispute_UnblocksExit(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")
	dispute, _ := contract.RaiseDispute(member, "obj-a", "contested")

	// Act
	err := contract.ResolveDispute(dispute.DisputeID, "moderator-1")
	require.NoError(t, err)

	// Assert — exit unblocked
	err = contract.Leave(member)
	require.NoError(t, err)
}

// An objective outside the contract terms cannot be disputed.
func TestRaiseDispute_UnknownObjectiveRejected(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")

	// Act
	_, err := contract.RaiseDispute(member, "obj-zzz", "no such objective")

	// Assert
	require.ErrorIs(t, err, domain.ErrObjectiveNotInContract)
}

// Resolving an unknown dispute id is rejected.
func TestResolveDispute_UnknownIDRejected(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)

	// Act
	err := contract.ResolveDispute("contract-1~DISPUTE~obj-a~player-9", "moderator-1")

	// Assert
	require.ErrorIs(t, err, domain.ErrDisputeNotFound)
}

// A contract without disputes leaves normally.
func TestLeave_NoDisputeAllowsExit(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")

	// Act
	err := contract.Leave(member)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 0, len(contract.Disputes()))
}

// A resolved dispute cannot be resolved again.
func TestResolveDispute_AlreadyResolvedRejected(t *testing.T) {
	// Arrange
	contract := startedContractFixture(t)
	member := contractMember(t, "player-1")
	dispute, _ := contract.RaiseDispute(member, "obj-a", "contested")
	_ = contract.ResolveDispute(dispute.DisputeID, "moderator-1")

	// Act
	err := contract.ResolveDispute(dispute.DisputeID, "moderator-1")

	// Assert
	require.ErrorIs(t, err, domain.ErrDisputeStateInvalid)
}
