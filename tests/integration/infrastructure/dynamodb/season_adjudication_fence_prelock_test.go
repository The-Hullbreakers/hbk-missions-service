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

// TEST-INF-FENCE-001: FREEZING alone does not block a confirmation;
// the lock is the point at which ranking-relevant writes must stop.
func TestAdjudicationFence_FreezingBeforeLockAllowsInProgressCredit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID, contractID = "season-before-lock", "contract-before-lock"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	contracts := missiondynamo.NewContractStore(client, table)
	offer := domain.NewMissionOffer("offer-before-lock", "BOUNTY_HUNTING", seasonID)
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
	handler := application.NewConfirmObjectiveHandlerWithEvidence(contracts, contracts, seasons, stubEvidenceReader{})

	// Act
	credit, err := handler.HandleWithMedia(member, contractID, "objective-1", "unit-1", "evidence/"+contractID+"/e1", seasonID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, seasonID, credit.SeasonID)
	stored, readErr := contracts.Get(contractID)
	require.NoError(t, readErr)
	require.Len(t, stored.PendingCredits(), 1)
	assert.Equal(t, credit.ConfirmationID, stored.PendingCredits()[0].ConfirmationID)
	season, readErr := seasons.Get(seasonID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.SeasonStateFreezing, season.State)
}
