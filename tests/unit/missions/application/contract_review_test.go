package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestApproveContract_ModeratorApprovesAndWritesCredits(t *testing.T) {
	contract := pendingContract(t, 200)
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	credits, err := handler.Handle(reviewer(t, "moderator"), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractCompleted)
	require.Len(t, store.credits, 2)
	require.Equal(t, len(store.credits), len(credits))
}

func TestApproveContract_AdminApproves(t *testing.T) {
	contract := pendingContract(t, 0)
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	_, err := handler.Handle(reviewer(t, "admin"), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractCompleted)
}

func TestApproveContract_MemberCannotAdjudicate(t *testing.T) {
	contract := pendingContract(t, 200)
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	_, err := handler.Handle(lifecycleMember(t), "c-1")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.False(t, store.saved != nil || store.credits != nil)
}

func TestApproveContract_RejectsNonPending(t *testing.T) {
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewApproveContractHandler(store, store)

	_, err := handler.Handle(reviewer(t, "moderator"), "c-1")

	require.ErrorIs(t, err, domain.ErrContractNotPendingValidation)
}
