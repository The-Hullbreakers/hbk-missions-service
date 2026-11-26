package domain

// ReasonActiveMissionLimitReached is the rejection reason when a player is at the active contract limit.
const ReasonActiveMissionLimitReached = "ACTIVE_MISSION_LIMIT_REACHED"

// ReasonMissionDefinitionAlreadyActive is the rejection reason when the player already has an active contract for the same definition.
const ReasonMissionDefinitionAlreadyActive = "MISSION_DEFINITION_ALREADY_ACTIVE"

// ReasonSeasonNotActive is the rejection reason when the offer's bound season is not ACTIVE.
const ReasonSeasonNotActive = "SEASON_NOT_ACTIVE"

// AcceptanceDecision is the outcome of an acceptance attempt.
type AcceptanceDecision struct {
	Accepted bool
	Reason   string
}

// DecideAcceptance returns whether a player may accept an offer given the
// bound season's state, their active contract summaries, and policy.
func DecideAcceptance(offer MissionOffer, seasonState SeasonState, active []ActiveContractSummary, policy ActiveLimitPolicy) AcceptanceDecision {
	if seasonState != SeasonStateActive {
		return AcceptanceDecision{
			Accepted: false,
			Reason:   ReasonSeasonNotActive,
		}
	}

	activeCount := CountActive(active)
	definitionActive := false
	for _, c := range active {
		if c.State.BlocksRetake() {
			if c.DefinitionCode != "" && offer.DefinitionCode != "" && c.DefinitionCode == offer.DefinitionCode {
				definitionActive = true
			}
		}
	}

	if definitionActive {
		return AcceptanceDecision{
			Accepted: false,
			Reason:   ReasonMissionDefinitionAlreadyActive,
		}
	}

	if activeCount >= policy.EffectiveLimit(offer.Category) {
		return AcceptanceDecision{
			Accepted: false,
			Reason:   ReasonActiveMissionLimitReached,
		}
	}

	return AcceptanceDecision{Accepted: true}
}

// CountActive returns how many summaries are in an active state.
func CountActive(active []ActiveContractSummary) int {
	count := 0
	for _, c := range active {
		if c.State.IsActive() {
			count++
		}
	}
	return count
}
