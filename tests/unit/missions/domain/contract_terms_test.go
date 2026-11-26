package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestNewContractTerms_RejectsEmpty(t *testing.T) {
	_, err := domain.NewContractTerms(nil)

	require.ErrorIs(t, err, domain.ErrContractTermsInvalid)
}

func TestContractTerms_LookupFindsObjective(t *testing.T) {
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 5, 2),
	})
	require.NoError(t, err)

	objective, ok := terms.Lookup("obj-a")

	require.True(t, ok, "expected obj-a to resolve")
	require.False(t, objective.MeritPerUnit != 10 || objective.Bonus != 5 || objective.UnitLimit != 2)
	_, ok = terms.Lookup("obj-z")
	require.False(t, ok)
}

func TestContractConfirm_ScoresFromFrozenTerms(t *testing.T) {
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("off-1:player-1:1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 5, 0),
	})
	contract.Terms = terms
	_ = contract.Start(member(t))
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	credit, err := contract.Confirm("obj-a", "unit-1", season)

	require.NoError(t, err)
	require.Equal(t, 15, credit.Amount)
	require.Equal(t, "obj-a", credit.ObjectiveID)
}

func TestContractConfirm_RejectsObjectiveOutsideTerms(t *testing.T) {
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 0, 0),
	})
	contract.Terms = terms
	_ = contract.Start(member(t))
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	_, err := contract.Confirm("obj-z", "unit-1", season)

	require.ErrorIs(t, err, domain.ErrObjectiveNotInContract)
}

func TestContractConfirm_EnforcesUnitLimitFromTerms(t *testing.T) {
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("c-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 0, 2),
	})
	contract.Terms = terms
	_ = contract.Start(member(t))
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	for _, unit := range []string{"u1", "u2"} {
		_, err := contract.Confirm("obj-a", unit, season)
		require.NoError(t, err)
	}

	_, err := contract.Confirm("obj-a", "u3", season)

	require.ErrorIs(t, err, domain.ErrObjectiveLimitReached)
}
