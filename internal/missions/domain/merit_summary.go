package domain

// MeritSummary is the derived maximum a contract on an offer can earn:
// the sum of objective credits plus the frozen completion reward.
type MeritSummary struct {
	ObjectiveMerits  int
	CompletionReward int
	TotalMerits      int
}
