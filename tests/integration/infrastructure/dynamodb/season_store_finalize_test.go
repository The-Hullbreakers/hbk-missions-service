package dynamodb_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func makePlacementFact(t *testing.T, recipient string, rank int) domain.PlacementFact {
	t.Helper()
	fact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: "season-finalize", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual,
		RecipientID: recipient, Rank: rank, FinalMerits: 100 * rank,
	})
	require.NoError(t, err)
	return fact
}

func TestSeasonStoreSaveFinalized_PersistsSeasonAndOutboxAtomically(t *testing.T) {
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

	got, err := store.Get("season-finalize")
	require.NoError(t, err)
	require.Equal(t, domain.SeasonStateFinalized, got.State)

	items := scanPartition(t, ctx, client, "hbk-missions-outbox", "OUTBOX")
	require.Len(t, items, len(facts))
	correlationIDs := map[string]bool{}
	for _, item := range items {
		require.Equal(t, "PENDING", item["Status"].(*types.AttributeValueMemberS).Value)
		require.Equal(t, "missions.season.placement-awarded", item["EventType"].(*types.AttributeValueMemberS).Value)
		var envelope map[string]any
		require.NoError(t, json.Unmarshal([]byte(item["Envelope"].(*types.AttributeValueMemberS).Value), &envelope))
		require.Equal(t, "1.0", envelope["event_version"])
		require.Equal(t, "hbk-missions-service", envelope["source"])
		require.Equal(t, item["SK"].(*types.AttributeValueMemberS).Value, "PLACEMENT#"+envelope["event_id"].(string))
		occurredAt, ok := envelope["occurred_at"].(string)
		require.True(t, ok, "envelope must carry occurred_at")
		_, err := time.Parse(time.RFC3339, occurredAt)
		require.NoError(t, err)
		correlationID, ok := envelope["correlation_id"].(string)
		require.True(t, ok, "envelope must carry correlation_id")
		require.NotEmpty(t, correlationID)
		correlationIDs[correlationID] = true
	}
	require.Len(t, correlationIDs, 1, "facts of one finalization share the correlation_id")
}

func TestSeasonStoreSaveFinalized_WithoutFactsWritesOnlySeason(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	season := domain.NewSeason("season-finalize", domain.SeasonStateFinalized)

	require.NoError(t, store.SaveFinalized(season, nil))
	require.Empty(t, scanPartition(t, ctx, client, "hbk-missions-outbox", "OUTBOX"))
}

func scanPartition(t *testing.T, ctx context.Context, client *missiondynamo.DynamoDBClient, table, pk string) []map[string]types.AttributeValue {
	t.Helper()
	result, err := client.Client().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(table),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
		},
	})
	require.NoError(t, err)
	return result.Items
}
