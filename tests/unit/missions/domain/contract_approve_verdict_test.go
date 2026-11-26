package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// limitedContract builds an IN_PROGRESS contract whose objective requires
// unitLimit confirmed units, with the completion reward set.
func limitedContract(t *testing.T, contractID string, reward, unitLimit int) *domain.MissionContract {
	t.Helper()
	definition, err := domain.NewMissionDefinition(
		"bounty-run", "Cacería", "Marca objetivos",
		"BOUNTY", domain.MissionModeSolo, reward,
		[]domain.Objective{domain.NewObjectiveNamed("obj-a", "Objetivo", "", "BOUNTY", 54, 0, unitLimit)}, nil,
	)
	require.NoError(t, err)
	definition.ContentEn = domain.LocalizedContent{
		Title: "Hunt", Description: "Tag targets",
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-a", Name: "Objective"},
		},
	}
	_ = definition.Publish("v1")
	offer, err := domain.NewMissionOfferFromDefinition("off-1", "season-1", definition)
	require.NoError(t, err)
	contract := domain.NewMissionContract(contractID, offer, []string{"player-1"})
	freezeTerms(t, contract, domain.NewObjectiveNamed("obj-a", "Objetivo", "", "BOUNTY", 54, 0, unitLimit))
	startContract(t, contract)
	return contract
}

func confirmUnits(t *testing.T, contract *domain.MissionContract, units ...string) {
	t.Helper()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	for _, unit := range units {
		_, err := contract.Confirm("obj-a", unit, season)
		require.NoError(t, err)
	}
}

func TestApproveVerdict_DropsRejectedAndCreditsApprovedSubset(t *testing.T) {
	contract := limitedContract(t, "c-1", 0, 3)
	confirmUnits(t, contract, "u1", "u2", "u3")
	err := contract.Submit(member(t))
	require.NoError(t, err)

	credits, err := contract.ApproveWithVerdict([]domain.UnitRef{
		{ObjectiveID: "obj-a", UnitID: "u2"},
	})

	require.NoError(t, err)
	require.Equal(t, domain.ContractCompleted, contract.State)
	require.Len(t, credits, 2)
	for _, credit := range credits {
		require.NotEqual(t, "u2", credit.UnitID)
	}
	require.Equal(t, []string{"u2"}, rejectedUnitIDs(t, contract.RejectedUnits(), "obj-a"))
}

func TestApproveVerdict_UnknownUnitFailsAndWritesNothing(t *testing.T) {
	contract := limitedContract(t, "c-2", 0, 3)
	confirmUnits(t, contract, "u1")
	err := contract.Submit(member(t))
	require.NoError(t, err)

	credits, err := contract.ApproveWithVerdict([]domain.UnitRef{
		{ObjectiveID: "obj-a", UnitID: "ghost"},
	})

	require.ErrorIs(t, err, domain.ErrUnitNotStaged)
	require.Empty(t, credits)
	require.Equal(t, domain.ContractPendingValidation, contract.State)
	require.Len(t, contract.PendingCredits(), 1)
}

func TestApproveVerdict_RejectedUnitBlocksCompletionBonus(t *testing.T) {
	contract := limitedContract(t, "c-3", 60, 3)
	confirmUnits(t, contract, "u1", "u2", "u3")
	err := contract.Submit(member(t))
	require.NoError(t, err)

	credits, err := contract.ApproveWithVerdict([]domain.UnitRef{
		{ObjectiveID: "obj-a", UnitID: "u3"},
	})

	require.NoError(t, err)
	_, found := completionCredit(credits)
	require.False(t, found, "rejected unit must block the completion bonus")
}

func TestApproveVerdict_FullSatisfactionGrantsBonus(t *testing.T) {
	contract := limitedContract(t, "c-4", 60, 3)
	confirmUnits(t, contract, "u1", "u2", "u3")
	err := contract.Submit(member(t))
	require.NoError(t, err)

	credits, err := contract.ApproveWithVerdict(nil)

	require.NoError(t, err)
	completion, found := completionCredit(credits)
	require.True(t, found)
	require.Equal(t, 60, completion.Amount)
}

func TestApproveVerdict_PartialStagingGrantsNoBonus(t *testing.T) {
	contract := limitedContract(t, "c-5", 60, 3)
	confirmUnits(t, contract, "u1", "u2")
	err := contract.Submit(member(t))
	require.NoError(t, err)

	credits, err := contract.ApproveWithVerdict(nil)

	require.NoError(t, err)
	require.Len(t, credits, 2)
	_, found := completionCredit(credits)
	require.False(t, found)
}

func TestApproveVerdict_DuplicateRejectsCollapse(t *testing.T) {
	contract := limitedContract(t, "c-6", 0, 3)
	confirmUnits(t, contract, "u1", "u2")
	err := contract.Submit(member(t))
	require.NoError(t, err)

	credits, err := contract.ApproveWithVerdict([]domain.UnitRef{
		{ObjectiveID: "obj-a", UnitID: "u1"},
		{ObjectiveID: "obj-a", UnitID: "u1"},
	})

	require.NoError(t, err)
	require.Len(t, credits, 1)
	require.Equal(t, "u2", credits[0].UnitID)
}

// rejectedUnitIDs extracts the unit ids of one objective from confirmation ids.
func rejectedUnitIDs(t *testing.T, confirmationIDs []string, objectiveID string) []string {
	t.Helper()
	units := make([]string, 0, len(confirmationIDs))
	for _, raw := range confirmationIDs {
		id, err := domain.ParseConfirmationID(raw)
		require.NoError(t, err)
		if id.ObjectiveID() == objectiveID {
			units = append(units, id.UnitID())
		}
	}
	return units
}
