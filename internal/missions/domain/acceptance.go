package domain

// ReasonActiveMissionLimitReached is the rejection reason when a player is at the active contract limit.
const ReasonActiveMissionLimitReached = "ACTIVE_MISSION_LIMIT_REACHED"

// ReasonOfferCapacityExhausted is the rejection reason when the offer has no remaining capacity.
const ReasonOfferCapacityExhausted = "OFFER_CAPACITY_EXHAUSTED"

// ReasonMissionDefinitionAlreadyActive is the rejection reason when the player already has an active contract for the same definition.
const ReasonMissionDefinitionAlreadyActive = "MISSION_DEFINITION_ALREADY_ACTIVE"

// AcceptanceDecision is the outcome of an acceptance attempt.
type AcceptanceDecision struct {
	Accepted bool
	Reason   string
}

// DecideAcceptance returns whether a player may accept an offer given their active contract summaries and policy.
func DecideAcceptance(offer MissionOffer, active []ActiveContractSummary, policy ActiveLimitPolicy) AcceptanceDecision {
	activeCount := 0
	offerActiveCount := 0
	definitionActive := false
	for _, c := range active {
		if c.State.IsActive() {
			activeCount++
			if c.OfferID == offer.OfferID {
				offerActiveCount++
			}
			if c.DefinitionID != "" && offer.DefinitionID != "" && c.DefinitionID == offer.DefinitionID {
				definitionActive = true
			}
		}
	}

	if offer.Capacity > 0 && offerActiveCount >= offer.Capacity {
		return AcceptanceDecision{
			Accepted: false,
			Reason:   ReasonOfferCapacityExhausted,
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
