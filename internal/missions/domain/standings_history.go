package domain

// StandingBucket identifies one category/track slice of a season's standings.
type StandingBucket struct {
	Category string
	Track    Track
}

// StandingsHistory is the immutable final-leaderboard projection of a
// FINALIZED season, grouped by bucket.
type StandingsHistory struct {
	SeasonID string
	Buckets  []StandingsBucketHistory
}

// StandingsBucketHistory is one ranked bucket inside the history.
type StandingsBucketHistory struct {
	Bucket  StandingBucket
	Entries []LeaderboardEntry
}
