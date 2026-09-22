package domain

// MeritLedger is an append-only collection of credits per bucket.
type MeritLedger struct {
	credits []MeritCredit
}

// NewMeritLedger creates an empty ledger.
func NewMeritLedger() *MeritLedger {
	return &MeritLedger{}
}

// Record adds a credit to the ledger.
func (l *MeritLedger) Record(credit MeritCredit) {
	l.credits = append(l.credits, credit)
}

// TotalFor returns the sum of credits for the given bucket.
func (l *MeritLedger) TotalFor(tenantID, seasonID, category string, track Track, recipientID string) int {
	total := 0
	for _, c := range l.CreditsFor(tenantID, seasonID, category, track, recipientID) {
		total += c.Amount
	}
	return total
}

// CreditsFor returns the credits matching the given bucket.
func (l *MeritLedger) CreditsFor(tenantID, seasonID, category string, track Track, recipientID string) []MeritCredit {
	var matches []MeritCredit
	for _, c := range l.credits {
		if c.TenantID == tenantID && c.SeasonID == seasonID && c.Category == category && c.Track == track && c.RecipientID == recipientID {
			matches = append(matches, c)
		}
	}
	return matches
}
