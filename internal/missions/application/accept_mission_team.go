package application

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TeamMembershipReader answers whether a player belongs to a team. It is the
// minimal seam to the external party/membership source.
type TeamMembershipReader interface {
	IsMember(teamID, playerID string) (bool, error)
}

// validateRoster builds the contract roster and verifies every declared
// teammate belongs to the declared team. Without a membership source only
// the requester can form the roster.
func (h *AcceptMissionHandler) validateRoster(playerID, teamID string, teammates []string) (domain.Roster, error) {
	if teamID == "" {
		return domain.NewRoster(playerID, nil)
	}
	roster, err := domain.NewRoster(playerID, teammates)
	if err != nil {
		return domain.Roster{}, err
	}
	for _, teammate := range teammates {
		member, err := h.isMember(teamID, teammate)
		if err != nil {
			return domain.Roster{}, err
		}
		if !member {
			return domain.Roster{}, domain.ErrTeammateNotMember
		}
	}
	return roster, nil
}

func (h *AcceptMissionHandler) isMember(teamID, playerID string) (bool, error) {
	if h.membership == nil {
		return false, nil
	}
	return h.membership.IsMember(teamID, playerID)
}

// memberSummaries loads each member's active contracts as policy summaries.
func (h *AcceptMissionHandler) memberSummaries(members []string) (map[string][]domain.ActiveContractSummary, error) {
	summaries := make(map[string][]domain.ActiveContractSummary, len(members))
	for _, member := range members {
		active, err := h.activeContracts.ListActiveByPlayer(member)
		if err != nil {
			return nil, err
		}
		memberSummaries := make([]domain.ActiveContractSummary, len(active))
		for i, c := range active {
			memberSummaries[i] = c.Summary()
		}
		summaries[member] = memberSummaries
	}
	return summaries, nil
}
