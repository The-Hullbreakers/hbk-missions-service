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

// TEST-INF-MAN-002: a FREEZING season without the matching persisted lock
// cannot become FINALIZED or create a publishable empty manifest.
func TestSeasonManifestCommit_WithoutAcquiredLockCannotCommit(t *testing.T) {
	t.Parallel()
	// Arrange
	client, table, err := manifestStoreFixture(t)
	require.NoError(t, err)
	const seasonID = "season-commit-unlocked"
	seasons := missiondynamo.NewSeasonStore(client, table)
	season := domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)
	require.NoError(t, seasons.Create(season))
	attempt, err := domain.NewAdjudicationAttempt(seasonID, time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NoError(t, missiondynamo.NewStagedFactVerifier(client, table).Verify(attempt, nil))
	committer := missiondynamo.NewFinalizationCommitter(client, table)

	// Act
	err = committer.Commit(season, attempt, nil)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidSeasonStateTransition)
	stored, readErr := seasons.Get(seasonID)
	require.NoError(t, readErr)
	assert.Equal(t, domain.SeasonStateFreezing, stored.State)
	intents, readErr := client.Client().Query(t.Context(), &dynamodb.QueryInput{
		TableName: aws.String(table), KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "OUTBOX"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, readErr)
	assert.Empty(t, intents.Items)
}
