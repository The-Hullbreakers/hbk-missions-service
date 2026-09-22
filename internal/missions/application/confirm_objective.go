package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// ContractReader loads a mission contract by identity.
type ContractReader interface {
	Get(contractID string) (*domain.MissionContract, error)
}

// ContractWriter persists an updated mission contract.
type ContractWriter interface {
	Save(contract *domain.MissionContract) error
}

// SeasonReader loads a season by identity.
type SeasonReader interface {
	Get(seasonID string) (domain.Season, error)
}

// ObjectiveReader loads an objective by identity.
type ObjectiveReader interface {
	Get(objectiveID string) (domain.Objective, error)
}

// MeritLedgerWriter persists a credit.
type MeritLedgerWriter interface {
	Save(credit domain.MeritCredit) error
}

// ConfirmObjectiveHandler orchestrates the confirm-objective use case.
type ConfirmObjectiveHandler struct {
	contracts      ContractReader
	contractWriter ContractWriter
	seasons        SeasonReader
	objectives     ObjectiveReader
	ledger         MeritLedgerWriter
}

// NewConfirmObjectiveHandler creates a handler with the required ports.
func NewConfirmObjectiveHandler(contracts ContractReader, contractWriter ContractWriter, seasons SeasonReader, objectives ObjectiveReader, ledger MeritLedgerWriter) *ConfirmObjectiveHandler {
	return &ConfirmObjectiveHandler{
		contracts:      contracts,
		contractWriter: contractWriter,
		seasons:        seasons,
		objectives:     objectives,
		ledger:         ledger,
	}
}

// Handle confirms a unit and records the resulting credit.
func (h *ConfirmObjectiveHandler) Handle(tenantID, contractID, objectiveID, unitID, seasonID string) (domain.MeritCredit, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	objective, err := h.objectives.Get(objectiveID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	credit, err := domain.ConfirmObjective(contract, objective, unitID, season, tenantID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if err := h.contractWriter.Save(contract); err != nil {
		return domain.MeritCredit{}, err
	}

	if err := h.ledger.Save(credit); err != nil {
		return domain.MeritCredit{}, err
	}

	return credit, nil
}
