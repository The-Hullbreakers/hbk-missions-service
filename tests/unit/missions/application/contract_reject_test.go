package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestRejectContract_ModeratorDisqualifiesWithoutCredits(t *testing.T) {
	contract := pendingContract(t, 200)
	store := &reviewStore{contract: contract}
	handler := application.NewRejectContractHandler(store, store)

	err := handler.Handle(reviewer(t, "moderator"), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractDisqualified)
	require.False(t, store.credits != nil || store.saveCalls != 1)
}

func TestRejectContract_AdminRejects(t *testing.T) {
	contract := pendingContract(t, 0)
	store := &reviewStore{contract: contract}
	handler := application.NewRejectContractHandler(store, store)

	err := handler.Handle(reviewer(t, "admin"), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractDisqualified)
}

func TestRejectContract_MemberCannotAdjudicate(t *testing.T) {
	contract := pendingContract(t, 200)
	store := &reviewStore{contract: contract}
	handler := application.NewRejectContractHandler(store, store)

	err := handler.Handle(lifecycleMember(t), "c-1")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Nil(t, store.saved)
}

func TestRejectContract_RejectsNonPending(t *testing.T) {
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewRejectContractHandler(store, store)

	err := handler.Handle(reviewer(t, "moderator"), "c-1")

	require.ErrorIs(t, err, domain.ErrContractNotPendingValidation)
}
