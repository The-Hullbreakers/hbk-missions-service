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

func verifierFacts(seasonID string) ([]domain.PlacementFact, error) {
	first, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-1", Rank: 1, FinalMerits: 100,
	})
	if err != nil {
		return nil, err
	}
	second, err := domain.NewPlacementFact(domain.PlacementFactParams{
		SeasonID: seasonID, Category: "BOUNTY_HUNTING",
		Track: domain.TrackIndividual, RecipientID: "p-2", Rank: 2, FinalMerits: 50,
	})
	return []domain.PlacementFact{first, second}, err
}

// TEST-INF-FACT-002: complete immutable staged facts pass verification;
// verification itself cannot create a publishable manifest.
func TestStagedFactVerifier_CompleteSetMatchesWithoutManifest(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-verified-complete"
	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	facts, err := verifierFacts(seasonID)
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, facts))
	verifier := missiondynamo.NewStagedFactVerifier(client, table)

	// Act
	err = verifier.Verify(attempt, facts)

	// Assert
	require.NoError(t, err)
	manifests, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "OUTBOX"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Empty(t, manifests.Items)
}

// TEST-INF-FACT-002: a missing staged fact must block a commit instead
// of accepting the caller's declared count and hash on trust.
func TestStagedFactVerifier_MissingRowBlocksCommit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-verified-missing"
	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	facts, err := verifierFacts(seasonID)
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, facts[:1]))
	verifier := missiondynamo.NewStagedFactVerifier(client, table)

	// Act
	err = verifier.Verify(attempt, facts)

	// Assert
	require.ErrorIs(t, err, domain.ErrStagedFactSetMismatch)
	manifests, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "OUTBOX"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Empty(t, manifests.Items)
}
