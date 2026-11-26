package domain

// ReasonTeammateLimitReached is the rejection reason when a roster member
// other than the requester is already at the active contract limit.
const ReasonTeammateLimitReached = "TEAMMATE_LIMIT_REACHED"

// DecideTeamParticipation enforces mode legality: a team_id is meaningful
// only on TEAM/FLEX offers, and TEAM offers always require one.
func DecideTeamParticipation(mode MissionMode, teamID string) error {
	switch {
	case mode == MissionModeTeam && teamID == "":
		return ErrTeamIDRequired
	case mode == MissionModeSolo && teamID != "":
		return ErrTeamNotAllowed
	default:
		return nil
	}
}

// DecideRosterAcceptance returns whether a team acceptance may proceed given
// the bound season's state and every member's active contract summaries.
// Each member is decided through DecideAcceptance: a same-definition conflict
// on any member resolves MISSION_DEFINITION_ALREADY_ACTIVE; a member at
// their personal limit resolves TEAMMATE_LIMIT_REACHED. The definition
// conflict is checked first so the outcome is deterministic regardless of
// map iteration order.
func DecideRosterAcceptance(offer MissionOffer, seasonState SeasonState, memberActive map[string][]ActiveContractSummary, policy ActiveLimitPolicy) AcceptanceDecision {
	if seasonState != SeasonStateActive {
		return AcceptanceDecision{Accepted: false, Reason: ReasonSeasonNotActive}
	}
	limitReached := false
	for _, active := range memberActive {
		decision := DecideAcceptance(offer, seasonState, active, policy)
		if decision.Accepted {
			continue
		}
		if decision.Reason == ReasonMissionDefinitionAlreadyActive {
			return AcceptanceDecision{Accepted: false, Reason: ReasonMissionDefinitionAlreadyActive}
		}
		limitReached = true
	}
	if limitReached {
		return AcceptanceDecision{Accepted: false, Reason: ReasonTeammateLimitReached}
	}
	return AcceptanceDecision{Accepted: true}
}
