package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// pendingTeamFixture seeds a persisted PENDING_VALIDATION team contract with
// one staged credit and frozen terms.
func pendingTeamFixture(t *testing.T, store *missiondynamo.ContractStore) {
	t.Helper()
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewTeamMissionContract("contract-1", offer, []string{"player-1", "player-2"}, "team-9")
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	require.NoError(t, err)
	contract.Terms = terms
	require.NoError(t, contract.Start(caller(t, "player-1")))
	media, err := domain.NewMediaURI("evidence/contract-1/unit-1", "contract-1")
	require.NoError(t, err)
	_, err = contract.ConfirmWithMedia("objective-1", "unit-1", media, domain.NewSeason("season-1", domain.SeasonStateActive))
	require.NoError(t, err)
	require.NoError(t, contract.Submit(caller(t, "player-1")))
	require.NoError(t, store.Save(contract))
}

// TEST-INF-REL-007 (REG-0053-05): an approval computed on a snapshot taken
// before a member's committed leave must conflict atomically — the stored
// contract keeps the departure and no credit items materialize.
func TestContractStore_StaleSaveApprovedDoesNotRestoreOrCredit(t *testing.T) {
	// Arrange
	store, client, table := chg0053Fixture(t, "t15-stale-approval")
	require.NoError(t, missiondynamo.NewSeasonStore(client, table).Create(domain.NewSeason("season-1", domain.SeasonStateActive)))
	pendingTeamFixture(t, store)
	stale, err := store.Get("contract-1")
	require.NoError(t, err)
	live, err := store.Get("contract-1")
	require.NoError(t, err)
	previousState := live.State
	require.NoError(t, live.Leave(caller(t, "player-1")))
	require.NoError(t, store.SaveDeparture(live, "player-1", previousState))
	credits, err := stale.Approve()
	require.NoError(t, err)

	// Act
	err = store.SaveApproved(stale, credits)

	// Assert
	require.Error(t, err)
	require.NotErrorIs(t, err, domain.ErrDuplicateConfirmation)
	stored, err := store.Get("contract-1")
	require.NoError(t, err)
	require.Equal(t, domain.ContractPendingValidation, stored.State)
	require.True(t, stored.IsDeparted("player-1"))
	ledger := missiondynamo.NewMeritLedgerStore(client, table)
	require.Equal(t, 0, ledger.TotalFor("season-1", "BOUNTY_HUNTING", domain.TrackTeam, "team-9"))
}
