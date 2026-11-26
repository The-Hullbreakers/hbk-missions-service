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

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-MAN-001: a retry resumes from persisted progress rather than
// re-publishing facts that have already been acknowledged.
func TestManifestPublisher_ResumesOnlyRemainingStagedFact(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-manifest-resume"
	cutoff := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewAdjudicationAttempt(seasonID, cutoff)
	require.NoError(t, err)
	first, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-1", Rank: 1, FinalMerits: 100,
	})
	require.NoError(t, err)
	second, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-2", Rank: 2, FinalMerits: 90,
	})
	require.NoError(t, err)
	facts := []domain.PlacementFact{first, second}
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, facts))
	set, err := domain.NewStagedFactSet(facts)
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
			"Cutoff":    &types.AttributeValueMemberS{Value: cutoff.Format(time.RFC3339Nano)},
			"FactCount": &types.AttributeValueMemberN{Value: "2"},
			"FactHash":  &types.AttributeValueMemberS{Value: set.Hash()},
			"NextIndex": &types.AttributeValueMemberN{Value: "1"},
			"Status":    &types.AttributeValueMemberS{Value: "PENDING"},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	require.NoError(t, err)
	store := missiondynamo.NewSeasonStore(client, table)
	events := &manifestEventRecorder{}
	publisher := application.NewPlacementFactPublisher(store, store, events, store, store)

	// Act
	err = publisher.PublishPending(t.Context(), 10)

	// Assert
	require.NoError(t, err)
	require.Len(t, events.published, 1)
	assert.Equal(t, second.EventID, events.published[0].EventID)
	assert.NotEqual(t, first.EventID, events.published[0].EventID)
	stored, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.IsType(t, &types.AttributeValueMemberN{}, stored.Item["NextIndex"])
	require.IsType(t, &types.AttributeValueMemberS{}, stored.Item["Status"])
	assert.Equal(t, "2", stored.Item["NextIndex"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "PUBLISHED", stored.Item["Status"].(*types.AttributeValueMemberS).Value)
}
