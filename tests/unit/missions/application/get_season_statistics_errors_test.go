package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestGetSeasonStatistics_CurrentWithoutActiveSeason(t *testing.T) {
	// Arrange
	handler := statsHandler(domain.Season{}, nil, nil)

	// Act
	_, err := handler.Handle("current", "", "", 3)

	// Assert
	require.ErrorIs(t, err, domain.ErrNoActiveSeason)
}

func TestGetSeasonStatistics_CurrentMultipleActiveSignalsInvariantBreach(t *testing.T) {
	// Arrange
	two := []domain.Season{
		domain.NewSeason("season-1", domain.SeasonStateActive),
		domain.NewSeason("season-2", domain.SeasonStateActive),
	}
	handler := statsHandler(domain.Season{}, two, nil)

	// Act
	_, err := handler.Handle("current", "", "", 3)

	// Assert
	require.ErrorIs(t, err, domain.ErrActiveSeasonAlreadyExists)
}

func TestGetSeasonStatistics_EmptySeasonReturnsEmptyBuckets(t *testing.T) {
	// Arrange
	handler := application.NewGetSeasonStatisticsHandler(
		&fakeLeaderboardSeasonReader{season: domain.NewSeason("season-2", domain.SeasonStateFinalized)},
		&activeSeasonsStub{},
		&bucketReaderStub{},
		&bucketStandingsStub{})

	// Act
	stats, err := handler.Handle("season-2", "", "", 3)

	// Assert
	require.NoError(t, err)
	require.Empty(t, stats.Buckets)
}

func TestGetSeasonStatistics_UnknownSeasonPropagates(t *testing.T) {
	// Arrange
	handler := application.NewGetSeasonStatisticsHandler(
		&fakeLeaderboardSeasonReader{err: domain.ErrSeasonNotFound},
		&activeSeasonsStub{},
		&bucketReaderStub{},
		&bucketStandingsStub{})

	// Act
	_, err := handler.Handle("season-x", "", "", 3)

	// Assert
	require.ErrorIs(t, err, domain.ErrSeasonNotFound)
}

func TestGetSeasonStatistics_InvalidTrackRejected(t *testing.T) {
	// Arrange
	handler := statsHandler(
		domain.NewSeason("season-1", domain.SeasonStateActive),
		nil, nil)

	// Act
	_, err := handler.Handle("season-1", "", "bogus", 3)

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidTrack)
}
