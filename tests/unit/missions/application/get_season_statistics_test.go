package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// bucketStandingsStub returns per-category standings for statistics tests.
type bucketStandingsStub struct {
	byCategory map[string][]domain.Standing
	err        error
}

func (s *bucketStandingsStub) ListStandings(_, category string, _ domain.Track, _ int) ([]domain.Standing, error) {
	return s.byCategory[category], s.err
}

func statsBuckets() []domain.StandingBucket {
	return []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
		{Category: "BOUNTY_HUNTING", Track: domain.TrackTeam},
		{Category: "RACING", Track: domain.TrackIndividual},
	}
}

func statsHandler(season domain.Season, active []domain.Season, standings map[string][]domain.Standing) *application.GetSeasonStatisticsHandler {
	return application.NewGetSeasonStatisticsHandler(
		&fakeLeaderboardSeasonReader{season: season},
		&activeSeasonsStub{seasons: active},
		&bucketReaderStub{buckets: statsBuckets()},
		&bucketStandingsStub{byCategory: standings})
}

func TestGetSeasonStatistics_GroupsBucketsWithTopEntries(t *testing.T) {
	// Arrange
	standings := map[string][]domain.Standing{
		"BOUNTY_HUNTING": {
			{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150, CompletedMissions: 2},
			{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Total: 100},
		},
	}
	handler := statsHandler(
		domain.NewSeason("season-1", domain.SeasonStateActive),
		nil, standings)

	// Act
	stats, err := handler.Handle("season-1", "", "", 3)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "season-1", stats.SeasonID)
	require.Equal(t, domain.SeasonStateActive, stats.State)
	require.True(t, stats.Official)
	require.Len(t, stats.Buckets, 3)

	var bounty domain.SeasonBucketStatistics
	for _, b := range stats.Buckets {
		if b.Bucket.Category == "BOUNTY_HUNTING" && b.Bucket.Track == domain.TrackIndividual {
			bounty = b
		}
	}
	require.Equal(t, 2, bounty.Participants)
	require.Equal(t, 250, bounty.TotalMerits)
	require.Len(t, bounty.Top, 2)
	require.Equal(t, "player-1", bounty.Top[0].RecipientID)
	require.Equal(t, 1, bounty.Top[0].Rank)
}

func TestGetSeasonStatistics_CurrentResolvesActiveSeason(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := statsHandler(season, []domain.Season{season}, nil)

	// Act
	stats, err := handler.Handle("current", "", "", 3)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "season-1", stats.SeasonID)
}

func TestGetSeasonStatistics_TopNSlicesEntries(t *testing.T) {
	// Arrange
	standings := map[string][]domain.Standing{
		"BOUNTY_HUNTING": {
			{SeasonID: "season-1", Category: "BOUNTY_HUNTING", RecipientID: "player-1", Total: 300},
			{SeasonID: "season-1", Category: "BOUNTY_HUNTING", RecipientID: "player-2", Total: 200},
			{SeasonID: "season-1", Category: "BOUNTY_HUNTING", RecipientID: "player-3", Total: 100},
		},
	}
	handler := statsHandler(
		domain.NewSeason("season-1", domain.SeasonStateActive),
		nil, standings)

	// Act
	stats, err := handler.Handle("season-1", "BOUNTY_HUNTING", "individual", 2)

	// Assert
	require.NoError(t, err)
	require.Len(t, stats.Buckets, 1)
	require.Equal(t, 3, stats.Buckets[0].Participants)
	require.Len(t, stats.Buckets[0].Top, 2)
	require.Equal(t, 2, stats.Buckets[0].Top[1].Rank)
}

func TestGetSeasonStatistics_FiltersByTrack(t *testing.T) {
	// Arrange
	handler := statsHandler(
		domain.NewSeason("season-1", domain.SeasonStateActive),
		nil, nil)

	// Act
	stats, err := handler.Handle("season-1", "", "team", 3)

	// Assert
	require.NoError(t, err)
	require.Len(t, stats.Buckets, 1)
	require.Equal(t, domain.TrackTeam, stats.Buckets[0].Bucket.Track)
}
