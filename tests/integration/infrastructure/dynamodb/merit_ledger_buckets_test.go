package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// ListBuckets must deduplicate the category/track pairs seen in a season's
// credits and ignore other seasons.
func TestMeritLedgerStore_ListBuckets_DeduplicatesAndScopes(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	// Arrange
	store := missiondynamo.NewMeritLedgerStore(client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	seedActiveSeason(t, client, "hbk-missions", "season-2")
	credits := []domain.MeritCredit{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 100, ContractID: "c-1", ObjectiveID: "o-1", UnitID: "u-1", ConfirmationID: "conf-1"},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Amount: 50, ContractID: "c-2", ObjectiveID: "o-1", UnitID: "u-2", ConfirmationID: "conf-2"},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam, RecipientID: "team-1", Amount: 200, ContractID: "c-3", ObjectiveID: "o-1", UnitID: "u-3", ConfirmationID: "conf-3"},
		{SeasonID: "season-2", Category: "EXPLORATION", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 10, ContractID: "c-4", ObjectiveID: "o-2", UnitID: "u-4", ConfirmationID: "conf-4"},
	}
	for _, credit := range credits {
		err := store.Save(credit)
		require.NoError(t, err)
	}

	// Act
	buckets, err := store.ListBuckets("season-1")

	// Assert
	require.NoError(t, err)
	require.Len(t, buckets, 2)
	seen := map[string]bool{}
	for _, b := range buckets {
		seen[b.Category+"/"+string(b.Track)] = true
	}
	require.False(t, !seen["BOUNTY_HUNTING/INDIVIDUAL"] || !seen["BOUNTY_HUNTING/TEAM"])
}
