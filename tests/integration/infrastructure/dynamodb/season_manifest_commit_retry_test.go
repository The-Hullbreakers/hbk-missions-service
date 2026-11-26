//go:build integration

package dynamodb_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-MAN-002: a retried commit cannot replace an existing PUBLISHED
// manifest or reset its cursor after the season was finalized.
func TestSeasonManifestCommit_ExistingPublishedManifestIsNotOverwritten(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-commit-retried"
	seasons := missiondynamo.NewSeasonStore(client, table)
	season := domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)
	require.NoError(t, seasons.Create(season))
	attempt, err := missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	set, err := domain.NewStagedFactSet(nil)
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewStagedFactVerifier(client, table).Verify(attempt, nil))
	_, err = client.Client().UpdateItem(t.Context(), &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
		UpdateExpression:         aws.String("SET #state = :finalized"),
		ConditionExpression:      aws.String("#state = :freezing"),
		ExpressionAttributeNames: map[string]string{"#state": "State"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":finalized": &types.AttributeValueMemberS{Value: string(domain.SeasonStateFinalized)},
			":freezing":  &types.AttributeValueMemberS{Value: string(domain.SeasonStateFreezing)},
		},
	})
	require.NoError(t, err)
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "OUTBOX"},
		"SK": &types.AttributeValueMemberS{Value: "MANIFEST#" + seasonID + "#" + attempt.ID},
	}
	_, err = client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"PK": key["PK"], "SK": key["SK"],
			"SeasonID":  &types.AttributeValueMemberS{Value: seasonID},
			"AttemptID": &types.AttributeValueMemberS{Value: attempt.ID},
			"Cutoff":    &types.AttributeValueMemberS{Value: attempt.Cutoff.Format(time.RFC3339Nano)},
			"FactCount": &types.AttributeValueMemberN{Value: "0"},
			"FactHash":  &types.AttributeValueMemberS{Value: set.Hash()},
			"NextIndex": &types.AttributeValueMemberN{Value: "0"},
			"Status":    &types.AttributeValueMemberS{Value: "PUBLISHED"},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	require.NoError(t, err)
	before, err := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, err)
	committer := missiondynamo.NewFinalizationCommitter(client, table)

	// Act
	err = committer.Commit(season, attempt, nil)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidSeasonStateTransition)
	after, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Equal(t, before.Item, after.Item, "retry cannot reset published progress")
}
