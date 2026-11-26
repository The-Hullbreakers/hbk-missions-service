package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// currentSeasonSelector resolves the "current" statistics request to the
// single ACTIVE season id.
const currentSeasonSelector = "current"

// GetSeasonStatisticsHandler serves the per-season dashboard summary:
// buckets grouped by category and track with participation facts and the
// ranked top entries, for any season state.
type GetSeasonStatisticsHandler struct {
	seasons      SeasonReader
	active       ActiveSeasonsReader
	buckets      StandingsBucketReader
	leaderboards LeaderboardReader
}

// NewGetSeasonStatisticsHandler creates a handler with the required readers.
func NewGetSeasonStatisticsHandler(
	seasons SeasonReader,
	active ActiveSeasonsReader,
	buckets StandingsBucketReader,
	leaderboards LeaderboardReader,
) *GetSeasonStatisticsHandler {
	return &GetSeasonStatisticsHandler{seasons: seasons, active: active, buckets: buckets, leaderboards: leaderboards}
}

// Handle returns the statistics projection for the requested season.
// seasonID "current" resolves to the single ACTIVE season.
func (h *GetSeasonStatisticsHandler) Handle(seasonID, category, track string, topN int) (domain.SeasonStatistics, error) {
	parsedTrack, err := parseOptionalTrack(track)
	if err != nil {
		return domain.SeasonStatistics{}, err
	}

	season, err := h.resolveSeason(seasonID)
	if err != nil {
		return domain.SeasonStatistics{}, err
	}

	buckets, err := h.buckets.ListBuckets(season.SeasonID)
	if err != nil {
		return domain.SeasonStatistics{}, err
	}

	stats := domain.SeasonStatistics{
		SeasonID: season.SeasonID,
		State:    season.State,
		Official: season.Official,
		Buckets:  make([]domain.SeasonBucketStatistics, 0),
	}
	for _, bucket := range buckets {
		if !matchesFilter(bucket, category, parsedTrack) {
			continue
		}
		bucketStats, err := h.bucketStatistics(season.SeasonID, bucket, topN)
		if err != nil {
			return domain.SeasonStatistics{}, err
		}
		stats.Buckets = append(stats.Buckets, bucketStats)
	}
	return stats, nil
}

// bucketStatistics ranks a single bucket and slices the top entries.
func (h *GetSeasonStatisticsHandler) bucketStatistics(seasonID string, bucket domain.StandingBucket, topN int) (domain.SeasonBucketStatistics, error) {
	standings, err := h.leaderboards.ListStandings(seasonID, bucket.Category, bucket.Track, 0)
	if err != nil {
		return domain.SeasonBucketStatistics{}, err
	}

	board, err := domain.BuildLeaderboard(standings, len(standings), "")
	if err != nil {
		return domain.SeasonBucketStatistics{}, err
	}

	top := board.Entries
	if topN > 0 && len(top) > topN {
		top = top[:topN]
	}

	return domain.SeasonBucketStatistics{
		Bucket:       bucket,
		Participants: len(standings),
		TotalMerits:  sumStandings(standings),
		Top:          top,
	}, nil
}

// resolveSeason returns the season by id, or the single ACTIVE season
// when the id is the "current" selector.
func (h *GetSeasonStatisticsHandler) resolveSeason(seasonID string) (domain.Season, error) {
	if seasonID != currentSeasonSelector {
		return h.seasons.Get(seasonID)
	}
	active, err := h.active.ListActive()
	if err != nil {
		return domain.Season{}, err
	}
	if len(active) == 0 {
		return domain.Season{}, domain.ErrNoActiveSeason
	}
	if len(active) > 1 {
		return domain.Season{}, domain.ErrActiveSeasonAlreadyExists
	}
	return active[0], nil
}

// parseOptionalTrack parses a track filter; empty means no filter.
func parseOptionalTrack(track string) (domain.Track, error) {
	if track == "" {
		return "", nil
	}
	return domain.ParseTrack(track)
}

func sumStandings(standings []domain.Standing) int {
	total := 0
	for _, s := range standings {
		total += s.Total
	}
	return total
}
