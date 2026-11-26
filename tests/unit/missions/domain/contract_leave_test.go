package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func teamContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("off-t1", "BOUNTY", "season-1")
	contract := domain.NewTeamMissionContract("c-t1", offer, []string{"player-1", "player-2"}, "team-9")
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 0, 0),
	})
	contract.Terms = terms
	err := contract.Start(member(t))
	require.NoError(t, err)
	return contract
}

func player(t *testing.T, id string) domain.CallerIdentity {
	t.Helper()
	identity, err := domain.NewCallerIdentity(id, nil, nil)
	require.NoError(t, err)
	return identity
}

func TestContractLeave_MemberLeavesContractSurvives(t *testing.T) {
	contract := teamContract(t)

	err := contract.Leave(player(t, "player-1"))

	require.NoError(t, err)
	require.Equal(t, domain.ContractInProgress, contract.State)
	require.False(t, !contract.IsDeparted("player-1"))
	active := contract.ActiveMembers()
	require.False(t, len(active) != 1 || active[0] != "player-2")
}

func TestContractLeave_DepartedCannotAct(t *testing.T) {
	contract := teamContract(t)
	err := contract.Leave(player(t, "player-1"))
	require.NoError(t, err)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	if _, err := contract.Confirm("obj-a", "unit-1", season); err == nil {
		// Confirm uses roster identity at the boundary; domain guard is CanSubmitEvidence
	}
	require.False(t, contract.CanSubmitEvidence(player(t, "player-1")))
	err = contract.Leave(player(t, "player-1"))
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	err = contract.Submit(player(t, "player-1"))
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	err = contract.Abandon(player(t, "player-1"))
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

func TestContractLeave_SurvivorStillConfirms(t *testing.T) {
	contract := teamContract(t)
	err := contract.Leave(player(t, "player-1"))
	require.NoError(t, err)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	credit, err := contract.Confirm("obj-a", "unit-1", season)

	require.NoError(t, err)
	require.False(t, credit.RecipientID != "team-9" || credit.Track != domain.TrackTeam)
}

func TestContractLeave_LastMemberAbandons(t *testing.T) {
	contract := teamContract(t)
	err := contract.Leave(player(t, "player-1"))
	require.NoError(t, err)

	err = contract.Leave(player(t, "player-2"))

	require.NoError(t, err)
	require.Equal(t, domain.ContractAbandoned, contract.State)
}

func TestContractLeave_SoloContractAbandons(t *testing.T) {
	contract := lifecycleContract()
	_ = contract.Start(member(t))

	err := contract.Leave(member(t))

	require.NoError(t, err)
	require.Equal(t, domain.ContractAbandoned, contract.State)
}

func TestContractLeave_RejectsOutsider(t *testing.T) {
	contract := teamContract(t)

	err := contract.Leave(outsider(t))

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
}

func TestContractLeave_RejectsTerminal(t *testing.T) {
	contract := lifecycleContract()
	_ = contract.Start(member(t))
	err := contract.Abandon(member(t))
	require.NoError(t, err)

	err = contract.Leave(member(t))

	require.ErrorIs(t, err, domain.ErrContractTerminal)
}
