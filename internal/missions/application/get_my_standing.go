package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// GetMyStandingHandler resolves the caller's own leaderboard entry for a
// season/category/track; absence surfaces ErrStandingNotFound.
type GetMyStandingHandler struct {
	seasons      SeasonReader
	leaderboards LeaderboardReader
}

// NewGetMyStandingHandler creates a handler with the required readers.
func NewGetMyStandingHandler(seasons SeasonReader, leaderboards LeaderboardReader) *GetMyStandingHandler {
	return &GetMyStandingHandler{seasons: seasons, leaderboards: leaderboards}
}

// Handle returns the caller's ranked entry in the given bucket.
func (h *GetMyStandingHandler) Handle(identity domain.CallerIdentity, seasonID, category, track string) (domain.LeaderboardEntry, error) {
	parsedTrack, err := domain.ParseTrack(track)
	if err != nil {
		return domain.LeaderboardEntry{}, domain.ErrInvalidTrack
	}

	if _, err := h.seasons.Get(seasonID); err != nil {
		return domain.LeaderboardEntry{}, err
	}

	standings, err := h.leaderboards.ListStandings(seasonID, category, parsedTrack, 0)
	if err != nil {
		return domain.LeaderboardEntry{}, err
	}

	board, err := domain.BuildLeaderboard(standings, len(standings), "")
	if err != nil {
		return domain.LeaderboardEntry{}, err
	}

	for _, entry := range board.Entries {
		if isCallerRecipient(identity, entry) {
			return entry, nil
		}
	}
	return domain.LeaderboardEntry{}, domain.ErrStandingNotFound
}

// isCallerRecipient matches the entry recipient to the caller: player id on
// the individual track, any claimed team on the team track.
func isCallerRecipient(identity domain.CallerIdentity, entry domain.LeaderboardEntry) bool {
	if entry.Track == domain.TrackTeam {
		return identity.IsTeamMember(entry.RecipientID)
	}
	return entry.RecipientID == identity.PlayerID()
}
