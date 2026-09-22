package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

const testOfferID = "offer-1"

func defaultLimitPolicy(limit int) domain.ActiveLimitPolicy {
	return domain.NewActiveLimitPolicy(limit, nil)
}

func aBountyOffer(capacity int) domain.MissionOffer {
	return domain.NewMissionOffer(testOfferID, "BOUNTY_HUNTING", capacity, "season-1")
}

func activeContracts(category string, states ...domain.ContractState) []domain.ActiveContractSummary {
	contracts := make([]domain.ActiveContractSummary, len(states))
	for i, state := range states {
		contracts[i] = domain.NewActiveContractSummary(testOfferID, category, state, "definition-1")
	}
	return contracts
}

func TestActiveMissionLimit_RejectsWhenDefaultLimitReached(t *testing.T) {
	// Arrange
	policy := defaultLimitPolicy(3)
	offer := aBountyOffer(10)
	contracts := activeContracts(
		"BOUNTY_HUNTING",
		domain.ContractAccepted,
		domain.ContractInProgress,
		domain.ContractPendingValidation,
	)

	// Act
	decision := domain.DecideAcceptance(offer, contracts, policy)

	// Assert
	if decision.Accepted {
		t.Fatalf("expected rejection when active count %d equals default limit %d", len(contracts), policy.Default())
	}
	if decision.Reason != domain.ReasonActiveMissionLimitReached {
		t.Fatalf("expected reason ACTIVE_MISSION_LIMIT_REACHED, got %q", decision.Reason)
	}
}

func TestActiveMissionLimit_RejectsWhenOfferCapacityExhausted(t *testing.T) {
	// Arrange
	policy := defaultLimitPolicy(3)
	offer := aBountyOffer(1)
	contracts := activeContracts("BOUNTY_HUNTING", domain.ContractAccepted)

	// Act
	decision := domain.DecideAcceptance(offer, contracts, policy)

	// Assert
	if decision.Accepted {
		t.Fatalf("expected rejection when offer capacity %d is exhausted", offer.Capacity)
	}
	if decision.Reason != domain.ReasonOfferCapacityExhausted {
		t.Fatalf("expected reason OFFER_CAPACITY_EXHAUSTED, got %q", decision.Reason)
	}
}

func TestActiveLimitPolicy_DefaultIsZero(t *testing.T) {
	policy := domain.NewActiveLimitPolicy(0, nil)
	if policy.Default() != 0 {
		t.Fatalf("expected default 0, got %d", policy.Default())
	}
}

func TestActiveLimitPolicy_EffectiveLimitUsesExplicit(t *testing.T) {
	policy := domain.NewActiveLimitPolicy(5, map[string]int{"BOUNTY_HUNTING": 2})
	if got := policy.EffectiveLimit("BOUNTY_HUNTING"); got != 2 {
		t.Fatalf("expected effective limit 2, got %d", got)
	}
}

func TestActiveLimitPolicy_EffectiveLimitFallsBackToDefault(t *testing.T) {
	policy := domain.NewActiveLimitPolicy(5, map[string]int{"BOUNTY_HUNTING": 2})
	if got := policy.EffectiveLimit("UNKNOWN"); got != 5 {
		t.Fatalf("expected effective limit 5, got %d", got)
	}
}
