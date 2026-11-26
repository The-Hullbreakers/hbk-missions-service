package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// REQ-SCC-008 / SCN-SCC-TIE-COMPLETED: the CompletedMissions tie-break counts
// only contracts that reached COMPLETED — credits from an IN_PROGRESS
// contract do not count.
func TestMeritLedgerStore_ListStandingsCountsOnlyCompletedContracts(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)

	tableName := "hbk-missions-" + t.Name()
	err = client.CreateMissionsTable(ctx, tableName)
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, tableName)

	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	completed := domain.NewMissionContract("contract-done", offer, []string{"player-1"})
	completed.State = domain.ContractCompleted
	inProgress := domain.NewMissionContract("contract-live", offer, []string{"player-2"})
	inProgress.State = domain.ContractInProgress

	contractStore := missiondynamo.NewContractStore(client, tableName)
	for _, contract := range []*domain.MissionContract{completed, inProgress} {
		err := contractStore.Save(contract)
		require.NoError(t, err)
	}

	ledgerStore := missiondynamo.NewMeritLedgerStore(client, tableName)
	seedActiveSeason(t, client, tableName, "season-1")
	for _, credit := range []domain.MeritCredit{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 100, ContractID: "contract-done", ObjectiveID: "o1", UnitID: "u1", ConfirmationID: "c1"},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Amount: 100, ContractID: "contract-live", ObjectiveID: "o1", UnitID: "u2", ConfirmationID: "c2"},
	} {
		err := ledgerStore.Save(credit)
		require.NoError(t, err)
	}

	standings, err := ledgerStore.ListStandings("season-1", "BOUNTY_HUNTING", domain.TrackIndividual, 0)
	require.NoError(t, err)
	require.Len(t, standings, 2)
	require.Equal(t, "player-1", standings[0].RecipientID)
	require.False(t, standings[0].CompletedMissions != 1 || standings[1].CompletedMissions != 0)
}
