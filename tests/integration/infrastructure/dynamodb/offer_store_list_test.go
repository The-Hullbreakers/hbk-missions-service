package dynamodb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func newOfferStoreForList(t *testing.T, table string) *missiondynamo.OfferStore {
	t.Helper()
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	_, _ = client.Client().DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: &table})
	err = client.CreateMissionsTable(ctx, table)
	require.NoError(t, err)
	return missiondynamo.NewOfferStore(client, table)
}

func TestOfferStore_List_FiltersSeasonCategoryAndAvailability(t *testing.T) {
	// Arrange
	store := newOfferStoreForList(t, "hbk-missions-offer-list-1")
	mine := domain.NewMissionOffer("offer-a1", "BOUNTY", "season-1")
	otherSeason := domain.NewMissionOffer("offer-a2", "BOUNTY", "season-2")
	otherCategory := domain.NewMissionOffer("offer-a3", "EXPLORATION", "season-1")
	for _, o := range []domain.MissionOffer{mine, otherSeason, otherCategory} {
		err := store.Save(o)
		require.NoError(t, err)
	}
	filter := domain.OfferListFilter{SeasonID: "season-1", Categories: []string{"BOUNTY"}, Status: domain.OfferStatusAvailable}

	// Act
	got, err := store.List(filter)

	// Assert
	require.NoError(t, err)
	require.False(t, len(got) != 1 || got[0].OfferID != "offer-a1")
}

func TestOfferStore_List_StatusAllListsEveryOffer(t *testing.T) {
	// Arrange
	store := newOfferStoreForList(t, "hbk-missions-offer-list-2")
	open := domain.NewMissionOffer("offer-open", "BOUNTY", "season-1")
	second := domain.NewMissionOffer("offer-full", "BOUNTY", "season-1")
	_ = store.Save(open)
	_ = store.Save(second)
	filter := domain.OfferListFilter{Status: domain.OfferStatusAll}

	// Act
	got, err := store.List(filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestOfferStore_List_ReturnsErrorWhenDynamoUnavailable(t *testing.T) {
	// Arrange
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", "http://localhost:1")
	require.NoError(t, err)
	store := missiondynamo.NewOfferStore(client, "hbk-missions-offer-list-err")

	// Act
	_, err = store.List(domain.OfferListFilter{})

	// Assert
	require.Error(t, err)
}
