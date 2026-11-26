package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// ContractCreditsWriter commits the adjudicated contract state and all of its
// materialized merit credits as one atomic unit: either everything reaches
// the ledger or nothing does.
type ContractCreditsWriter interface {
	SaveApproved(contract *domain.MissionContract, credits []domain.MeritCredit) error
}

// ApproveContractHandler orchestrates reviewer approval of a pending contract:
// the staged credits plus the completion reward (when every required
// objective is confirmed) commit to the ledger atomically.
type ApproveContractHandler struct {
	contracts ContractReader
	writer    ContractCreditsWriter
}

// NewApproveContractHandler creates a handler with the required ports.
func NewApproveContractHandler(contracts ContractReader, writer ContractCreditsWriter) *ApproveContractHandler {
	return &ApproveContractHandler{contracts: contracts, writer: writer}
}

// Handle approves a PENDING_VALIDATION contract. Only callers with an
// adjudicator role may approve; the contract state guard stays in the domain.
// The written credits are returned for the response projection.
func (h *ApproveContractHandler) Handle(identity domain.CallerIdentity, contractID string) ([]domain.MeritCredit, error) {
	return h.HandleWithVerdict(identity, contractID, nil)
}

// HandleWithVerdict behaves like Handle and additionally drops the staged
// units named in rejected before materializing credits — the contract
// completes with the approved subset and rejected units are recorded.
func (h *ApproveContractHandler) HandleWithVerdict(identity domain.CallerIdentity, contractID string, rejected []domain.UnitRef) ([]domain.MeritCredit, error) {
	if !isReviewer(identity) {
		return nil, domain.ErrCallerNotAuthorized
	}
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return nil, err
	}
	credits, err := contract.ApproveWithVerdict(rejected)
	if err != nil {
		return nil, err
	}
	if err := h.writer.SaveApproved(contract, credits); err != nil {
		return nil, err
	}
	return credits, nil
}

// RejectContractHandler orchestrates reviewer rejection of a pending contract:
// the contract lands in the terminal DISQUALIFIED state and no staged credit
// reaches the ledger.
type RejectContractHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewRejectContractHandler creates a handler with the required ports.
func NewRejectContractHandler(contracts ContractReader, writer ContractWriter) *RejectContractHandler {
	return &RejectContractHandler{contracts: contracts, writer: writer}
}

// Handle rejects a PENDING_VALIDATION contract. Only callers with an
// adjudicator role may reject.
func (h *RejectContractHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	if !isReviewer(identity) {
		return domain.ErrCallerNotAuthorized
	}
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.Reject(); err != nil {
		return err
	}
	return h.writer.Save(contract)
}
