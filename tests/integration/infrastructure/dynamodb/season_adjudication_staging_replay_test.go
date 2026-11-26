//go:build integration

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-FACT-001/-002 — staged rows are immutable: an identical replay
// is a no-op and a divergent replay under the same fact identity is a
// conflict that never overwrites. require.NoError is fixture-seeding only.
func TestSeasonAdjudicationStaging_GivenIdenticalStagedFact_WhenReplayed_ThenNoOp(t *testing.T) {
	// Arrange — a staged row already carrying the exact canonical digest.
	ctx := context.Background()
	tableName := fmt.Sprintf("hbk-chg0051-replay-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	const seasonID = "season-stage-replay"
	store := missiondynamo.NewSeasonStore(client, tableName)
	require.NoError(t, store.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))

	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	seededFact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual,
		RecipientID: "p-001", Rank: 1, FinalMerits: 100,
	})
	require.NoError(t, err)
	_, seedErr := client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "FACT#" + seasonID},
			"SK":           &types.AttributeValueMemberS{Value: attempt.ID + "#BOUNTY_HUNTING#" + string(domain.TrackIndividual) + "#p-001"},
			"Digest":       &types.AttributeValueMemberS{Value: seededFact.Digest()},
			"ReplayMarker": &types.AttributeValueMemberS{Value: "original"},
		},
	})
	require.NoError(t, seedErr)

	snapshot := &stageTestSnapshotStub{snap: application.AdjudicationSnapshot{
		Buckets: []application.AdjudicationBucket{
			{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, Entries: []application.AdjudicationEntry{
				{RecipientID: "p-001", Total: 100},
			}},
		},
	}}
	committer := &lockTestCommitStub{}
	stager := missiondynamo.NewPlacementFactStager(client, tableName)
	finalizer := application.NewSeasonFinalizer(store, &stageTestLockStub{attempt: attempt},
		snapshot, &lockTestRosterStub{}, stager, &lockTestVerifyStub{}, committer)

	// Act
	finalized, err := finalizer.Handle(seasonID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, finalized.State)
	require.Len(t, committer.facts, 1)
	assert.Equal(t, seededFact.Digest(), committer.facts[0].Digest())

	result, readErr := client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "FACT#" + seasonID},
			"SK": &types.AttributeValueMemberS{Value: attempt.ID + "#BOUNTY_HUNTING#" + string(domain.TrackIndividual) + "#p-001"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.NotNil(t, result.Item)
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["ReplayMarker"])
	assert.Equal(t, "original", result.Item["ReplayMarker"].(*types.AttributeValueMemberS).Value,
		"an identical replay is a no-op — the staged row is untouched")
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["Digest"])
	assert.Equal(t, seededFact.Digest(), result.Item["Digest"].(*types.AttributeValueMemberS).Value)
}
