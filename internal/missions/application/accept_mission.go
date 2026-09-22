package application

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// OfferReader loads a published mission offer by identity.
type OfferReader interface {
	Get(tenantID, offerID string) (domain.MissionOffer, error)
}

// ActiveContractsReader lists a player's currently active mission contracts.
type ActiveContractsReader interface {
	ListActiveByPlayer(tenantID, playerID, category string) ([]domain.MissionContract, error)
}

// AcceptMissionHandler orchestrates the accept-mission use case.
type AcceptMissionHandler struct {
	policy          domain.ActiveLimitPolicy
	offers          OfferReader
	activeContracts ActiveContractsReader
	contracts       ContractWriter
}

// NewAcceptMissionHandler creates a handler with the required ports.
func NewAcceptMissionHandler(policy domain.ActiveLimitPolicy, offers OfferReader, activeContracts ActiveContractsReader, contracts ContractWriter) *AcceptMissionHandler {
	return &AcceptMissionHandler{
		policy:          policy,
		offers:          offers,
		activeContracts: activeContracts,
		contracts:       contracts,
	}
}

// Handle attempts to accept an offer for a player or team.
func (h *AcceptMissionHandler) Handle(tenantID, playerID, offerID, teamID string) (*domain.MissionContract, error) {
	offer, err := h.offers.Get(tenantID, offerID)
	if err != nil {
		return nil, err
	}

	if offer.Mode == domain.MissionModeTeam && teamID == "" {
		return nil, domain.ErrTeamIDRequired
	}

	active, err := h.activeContracts.ListActiveByPlayer(tenantID, playerID, offer.Category)
	if err != nil {
		return nil, err
	}

	summaries := make([]domain.ActiveContractSummary, len(active))
	for i, c := range active {
		summaries[i] = c.Summary()
	}

	decision := domain.DecideAcceptance(offer, summaries, h.policy)
	if !decision.Accepted {
		if decision.Reason == domain.ReasonOfferCapacityExhausted {
			return nil, domain.ErrOfferCapacityExhausted
		}
		if decision.Reason == domain.ReasonMissionDefinitionAlreadyActive {
			return nil, domain.ErrMissionDefinitionAlreadyActive
		}
		return nil, domain.ErrActiveMissionLimitReached
	}

	contractID := tenantID + ":" + offerID + ":" + playerID
	var contract *domain.MissionContract
	switch offer.Mode {
	case domain.MissionModeTeam:
		contract = domain.NewTeamMissionContract(contractID, offer, []string{playerID}, teamID)
	case domain.MissionModeSolo, domain.MissionModeFlex:
		if teamID != "" {
			contract = domain.NewTeamMissionContract(contractID, offer, []string{playerID}, teamID)
		} else {
			contract = domain.NewMissionContract(contractID, offer, []string{playerID})
		}
	default:
		contract = domain.NewMissionContract(contractID, offer, []string{playerID})
	}

	if err := h.contracts.Save(contract); err != nil {
		return nil, err
	}

	return contract, nil
}
