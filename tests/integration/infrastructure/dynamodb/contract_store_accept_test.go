package dynamodb_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"strconv"
	"sync"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func acceptanceFixture(t *testing.T) (*missiondynamo.ContractStore, *missiondynamo.DynamoDBClient) {
	t.Helper()
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions")
	seedActiveSeason(t, client, "hbk-missions", "season-1")
	return missiondynamo.NewContractStore(client, "hbk-missions"), client
}

func commitAttempt(offer domain.MissionOffer, playerID string, attempt int) *domain.MissionContract {
	contract := domain.NewMissionContract(offer.OfferID+":"+playerID+":"+itoa(attempt), offer, []string{playerID})
	return contract
}

func itoa(n int) string { return strconv.Itoa(n) }

// SCN-MB-ATOMIC-LIMIT: two concurrent acceptances when the player is one below
// the effective limit — exactly one wins, the other gets
// ErrActiveMissionLimitReached.
func TestCommitAcceptance_ConcurrentLimitRace(t *testing.T) {
	store, client := acceptanceFixture(t)

	offerStore := missiondynamo.NewOfferStore(client, "hbk-missions")
	for _, id := range []string{"offer-a", "offer-b"} {
		offer := domain.NewMissionOffer(id, "BOUNTY_HUNTING", "season-1")
		err := offerStore.Save(offer)
		require.NoError(t, err)
	}

	// Player already holds one active contract; effective limit is 2.
	existing := domain.NewMissionOffer("offer-old", "TRADE", "season-1")
	old := commitAttempt(existing, "player-1", 1)
	old.State = domain.ContractInProgress
	err := store.Save(old)
	require.NoError(t, err)
	err = store.CommitAcceptance(commitAttempt(existing, "player-1", 1), existing, 2)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, id := range []string{"offer-a", "offer-b"} {
		offer := domain.NewMissionOffer(id, "BOUNTY_HUNTING", "season-1")
		wg.Add(1)
		go func(o domain.MissionOffer, n int) {
			defer wg.Done()
			errs[n] = store.CommitAcceptance(commitAttempt(o, "player-1", 2), o, 2)
		}(offer, i)
	}
	wg.Wait()

	succeeded, limitRejected := 0, 0
	for _, err := range errs {
		if err == nil {
			succeeded++
		} else if errors.Is(err, domain.ErrActiveMissionLimitReached) {
			limitRejected++
		}
	}
	require.False(t, succeeded != 1 || limitRejected != 1)
}
