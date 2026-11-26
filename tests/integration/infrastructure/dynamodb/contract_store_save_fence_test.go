package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// FIND-0053-02 candidate: ContractStore.Save writes a non-terminal contract
// with an unconditional put. A snapshot taken before a member's committed
// departure — e.g. an in-flight confirmation or invitation flow — can then
// restore the departed player as an active roster member.
func TestContractStore_StaleSaveDoesNotRestoreDepartedMember(t *testing.T) {
	// Arrange: c-1 persisted with no departures; the stale snapshot is
	// fetched before player-1's leave commits through SaveDeparture.
	store, client, table := chg0053Fixture(t, "t13-stale-save")
	contract := startedTeamContract(t)
	require.NoError(t, store.Save(contract))
	stale, err := store.Get("c-1")
	require.NoError(t, err)
	live, err := store.Get("c-1")
	require.NoError(t, err)
	seedReservation(t, client, table, "player-1", "def-x", "c-1")
	seedActiveSlot(t, client, table, "player-1", "c-1")
	previousState := live.State
	require.NoError(t, live.Leave(caller(t, "player-1")))
	require.NoError(t, store.SaveDeparture(live, "player-1", previousState))

	// Act
	err = store.Save(stale)

	// Assert: the stale write must not resurrect the departed member or
	// recreate the released reservation.
	require.Error(t, err)
	loaded, getErr := store.Get("c-1")
	require.NoError(t, getErr)
	require.True(t, loaded.IsDeparted("player-1"))
	require.Equal(t, "", reservationOwner(t, client, table, "player-1", "def-x"))
}
