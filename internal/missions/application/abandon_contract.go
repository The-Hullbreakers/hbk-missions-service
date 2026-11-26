package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// AbandonContractHandler orchestrates the contract-abandon use case.
type AbandonContractHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewAbandonContractHandler creates a handler with the required ports.
func NewAbandonContractHandler(contracts ContractReader, writer ContractWriter) *AbandonContractHandler {
	return &AbandonContractHandler{contracts: contracts, writer: writer}
}

// Handle transitions a non-terminal contract to ABANDONED for a roster member,
// releasing the active slot without deleting the attempt history.
func (h *AbandonContractHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.Abandon(identity); err != nil {
		return err
	}
	return h.writer.Save(contract)
}
