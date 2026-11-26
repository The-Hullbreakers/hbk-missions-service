package dynamodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestSeasonStoreSaveFinalized_UnknownSeasonWritesNothing(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	season := domain.NewSeason("season-ghost", domain.SeasonStateFinalized)
	facts := []domain.PlacementFact{makePlacementFact(t, "player-1", 1)}

	err = store.SaveFinalized(season, facts)

	require.ErrorIs(t, err, domain.ErrSeasonNotFound)
	require.Empty(t, scanPartition(t, ctx, client, "hbk-missions-outbox", "OUTBOX"))
	require.Empty(t, scanPartition(t, ctx, client, "hbk-missions-outbox", "SEASON"))
}

func TestSeasonStoreSaveFinalized_RefinalizeIsRejectedAndKeepsState(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	season := domain.NewSeason("season-finalize", domain.SeasonStateFinalized)
	facts := []domain.PlacementFact{makePlacementFact(t, "player-1", 1), makePlacementFact(t, "player-2", 2)}
	require.NoError(t, store.SaveFinalized(season, facts))
	require.NoError(t, store.MarkPublished(ctx, facts[0].EventID))

	err = store.SaveFinalized(season, facts)

	require.ErrorIs(t, err, domain.ErrInvalidSeasonStateTransition)
	got, getErr := store.Get("season-finalize")
	require.NoError(t, getErr)
	require.Equal(t, domain.SeasonStateFinalized, got.State)
	items := scanPartition(t, ctx, client, "hbk-missions-outbox", "OUTBOX")
	require.Len(t, items, 2)
	statuses := map[string]string{}
	for _, item := range items {
		statuses[item["SK"].(*types.AttributeValueMemberS).Value] = item["Status"].(*types.AttributeValueMemberS).Value
	}
	require.Equal(t, "PUBLISHED", statuses["PLACEMENT#"+facts[0].EventID], "refinalize must not reset a published intent")
	require.Equal(t, "PENDING", statuses["PLACEMENT#"+facts[1].EventID])
}

func TestSeasonStoreSaveFinalized_RejectsOversizedTransaction(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	facts := make([]domain.PlacementFact, 0, 200)
	for i := 1; i <= 200; i++ {
		facts = append(facts, makePlacementFact(t, fmt.Sprintf("player-%d", i), i))
	}

	err = store.SaveFinalized(domain.NewSeason("season-finalize", domain.SeasonStateFinalized), facts)

	require.ErrorIs(t, err, missiondynamo.ErrOutboxBatchTooLarge)
	got, getErr := store.Get("season-finalize")
	require.NoError(t, getErr)
	require.Equal(t, domain.SeasonStateFreezing, got.State, "oversized batch must leave the season untouched")
	require.Empty(t, scanPartition(t, ctx, client, "hbk-missions-outbox", "OUTBOX"))
}
