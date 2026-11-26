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

// ConfirmObjectiveHandler orchestrates the confirm-objective use case.
type ConfirmObjectiveHandler struct {
	contracts ContractReader
	writer    ContractWriter
	seasons   SeasonReader
	evidence  EvidenceObjectReader
}

// NewConfirmObjectiveHandler creates a handler with the required ports.
func NewConfirmObjectiveHandler(contracts ContractReader, writer ContractWriter, seasons SeasonReader) *ConfirmObjectiveHandler {
	return &ConfirmObjectiveHandler{
		contracts: contracts,
		writer:    writer,
		seasons:   seasons,
	}
}

// NewConfirmObjectiveHandlerWithEvidence creates a handler that can also
// validate and attach stored media references through the evidence port.
func NewConfirmObjectiveHandlerWithEvidence(contracts ContractReader, writer ContractWriter, seasons SeasonReader, evidence EvidenceObjectReader) *ConfirmObjectiveHandler {
	h := NewConfirmObjectiveHandler(contracts, writer, seasons)
	h.evidence = evidence
	return h
}

// Handle confirms a unit submitted by the authenticated caller and stages
// the resulting credit on the contract; reviewer approval commits it to the
// ledger. Authorization requires roster membership or a privileged role.
func (h *ConfirmObjectiveHandler) Handle(identity domain.CallerIdentity, contractID, objectiveID, unitID, seasonID string) (domain.MeritCredit, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if !canAdjudicate(identity, contract) {
		return domain.MeritCredit{}, domain.ErrCallerNotAuthorized
	}

	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	credit, err := domain.ConfirmObjective(contract, objectiveID, unitID, season)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if err := h.writer.Save(contract); err != nil {
		return domain.MeritCredit{}, err
	}

	return credit, nil
}

// HandleWithMedia behaves like Handle and additionally validates that the
// submitted media reference belongs to the contract and exists in storage,
// then attaches it to the staged credit.
func (h *ConfirmObjectiveHandler) HandleWithMedia(identity domain.CallerIdentity, contractID, objectiveID, unitID, mediaURI, seasonID string) (domain.MeritCredit, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if !canAdjudicate(identity, contract) {
		return domain.MeritCredit{}, domain.ErrCallerNotAuthorized
	}

	uri, err := domain.NewMediaURI(mediaURI, contract.ContractID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if h.evidence == nil {
		return domain.MeritCredit{}, domain.ErrEvidenceMediaInvalid
	}
	exists, err := h.evidence.Exists(uri)
	if err != nil {
		return domain.MeritCredit{}, err
	}
	if !exists {
		return domain.MeritCredit{}, domain.ErrEvidenceMediaInvalid
	}

	season, err := h.seasons.Get(seasonID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	credit, err := contract.ConfirmWithMedia(objectiveID, unitID, uri, season)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if err := h.writer.Save(contract); err != nil {
		return domain.MeritCredit{}, err
	}

	return credit, nil
}

// canAdjudicate allows roster members or callers with an admin/moderator role.
func canAdjudicate(identity domain.CallerIdentity, contract *domain.MissionContract) bool {
	if contract.CanSubmitEvidence(identity) {
		return true
	}
	return identity.HasRole("admin") || identity.HasRole("moderator")
}
