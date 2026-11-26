package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func leaveFixture(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewTeamMissionContract("c-1", offer, []string{"player-1", "player-2"}, "team-9")
	err := contract.Start(callerID(t, "player-1"))
	require.NoError(t, err)
	return contract
}

func callerID(t *testing.T, player string) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity(player, nil, nil)
	require.NoError(t, err)
	return identity
}

func TestLeaveContract_PersistsDepartureAndContractSurvives(t *testing.T) {
	store := &lifecycleStore{contract: leaveFixture(t)}
	handler := application.NewLeaveContractHandler(store, store)

	err := handler.Handle(callerID(t, "player-1"), "c-1")

	require.NoError(t, err)
	require.NotNil(t, store.saved)
	require.Equal(t, domain.ContractInProgress, store.saved.State)
	require.False(t, !store.saved.IsDeparted("player-1"))
}

func TestLeaveContract_OutsideRosterRejected(t *testing.T) {
	store := &lifecycleStore{contract: leaveFixture(t)}
	handler := application.NewLeaveContractHandler(store, store)

	err := handler.Handle(callerID(t, "outsider-9"), "c-1")

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Nil(t, store.saved)
}

// departureStore records both write paths so the leave use case must route
// every departure — including a last-member terminal exit — through
// SaveDeparture (REG-0053-01), never the generic Save.
type departureStore struct {
	contract       *domain.MissionContract
	saved          *domain.MissionContract
	saveCalls      int
	departureCalls int
	departedPlayer string
	prevState      domain.ContractState
}

func (s *departureStore) Get(string) (*domain.MissionContract, error) {
	return s.contract, nil
}

func (s *departureStore) Save(c *domain.MissionContract) error {
	s.saveCalls++
	s.saved = c
	return nil
}

func (s *departureStore) SaveDeparture(c *domain.MissionContract, departingPlayerID string, previousState domain.ContractState) error {
	s.departureCalls++
	s.saved = c
	s.departedPlayer = departingPlayerID
	s.prevState = previousState
	return nil
}

// REG-0053-01: a non-terminal leave must commit through SaveDeparture so the
// departing member's guards are released atomically with the write.
func TestLeaveContract_ReleasesThroughDepartureWriter(t *testing.T) {
	// Arrange: in-progress team contract; player-1 leaves while player-2 stays.
	store := &departureStore{contract: leaveFixture(t)}
	handler := application.NewLeaveContractHandler(store, store)

	// Act
	err := handler.Handle(callerID(t, "player-1"), "c-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, store.departureCalls)
	require.Equal(t, "player-1", store.departedPlayer)
	require.Equal(t, 0, store.saveCalls)
}

// TEST-APP-LVE-002 (REG-0053-01): last-member leave also commits via
// SaveDeparture; supersedes LastMemberStillSaves (tasks.md T-3 note).
func TestLeaveContract_LastMemberUsesDepartureWriter(t *testing.T) {
	// Arrange: solo contract already in progress; leaving abandons it.
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	if err := contract.Start(callerID(t, "player-1")); err != nil {
		t.Fatalf("arrange started contract: %v", err)
	}
	store := &departureStore{contract: contract}
	handler := application.NewLeaveContractHandler(store, store)

	// Act
	err := handler.Handle(callerID(t, "player-1"), "c-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, store.departureCalls)
	require.Equal(t, "player-1", store.departedPlayer)
	require.Equal(t, domain.ContractInProgress, store.prevState)
	require.Equal(t, 0, store.saveCalls)
	require.Equal(t, domain.ContractAbandoned, store.saved.State)
}

func TestLeaveContract_TerminalRejected(t *testing.T) {
	contract := leaveFixture(t)
	_ = contract.Abandon(callerID(t, "player-1"))
	store := &lifecycleStore{contract: contract}
	handler := application.NewLeaveContractHandler(store, store)

	err := handler.Handle(callerID(t, "player-1"), "c-1")

	require.ErrorIs(t, err, domain.ErrContractTerminal)
}
