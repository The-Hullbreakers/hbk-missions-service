package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// phasedFixture builds a two-phase contract: phase 1 has a mandatory obj-a and
// optional obj-x; phase 2 has mandatory obj-b.
func phasedFixture(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})

	objA := domain.NewObjectiveNamed("obj-a", "mandatory a", "", "BOUNTY", 10, 0, 0)
	objX := domain.NewObjectiveNamed("obj-x", "optional x", "", "BOUNTY", 5, 0, 0)
	objB := domain.NewObjectiveNamed("obj-b", "mandatory b", "", "BOUNTY", 20, 0, 0)
	terms, err := domain.NewContractTermsPhased([]domain.Phase{
		{Mandatory: []domain.Objective{objA}, Optional: []domain.Objective{objX}},
		{Mandatory: []domain.Objective{objB}},
	})
	require.NoError(t, err)
	contract.Terms = terms

	startContract(t, contract)
	return contract
}

func TestConfirm_Phase2ObjectiveRejectedBeforeUnlock(t *testing.T) {
	// Arrange
	contract := phasedFixture(t)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	// Act
	_, err := contract.Confirm("obj-b", "u1", season)

	// Assert
	require.ErrorIs(t, err, domain.ErrPhaseNotUnlocked)
}

func TestConfirm_MandatoryCompletionUnlocksNextPhase(t *testing.T) {
	// Arrange
	contract := phasedFixture(t)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	_, err := contract.Confirm("obj-a", "u1", season)
	require.NoError(t, err)

	// Act
	_, err = contract.Confirm("obj-b", "u1", season)

	// Assert
	require.NoError(t, err)
}

func TestConfirm_OptionalObjectiveDoesNotGateUnlock(t *testing.T) {
	// Arrange — obj-x is optional and never confirmed; obj-a is mandatory.
	contract := phasedFixture(t)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	_, err := contract.Confirm("obj-a", "u1", season)
	require.NoError(t, err)

	// Act — phase 2 must be reachable without obj-x
	_, err = contract.Confirm("obj-b", "u1", season)

	// Assert
	require.NoError(t, err)
}

func TestConfirm_CurrentPhaseObjectivesStayEligible(t *testing.T) {
	// Arrange — the optional objective is still in phase 1 while it is current.
	contract := phasedFixture(t)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	// Act
	_, err := contract.Confirm("obj-x", "u1", season)

	// Assert
	require.NoError(t, err)
}
