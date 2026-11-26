package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// GetContractHandler orchestrates the contract-detail use case: adjudicators
// see the full projection; active roster members read their own contract;
// anyone else is rejected (REQ-MCF-022).
type GetContractHandler struct {
	contracts ContractReader
}

// NewGetContractHandler creates a handler with the contract reader port.
func NewGetContractHandler(contracts ContractReader) *GetContractHandler {
	return &GetContractHandler{contracts: contracts}
}

// Handle returns the stored contract for a reviewer or an active roster
// member, projected into the requested language; departed members and
// outsiders are rejected before any localization happens.
func (h *GetContractHandler) Handle(identity domain.CallerIdentity, contractID string, language domain.Language) (*LocalizedContract, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return nil, err
	}
	if isReviewer(identity) || contract.CanSubmitEvidence(identity) {
		served := LocalizeContract(*contract, language)
		return &served, nil
	}
	return nil, domain.ErrCallerNotAuthorized
}
