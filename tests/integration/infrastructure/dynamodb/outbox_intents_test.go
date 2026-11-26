package dynamodb_test

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func seedFreezingSeason(t *testing.T, store *missiondynamo.SeasonStore, seasonID string) {
	t.Helper()
	require.NoError(t, store.Create(domain.NewSeason(seasonID, domain.SeasonStateFreezing)))
}

func TestSeasonStoreOutbox_ListPendingReturnsOnlyPendingIntents(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	facts := []domain.PlacementFact{makePlacementFact(t, "player-1", 1), makePlacementFact(t, "player-2", 2)}
	require.NoError(t, store.SaveFinalized(domain.NewSeason("season-finalize", domain.SeasonStateFinalized), facts))
	require.NoError(t, store.MarkPublished(ctx, facts[0].EventID))

	pending, err := store.ListPending(ctx, 10)

	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, facts[1].EventID, pending[0].EventID)
	require.Equal(t, "missions.season.placement-awarded", pending[0].EventType)
	require.NotEmpty(t, pending[0].Envelope)
}

func TestSeasonStoreOutbox_ListPendingHonorsLimit(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	facts := []domain.PlacementFact{
		makePlacementFact(t, "player-1", 1),
		makePlacementFact(t, "player-2", 2),
		makePlacementFact(t, "player-3", 3),
	}
	require.NoError(t, store.SaveFinalized(domain.NewSeason("season-finalize", domain.SeasonStateFinalized), facts))

	pending, err := store.ListPending(ctx, 2)

	require.NoError(t, err)
	require.Len(t, pending, 2)
}

func TestSeasonStoreOutbox_MarkPublishedIsIdempotent(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	fact := makePlacementFact(t, "player-1", 1)
	require.NoError(t, store.SaveFinalized(domain.NewSeason("season-finalize", domain.SeasonStateFinalized), []domain.PlacementFact{fact}))

	require.NoError(t, store.MarkPublished(ctx, fact.EventID))
	require.NoError(t, store.MarkPublished(ctx, fact.EventID))
	pending, err := store.ListPending(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, pending)
}

func TestSeasonStoreOutbox_ListPendingPagesPastPublishedPrefix(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")
	seedFreezingSeason(t, store, "season-finalize")
	facts := []domain.PlacementFact{
		makePlacementFact(t, "player-1", 1),
		makePlacementFact(t, "player-2", 2),
		makePlacementFact(t, "player-3", 3),
		makePlacementFact(t, "player-4", 4),
	}
	sort.Slice(facts, func(i, j int) bool { return facts[i].EventID < facts[j].EventID })
	require.NoError(t, store.SaveFinalized(domain.NewSeason("season-finalize", domain.SeasonStateFinalized), facts))
	require.NoError(t, store.MarkPublished(ctx, facts[0].EventID))
	require.NoError(t, store.MarkPublished(ctx, facts[1].EventID))

	// limit=1 must page past the two PUBLISHED intents at the head of the
	// partition: DynamoDB applies Limit before the status filter.
	pending, err := store.ListPending(ctx, 1)

	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, facts[2].EventID, pending[0].EventID)
}

func TestSeasonStoreOutbox_MarkPublishedMissingIntentCreatesNoRow(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewSeasonStore(client, "hbk-missions-outbox")

	err = store.MarkPublished(ctx, "evt-does-not-exist")

	require.Error(t, err)
	require.Empty(t, scanPartition(t, ctx, client, "hbk-missions-outbox", "OUTBOX"))
}
