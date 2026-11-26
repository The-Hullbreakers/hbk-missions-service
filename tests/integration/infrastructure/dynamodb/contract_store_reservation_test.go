package dynamodb_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// raceCommits runs one CommitAcceptance per (player, offer) pair in
// parallel and returns each attempt's error — no assertions inside.
func raceCommits(store *missiondynamo.ContractStore, offers []domain.MissionOffer, playerIDs []string, limit int) []error {
	var wg sync.WaitGroup
	errs := make([]error, len(offers))
	for i := range offers {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			errs[n] = store.CommitAcceptance(commitAttempt(offers[n], playerIDs[n], n+1), offers[n], limit)
		}(i)
	}
	wg.Wait()
	return errs
}

// REG-0053-02: one player racing two published versions of the same
// definition_code must serialize — one commit, one
// ErrMissionDefinitionAlreadyActive, and no partial contract or
// PLAYER_ACTIVE write from the loser.
func TestCommitAcceptance_ConcurrentSameDefinitionSerializes(t *testing.T) {
	// Arrange: two versions of def-x, player-1 well under the active limit.
	store, client, table := chg0053Fixture(t, "t7-serializes")
	seedActiveSeason(t, client, table, "season-1")
	offerV1 := domain.NewMissionOffer("offer-x-v1", "BOUNTY_HUNTING", "season-1")
	offerV1.DefinitionCode = "def-x"
	offerV1.DefinitionVersion = "1"
	offerV2 := domain.NewMissionOffer("offer-x-v2", "BOUNTY_HUNTING", "season-1")
	offerV2.DefinitionCode = "def-x"
	offerV2.DefinitionVersion = "2"

	// Act
	errs := raceCommits(store, []domain.MissionOffer{offerV1, offerV2}, []string{"player-1", "player-1"}, 5)

	// Assert
	succeeded, conflicts := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, domain.ErrMissionDefinitionAlreadyActive):
			conflicts++
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, conflicts)
	active, err := store.ListActiveByPlayer("player-1")
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Equal(t, 1, activeSetSize(t, client, table, "player-1"))
}

// REG-0053-02 guard: per-player reservations are independent — different
// players may hold the same definition_code concurrently.
func TestCommitAcceptance_DifferentPlayersSameDefinitionBothCommit(t *testing.T) {
	// Arrange: the two def-x versions accepted by two unrelated players.
	store, client, table := chg0053Fixture(t, "t7-different-players")
	seedActiveSeason(t, client, table, "season-1")
	offerV1 := domain.NewMissionOffer("offer-x-v1", "BOUNTY_HUNTING", "season-1")
	offerV1.DefinitionCode = "def-x"
	offerV1.DefinitionVersion = "1"
	offerV2 := domain.NewMissionOffer("offer-x-v2", "BOUNTY_HUNTING", "season-1")
	offerV2.DefinitionCode = "def-x"
	offerV2.DefinitionVersion = "2"

	// Act
	errs := raceCommits(store, []domain.MissionOffer{offerV1, offerV2}, []string{"player-1", "player-2"}, 5)

	// Assert
	require.NoError(t, errs[0])
	require.NoError(t, errs[1])
}
