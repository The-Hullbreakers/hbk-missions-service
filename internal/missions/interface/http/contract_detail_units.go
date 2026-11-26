package http

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// detailUnitRef projects an adjudication-rejected unit for reviewer and
// member inspection.
type detailUnitRef struct {
	ObjectiveID string `json:"objective_id"`
	UnitID      string `json:"unit_id"`
	RecipientID string `json:"recipient_id"`
}

// rejectedUnitsProjection parses stored confirmation ids into reviewer-facing
// unit refs; unparseable entries are skipped defensively.
func rejectedUnitsProjection(confirmationIDs []string) []detailUnitRef {
	out := make([]detailUnitRef, 0, len(confirmationIDs))
	for _, raw := range confirmationIDs {
		id, err := domain.ParseConfirmationID(raw)
		if err != nil {
			continue
		}
		out = append(out, detailUnitRef{
			ObjectiveID: id.ObjectiveID(),
			UnitID:      id.UnitID(),
			RecipientID: id.RecipientID(),
		})
	}
	return out
}
