package dynamodb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// pendingStoreFixture seeds an IN_PROGRESS contract with one confirmed unit
// submitted for review, returning the persisted pending contract.
func pendingStoreFixture(t *testing.T, store *missiondynamo.ContractStore) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.CompletionReward = 200
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	_, err := contract.Confirm("objective-1", "unit-1", domain.NewSeason("season-1", domain.SeasonStateActive))
	require.NoError(t, err)
	err = contract.Submit(member)
	require.NoError(t, err)
	err = store.Save(contract)
	require.NoError(t, err)
	return contract
}

// Save + Get must round-trip the staged credits inside the contract item so a
// reviewer reloads the full adjudication context.
func TestContractStore_SaveRestoresPendingCredits(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	store := missiondynamo.NewContractStore(client, "hbk-missions")

	pendingStoreFixture(t, store)

	loaded, err := store.Get("contract-1")
	require.NoError(t, err)
	require.Equal(t, domain.ContractPendingValidation, loaded.State)
	credits := loaded.PendingCredits()
	require.Len(t, credits, 1)
	require.False(t, credits[0].Amount != 100 || credits[0].ObjectiveID != "objective-1" || credits[0].RecipientID != "player-1")
}

// SaveApproved commits the adjudicated contract and every materialized credit
// in one transaction, idempotently under retry.
func TestContractStore_SaveApprovedCommitsContractAndCredits(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	store := missiondynamo.NewContractStore(client, "hbk-missions")
	ledgerStore := missiondynamo.NewMeritLedgerStore(client, "hbk-missions")

	contract := pendingStoreFixture(t, store)
	credits, err := contract.Approve()
	require.NoError(t, err)
	err = store.SaveApproved(contract, credits)
	require.NoError(t, err)

	loaded, err := store.Get("contract-1")
	require.NoError(t, err)
	require.Equal(t, domain.ContractCompleted, loaded.State)
	total := ledgerStore.TotalFor("season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1")
	require.False(t, total != 300)
}

// A verdict rejecting one staged unit must persist the rejection record on the
// contract and never write that unit's credit to the ledger.
func TestContractStore_SaveApprovedWithVerdictDropsRejectedUnit(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	store := missiondynamo.NewContractStore(client, "hbk-missions")
	ledgerStore := missiondynamo.NewMeritLedgerStore(client, "hbk-missions")

	contract := pendingStoreFixture(t, store)
	credits, err := contract.ApproveWithVerdict([]domain.UnitRef{
		{ObjectiveID: "objective-1", UnitID: "unit-1"},
	})
	require.NoError(t, err)
	err = store.SaveApproved(contract, credits)
	require.NoError(t, err)

	loaded, err := store.Get("contract-1")
	require.NoError(t, err)
	require.Equal(t, domain.ContractCompleted, loaded.State)
	require.Len(t, loaded.RejectedUnits(), 1)
	require.Empty(t, loaded.PendingCredits())
	total := ledgerStore.TotalFor("season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1")
	require.Equal(t, 0, total)
}

// A retried approval under a stale read must hit the conditional credit put
// and surface the duplicate signal — the ledger stays exactly-once.
func TestContractStore_SaveApprovedRejectsDuplicateRetry(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	store := missiondynamo.NewContractStore(client, "hbk-missions")

	contract := pendingStoreFixture(t, store)
	credits, err := contract.Approve()
	require.NoError(t, err)
	err = store.SaveApproved(contract, credits)
	require.NoError(t, err)
	err = store.SaveApproved(contract, credits)
	require.ErrorIs(t, err, domain.ErrDuplicateConfirmation)
}
