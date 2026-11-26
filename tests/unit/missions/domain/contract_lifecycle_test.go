package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func lifecycleContract() *domain.MissionContract {
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	return contract
}

func member(t *testing.T) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, err)
	return identity
}

func outsider(t *testing.T) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity("player-9", nil, nil)
	require.NoError(t, err)
	return identity
}

func TestContractStart_AcceptedBecomesInProgress(t *testing.T) {
	contract := lifecycleContract()

	err := contract.Start(member(t))

	require.NoError(t, err)
	require.Equal(t, domain.ContractInProgress, contract.State)
}

func TestContractStart_RejectsNonRosterCaller(t *testing.T) {
	contract := lifecycleContract()

	err := contract.Start(outsider(t))

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Equal(t, domain.ContractAccepted, contract.State)
}

func TestContractStart_RejectsNonAccepted(t *testing.T) {
	contract := lifecycleContract()
	contract.State = domain.ContractInProgress

	err := contract.Start(member(t))

	require.ErrorIs(t, err, domain.ErrContractStateInvalid)
}

func TestContractAbandon_ReleasesSlotAndRejectsOutsider(t *testing.T) {
	contract := lifecycleContract()
	_ = contract.Start(member(t))

	err := contract.Abandon(outsider(t))
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	err = contract.Abandon(member(t))
	require.NoError(t, err)
	require.Equal(t, domain.ContractAbandoned, contract.State)
}
