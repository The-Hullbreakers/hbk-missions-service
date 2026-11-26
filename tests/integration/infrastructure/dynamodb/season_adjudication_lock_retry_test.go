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

// TEST-INF-ADJ-001 (retry partition) — a persisted lock must be RESUMED, not
// reminted: a retried finalization reuses the existing attempt ID and cutoff
// verbatim, the cutoff is never re-minted or reset, and the public state
// stays FREEZING. require.NoError is used only for fixture seeding.
func TestSeasonAdjudicationLock_GivenPersistedLock_WhenFinalizationRetried_ThenReusesAttemptAndCutoff(t *testing.T) {
	// Arrange — fresh per-run table, FREEZING official season, and a
	// pre-seeded stable lock/cutoff simulating an orphaned prior attempt.
	ctx := context.Background()
	tableName := fmt.Sprintf("hbk-chg0051-adj-retry-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	const seasonID = "season-lock-retry"
	store := missiondynamo.NewSeasonStore(client, tableName)
	require.NoError(t, store.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))

	knownCutoff := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	_, seedErr := client.Client().UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		UpdateExpression: aws.String("SET AdjudicationLock = :lock, AdjudicationCutoff = :cutoff"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lock":   &types.AttributeValueMemberS{Value: "ADJ#" + seasonID},
			":cutoff": &types.AttributeValueMemberS{Value: knownCutoff.Format(time.RFC3339Nano)},
		},
	})
	require.NoError(t, seedErr)

	snapshot := &lockTestSnapshotStub{}
	roster := &lockTestRosterStub{}
	stager := &lockTestStageStub{}
	verifier := &lockTestVerifyStub{}
	committer := &lockTestCommitStub{}
	lock := missiondynamo.NewAdjudicationLock(client, tableName)
	finalizer := application.NewSeasonFinalizer(store, lock, snapshot, roster, stager, verifier, committer)

	// Act
	finalized, err := finalizer.Handle(seasonID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, finalized.State)
	for name, got := range map[string]domain.AdjudicationAttempt{
		"snapshot": snapshot.got, "stager": stager.attempt,
		"verifier": verifier.attempt, "committer": committer.attempt,
	} {
		assert.Equal(t, "ADJ#"+seasonID, got.ID, "%s must resume the persisted attempt", name)
		assert.True(t, got.Cutoff.Equal(knownCutoff), "%s must reuse the persisted cutoff", name)
	}
	assert.Equal(t, domain.SeasonStateFreezing, committer.season.State)
	assert.False(t, roster.called)

	result, readErr := client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.NotNil(t, result.Item)
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["State"])
	assert.Equal(t, string(domain.SeasonStateFreezing),
		result.Item["State"].(*types.AttributeValueMemberS).Value)
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["AdjudicationLock"])
	assert.Equal(t, "ADJ#"+seasonID,
		result.Item["AdjudicationLock"].(*types.AttributeValueMemberS).Value)
	require.IsType(t, &types.AttributeValueMemberS{}, result.Item["AdjudicationCutoff"])
	assert.Equal(t, knownCutoff.Format(time.RFC3339Nano),
		result.Item["AdjudicationCutoff"].(*types.AttributeValueMemberS).Value,
		"the persisted cutoff must survive the retry verbatim — no re-minting")
}
