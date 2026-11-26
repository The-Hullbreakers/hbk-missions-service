package dynamodb_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// Phased terms and phase progression must survive a save/load round-trip.
func TestContractStore_PhasedTermsRoundTrip(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	// Arrange
	store := missiondynamo.NewContractStore(client, "hbk-missions")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-p1", offer, []string{"player-1"})

	objA := domain.NewObjectiveNamed("obj-a", "mandatory a", "", "BOUNTY", 10, 0, 0)
	objX := domain.NewObjectiveNamed("obj-x", "optional x", "", "BOUNTY", 5, 0, 0)
	objB := domain.NewObjectiveNamed("obj-b", "mandatory b", "", "BOUNTY", 20, 0, 0)
	terms, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{objA}, Optional: []domain.Objective{objX}},
		{Mandatory: []domain.Objective{objB}},
	})
	require.NoError(t, err)
	contract.Terms = terms
	contract.RestoreCurrentPhase(1)

	// Act
	err = store.Save(contract)
	require.NoError(t, err)
	loaded, err := store.Get("contract-p1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, loaded.CurrentPhase())
	phases := loaded.Terms.Phases()
	require.Len(t, phases, 2)
	require.False(t, len(phases[0].Mandatory) != 1 || phases[0].Mandatory[0].ObjectiveID != "obj-a")
	require.False(t, len(phases[0].Optional) != 1 || phases[0].Optional[0].ObjectiveID != "obj-x")
}

// A flat contract restores as a single implicit all-mandatory phase.
func TestContractStore_FlatTermsImplicitPhase(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")

	// Arrange
	store := missiondynamo.NewContractStore(client, "hbk-missions")
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-f1", offer, []string{"player-1"})
	objA := domain.NewObjectiveNamed("obj-a", "flat a", "", "BOUNTY", 10, 0, 0)
	objB := domain.NewObjectiveNamed("obj-b", "flat b", "", "BOUNTY", 20, 0, 0)
	terms, err := domain.NewContractTerms([]domain.Objective{objA, objB})
	require.NoError(t, err)
	contract.Terms = terms

	// Act
	err = store.Save(contract)
	require.NoError(t, err)
	loaded, err := store.Get("contract-f1")

	// Assert
	require.NoError(t, err)
	phases := loaded.Terms.Phases()
	require.False(t, len(phases) != 1 || len(phases[0].Mandatory) != 2 || len(phases[0].Optional) != 0)
}
