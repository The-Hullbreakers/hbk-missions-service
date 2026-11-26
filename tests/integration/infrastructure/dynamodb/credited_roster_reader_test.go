//go:build integration

package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// SCN-SCC-AWARD-003 / TEST-INF-SNAP-001: only the contracts that actually
// contributed credits feed the team fact's member_ids — an uncredited
// contract of the same team and season is excluded.
func TestCreditedRosterReader_OnlyCreditedFrozenRostersFeedMembers(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-roster-1"
	seasons := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasons.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))
	offer := domain.NewMissionOffer("offer-team", "BOUNTY_HUNTING", seasonID)
	contracts := missiondynamo.NewContractStore(client, table)
	credited := domain.NewMissionContractFromState("c-1", offer, []string{"p-1", "p-2"},
		domain.ContractInProgress, true, "team-alpha", seasonID, nil, nil, nil)
	uncredited := domain.NewMissionContractFromState("c-2", offer, []string{"p-3"},
		domain.ContractInProgress, true, "team-alpha", seasonID, nil, nil, nil)
	require.NoError(t, contracts.Save(credited))
	require.NoError(t, contracts.Save(uncredited))
	reader := missiondynamo.NewCreditedRosterReader(client, table)

	// Act
	members, err := reader.Members(seasonID, "BOUNTY_HUNTING", "team-alpha", []string{"c-1"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"p-1", "p-2"}, members)
}
