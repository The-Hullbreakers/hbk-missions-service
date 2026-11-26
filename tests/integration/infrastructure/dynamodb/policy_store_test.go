package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// SaveActiveLimit + GetActiveLimit round-trip a persisted policy; a missing
// document falls back to the shipped default.
func TestPolicyStore_ActiveLimitRoundTrip(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	// Arrange
	store := missiondynamo.NewPolicyStore(client, "hbk-missions")

	// Act — missing document reads the shipped default
	policy, err := store.GetActiveLimit()

	// Assert
	require.NoError(t, err)
	require.False(t, policy.Default() != 5 || policy.EffectiveLimit("BOUNTY_HUNTING") != 5)

	// Act — persist a tuned policy with an override
	updated := domain.NewActiveLimitPolicy(3, map[string]int{"BOUNTY_HUNTING": 2})
	err = store.SaveActiveLimit(updated)
	require.NoError(t, err)
	got, err := store.GetActiveLimit()

	// Assert
	require.NoError(t, err)
	require.False(t, got.Default() != 3 || got.EffectiveLimit("BOUNTY_HUNTING") != 2 || got.EffectiveLimit("EXPLORATION") != 3)
}
