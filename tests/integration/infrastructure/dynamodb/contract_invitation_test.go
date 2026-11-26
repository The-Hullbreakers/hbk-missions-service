package dynamodb_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/stretchr/testify/require"
)

// Invitation state must round-trip: a pending invitee stays pending after
// save/load, and the freeze rule still sees it.
func TestContractStore_RoundTripPreservesInvitations(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	store := missiondynamo.NewContractStore(client, "hbk-missions")

	offer := domain.NewTeamMissionOffer("offer-inv", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewTeamMissionContract("offer-inv:player-1:1", offer,
		[]string{"player-1", "player-2"}, "team-1")

	err = store.Save(contract)
	require.NoError(t, err)
	loaded, err := store.Get("offer-inv:player-1:1")
	require.NoError(t, err)

	require.Equal(t, domain.InvitationPending, loaded.InvitationState("player-2"))
	require.Equal(t, domain.InvitationConfirmed, loaded.InvitationState("player-1"))
	require.False(t, loaded.RosterFrozen(), "roster must stay unfrozen while invitations are pending")
	require.False(t, loaded.InvitedAt().IsZero(), "invitedAt must round-trip for expiry to work")

	// Expiry must work on the reloaded contract (lazy resolution).
	err = loaded.ResolveExpirations(loaded.InvitedAt().Add(domain.InvitationWindow + time.Minute))
	require.NoError(t, err)
	require.Equal(t, domain.InvitationExpired, loaded.InvitationState("player-2"))
}
