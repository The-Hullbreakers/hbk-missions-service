package domain

import (
	"strings"
)

// ConfirmationID is a composite identity for a confirmed objective unit.
// It encodes contract, objective, unit, and recipient in a single stable string.
type ConfirmationID struct {
	contractID  string
	objectiveID string
	unitID      string
	recipientID string
}

// NewConfirmationID creates a confirmation identity from its four required parts.
func NewConfirmationID(contractID, objectiveID, unitID, recipientID string) (ConfirmationID, error) {
	if contractID == "" || objectiveID == "" || unitID == "" || recipientID == "" {
		return ConfirmationID{}, ErrInvalidConfirmationID
	}
	if strings.ContainsAny(contractID+objectiveID+unitID+recipientID, "~") {
		return ConfirmationID{}, ErrInvalidConfirmationID
	}
	return ConfirmationID{
		contractID:  contractID,
		objectiveID: objectiveID,
		unitID:      unitID,
		recipientID: recipientID,
	}, nil
}

// ParseConfirmationID parses a previously encoded confirmation identity.
// The delimiter is "~" because generated contract ids contain colons.
func ParseConfirmationID(raw string) (ConfirmationID, error) {
	parts := strings.Split(raw, "~")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return ConfirmationID{}, ErrInvalidConfirmationID
	}
	return ConfirmationID{
		contractID:  parts[0],
		objectiveID: parts[1],
		unitID:      parts[2],
		recipientID: parts[3],
	}, nil
}

// String returns the canonical "contract~objective~unit~recipient" form.
func (id ConfirmationID) String() string {
	return id.contractID + "~" + id.objectiveID + "~" + id.unitID + "~" + id.recipientID
}

// ContractID returns the contract part of the identity.
func (id ConfirmationID) ContractID() string { return id.contractID }

// ObjectiveID returns the objective part of the identity.
func (id ConfirmationID) ObjectiveID() string { return id.objectiveID }

// UnitID returns the unit part of the identity.
func (id ConfirmationID) UnitID() string { return id.unitID }

// RecipientID returns the recipient part of the identity.
func (id ConfirmationID) RecipientID() string { return id.recipientID }

// IsZero reports whether the identity has no value.
func (id ConfirmationID) IsZero() bool {
	return id.contractID == ""
}
