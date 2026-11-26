package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// StartContractHandler orchestrates the contract-start use case.
type StartContractHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewStartContractHandler creates a handler with the required ports.
func NewStartContractHandler(contracts ContractReader, writer ContractWriter) *StartContractHandler {
	return &StartContractHandler{contracts: contracts, writer: writer}
}

// Handle transitions an ACCEPTED contract to IN_PROGRESS for a roster member.
func (h *StartContractHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.Start(identity); err != nil {
		return err
	}
	return h.writer.Save(contract)
}
