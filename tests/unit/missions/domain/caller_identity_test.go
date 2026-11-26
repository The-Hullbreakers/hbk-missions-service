package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestCallerIdentity_RejectsEmptyPlayer(t *testing.T) {
	_, err := domain.NewCallerIdentity("", []string{"member"}, nil)
	require.Error(t, err)
}

func TestCallerIdentity_SystemIdentityNeedsNoTenant(t *testing.T) {
	// Act
	identity := domain.NewSystemIdentity()

	// Assert
	require.False(t, identity.PlayerID() != "system" || !identity.HasRole("system"))
}

func TestCallerIdentity_TeamIDsAreDefensivelyCopied(t *testing.T) {
	// Arrange
	teams := []string{"team-1"}

	// Act
	identity, err := domain.NewCallerIdentity("player-1", []string{"member"}, teams)
	require.NoError(t, err)
	teams[0] = "tampered"

	// Assert
	got := identity.TeamIDs()
	require.False(t, len(got) != 1 || got[0] != "team-1")
}

func TestMissionContract_CanSubmitEvidence_RosterMember(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	caller, _ := domain.NewCallerIdentity("player-1", []string{"member"}, nil)

	// Act
	allowed := contract.CanSubmitEvidence(caller)

	// Assert
	require.True(t, allowed, "expected roster member to submit evidence")
}

func TestMissionContract_CanSubmitEvidence_RejectsOutsider(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	caller, _ := domain.NewCallerIdentity("player-2", []string{"member"}, nil)

	// Act
	allowed := contract.CanSubmitEvidence(caller)

	// Assert
	require.False(t, allowed, "expected outsider to be denied evidence submission")
}
