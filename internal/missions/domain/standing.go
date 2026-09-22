package domain

// Standing is a read-only projection of a recipient's score in a bucket.
type Standing struct {
	TenantID          string
	SeasonID          string
	Category          string
	Track             Track
	RecipientID       string
	Total             int
	CompletedMissions int
	UniqueObjectives  int
}

// ProjectStanding computes the total merits and public tie-break facts for a recipient in a bucket.
func ProjectStanding(ledger *MeritLedger, tenantID, seasonID, category string, track Track, recipientID string) Standing {
	credits := ledger.CreditsFor(tenantID, seasonID, category, track, recipientID)

	contracts := make(map[string]bool)
	objectives := make(map[string]bool)
	total := 0
	for _, c := range credits {
		total += c.Amount
		contracts[c.ContractID] = true
		objectives[c.ObjectiveID] = true
	}

	return Standing{
		TenantID:          tenantID,
		SeasonID:          seasonID,
		Category:          category,
		Track:             track,
		RecipientID:       recipientID,
		Total:             total,
		CompletedMissions: len(contracts),
		UniqueObjectives:  len(objectives),
	}
}
