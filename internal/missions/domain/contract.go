package domain

// ContractState represents the lifecycle state of a contract.
type ContractState string

const (
	ContractAccepted          ContractState = "ACCEPTED"
	ContractInProgress        ContractState = "IN_PROGRESS"
	ContractPendingValidation ContractState = "PENDING_VALIDATION"
	ContractCompleted         ContractState = "COMPLETED"
	ContractFailed            ContractState = "FAILED"
	ContractAbandoned         ContractState = "ABANDONED"
	ContractExpired           ContractState = "EXPIRED"
	ContractCancelled         ContractState = "CANCELLED"
	ContractDisqualified      ContractState = "DISQUALIFIED"
)

// IsTerminal returns true for states that end a contract and cannot return to active.
func (s ContractState) IsTerminal() bool {
	return s == ContractCompleted || s == ContractFailed || s == ContractAbandoned ||
		s == ContractExpired || s == ContractCancelled || s == ContractDisqualified
}

// IsActive returns true for states that count against the active contract limit.
func (s ContractState) IsActive() bool {
	return s == ContractAccepted || s == ContractInProgress || s == ContractPendingValidation
}

// ActiveContractSummary is a lightweight, read-only value object used by acceptance policy.
type ActiveContractSummary struct {
	OfferID      string
	Category     string
	State        ContractState
	DefinitionID string
}

// NewActiveContractSummary creates a summary from the offer and state of an active contract.
func NewActiveContractSummary(offerID, category string, state ContractState, definitionID string) ActiveContractSummary {
	return ActiveContractSummary{
		OfferID:      offerID,
		Category:     category,
		State:        state,
		DefinitionID: definitionID,
	}
}
