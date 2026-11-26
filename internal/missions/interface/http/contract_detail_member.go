package http

// memberContractDetail minimizes the member projection: staged credits and
// rejected units are scoped to the caller and never carry evidence media
// references.
func memberContractDetail(detail contractDetail, playerID string) contractDetail {
	credits := make([]detailCredit, 0, len(detail.PendingCredits))
	for _, credit := range detail.PendingCredits {
		if credit.RecipientID != playerID {
			continue
		}
		credit.MediaURI = ""
		credits = append(credits, credit)
	}
	detail.PendingCredits = credits
	rejected := make([]detailUnitRef, 0, len(detail.RejectedUnits))
	for _, unit := range detail.RejectedUnits {
		if unit.RecipientID == playerID {
			rejected = append(rejected, unit)
		}
	}
	detail.RejectedUnits = rejected
	return detail
}
