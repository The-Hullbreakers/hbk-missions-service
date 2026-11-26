package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type bucketReaderStub struct {
	buckets []domain.StandingBucket
	err     error
	gotID   string
}

func (s *bucketReaderStub) ListBuckets(seasonID string) ([]domain.StandingBucket, error) {
	s.gotID = seasonID
	return s.buckets, s.err
}

func historyFixture() []domain.StandingBucket {
	return []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
		{Category: "BOUNTY_HUNTING", Track: domain.TrackTeam},
		{Category: "EXPLORATION", Track: domain.TrackIndividual},
	}
}

func historyStandings(category string, track domain.Track) []domain.Standing {
	return []domain.Standing{
		{SeasonID: "season-1", Category: category, Track: track, RecipientID: "player-2", Total: 200},
		{SeasonID: "season-1", Category: category, Track: track, RecipientID: "player-1", Total: 150},
	}
}

func TestListStandingsHistory_FinalizedSeasonReturnsRankedBuckets(t *testing.T) {
	// Arrange
	handler := application.NewListStandingsHistoryHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateFinalized)},
		&bucketReaderStub{buckets: historyFixture()},
		&fakeLeaderboardReader{standings: historyStandings("BOUNTY_HUNTING", domain.TrackIndividual)})

	// Act
	history, err := handler.Handle("season-1", "", "")

	// Assert
	require.NoError(t, err)
	require.Len(t, history.Buckets, 3)
}

func TestListStandingsHistory_ActiveSeasonRejected(t *testing.T) {
	// Arrange
	handler := application.NewListStandingsHistoryHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&bucketReaderStub{},
		&fakeLeaderboardReader{})

	// Act
	_, err := handler.Handle("season-1", "", "")

	// Assert
	require.ErrorIs(t, err, domain.ErrSeasonStateInvalid)
}

func TestListStandingsHistory_UnknownSeasonPropagates(t *testing.T) {
	// Arrange
	handler := application.NewListStandingsHistoryHandler(
		&fakeLeaderboardSeasonReader{err: domain.ErrSeasonNotFound},
		&bucketReaderStub{},
		&fakeLeaderboardReader{})

	// Act
	_, err := handler.Handle("season-x", "", "")

	// Assert
	require.ErrorIs(t, err, domain.ErrSeasonNotFound)
}

func TestListStandingsHistory_FiltersByCategoryAndTrack(t *testing.T) {
	// Arrange
	handler := application.NewListStandingsHistoryHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateFinalized)},
		&bucketReaderStub{buckets: historyFixture()},
		&fakeLeaderboardReader{standings: historyStandings("BOUNTY_HUNTING", domain.TrackIndividual)})

	// Act
	history, err := handler.Handle("season-1", "BOUNTY_HUNTING", "individual")

	// Assert
	require.NoError(t, err)
	require.Len(t, history.Buckets, 1)
}

func TestListStandingsHistory_InvalidTrackRejected(t *testing.T) {
	// Arrange
	handler := application.NewListStandingsHistoryHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateFinalized)},
		&bucketReaderStub{},
		&fakeLeaderboardReader{})

	// Act
	_, err := handler.Handle("season-1", "", "bogus")

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidTrack)
}
