//go:build integration

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-E2E-AWD-002 — critical journey: an official FREEZING season with 200
// ranked merit entries finalizes through the real two-phase pipeline; the
// constant-size commit leaves one pending manifest declaring all 200 staged
// facts, each publishable once under its stable event identity.
func TestSeasonAdjudicationE2E_Given200RankedEntries_WhenFinalized_ThenManifestDeclares200Facts(t *testing.T) {
	// Arrange — fresh test-owned table; FREEZING official season; 200
	// distinct credited recipients so the snapshot ranks 200 entries.
	ctx := context.Background()
	tableName := fmt.Sprintf("hbk-chg0051-e2e-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	const seasonID = "season-e2e-200"
	seasonStore := missiondynamo.NewSeasonStore(client, tableName)
	require.NoError(t, seasonStore.Create(domain.NewSeasonWithOfficial(seasonID, domain.SeasonStateFreezing, true)))

	ledger := missiondynamo.NewMeritLedgerStore(client, tableName)
	for i := 1; i <= 200; i++ {
		require.NoError(t, ledger.Save(domain.MeritCredit{
			SeasonID:       seasonID,
			Category:       "BOUNTY_HUNTING",
			Track:          domain.TrackIndividual,
			RecipientID:    fmt.Sprintf("player-%03d", i),
			Amount:         i,
			ContractID:     fmt.Sprintf("contract-%03d", i),
			ObjectiveID:    "objective-1",
			UnitID:         fmt.Sprintf("unit-%03d", i),
			ConfirmationID: fmt.Sprintf("conf-%03d", i),
		}))
	}

	finalizer := application.NewSeasonFinalizer(
		seasonStore,
		missiondynamo.NewAdjudicationLock(client, tableName),
		missiondynamo.NewAdjudicationSnapshotReader(client, tableName),
		missiondynamo.NewCreditedRosterReader(client, tableName),
		missiondynamo.NewPlacementFactStager(client, tableName),
		missiondynamo.NewStagedFactVerifier(client, tableName),
		missiondynamo.NewFinalizationCommitter(client, tableName),
	)

	// Act
	finalized, err := finalizer.Handle(seasonID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, finalized.State)

	stored, err := seasonStore.Get(seasonID)
	require.NoError(t, err)
	assert.Equal(t, domain.SeasonStateFinalized, stored.State)

	manifests, err := seasonStore.ListPendingManifests(ctx, 10)
	require.NoError(t, err)
	require.Len(t, manifests, 1)
	manifest := manifests[0]
	assert.Equal(t, seasonID, manifest.SeasonID)
	assert.Equal(t, 200, manifest.FactCount)
	assert.NotEmpty(t, manifest.FactHash)
	assert.Equal(t, 0, manifest.NextIndex)

	remaining, err := seasonStore.ListRemaining(ctx, manifest)
	require.NoError(t, err)
	require.Len(t, remaining, 200)
	seen := map[string]bool{}
	for _, intent := range remaining {
		require.NotEmpty(t, intent.EventID)
		require.False(t, seen[intent.EventID], "duplicate event id %s", intent.EventID)
		seen[intent.EventID] = true
	}
}
