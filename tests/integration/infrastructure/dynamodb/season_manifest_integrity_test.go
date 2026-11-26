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

// TEST-INF-MAN-001: the digest alone cannot authorize a staged row whose
// sort key would change the ordered progress cursor.
func TestManifestPublisher_TamperedStagedKeyCannotPublish(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-manifest-tampered-key"
	cutoff := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewAdjudicationAttempt(seasonID, cutoff)
	require.NoError(t, err)
	fact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-1", Rank: 1, FinalMerits: 100,
	})
	require.NoError(t, err)
	set, err := domain.NewStagedFactSet([]domain.PlacementFact{fact})
	require.NoError(t, err)
	_, err = client.Client().PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "FACT#" + seasonID},
			"SK":          &types.AttributeValueMemberS{Value: attempt.ID + "#BOUNTY_HUNTING#INDIVIDUAL#wrong"},
			"SeasonID":    &types.AttributeValueMemberS{Value: seasonID},
			"AttemptID":   &types.AttributeValueMemberS{Value: attempt.ID},
			"Category":    &types.AttributeValueMemberS{Value: fact.Category},
			"Track":       &types.AttributeValueMemberS{Value: string(fact.Track)},
			"RecipientID": &types.AttributeValueMemberS{Value: fact.RecipientID},
			"EventID":     &types.AttributeValueMemberS{Value: fact.EventID},
			"Rank":        &types.AttributeValueMemberN{Value: "1"},
			"FinalMerits": &types.AttributeValueMemberN{Value: "100"},
			"MemberIDs":   &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
			"Digest":      &types.AttributeValueMemberS{Value: fact.Digest()},
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
			"Cutoff":    &types.AttributeValueMemberS{Value: cutoff.Format(time.RFC3339Nano)},
			"FactCount": &types.AttributeValueMemberN{Value: "1"},
			"FactHash":  &types.AttributeValueMemberS{Value: set.Hash()},
			"NextIndex": &types.AttributeValueMemberN{Value: "0"},
			"Status":    &types.AttributeValueMemberS{Value: "PENDING"},
		},
	})
	require.NoError(t, err)
	store := missiondynamo.NewSeasonStore(client, table)
	events := &manifestEventRecorder{}
	publisher := application.NewPlacementFactPublisher(store, store, events, store, store)

	// Act
	err = publisher.PublishPending(t.Context(), 10)

	// Assert
	require.Error(t, err)
	assert.Empty(t, events.published)
	stored, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Equal(t, "0", stored.Item["NextIndex"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "PENDING", stored.Item["Status"].(*types.AttributeValueMemberS).Value)
}
