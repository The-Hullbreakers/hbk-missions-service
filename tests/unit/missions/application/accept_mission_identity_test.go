package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// testIdentity builds a member CallerIdentity for the given player.
func testIdentity(t *testing.T, playerID string, teamIDs []string) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity(playerID, []string{"member"}, teamIDs)
	require.NoError(t, err)
	return identity
}

func TestAcceptMissionHandler_UsesCallerIdentityPlayer(t *testing.T) {
	// Arrange
	identity := testIdentity(t, "player-1", nil)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	writer := &recordingContractWriter{}
	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, historyContracts{}, versionedDefinitions{}, fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, writer, nil)

	// Act
	contract, err := handler.Handle(identity, "offer-1", "", nil)

	// Assert
	require.NoError(t, err)
	roster := contract.Roster()
	require.False(t, len(roster) != 1 || roster[0] != "player-1")
}

func TestAcceptMissionHandler_RejectsTeamOutsideCallerMembership(t *testing.T) {
	// Arrange
	identity := testIdentity(t, "player-1", []string{"team-1"})
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	writer := &recordingContractWriter{}
	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, historyContracts{}, versionedDefinitions{}, fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, writer, nil)

	// Act
	_, err := handler.Handle(identity, "offer-1", "team-9", nil)

	// Assert
	require.ErrorIs(t, err, domain.ErrNotTeamMember)
	require.Len(t, writer.saved, 0)
}

func TestAcceptMissionHandler_AcceptsTeamWithinCallerMembership(t *testing.T) {
	// Arrange
	identity := testIdentity(t, "player-1", []string{"team-1"})
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, historyContracts{}, versionedDefinitions{}, fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)}, &recordingContractWriter{}, nil)

	// Act
	contract, err := handler.Handle(identity, "offer-1", "team-1", nil)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "team-1", contract.TeamID)
}
