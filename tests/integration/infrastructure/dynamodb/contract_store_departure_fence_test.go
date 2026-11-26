package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func defXTeamOffer() domain.MissionOffer {
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	return offer
}

func startedTeamContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := domain.NewTeamMissionContract("c-1", defXTeamOffer(), []string{"player-1", "player-2", "player-3"}, "team-9")
	require.NoError(t, contract.Start(caller(t, "player-1")))
	return contract
}

// removeContractAttribute strips an attribute from a stored contract row,
// simulating a legacy item written before the attribute existed.
func removeContractAttribute(t *testing.T, client *missiondynamo.DynamoDBClient, table, contractID, attribute string) {
	t.Helper()
	_, err := client.Client().UpdateItem(t.Context(), &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
			"SK": &types.AttributeValueMemberS{Value: contractID},
		},
		UpdateExpression:         aws.String("REMOVE #a"),
		ExpressionAttributeNames: map[string]string{"#a": attribute},
	})
	require.NoError(t, err)
}

// REG-0053-03 edge A: a legacy row missing the Departed attribute may only
// satisfy the stale-write fence when the leaving snapshot expects zero prior
// departures. A snapshot carrying a prior departure must still be rejected.
func TestContractStore_SaveDepartureRejectsMissingDepartedWithPriorCount(t *testing.T) {
	// Arrange: persisted c-1 has player-1 departed; the snapshot was fetched
	// while the row still carried Departed, then the attribute is stripped to
	// simulate a legacy item.
	store, client, table := chg0053Fixture(t, "t13-depart-missing")
	contract := startedTeamContract(t)
	require.NoError(t, store.Save(contract))
	require.NoError(t, contract.Leave(caller(t, "player-1")))
	require.NoError(t, store.SaveDeparture(contract, "player-1", domain.ContractInProgress))
	snapshot, err := store.Get("c-1")
	require.NoError(t, err)
	removeContractAttribute(t, client, table, "c-1", "Departed")
	previousState := snapshot.State
	require.NoError(t, snapshot.Leave(caller(t, "player-2")))

	// Act
	err = store.SaveDeparture(snapshot, "player-2", previousState)

	// Assert
	require.Error(t, err)
	loaded, getErr := store.Get("c-1")
	require.NoError(t, getErr)
	require.False(t, loaded.IsDeparted("player-2"))
}

// REG-0053-03 edge B: departed-count equality alone is not enough — a
// candidate carrying a different departed set of the same size must conflict
// rather than overwrite the persisted departures.
func TestContractStore_SaveDepartureRejectsSameCountDifferentDepartedSet(t *testing.T) {
	// Arrange: persisted c-1 has player-1 departed; the candidate carries
	// player-2 departed instead — same count, different member — and
	// player-3 leaves on it.
	store, client, table := chg0053Fixture(t, "t13-depart-set")
	contract := startedTeamContract(t)
	require.NoError(t, store.Save(contract))
	require.NoError(t, contract.Leave(caller(t, "player-1")))
	require.NoError(t, store.SaveDeparture(contract, "player-1", domain.ContractInProgress))
	seedReservation(t, client, table, "player-3", "def-x", "c-1")
	seedActiveSlot(t, client, table, "player-3", "c-1")
	candidate := domain.NewMissionContractFromState("c-1", defXTeamOffer(),
		[]string{"player-1", "player-2", "player-3"}, domain.ContractInProgress,
		true, "team-9", "season-1", nil, []string{"player-2"}, nil)
	require.NoError(t, candidate.Leave(caller(t, "player-3")))

	// Act
	err := store.SaveDeparture(candidate, "player-3", domain.ContractInProgress)

	// Assert: the contract put at index 0 must be the canceled item — a
	// ConditionalCheckFailed, not a validation or guard-release failure.
	require.Error(t, err)
	var txCanceled *types.TransactionCanceledException
	require.ErrorAs(t, err, &txCanceled)
	require.NotEmpty(t, txCanceled.CancellationReasons)
	require.NotNil(t, txCanceled.CancellationReasons[0].Code)
	require.Equal(t, "ConditionalCheckFailed", *txCanceled.CancellationReasons[0].Code)
	loaded, getErr := store.Get("c-1")
	require.NoError(t, getErr)
	require.True(t, loaded.IsDeparted("player-1"))
	require.False(t, loaded.IsDeparted("player-2"))
	require.False(t, loaded.IsDeparted("player-3"))
	require.Equal(t, "c-1", reservationOwner(t, client, table, "player-3", "def-x"))
}

// Guard: a genuinely legacy row that never carried Departed must still accept
// a first departure — the missing-attribute branch is valid only when the
// leaving snapshot expects zero prior departures.
func TestContractStore_SaveDepartureAcceptsLegacyMissingDeparted(t *testing.T) {
	// Arrange
	store, client, table := chg0053Fixture(t, "t13-depart-legacy")
	contract := startedTeamContract(t)
	require.NoError(t, store.Save(contract))
	removeContractAttribute(t, client, table, "c-1", "Departed")
	seedReservation(t, client, table, "player-1", "def-x", "c-1")
	snapshot, err := store.Get("c-1")
	require.NoError(t, err)
	previousState := snapshot.State
	require.NoError(t, snapshot.Leave(caller(t, "player-1")))

	// Act
	err = store.SaveDeparture(snapshot, "player-1", previousState)

	// Assert
	require.NoError(t, err)
	loaded, getErr := store.Get("c-1")
	require.NoError(t, getErr)
	require.True(t, loaded.IsDeparted("player-1"))
	require.Equal(t, "", reservationOwner(t, client, table, "player-1", "def-x"))
}
