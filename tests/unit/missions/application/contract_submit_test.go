package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSubmitContract_MemberSubmitsAndPersists(t *testing.T) {
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewSubmitContractHandler(store, store)

	err := handler.Handle(lifecycleMember(t), "c-1")

	require.NoError(t, err)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractPendingValidation)
	require.False(t, store.credits != nil || store.saveCalls != 1)
}

func TestSubmitContract_RejectsOutsider(t *testing.T) {
	contract := inProgressContract(t)
	store := &reviewStore{contract: contract}
	handler := application.NewSubmitContractHandler(store, store)

	err := handler.Handle(lifecycleOutsider(t), "c-1")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Nil(t, store.saved)
}

func TestSubmitContract_RejectsNonInProgress(t *testing.T) {
	contract := lifecycleFixture(t, nil)
	store := &reviewStore{contract: contract}
	handler := application.NewSubmitContractHandler(store, store)

	err := handler.Handle(lifecycleMember(t), "c-1")

	require.ErrorIs(t, err, domain.ErrContractStateInvalid)
}
