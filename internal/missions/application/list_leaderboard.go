package application

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// LeaderboardReader loads standings for a leaderboard bucket.
type LeaderboardReader interface {
	ListStandings(tenantID, seasonID, category string, track domain.Track, minMerits int) ([]domain.Standing, error)
}

// ListLeaderboardHandler orchestrates the leaderboard query.
type ListLeaderboardHandler struct {
	seasons      SeasonReader
	leaderboards LeaderboardReader
}

// NewListLeaderboardHandler creates a handler with the required readers.
func NewListLeaderboardHandler(seasons SeasonReader, leaderboards LeaderboardReader) *ListLeaderboardHandler {
	return &ListLeaderboardHandler{
		seasons:      seasons,
		leaderboards: leaderboards,
	}
}

// Handle returns a paginated leaderboard for the given bucket.
func (h *ListLeaderboardHandler) Handle(tenantID, seasonID, category, track string, limit, minMerits int, cursor string) (domain.Leaderboard, error) {
	parsedTrack, err := domain.ParseTrack(track)
	if err != nil {
		return domain.Leaderboard{}, domain.ErrInvalidTrack
	}

	if _, err := h.seasons.Get(seasonID); err != nil {
		return domain.Leaderboard{}, err
	}

	standings, err := h.leaderboards.ListStandings(tenantID, seasonID, category, parsedTrack, minMerits)
	if err != nil {
		return domain.Leaderboard{}, err
	}

	return domain.BuildLeaderboard(standings, limit, cursor)
}
