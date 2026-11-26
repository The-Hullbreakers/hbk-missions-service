package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-INF-REL-008 (REG-0053-01): the last member's departure also commits
// through the focused SaveDeparture write — the terminal contract persists
// and the leaver's reservation and PLAYER_ACTIVE entry release atomically.
func TestContractStore_LastMemberSaveDepartureReleasesGuards(t *testing.T) {
	// Arrange: persisted solo IN_PROGRESS contract with seeded guards.
	store, client, table := chg0053Fixture(t, "t15-last-member")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	require.NoError(t, contract.Start(caller(t, "player-1")))
	require.NoError(t, store.Save(contract))
	seedReservation(t, client, table, "player-1", "def-x", "contract-1")
	seedActiveSlot(t, client, table, "player-1", "contract-1")
	previousState := contract.State
	require.NoError(t, contract.Leave(caller(t, "player-1")))
	require.Equal(t, domain.ContractAbandoned, contract.State)

	// Act
	err := store.SaveDeparture(contract, "player-1", previousState)

	// Assert
	require.NoError(t, err)
	stored, err := store.Get("contract-1")
	require.NoError(t, err)
	require.Equal(t, domain.ContractAbandoned, stored.State)
	require.True(t, stored.IsDeparted("player-1"))
	require.Equal(t, "", reservationOwner(t, client, table, "player-1", "def-x"))
	require.Equal(t, 0, activeSetSize(t, client, table, "player-1"))
}
