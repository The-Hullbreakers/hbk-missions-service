//go:build integration

package dynamodb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-MAN-002: 200 staged and verified facts still commit with one
// deterministic manifest and a FINALIZED season, not 200 outbox writes.
func TestSeasonManifestCommit_Given200VerifiedFacts_WhenCommitted_ThenOneManifestAndFinalizedSeason(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-commit-many"
	seasons := missiondynamo.NewSeasonStore(client, table)
	season := domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)
	require.NoError(t, seasons.Create(season))
	attempt, err := missiondynamo.NewAdjudicationLock(client, table).Acquire(seasonID)
	require.NoError(t, err)
	facts := make([]domain.PlacementFact, 200)
	for i := range facts {
		facts[i], err = domain.NewPlacementFact(domain.PlacementFactParams{
			SeasonID: seasonID, Category: "BOUNTY_HUNTING",
			Track: domain.TrackIndividual, RecipientID: fmt.Sprintf("p-%03d", i+1),
			Rank: i + 1, FinalMerits: 200 - i,
		})
		require.NoError(t, err)
	}
	require.NoError(t, missiondynamo.NewPlacementFactStager(client, table).Stage(attempt, facts))
	require.NoError(t, missiondynamo.NewStagedFactVerifier(client, table).Verify(attempt, facts))
	set, err := domain.NewStagedFactSet(facts)
	require.NoError(t, err)
	committer := missiondynamo.NewFinalizationCommitter(client, table)

	// Act
	err = committer.Commit(season, attempt, facts)

	// Assert
	require.NoError(t, err)
	finalized, readErr := seasons.Get(seasonID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.SeasonStateFinalized, finalized.State)
	seasonRow, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "SEASON"},
			"SK": &types.AttributeValueMemberS{Value: seasonID},
		},
	})
	require.NoError(t, readErr)
	assert.Equal(t, attempt.ID, seasonRow.Item["AdjudicationLock"].(*types.AttributeValueMemberS).Value)
	manifest, readErr := client.Client().GetItem(t.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(table), ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OUTBOX"},
			"SK": &types.AttributeValueMemberS{Value: "MANIFEST#" + seasonID + "#" + attempt.ID},
		},
	})
	require.NoError(t, readErr)
	require.NotEmpty(t, manifest.Item)
	assert.Equal(t, "PENDING", manifest.Item["Status"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, attempt.ID, manifest.Item["AttemptID"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "200", manifest.Item["FactCount"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, set.Hash(), manifest.Item["FactHash"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "0", manifest.Item["NextIndex"].(*types.AttributeValueMemberN).Value)
	intents, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "OUTBOX"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Len(t, intents.Items, 1, "only the constant-size manifest is publishable")
}
