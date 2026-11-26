package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func standingsMother() []domain.Standing {
	return []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Total: 200},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-3", Total: 100},
	}
}

func callerMother(t *testing.T, playerID string, teamIDs []string) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity(playerID, nil, teamIDs)
	require.NoError(t, err)
	return identity
}

func TestGetMyStanding_ReturnsCallerRank(t *testing.T) {
	// Arrange
	handler := application.NewGetMyStandingHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&fakeLeaderboardReader{standings: standingsMother()})

	// Act
	entry, err := handler.Handle(callerMother(t, "player-1", nil), "season-1", "BOUNTY_HUNTING", "individual")

	// Assert
	require.NoError(t, err)
	require.False(t, entry.Rank != 2 || entry.RecipientID != "player-1" || entry.Total != 150)
}

func TestGetMyStanding_AbsentReturnsErrStandingNotFound(t *testing.T) {
	// Arrange
	handler := application.NewGetMyStandingHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&fakeLeaderboardReader{standings: standingsMother()})

	// Act
	_, err := handler.Handle(callerMother(t, "player-9", nil), "season-1", "BOUNTY_HUNTING", "individual")

	// Assert
	require.ErrorIs(t, err, domain.ErrStandingNotFound)
}

func TestGetMyStanding_TeamTrackResolvesTeamMembership(t *testing.T) {
	// Arrange
	standings := []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam, RecipientID: "team-2", Total: 500},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackTeam, RecipientID: "team-1", Total: 300},
	}
	handler := application.NewGetMyStandingHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&fakeLeaderboardReader{standings: standings})

	// Act
	entry, err := handler.Handle(callerMother(t, "player-1", []string{"team-1"}), "season-1", "BOUNTY_HUNTING", "team")

	// Assert
	require.NoError(t, err)
	require.False(t, entry.Rank != 2 || entry.RecipientID != "team-1")
}

func TestGetMyStanding_UnknownSeasonPropagates(t *testing.T) {
	// Arrange
	handler := application.NewGetMyStandingHandler(
		&fakeLeaderboardSeasonReader{err: domain.ErrSeasonNotFound},
		&fakeLeaderboardReader{})

	// Act
	_, err := handler.Handle(callerMother(t, "player-1", nil), "season-x", "BOUNTY_HUNTING", "individual")

	// Assert
	require.ErrorIs(t, err, domain.ErrSeasonNotFound)
}

func TestGetMyStanding_InvalidTrackRejected(t *testing.T) {
	// Arrange
	handler := application.NewGetMyStandingHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&fakeLeaderboardReader{})

	// Act
	_, err := handler.Handle(callerMother(t, "player-1", nil), "season-1", "BOUNTY_HUNTING", "bogus")

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidTrack)
}
