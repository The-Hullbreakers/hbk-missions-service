package dynamodb_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// The review queue surfaces every PENDING_VALIDATION contract and nothing
// else — a wrong-state leak would let an adjudicator act on a contract the
// member did not submit.
func TestContractStore_ListPendingReviewFiltersState(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	store := missiondynamo.NewContractStore(client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")

	pendingStoreFixture(t, store)
	seedInProgressContract(t, store, "contract-2")
	seedInProgressContract(t, store, "contract-3")

	got, err := store.ListPendingReview()
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "contract-1", got[0].ContractID)
}

// seedInProgressContract persists a started but unsubmitted contract.
func seedInProgressContract(t *testing.T, store *missiondynamo.ContractStore, id string) {
	t.Helper()
	offer := domain.NewMissionOffer("offer-"+id, "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract(id, offer, []string{"player-9"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-9", nil, nil)
	_ = contract.Start(member)
	err := store.Save(contract)
	require.NoError(t, err)
}
