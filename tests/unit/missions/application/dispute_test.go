package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// A roster member raises a dispute; the handler persists the opened dispute on
// the contract aggregate.
func TestRaiseDispute_PersistsOpenedDispute(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewRaiseDisputeHandler(store, store)

	// Act
	dispute, err := handler.Handle(lifecycleMember(t), contract.ContractID, "obj-a", "wrongly rejected")

	// Assert
	require.NoError(t, err)
	require.False(t, dispute.State != domain.DisputeOpen || dispute.ObjectiveID != "obj-a")
	require.Equal(t, 1, store.saveCalls)
}

// Outsiders cannot raise disputes and nothing is persisted.
func TestRaiseDispute_OutsiderRejected(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewRaiseDisputeHandler(store, store)
	outsider, _ := domain.NewCallerIdentity("player-9", nil, nil)

	// Act
	_, err := handler.Handle(outsider, contract.ContractID, "obj-a", "no rights")

	// Assert
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Equal(t, 0, store.saveCalls)
}

// A moderator resolves an open dispute; the resolved state is persisted.
func TestResolveDispute_PersistsResolvedState(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	member := lifecycleMember(t)
	dispute, err := contract.RaiseDispute(member, "obj-a", "contested")
	require.NoError(t, err)
	store := &reviewStore{contract: contract}
	handler := application.NewResolveDisputeHandler(store, store)

	// Act
	err = handler.Handle(reviewer(t, "moderator"), contract.ContractID, dispute.DisputeID)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, store.saveCalls)
	saved := store.saved.Disputes()
	require.False(t, saved[0].State != domain.DisputeResolved || saved[0].ResolvedBy != "mod-1")
}

// Non-adjudicators cannot resolve disputes.
func TestResolveDispute_MemberRejected(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	member := lifecycleMember(t)
	dispute, _ := contract.RaiseDispute(member, "obj-a", "contested")
	store := &reviewStore{contract: contract}
	handler := application.NewResolveDisputeHandler(store, store)

	// Act
	err := handler.Handle(member, contract.ContractID, dispute.DisputeID)

	// Assert
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

// Reviewers list the contract's disputes; members are denied.
func TestListDisputes_ReviewerSeesDisputes(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	member := lifecycleMember(t)
	_, err := contract.RaiseDispute(member, "obj-a", "contested")
	require.NoError(t, err)
	store := &reviewStore{contract: contract}
	handler := application.NewListDisputesHandler(store)

	// Act
	disputes, err := handler.Handle(reviewer(t, "moderator"), contract.ContractID)

	// Assert
	require.NoError(t, err)
	require.False(t, len(disputes) != 1 || disputes[0].State != domain.DisputeOpen)
}

// Members cannot list disputes.
func TestListDisputes_MemberRejected(t *testing.T) {
	// Arrange
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewListDisputesHandler(store)

	// Act
	_, err := handler.Handle(lifecycleMember(t), contract.ContractID)

	// Assert
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}
