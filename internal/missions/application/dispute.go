package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// RaiseDisputeHandler orchestrates the member-raise-dispute use case.
type RaiseDisputeHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewRaiseDisputeHandler creates the handler with the required ports.
func NewRaiseDisputeHandler(contracts ContractReader, writer ContractWriter) *RaiseDisputeHandler {
	return &RaiseDisputeHandler{contracts: contracts, writer: writer}
}

// Handle opens a dispute on the contract for the target objective and persists
// the aggregate.
func (h *RaiseDisputeHandler) Handle(identity domain.CallerIdentity, contractID, objectiveID, reason string) (domain.Dispute, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return domain.Dispute{}, err
	}
	dispute, err := contract.RaiseDispute(identity, objectiveID, reason)
	if err != nil {
		return domain.Dispute{}, err
	}
	return dispute, h.writer.Save(contract)
}

// ResolveDisputeHandler orchestrates the moderator-resolve-dispute use case.
type ResolveDisputeHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewResolveDisputeHandler creates the handler with the required ports.
func NewResolveDisputeHandler(contracts ContractReader, writer ContractWriter) *ResolveDisputeHandler {
	return &ResolveDisputeHandler{contracts: contracts, writer: writer}
}

// Handle resolves the dispute; only reviewers may resolve, and the resolved
// outcome is persisted atomically with the contract.
func (h *ResolveDisputeHandler) Handle(identity domain.CallerIdentity, contractID, disputeID string) error {
	if !isReviewer(identity) {
		return domain.ErrCallerNotAuthorized
	}
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.ResolveDispute(disputeID, identity.PlayerID()); err != nil {
		return err
	}
	return h.writer.Save(contract)
}

// ListDisputesHandler serves the reviewer dispute-listing use case.
type ListDisputesHandler struct {
	contracts ContractReader
}

// NewListDisputesHandler creates the handler with the required port.
func NewListDisputesHandler(contracts ContractReader) *ListDisputesHandler {
	return &ListDisputesHandler{contracts: contracts}
}

// Handle returns the contract's disputes; restricted to reviewers.
func (h *ListDisputesHandler) Handle(identity domain.CallerIdentity, contractID string) ([]domain.Dispute, error) {
	if !isReviewer(identity) {
		return nil, domain.ErrCallerNotAuthorized
	}
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return nil, err
	}
	return contract.Disputes(), nil
}
