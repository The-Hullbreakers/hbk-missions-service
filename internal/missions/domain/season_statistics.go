package domain

// SeasonStatistics is the dashboard projection of one season: every
// category/track bucket that carries credits, with participation totals
// and the ranked top entries. It is recomputed from the MeritLedger —
// never stored as a second source of truth.
type SeasonStatistics struct {
	SeasonID string
	State    SeasonState
	Official bool
	Buckets  []SeasonBucketStatistics
}

// SeasonBucketStatistics aggregates one {category, track} bucket.
type SeasonBucketStatistics struct {
	Bucket       StandingBucket
	Participants int
	TotalMerits  int
	Top          []LeaderboardEntry
}
