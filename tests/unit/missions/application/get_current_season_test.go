package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestGetCurrentSeason_ReturnsActiveSeason(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewGetCurrentSeasonHandler(&activeSeasonsStub{seasons: []domain.Season{season}})

	// Act
	got, err := handler.Handle()

	// Assert
	require.NoError(t, err)
	require.Equal(t, "season-1", got.SeasonID)
}

func TestGetCurrentSeason_NoActiveSeason(t *testing.T) {
	// Arrange
	handler := application.NewGetCurrentSeasonHandler(&activeSeasonsStub{})

	// Act
	_, err := handler.Handle()

	// Assert
	require.ErrorIs(t, err, domain.ErrNoActiveSeason)
}

func TestGetCurrentSeason_MultipleActiveSignalsInvariantBreach(t *testing.T) {
	// Arrange
	two := []domain.Season{
		domain.NewSeason("season-1", domain.SeasonStateActive),
		domain.NewSeason("season-2", domain.SeasonStateActive),
	}
	handler := application.NewGetCurrentSeasonHandler(&activeSeasonsStub{seasons: two})

	// Act
	_, err := handler.Handle()

	// Assert
	require.ErrorIs(t, err, domain.ErrActiveSeasonAlreadyExists)
}
