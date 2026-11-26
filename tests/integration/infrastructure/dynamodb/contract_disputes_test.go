package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// Disputes persist on the contract item and round-trip with state intact.
func TestContractStore_DisputesRoundTrip(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	// Arrange — contract carrying one open + one resolved dispute
	store := missiondynamo.NewContractStore(client, "hbk-missions")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-d1", offer, []string{"player-1"})
	objA := domain.NewObjectiveNamed("obj-a", "a", "", "BOUNTY", 10, 0, 0)
	objB := domain.NewObjectiveNamed("obj-b", "b", "", "BOUNTY", 20, 0, 0)
	terms, _ := domain.NewContractTerms([]domain.Objective{objA, objB})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)

	open, err := contract.RaiseDispute(member, "obj-a", "still contested")
	require.NoError(t, err)
	resolved, err := contract.RaiseDispute(member, "obj-b", "was contested")
	require.NoError(t, err)
	err = contract.ResolveDispute(resolved.DisputeID, "mod-1")
	require.NoError(t, err)

	// Act
	err = store.Save(contract)
	require.NoError(t, err)
	loaded, err := store.Get("contract-d1")

	// Assert
	require.NoError(t, err)
	disputes := loaded.Disputes()
	require.Len(t, disputes, 2)
	states := map[string]domain.DisputeState{}
	for _, d := range disputes {
		states[d.ObjectiveID] = d.State
	}
	require.False(t, states["obj-a"] != domain.DisputeOpen || states["obj-b"] != domain.DisputeResolved)
	for _, d := range disputes {
		require.False(t, d.DisputeID == open.DisputeID && d.Reason != "still contested")
	}
}
