package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// ListPlayerContractsHandler orchestrates the my-contracts listing use case.
type ListPlayerContractsHandler struct {
	contracts PlayerContractsReader
}

// NewListPlayerContractsHandler creates a handler with the reader port.
func NewListPlayerContractsHandler(contracts PlayerContractsReader) *ListPlayerContractsHandler {
	return &ListPlayerContractsHandler{contracts: contracts}
}

// Handle returns the contracts where the player is in the roster, filtered
// by season and status bucket, projected into the requested language —
// each item's ContentLanguage covers its embedded offer and terms.
func (h *ListPlayerContractsHandler) Handle(playerID string, filter domain.ContractListFilter, language domain.Language) ([]LocalizedContract, error) {
	contracts, err := h.contracts.ListByPlayer(playerID, filter)
	if err != nil {
		return nil, err
	}
	mine := make([]LocalizedContract, 0, len(contracts))
	for _, c := range contracts {
		if c.IncludesPlayer(playerID) {
			mine = append(mine, LocalizeContract(c, language))
		}
	}
	return mine, nil
}
