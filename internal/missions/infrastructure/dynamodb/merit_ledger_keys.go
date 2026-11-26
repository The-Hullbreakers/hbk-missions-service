package dynamodb

import (
	"fmt"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func creditBucketKey(credit domain.MeritCredit) string {
	return bucketKey(credit.SeasonID, credit.Category, credit.Track, credit.RecipientID)
}

type standingBuilder struct {
	total      int
	contracts  map[string]bool
	objectives map[string]bool
}

func newStandingBuilder() *standingBuilder {
	return &standingBuilder{
		contracts:  make(map[string]bool),
		objectives: make(map[string]bool),
	}
}

func leaderboardBucketKey(credit domain.MeritCredit) string {
	return leaderboardBucketKeyFromParts(credit.SeasonID, credit.Category, credit.Track)
}

func leaderboardBucketKeyFromParts(seasonID, category string, track domain.Track) string {
	return fmt.Sprintf("CREDIT#%s#%s#%s", seasonID, category, track)
}

func leaderboardSortKey(credit domain.MeritCredit) string {
	return fmt.Sprintf("RECIPIENT#%s#%s", credit.RecipientID, credit.ConfirmationID)
}

func bucketKey(seasonID, category string, track domain.Track, recipientID string) string {
	return fmt.Sprintf("CREDIT#%s#%s#%s#%s", seasonID, category, track, recipientID)
}
