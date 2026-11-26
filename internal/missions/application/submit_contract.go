package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// SubmitContractHandler orchestrates the contract-submit use case: a roster
// member hands the contract over to reviewer adjudication.
type SubmitContractHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewSubmitContractHandler creates a handler with the required ports.
func NewSubmitContractHandler(contracts ContractReader, writer ContractWriter) *SubmitContractHandler {
	return &SubmitContractHandler{contracts: contracts, writer: writer}
}

// Handle transitions an IN_PROGRESS contract to PENDING_VALIDATION. The active
// slot frees immediately and the staged credits wait for reviewer
// adjudication — nothing reaches the ledger here.
func (h *SubmitContractHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.Submit(identity); err != nil {
		return err
	}
	return h.writer.Save(contract)
}
