package domain

// UnitRef identifies one staged unit by objective and unit id — the
// reviewer-facing verdict reference. The full ConfirmationID additionally
// embeds contract and recipient, which adjudication already knows.
type UnitRef struct {
	ObjectiveID string
	UnitID      string
}
