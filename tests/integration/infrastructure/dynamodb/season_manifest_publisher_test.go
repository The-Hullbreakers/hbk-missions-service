//go:build integration

package dynamodb_test

import (
	"context"
	"encoding/json"
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

type manifestEventRecorder struct{ published []application.PendingIntent }

func (r *manifestEventRecorder) Publish(_ context.Context, intent application.PendingIntent) error {
	r.published = append(r.published, intent)
	return nil
}

// TEST-INF-MAN-001: an uncommitted staged fact is not publishable.
func TestManifestPublisher_StagedFactWithoutManifestIsNotPublished(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-manifest-orphan"
	cutoff := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewAdjudicationAttempt(seasonID, cutoff)
	require.NoError(t, err)
	fact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-orphan", Rank: 1, FinalMerits: 100,
	})
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, []domain.PlacementFact{fact}))
	store := missiondynamo.NewSeasonStore(client, table)
	events := &manifestEventRecorder{}
	publisher := application.NewPlacementFactPublisher(store, store, events, store, store)

	// Act
	err = publisher.PublishPending(t.Context(), 10)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, events.published)
}

// TEST-INF-MAN-001: one committed manifest publishes only its staged fact,
// with the frozen cutoff, and persists its progress before completion.
func TestManifestPublisher_CommittedManifestPublishesAndCompletes(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-manifest-committed"
	cutoff := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewAdjudicationAttempt(seasonID, cutoff)
	require.NoError(t, err)
	fact, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-1", Rank: 1, FinalMerits: 100,
	})
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, []domain.PlacementFact{fact}))
	set, err := domain.NewStagedFactSet([]domain.PlacementFact{fact})
	require.NoError(t, err)
	manifestSK := "MANIFEST#" + seasonID + "#" + attempt.ID
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "OUTBOX"},
		"SK": &types.AttributeValueMemberS{Value: manifestSK},
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
	assert.Equal(t, fact.EventID, events.published[0].EventID)
	assert.Equal(t, "missions.season.placement-awarded", events.published[0].EventType)
	var envelope struct {
		EventID    string `json:"event_id"`
		OccurredAt string `json:"occurred_at"`
		Payload    struct {
			RecipientID string `json:"recipient_id"`
			Rank        int    `json:"rank"`
		} `json:"payload"`
	}
	require.NoError(t, json.Unmarshal([]byte(events.published[0].Envelope), &envelope))
	assert.Equal(t, fact.EventID, envelope.EventID)
	assert.Equal(t, cutoff.Format(time.RFC3339Nano), envelope.OccurredAt)
	assert.Equal(t, "p-1", envelope.Payload.RecipientID)
	assert.Equal(t, 1, envelope.Payload.Rank)
	stored, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), Key: key, ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	require.IsType(t, &types.AttributeValueMemberN{}, stored.Item["NextIndex"])
	require.IsType(t, &types.AttributeValueMemberS{}, stored.Item["Status"])
	assert.Equal(t, "1", stored.Item["NextIndex"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "PUBLISHED", stored.Item["Status"].(*types.AttributeValueMemberS).Value)
}
