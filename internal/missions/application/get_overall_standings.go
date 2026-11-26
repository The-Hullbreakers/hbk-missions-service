package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// OverallCreditsReader loads season credit facts and completed contract IDs for one ranking query.
type OverallCreditsReader interface {
	ListOverallCredits(seasonID string) ([]domain.MeritCredit, map[string]bool, error)
}

// OverallStandingsResult combines the requested page with the caller's full-board position.
type OverallStandingsResult struct {
	SeasonID     string
	Participants int
	Board        domain.Leaderboard
	MyStanding   *domain.LeaderboardEntry
}

// GetOverallStandingsHandler reads the authenticated individual standings of one season.
type GetOverallStandingsHandler struct {
	seasons SeasonReader
	credits OverallCreditsReader
}

func NewGetOverallStandingsHandler(seasons SeasonReader, credits OverallCreditsReader) *GetOverallStandingsHandler {
	return &GetOverallStandingsHandler{seasons: seasons, credits: credits}
}

func (h *GetOverallStandingsHandler) Handle(identity domain.CallerIdentity, seasonID string, limit int, cursor string) (OverallStandingsResult, error) {
	if _, err := h.seasons.Get(seasonID); err != nil {
		return OverallStandingsResult{}, err
	}
	credits, completed, err := h.credits.ListOverallCredits(seasonID)
	if err != nil {
		return OverallStandingsResult{}, err
	}
	standings := domain.ProjectOverallStandings(credits, completed, seasonID)
	board, err := domain.BuildLeaderboard(standings, limit, cursor)
	if err != nil {
		return OverallStandingsResult{}, err
	}
	result := OverallStandingsResult{SeasonID: seasonID, Participants: len(standings), Board: board}
	for rank, standing := range standings {
		if standing.RecipientID != identity.PlayerID() {
			continue
		}
		result.MyStanding = &domain.LeaderboardEntry{
			Rank:              rank + 1,
			RecipientID:       standing.RecipientID,
			Track:             standing.Track,
			Total:             standing.Total,
			CompletedMissions: standing.CompletedMissions,
			UniqueObjectives:  standing.UniqueObjectives,
		}
		break
	}
	return result, nil
}
