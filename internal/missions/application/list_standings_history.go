package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// StandingsBucketReader lists the category/track buckets that carry credits
// for a season.
type StandingsBucketReader interface {
	ListBuckets(seasonID string) ([]domain.StandingBucket, error)
}

// ListStandingsHistoryHandler serves the finalized standings projection; only
// FINALIZED seasons qualify — other states surface ErrSeasonStateInvalid.
type ListStandingsHistoryHandler struct {
	seasons      SeasonReader
	buckets      StandingsBucketReader
	leaderboards LeaderboardReader
}

// NewListStandingsHistoryHandler creates a handler with the required readers.
func NewListStandingsHistoryHandler(seasons SeasonReader, buckets StandingsBucketReader, leaderboards LeaderboardReader) *ListStandingsHistoryHandler {
	return &ListStandingsHistoryHandler{seasons: seasons, buckets: buckets, leaderboards: leaderboards}
}

// Handle returns the ranked buckets for a finalized season, optionally
// narrowed by category and track.
func (h *ListStandingsHistoryHandler) Handle(seasonID, category, track string) (domain.StandingsHistory, error) {
	var parsedTrack domain.Track
	if track != "" {
		parsed, err := domain.ParseTrack(track)
		if err != nil {
			return domain.StandingsHistory{}, domain.ErrInvalidTrack
		}
		parsedTrack = parsed
	}

	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.StandingsHistory{}, err
	}
	if season.State != domain.SeasonStateFinalized {
		return domain.StandingsHistory{}, domain.ErrSeasonStateInvalid
	}

	buckets, err := h.buckets.ListBuckets(seasonID)
	if err != nil {
		return domain.StandingsHistory{}, err
	}

	history := domain.StandingsHistory{SeasonID: seasonID, Buckets: make([]domain.StandingsBucketHistory, 0)}
	for _, bucket := range buckets {
		if !matchesFilter(bucket, category, parsedTrack) {
			continue
		}
		standings, err := h.leaderboards.ListStandings(seasonID, bucket.Category, bucket.Track, 0)
		if err != nil {
			return domain.StandingsHistory{}, err
		}
		board, err := domain.BuildLeaderboard(standings, len(standings), "")
		if err != nil {
			return domain.StandingsHistory{}, err
		}
		history.Buckets = append(history.Buckets, domain.StandingsBucketHistory{Bucket: bucket, Entries: board.Entries})
	}
	return history, nil
}

func matchesFilter(bucket domain.StandingBucket, category string, track domain.Track) bool {
	if category != "" && bucket.Category != category {
		return false
	}
	if track != "" && bucket.Track != track {
		return false
	}
	return true
}
