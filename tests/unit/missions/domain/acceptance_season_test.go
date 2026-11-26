package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// REQ-MB-014: an offer bound to a non-ACTIVE season cannot be accepted.
func TestDecideAcceptance_RejectsWhenBoundSeasonNotActive(t *testing.T) {
	policy := defaultLimitPolicy(3)
	offer := aBountyOffer()

	for _, state := range []domain.SeasonState{
		domain.SeasonStateFreezing,
		domain.SeasonStateFinalized,
		domain.SeasonStateArchived,
	} {
		decision := domain.DecideAcceptance(offer, state, nil, policy)
		require.False(t, decision.Accepted, "expected rejection for season state %s", state)
		require.Equal(t, domain.ReasonSeasonNotActive, decision.Reason)
	}
}

func TestDecideAcceptance_AcceptsWhenBoundSeasonActive(t *testing.T) {
	policy := defaultLimitPolicy(3)
	offer := aBountyOffer()

	decision := domain.DecideAcceptance(offer, domain.SeasonStateActive, nil, policy)
	require.True(t, decision.Accepted, "expected acceptance on ACTIVE season, got %q", decision.Reason)
}

// REQ-MB-003: the effective limit applies to the player's total active set,
// regardless of each contract's category.
func TestDecideAcceptance_CountsTotalActiveAcrossCategories(t *testing.T) {
	policy := defaultLimitPolicy(3)
	offer := aBountyOffer()
	active := []domain.ActiveContractSummary{
		domain.NewActiveContractSummary("offer-a", "BOUNTY_HUNTING", domain.ContractInProgress, "def-1"),
		domain.NewActiveContractSummary("offer-b", "TRADE", domain.ContractAccepted, "def-2"),
		domain.NewActiveContractSummary("offer-c", "EXPLORATION", domain.ContractAccepted, "def-3"),
	}

	decision := domain.DecideAcceptance(offer, domain.SeasonStateActive, active, policy)
	require.False(t, decision.Accepted, "expected rejection: 3 active contracts across categories reach the default limit")
	require.Equal(t, domain.ReasonActiveMissionLimitReached, decision.Reason)
}

// Category override changes the limit threshold, not the scope of the count.
func TestDecideAcceptance_OverrideRaisesThresholdForTotalSet(t *testing.T) {
	policy := domain.NewActiveLimitPolicy(3, map[string]int{"BOUNTY_HUNTING": 5})
	offer := aBountyOffer()
	active := []domain.ActiveContractSummary{
		domain.NewActiveContractSummary("offer-a", "BOUNTY_HUNTING", domain.ContractInProgress, "def-1"),
		domain.NewActiveContractSummary("offer-b", "TRADE", domain.ContractAccepted, "def-2"),
		domain.NewActiveContractSummary("offer-c", "EXPLORATION", domain.ContractAccepted, "def-3"),
	}

	decision := domain.DecideAcceptance(offer, domain.SeasonStateActive, active, policy)
	require.True(t, decision.Accepted, "expected acceptance: BOUNTY_HUNTING override raises limit to 5, got %q", decision.Reason)
}
