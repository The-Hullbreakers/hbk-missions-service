package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

func TestContractExpire_NonTerminalBecomesExpired(t *testing.T) {
	contract := lifecycleContract()
	_ = contract.Start(member(t))

	err := contract.Expire()

	require.NoError(t, err)
	require.Equal(t, domain.ContractExpired, contract.State)
}

func TestContractExpire_TerminalIsRejected(t *testing.T) {
	contract := lifecycleContract()
	contract.State = domain.ContractCompleted

	err := contract.Expire()

	require.ErrorIs(t, err, domain.ErrContractTerminal)
}
