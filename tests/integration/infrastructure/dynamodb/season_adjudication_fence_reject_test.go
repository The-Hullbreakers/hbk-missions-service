//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-FENCE-001: a rejection discards staged credits and stays legal
// after the lock; fencing every contract save would block this outcome.
func TestAdjudicationFence_LockedSeasonAllowsZeroCreditRejection(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID = "season-reject-locked", "contract-reject-locked"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	contracts := missiondynamo.NewContractStore(client, table)
	offer := domain.NewMissionOffer("offer-reject-locked", "BOUNTY_HUNTING", seasonID)
	contract := domain.NewMissionContract(contractID, offer, []string{"player-1"})
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "Capture", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	require.NoError(t, err)
	contract.Terms = terms
	member, err := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, err)
	require.NoError(t, contract.Start(member))
	require.NoError(t, contracts.Save(contract))
	confirmer := application.NewConfirmObjectiveHandler(contracts, contracts, seasons)
	_, err = confirmer.Handle(member, contractID, "objective-1", "unit-1", seasonID)
	require.NoError(t, err)
	submitter := application.NewSubmitContractHandler(contracts, contracts)
	require.NoError(t, submitter.Handle(member, contractID))
	_, err = missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	reviewer, err := domain.NewCallerIdentity("mod-1", []string{"admin"}, nil)
	require.NoError(t, err)
	rejector := application.NewRejectContractHandler(contracts, contracts)

	// Act
	err = rejector.Handle(reviewer, contractID)

	// Assert
	require.NoError(t, err)
	stored, readErr := contracts.Get(contractID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.ContractDisqualified, stored.State)
	assert.Empty(t, stored.PendingCredits())
	ledger := missiondynamo.NewMeritLedgerStore(client, table)
	assert.Equal(t, 0, ledger.TotalFor(seasonID, "BOUNTY_HUNTING", domain.TrackIndividual, "player-1"))
}
