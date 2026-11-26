package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func newContractStoreForList(t *testing.T, table string) *missiondynamo.ContractStore {
	t.Helper()
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, table)
	require.NoError(t, err)
	return missiondynamo.NewContractStore(client, table)
}

func TestContractStore_ListByPlayer_FiltersRosterSeasonAndStatus(t *testing.T) {
	// Arrange
	store := newContractStoreForList(t, "hbk-missions-contract-list-1")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	mine := domain.NewMissionContract("c1", offer, []string{"player-1"})
	otherPlayer := domain.NewMissionContract("c2", offer, []string{"player-2"})
	otherSeasonOffer := domain.NewMissionOffer("offer-2", "BOUNTY", "season-2")
	otherSeason := domain.NewMissionContract("c4", otherSeasonOffer, []string{"player-1"})
	done := domain.NewMissionContractFromState("c5", offer, []string{"player-1"}, domain.ContractCompleted, true, "", "season-1", nil, nil, nil)
	for _, c := range []*domain.MissionContract{mine, otherPlayer, otherSeason, done} {
		err := store.Save(c)
		require.NoError(t, err)
	}
	filter := domain.ContractListFilter{SeasonID: "season-1", Status: domain.ContractStatusActive}

	// Act
	got, err := store.ListByPlayer("player-1", filter)

	// Assert
	require.NoError(t, err)
	require.False(t, len(got) != 1 || got[0].ContractID != "c1")
}

func TestContractStore_ListByPlayer_StatusAllReturnsTerminal(t *testing.T) {
	// Arrange
	store := newContractStoreForList(t, "hbk-missions-contract-list-2")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	active := domain.NewMissionContract("c1", offer, []string{"player-1"})
	done := domain.NewMissionContractFromState("c2", offer, []string{"player-1"}, domain.ContractCompleted, true, "", "season-1", nil, nil, nil)
	_ = store.Save(active)
	_ = store.Save(done)
	filter := domain.ContractListFilter{Status: domain.ContractStatusAll}

	// Act
	got, err := store.ListByPlayer("player-1", filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestContractStore_ListByPlayer_ReturnsErrorWhenDynamoUnavailable(t *testing.T) {
	// Arrange
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", "http://localhost:1")
	require.NoError(t, err)
	store := missiondynamo.NewContractStore(client, "hbk-missions-contract-list-err")

	// Act
	_, err = store.ListByPlayer("player-1", domain.ContractListFilter{})

	// Assert
	require.Error(t, err)
}
