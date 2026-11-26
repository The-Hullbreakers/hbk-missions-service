package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// REQ-SCC-005 / SCN-SCC-CREATE-EXISTS: the store itself must refuse to
// overwrite an existing season id — the invariant cannot be bypassed by a
// buggy caller.
func TestSeasonStore_CreateRejectsExistingSeasonID(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)

	tableName := "hbk-missions-" + t.Name()
	err = client.CreateMissionsTable(ctx, tableName)
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, tableName)

	store := missiondynamo.NewSeasonStore(client, tableName)
	finalized := domain.NewSeasonWithOfficial("season-x", domain.SeasonStateFinalized, true)
	err = store.Create(finalized)
	require.NoError(t, err)

	err = store.Create(domain.NewSeasonWithOfficial("season-x", domain.SeasonStateActive, false))
	require.ErrorIs(t, err, domain.ErrSeasonAlreadyExists)

	stored, err := store.Get("season-x")
	require.NoError(t, err)
	require.False(t, stored.State != domain.SeasonStateFinalized || !stored.Official)
}

// Save stays an upsert for lifecycle transitions — only Create is conditional.
func TestSeasonStore_SaveTransitionsExistingSeason(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)

	tableName := "hbk-missions-" + t.Name()
	err = client.CreateMissionsTable(ctx, tableName)
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, tableName)

	store := missiondynamo.NewSeasonStore(client, tableName)
	err = store.Create(domain.NewSeasonWithOfficial("season-x", domain.SeasonStateActive, true))
	require.NoError(t, err)

	err = store.Save(domain.NewSeasonWithOfficial("season-x", domain.SeasonStateFreezing, true))
	require.NoError(t, err)

	stored, err := store.Get("season-x")
	require.NoError(t, err)
	require.Equal(t, domain.SeasonStateFreezing, stored.State)
}
