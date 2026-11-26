package dynamodb_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// DefinitionStore.List returns every stored entry for a code, including
// published version snapshots — the query feeds the definitions listing.
func TestDefinitionStore_ListReturnsAllVersions(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	store := missiondynamo.NewDefinitionStore(client, "hbk-missions")

	definition, _ := domain.NewMissionDefinition(
		"bounty-template", "Caza de recompensas", "Elimina al objetivo",
		"BOUNTY_HUNTING", domain.MissionModeSolo, 200,
		[]domain.Objective{domain.NewObjectiveNamed("obj-kills", "Kills", "", "BOUNTY_HUNTING", 10, 5, 100)}, nil,
	)
	definition = englishize(definition)
	err = definition.Publish("v1")
	require.NoError(t, err)
	err = store.Save(definition)
	require.NoError(t, err)

	definitions, err := store.List()
	require.NoError(t, err)
	require.False(t, len(definitions) < 2)
}
