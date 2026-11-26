//go:build integration

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-INF-SWP-001 — the sweep query returns exactly the seasons holding an
// adjudication lock, with their persisted attempt identity and cutoff, and
// never exposes unlocked or missing seasons.
func TestAdjudicationLockSweep_GivenMixedSeasons_WhenListLocked_ThenOnlyLockedReturned(t *testing.T) {
	// Arrange — fresh test-owned table; one FREEZING season acquires the
	// lock, a second FREEZING season stays unlocked.
	ctx := context.Background()
	tableName := fmt.Sprintf("hbk-chg0051-swp-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	seasonStore := missiondynamo.NewSeasonStore(client, tableName)
	lock := missiondynamo.NewAdjudicationLock(client, tableName)
	require.NoError(t, seasonStore.Create(domain.NewSeason("season-locked", domain.SeasonStateFreezing)))
	require.NoError(t, seasonStore.Create(domain.NewSeason("season-open", domain.SeasonStateFreezing)))

	attempt, err := lock.Acquire("season-locked")
	require.NoError(t, err)

	// Act
	locked, err := lock.ListLocked(ctx)

	// Assert
	require.NoError(t, err)
	require.Len(t, locked, 1)
	assert.Equal(t, "season-locked", locked[0].SeasonID)
	assert.Equal(t, attempt.ID, locked[0].AttemptID)
	assert.WithinDuration(t, attempt.Cutoff, locked[0].Cutoff, time.Second)
}
