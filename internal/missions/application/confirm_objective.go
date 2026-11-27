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

// NewConfirmObjectiveHandlerWithEvidence creates a handler that validates
// and attaches stored media references through the evidence port.
func NewConfirmObjectiveHandlerWithEvidence(contracts ContractReader, writer ContractWriter, seasons SeasonReader, evidence EvidenceObjectReader) *ConfirmObjectiveHandler {
	return &ConfirmObjectiveHandler{
		contracts: contracts,
		writer:    writer,
		seasons:   seasons,
		evidence:  evidence,
	}
}

// HandleWithMedia confirms a unit submitted by the authenticated caller and
// stages the resulting credit on the contract; reviewer approval commits it
// to the ledger. A stored media reference owned by the contract is required
// for every confirmation. Authorization requires roster membership or a
// privileged role.
func (h *ConfirmObjectiveHandler) HandleWithMedia(identity domain.CallerIdentity, contractID, objectiveID, unitID, mediaURI, seasonID string) (domain.MeritCredit, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return domain.MeritCredit{}, err
	}

	if !canAdjudicate(identity, contract) {
		return domain.MeritCredit{}, domain.ErrCallerNotAuthorized
	}

	if mediaURI == "" {
		return domain.MeritCredit{}, domain.ErrContractEvidenceRequired
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
