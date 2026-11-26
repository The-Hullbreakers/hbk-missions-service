package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// contractGetter is a hand-written ContractReader double recording calls and
// returning a fixture contract or a not-found error.
type contractGetter struct {
	contract *domain.MissionContract
	err      error
	calls    int
}

func (g *contractGetter) Get(string) (*domain.MissionContract, error) {
	g.calls++
	return g.contract, g.err
}

func TestGetContract_ReviewerGetsDetail(t *testing.T) {
	contract := pendingContract(t, 0)
	reader := &contractGetter{contract: contract}
	handler := application.NewGetContractHandler(reader)

	got, err := handler.Handle(reviewer(t, "moderator"), contract.ContractID, domain.LanguageSpanish)

	require.NoError(t, err)
	require.Equal(t, 1, reader.calls)
	require.False(t, got == nil || got.ContractID != contract.ContractID)
}

func TestGetContract_RosterMemberGetsDetail(t *testing.T) {
	contract := pendingContract(t, 0)
	reader := &contractGetter{contract: contract}
	handler := application.NewGetContractHandler(reader)

	got, err := handler.Handle(lifecycleMember(t), contract.ContractID, domain.LanguageSpanish)

	require.NoError(t, err)
	require.False(t, got == nil || got.ContractID != contract.ContractID)
}

func TestGetContract_OutsiderDenied(t *testing.T) {
	reader := &contractGetter{contract: pendingContract(t, 0)}
	handler := application.NewGetContractHandler(reader)
	outsider, _ := domain.NewCallerIdentity("player-9", nil, nil)

	_, err := handler.Handle(outsider, "contract-1", domain.LanguageSpanish)

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

func TestGetContract_DepartedMemberDenied(t *testing.T) {
	contract := rewardedLifecycleFixture(t, 0)
	member := lifecycleMember(t)
	err := contract.Leave(member)
	require.NoError(t, err)
	reader := &contractGetter{contract: contract}
	handler := application.NewGetContractHandler(reader)

	_, err = handler.Handle(member, contract.ContractID, domain.LanguageSpanish)

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

func TestGetContract_UnknownContractPropagates(t *testing.T) {
	reader := &contractGetter{err: domain.ErrContractNil}
	handler := application.NewGetContractHandler(reader)

	_, err := handler.Handle(reviewer(t, "admin"), "contract-9", domain.LanguageSpanish)

	require.ErrorIs(t, err, domain.ErrContractNil)
}
