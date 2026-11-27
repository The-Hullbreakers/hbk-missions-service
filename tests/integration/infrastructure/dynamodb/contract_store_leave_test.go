package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func departedContractFixture(t *testing.T, store *missiondynamo.ContractStore) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-t1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewTeamMissionContract("contract-t1", offer, []string{"player-1", "player-2"}, "team-9")
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, contract.Start(member))
	require.NoError(t, store.Save(contract))
	require.NoError(t, contract.Leave(member))
	require.NoError(t, store.SaveDeparture(contract, "player-1", domain.ContractInProgress))
	return contract
}

func TestContractStore_PersistsDeparturesAndFreesLeaverSlot(t *testing.T) {
	store, _, _ := chg0053Fixture(t, "persists-departures")

	departedContractFixture(t, store)

	loaded, err := store.Get("contract-t1")
	require.NoError(t, err)
	require.False(t, !loaded.IsDeparted("player-1"))
	require.False(t, loaded.CanSubmitEvidence(caller(t, "player-1")))

	leaverActive, err := store.ListActiveByPlayer("player-1")
	require.NoError(t, err)
	require.Len(t, leaverActive, 0)

	survivorActive, err := store.ListActiveByPlayer("player-2")
	require.NoError(t, err)
	require.Len(t, survivorActive, 1)
}

func caller(t *testing.T, id string) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity(id, nil, nil)
	require.NoError(t, err)
	return identity
}

// stubEvidenceReader reports every well-formed reference as uploaded —
// infrastructure tests exercise the storage adapters, not the blob port.
type stubEvidenceReader struct{}

func (stubEvidenceReader) Exists(domain.MediaURI) (bool, error) { return true, nil }

func (stubEvidenceReader) PresignRead(uri domain.MediaURI) (application.ReadGrant, error) {
	return application.ReadGrant{URL: "https://signed.example/read?key=" + string(uri)}, nil
}

// REG-0053-01: SaveDeparture releases the departing member's reservation and
// PLAYER_ACTIVE entry atomically with the non-terminal contract write; the
// survivor's guards stay.
func TestContractStore_SaveDepartureReleasesDepartedGuards(t *testing.T) {
	// Arrange: player-1's guards exist for def-x under c-1; they then leave
	// while player-2 stays.
	store, client, table := chg0053Fixture(t, "t8-departure")
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewTeamMissionContract("c-1", offer, []string{"player-1", "player-2"}, "team-9")
	require.NoError(t, contract.Start(caller(t, "player-1")))
	seedReservation(t, client, table, "player-1", "def-x", "c-1")
	seedReservation(t, client, table, "player-2", "def-x", "c-1")
	seedActiveSlot(t, client, table, "player-1", "c-1")
	require.NoError(t, store.Save(contract))
	previousState := contract.State
	require.NoError(t, contract.Leave(caller(t, "player-1")))

	// Act
	err := store.SaveDeparture(contract, "player-1", previousState)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "", reservationOwner(t, client, table, "player-1", "def-x"))
	require.Equal(t, "c-1", reservationOwner(t, client, table, "player-2", "def-x"))
	require.Equal(t, 0, activeSetSize(t, client, table, "player-1"))
	loaded, err := store.Get("c-1")
	require.NoError(t, err)
	require.True(t, loaded.IsDeparted("player-1"))
}

// REG-0053-03 candidate: two members leaving on separate snapshots of the same
// stored contract must not resurrect the first leaver. A stale SaveDeparture
// has to fail closed; the unconditional contract put must not overwrite the
// newer departure state.
func TestContractStore_ConcurrentDeparturesDoNotRestoreFormerMember(t *testing.T) {
	// Arrange: c-1 has a three-member roster with all guards seeded; two
	// snapshots are fetched before either departure is persisted.
	store, client, table := chg0053Fixture(t, "t13-departure-race")
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewTeamMissionContract("c-1", offer, []string{"player-1", "player-2", "player-3"}, "team-9")
	require.NoError(t, contract.Start(caller(t, "player-1")))
	for _, player := range []string{"player-1", "player-2", "player-3"} {
		seedReservation(t, client, table, player, "def-x", "c-1")
		seedActiveSlot(t, client, table, player, "c-1")
	}
	require.NoError(t, store.Save(contract))
	first, err := store.Get("c-1")
	require.NoError(t, err)
	second, err := store.Get("c-1")
	require.NoError(t, err)
	firstCaller, secondCaller := caller(t, "player-1"), caller(t, "player-2")

	// Act: each snapshot carries only its own departure; the second write is
	// stale by the time it runs.
	firstLeaveErr, secondLeaveErr, firstSaveErr, staleSaveErr := attemptStaleDeparture(store, first, second, firstCaller, secondCaller)

	// Assert: the stale write is rejected and the newer departure survives.
	require.NoError(t, firstLeaveErr)
	require.NoError(t, secondLeaveErr)
	require.NoError(t, firstSaveErr)
	require.Error(t, staleSaveErr)
	loaded, getErr := store.Get("c-1")
	require.NoError(t, getErr)
	require.True(t, loaded.IsDeparted("player-1"))
	require.False(t, loaded.IsDeparted("player-2"))
	require.Equal(t, "", reservationOwner(t, client, table, "player-1", "def-x"))
	require.Equal(t, "c-1", reservationOwner(t, client, table, "player-2", "def-x"))
	require.Equal(t, "c-1", reservationOwner(t, client, table, "player-3", "def-x"))
}

// attemptStaleDeparture drives the two-snapshot departure race: each snapshot
// leaves for its own member, then both write in order. It returns every step's
// error so the caller can assert the ordering that was exercised.
func attemptStaleDeparture(store *missiondynamo.ContractStore, first, second *domain.MissionContract, firstCaller, secondCaller domain.CallerIdentity) (firstLeaveErr, secondLeaveErr, firstSaveErr, staleSaveErr error) {
	firstPrev, secondPrev := first.State, second.State
	firstLeaveErr = first.Leave(firstCaller)
	secondLeaveErr = second.Leave(secondCaller)
	firstSaveErr = store.SaveDeparture(first, firstCaller.PlayerID(), firstPrev)
	if firstSaveErr != nil {
		return
	}
	staleSaveErr = store.SaveDeparture(second, secondCaller.PlayerID(), secondPrev)
	return
}
