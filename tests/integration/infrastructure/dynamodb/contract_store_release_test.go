package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// Terminal save must free the active-set slot idempotently.
func TestContractStore_TerminalSaveReleasesActiveSlot(t *testing.T) {
	store, client := acceptanceFixture(t)

	offerStore := missiondynamo.NewOfferStore(client, "hbk-missions")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	err := offerStore.Save(offer)
	require.NoError(t, err)

	contract := commitAttempt(offer, "player-1", 1)
	err = store.CommitAcceptance(contract, offer, 1)
	require.NoError(t, err)

	// At limit 1, a second acceptance of a different definition is rejected
	// while the first is active — the limit guard, not the definition guard.
	other := domain.NewMissionOffer("offer-2", "BOUNTY_HUNTING", "season-1")
	err = store.CommitAcceptance(commitAttempt(other, "player-1", 2), other, 1)
	require.ErrorIs(t, err, domain.ErrActiveMissionLimitReached)

	// Releasing the slot lets the next attempt commit.
	contract.State = domain.ContractAbandoned
	err = store.Save(contract)
	require.NoError(t, err)
	err = store.CommitAcceptance(commitAttempt(offer, "player-1", 2), offer, 1)
	require.NoError(t, err)
}

// REG-0053-02: a terminal save releases reservations only for still-active
// members — a departed member's reservation may belong to a newer attempt
// and must never be touched.
func TestContractStore_TerminalSaveReleasesOnlyActiveMembers(t *testing.T) {
	// Arrange: team contract c-1 on def-x; player-2 departed and their def-x
	// reservation is owned by the newer contract c-2.
	store, client, table := chg0053Fixture(t, "t8-members")
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewTeamMissionContract("c-1", offer, []string{"player-1", "player-2"}, "team-9")
	require.NoError(t, contract.Start(caller(t, "player-1")))
	require.NoError(t, store.Save(contract))
	require.NoError(t, contract.Leave(caller(t, "player-2")))
	require.NoError(t, store.SaveDeparture(contract, "player-2", domain.ContractInProgress))
	seedReservation(t, client, table, "player-1", "def-x", "c-1")
	seedReservation(t, client, table, "player-2", "def-x", "c-2")
	require.NoError(t, contract.Abandon(caller(t, "player-1")))

	// Act
	err := store.Save(contract)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "", reservationOwner(t, client, table, "player-1", "def-x"))
	require.Equal(t, "c-2", reservationOwner(t, client, table, "player-2", "def-x"))
}

// REG-0053-02: a stale terminal save retried after the member already holds
// a newer reservation owned by another contract must conflict — it may
// never delete or clean up the newer guard.
func TestContractStore_StaleReleaseKeepsNewerReservation(t *testing.T) {
	// Arrange: c-1 was already saved terminal once; player-1 retook def-x as
	// c-2, so the reservation owner is c-2 while c-1 still lists player-1.
	store, client, table := chg0053Fixture(t, "t8-stale")
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	require.NoError(t, contract.Start(caller(t, "player-1")))
	require.NoError(t, contract.Abandon(caller(t, "player-1")))
	require.NoError(t, store.Save(contract))
	seedReservation(t, client, table, "player-1", "def-x", "c-2")

	// Act: a stale retry of the same terminal save.
	err := store.Save(contract)

	// Assert
	require.Error(t, err)
	require.Equal(t, "c-2", reservationOwner(t, client, table, "player-1", "def-x"))
}

// REG-0053-02: a stale SaveApproved whose release would touch a reservation
// owned by a newer contract must fail closed — the newer guard and the
// stored contract survive, and the conflict must not masquerade as a
// duplicate-confirmation rejection.
func TestContractStore_SaveApprovedStaleReleaseIsNotDuplicate(t *testing.T) {
	// Arrange: persisted in-progress c-1 with still-active player-1; the
	// approval snapshot marks it completed but is stale — player-1's def-x
	// reservation is already owned by the newer contract c-2. The conflict
	// must come from the foreign-owner release guard, not the contract put.
	store, client, table := chg0053Fixture(t, "t8-approve-stale")
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	require.NoError(t, contract.Start(caller(t, "player-1")))
	require.NoError(t, store.Save(contract))
	contract.State = domain.ContractCompleted
	seedReservation(t, client, table, "player-1", "def-x", "c-2")

	// Act
	err := store.SaveApproved(contract, nil)

	// Assert
	require.Error(t, err)
	require.NotErrorIs(t, err, domain.ErrDuplicateConfirmation)
	require.Equal(t, "c-2", reservationOwner(t, client, table, "player-1", "def-x"))
	stored, getErr := store.Get("c-1")
	require.NoError(t, getErr)
	require.Equal(t, domain.ContractInProgress, stored.State)
}

// REG-0053-02 guard: PENDING_VALIDATION keeps the reservation — the guard
// outlives the active-limit slot and blocks retake.
func TestContractStore_PendingValidationKeepsReservation(t *testing.T) {
	// Arrange: contract c-1 submitted for review while its def-x reservation
	// is still owned by c-1.
	store, client, table := chg0053Fixture(t, "t8-pending")
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	contract.State = domain.ContractPendingValidation
	seedReservation(t, client, table, "player-1", "def-x", "c-1")

	// Act
	err := store.Save(contract)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "c-1", reservationOwner(t, client, table, "player-1", "def-x"))
}
