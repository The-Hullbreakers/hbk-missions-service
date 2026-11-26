package domain

import (
	"strings"
	"time"
)

// AdjudicationAttempt is the internal identity of one adjudication run
// against a season. The ID is deterministic per season so a retry after an
// orphaned lock resumes the same attempt instead of minting a parallel one,
// and the Cutoff is the rank-relevance boundary captured once by the lock —
// every later fact and event reuses it verbatim.
type AdjudicationAttempt struct {
	ID       string
	SeasonID string
	Cutoff   time.Time
}

// NewAdjudicationAttempt mints the deterministic attempt for a season. The
// cutoff must be supplied by the caller (the lock), never defaulted here.
func NewAdjudicationAttempt(seasonID string, cutoff time.Time) (AdjudicationAttempt, error) {
	if strings.TrimSpace(seasonID) == "" || cutoff.IsZero() {
		return AdjudicationAttempt{}, ErrAdjudicationAttemptInvalid
	}
	return AdjudicationAttempt{
		ID:       "ADJ#" + seasonID,
		SeasonID: seasonID,
		Cutoff:   cutoff,
	}, nil
}
