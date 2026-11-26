package dynamodb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func frozenTeamContract(t *testing.T, id, teamID, seasonID string, roster []string) *domain.MissionContract {
	t.Helper()
	definition, err := domain.NewMissionDefinition("bounty-template", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	require.NoError(t, err)
	definition = englishize(definition)
	require.NoError(t, definition.Publish("v1"))
	offer, err := domain.NewMissionOfferFromDefinition(id+"-offer", seasonID, definition)
	require.NoError(t, err)
	return domain.NewMissionContractFromState(id, offer, roster, domain.ContractInProgress, true, teamID, seasonID, nil, nil, nil)
}

func TestContractStore_MembersReturnsFrozenSeasonRoster(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewContractStore(client, "hbk-missions-outbox")
	require.NoError(t, store.Save(frozenTeamContract(t, "c-1", "team-alpha", "season-finalize", []string{"user-2", "user-1"})))
	require.NoError(t, store.Save(frozenTeamContract(t, "c-2", "team-alpha", "season-finalize", []string{"user-3"})))
	require.NoError(t, store.Save(frozenTeamContract(t, "c-3", "team-alpha", "season-old", []string{"user-9"})))

	members, err := store.Members("season-finalize", "team-alpha")

	require.NoError(t, err)
	require.Equal(t, []string{"user-1", "user-2", "user-3"}, members)
}

func TestContractStore_MembersUnknownTeamOrSeasonIsEmpty(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, "hbk-missions-outbox"))
	clearMissionsTable(t, ctx, client, "hbk-missions-outbox")

	store := missiondynamo.NewContractStore(client, "hbk-missions-outbox")
	require.NoError(t, store.Save(frozenTeamContract(t, "c-1", "team-alpha", "season-finalize", []string{"user-1"})))

	members, err := store.Members("season-finalize", "team-ghost")
	require.NoError(t, err)
	require.Empty(t, members)

	members, err = store.Members("season-old", "team-alpha")
	require.NoError(t, err)
	require.Empty(t, members)
}
