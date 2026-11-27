package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// rewardedContract builds an IN_PROGRESS contract on a published definition
// with the given completion reward and one frozen objective.
func rewardedContract(t *testing.T, contractID string, reward int) *domain.MissionContract {
	t.Helper()
	definition, err := domain.NewMissionDefinition(
		"rescue-siege", "Rescate en el asedio", "Extrae al piloto",
		"BOUNTY", domain.MissionModeSolo, reward,
		[]domain.Objective{domain.NewObjectiveNamed("obj-a", "Llega al piloto", "", "BOUNTY", 50, 0, 0)}, nil,
	)
	require.NoError(t, err)
	definition.ContentEn = domain.LocalizedContent{
		Title: "Rescue", Description: "Extract the pilot",
		Objectives: []domain.LocalizedObjective{
			{ObjectiveID: "obj-a", Name: "Reach the pilot"},
		},
	}
	_ = definition.Publish("v1")
	offer, err := domain.NewMissionOfferFromDefinition("off-1", "season-1", definition)
	require.NoError(t, err)
	contract := domain.NewMissionContract(contractID, offer, []string{"player-1"})
	freezeTerms(t, contract, domain.NewObjectiveNamed("obj-a", "Llega al piloto", "", "BOUNTY", 50, 0, 0))
	startContract(t, contract)
	return contract
}

func confirmObjective(t *testing.T, contract *domain.MissionContract, objectiveID string) {
	t.Helper()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	media, err := domain.NewMediaURI("evidence/"+contract.ContractID+"/unit-1", contract.ContractID)
	require.NoError(t, err)
	_, err = contract.ConfirmWithMedia(objectiveID, "unit-1", media, season)
	require.NoError(t, err)
}

func completionCredit(credits []domain.MeritCredit) (domain.MeritCredit, bool) {
	for _, credit := range credits {
		if credit.ObjectiveID == domain.CompletionObjectiveID {
			return credit, true
		}
	}
	return domain.MeritCredit{}, false
}

func approveContract(t *testing.T, contract *domain.MissionContract) []domain.MeritCredit {
	t.Helper()
	err := contract.Submit(member(t))
	require.NoError(t, err)
	credits, err := contract.Approve()
	require.NoError(t, err)
	return credits
}

func TestContractApprove_EmitsCompletionCreditWhenAllConfirmed(t *testing.T) {
	contract := rewardedContract(t, "c-1", 200)
	confirmObjective(t, contract, "obj-a")

	credits := approveContract(t, contract)

	require.Equal(t, domain.ContractCompleted, contract.State)
	credit, found := completionCredit(credits)
	require.True(t, found, "expected COMPLETION credit, got %+v", credits)
	require.Equal(t, 200, credit.Amount)
	require.False(t, credit.SeasonID != "season-1" || credit.Category != "BOUNTY")
	require.False(t, credit.RecipientID != "player-1" || credit.Track != domain.TrackIndividual)
	require.False(t, credit.ContractID != "c-1" || credit.ConfirmationID == "")
}

func TestContractApprove_TeamContractCreditsTeamTrack(t *testing.T) {
	contract := rewardedContract(t, "c-2", 150)
	contract.TeamID = "team-9"
	confirmObjective(t, contract, "obj-a")

	credit, found := completionCredit(approveContract(t, contract))

	require.True(t, found, "expected COMPLETION credit")
	require.False(t, credit.RecipientID != "team-9" || credit.Track != domain.TrackTeam)
}

func TestContractApprove_ZeroRewardRecordsNoCredit(t *testing.T) {
	contract := rewardedContract(t, "c-3", 0)
	confirmObjective(t, contract, "obj-a")

	credits := approveContract(t, contract)

	require.Equal(t, domain.ContractCompleted, contract.State)
	_, found := completionCredit(credits)
	require.False(t, found)
}

func TestContractApprove_PartialFulfillmentSkipsCompletionCredit(t *testing.T) {
	contract := rewardedContract(t, "c-4", 200)
	extra := domain.NewObjectiveNamed("obj-b", "Segundo objetivo", "", "BOUNTY", 20, 0, 0)
	terms, err := domain.NewContractTerms(append(contract.Terms.Objectives(), extra))
	require.NoError(t, err)
	contract.Terms = terms
	confirmObjective(t, contract, "obj-a")

	credits := approveContract(t, contract)

	require.Equal(t, domain.ContractCompleted, contract.State)
	_, found := completionCredit(credits)
	require.False(t, found)
}

func TestContractApprove_TerminalContractRejects(t *testing.T) {
	contract := rewardedContract(t, "c-5", 200)
	confirmObjective(t, contract, "obj-a")
	approveContract(t, contract)

	_, err := contract.Approve()

	require.ErrorIs(t, err, domain.ErrContractTerminal)
}
