package application

import (
	"errors"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// OfferReader loads a published mission offer by identity.
type OfferReader interface {
	Get(offerID string) (domain.MissionOffer, error)
}

// ActiveContractsReader lists a player's currently active mission contracts
// across all categories; the limit applies to the total active set.
type ActiveContractsReader interface {
	ListActiveByPlayer(playerID string) ([]domain.MissionContract, error)
}

// AtomicAcceptanceWriter commits the contract and the active-set guard
// in one conditional write, so concurrent acceptances serialize without
// overspending the player's effective limit.
type AtomicAcceptanceWriter interface {
	CommitAcceptance(contract *domain.MissionContract, offer domain.MissionOffer, effectiveLimit int) error
}

// AcceptMissionHandler orchestrates the accept-mission use case.
type AcceptMissionHandler struct {
	policies        ActiveLimitPolicyReader
	offers          OfferReader
	activeContracts ActiveContractsReader
	history         PlayerContractsReader
	definitions     DefinitionReader
	seasons         SeasonReader
	contracts       AtomicAcceptanceWriter
	membership      TeamMembershipReader
}

// NewAcceptMissionHandler creates a handler with the required ports. The
// membership reader may be nil when no team source is wired; declared
// teammates are then always rejected.
func NewAcceptMissionHandler(policies ActiveLimitPolicyReader, offers OfferReader, activeContracts ActiveContractsReader, history PlayerContractsReader, definitions DefinitionReader, seasons SeasonReader, contracts AtomicAcceptanceWriter, membership TeamMembershipReader) *AcceptMissionHandler {
	return &AcceptMissionHandler{
		policies:        policies,
		offers:          offers,
		activeContracts: activeContracts,
		history:         history,
		definitions:     definitions,
		seasons:         seasons,
		contracts:       contracts,
		membership:      membership,
	}
}

// Handle attempts to accept an offer for the authenticated caller.
// Identity (player, team membership) comes only from CallerIdentity;
// request-supplied identity fields are never trusted. Declared teammates
// are validated against the team-membership source and consume one slot
// of their own active-contract limit.
func (h *AcceptMissionHandler) Handle(identity domain.CallerIdentity, offerID, teamID string, teammates []string) (*domain.MissionContract, error) {
	playerID := identity.PlayerID()

	offer, err := h.offers.Get(offerID)
	if err != nil {
		return nil, err
	}

	if len(teammates) > 0 && teamID == "" {
		return nil, domain.ErrTeamIDRequired
	}
	if err := domain.DecideTeamParticipation(offer.Mode, teamID); err != nil {
		return nil, err
	}
	if teamID != "" && !identity.IsTeamMember(teamID) {
		return nil, domain.ErrNotTeamMember
	}

	roster, err := h.validateRoster(playerID, teamID, teammates)
	if err != nil {
		return nil, err
	}

	season, err := h.seasons.Get(offer.SeasonID)
	if err != nil {
		if errors.Is(err, domain.ErrSeasonNotFound) {
			return nil, domain.ErrSeasonNotActive
		}
		return nil, err
	}

	memberActive, err := h.memberSummaries(roster.Members())
	if err != nil {
		return nil, err
	}

	policy, err := h.policies.GetActiveLimit()
	if err != nil {
		return nil, err
	}

	decision := domain.DecideAcceptance(offer, season.State, memberActive[playerID], policy)
	if !decision.Accepted {
		return nil, acceptanceError(decision.Reason)
	}
	if len(roster.Members()) > 1 {
		teamDecision := domain.DecideRosterAcceptance(offer, season.State, memberActive, policy)
		if !teamDecision.Accepted {
			return nil, acceptanceError(teamDecision.Reason)
		}
	}

	contractID, terms, err := h.nextAttempt(offer, playerID)
	if err != nil {
		return nil, err
	}
	var contract *domain.MissionContract
	if teamID != "" {
		contract = domain.NewTeamMissionContract(contractID, offer, roster.Members(), teamID)
	} else {
		contract = domain.NewMissionContract(contractID, offer, roster.Members())
	}
	contract.Terms = terms

	if err := h.contracts.CommitAcceptance(contract, offer, policy.EffectiveLimit(offer.Category)); err != nil {
		return nil, err
	}

	return contract, nil
}

// acceptanceError maps the domain rejection reason to its stable error.
func acceptanceError(reason string) error {
	switch reason {
	case domain.ReasonSeasonNotActive:
		return domain.ErrSeasonNotActive
	case domain.ReasonMissionDefinitionAlreadyActive:
		return domain.ErrMissionDefinitionAlreadyActive
	case domain.ReasonTeammateLimitReached:
		return domain.ErrTeammateLimitReached
	default:
		return domain.ErrActiveMissionLimitReached
	}
}
