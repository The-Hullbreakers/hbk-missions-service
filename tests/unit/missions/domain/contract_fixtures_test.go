package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// contractMember builds the identity of the first roster member.
func contractMember(t *testing.T, playerID string) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity(playerID, nil, nil)
	require.NoError(t, err)
	return identity
}

// startContract puts the contract in IN_PROGRESS through the real transition.
func startContract(t *testing.T, contract *domain.MissionContract) {
	t.Helper()
	roster := contract.Roster()
	require.NoError(t, contract.Start(contractMember(t, roster[0])))
}

// freezeTerms sets the contract's frozen terms from the given objectives.
func freezeTerms(t *testing.T, contract *domain.MissionContract, objectives ...domain.Objective) {
	t.Helper()
	terms, err := domain.NewContractTerms(objectives)
	require.NoError(t, err)
	contract.Terms = terms
}
