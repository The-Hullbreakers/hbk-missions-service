package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// REQ-MCF-001 / SCN-MCF-TEAM-SOLO-REJECT: a team_id on a SOLO offer is rejected.
func TestDecideTeamParticipation_RejectsTeamOnSoloOffer(t *testing.T) {
	err := domain.DecideTeamParticipation(domain.MissionModeSolo, "team-1")
	require.ErrorIs(t, err, domain.ErrTeamNotAllowed)
}

func TestDecideTeamParticipation_AllowsTeamOnTeamAndFlex(t *testing.T) {
	for _, mode := range []domain.MissionMode{domain.MissionModeTeam, domain.MissionModeFlex} {
		err := domain.DecideTeamParticipation(mode, "team-1")
		require.NoError(t, err)
	}
}

func TestDecideTeamParticipation_TeamRequiresTeamID(t *testing.T) {
	err := domain.DecideTeamParticipation(domain.MissionModeTeam, "")
	require.ErrorIs(t, err, domain.ErrTeamIDRequired)
}

func TestDecideTeamParticipation_FlexMayPlaySolo(t *testing.T) {
	err := domain.DecideTeamParticipation(domain.MissionModeFlex, "")
	require.NoError(t, err)
}

// Roster VO: requester plus distinct teammates.
func TestNewRoster_JoinsRequesterAndTeammates(t *testing.T) {
	roster, err := domain.NewRoster("player-1", []string{"player-2", "player-3"})
	require.NoError(t, err)
	members := roster.Members()
	require.False(t, len(members) != 3 || members[0] != "player-1")
}

func TestNewRoster_RejectsEmptyRequester(t *testing.T) {
	_, err := domain.NewRoster("", []string{"player-2"})
	require.ErrorIs(t, err, domain.ErrRosterEmpty)
}

func TestNewRoster_RejectsDuplicateMember(t *testing.T) {
	for _, teammates := range [][]string{
		{"player-1", "player-2"}, // duplicate of requester
		{"player-2", "player-2"}, // duplicate teammate
	} {
		_, err := domain.NewRoster("player-1", teammates)
		require.ErrorIs(t, err, domain.ErrRosterDuplicate)
	}
}

// SCN-MCF-TEAM-ROSTER: a teammate already at the limit rejects the acceptance.
func TestDecideRosterAcceptance_RejectsWhenTeammateAtLimit(t *testing.T) {
	policy := defaultLimitPolicy(2)
	offer := aBountyOffer()
	memberActive := map[string][]domain.ActiveContractSummary{
		"player-1": {},
		"player-2": {
			domain.NewActiveContractSummary("offer-a", "BOUNTY_HUNTING", domain.ContractInProgress, "def-1"),
			domain.NewActiveContractSummary("offer-b", "TRADE", domain.ContractAccepted, "def-2"),
		},
	}

	decision := domain.DecideRosterAcceptance(offer, domain.SeasonStateActive, memberActive, policy)
	require.False(t, decision.Accepted, "expected rejection: player-2 is at the active limit")
	require.Equal(t, domain.ReasonTeammateLimitReached, decision.Reason)
}

// SCN-MCF-TEAM-CONTRACT: every member under the limit accepts; each consumes a slot.
func TestDecideRosterAcceptance_AcceptsWhenEveryMemberFits(t *testing.T) {
	policy := defaultLimitPolicy(2)
	offer := aBountyOffer()
	memberActive := map[string][]domain.ActiveContractSummary{
		"player-1": {domain.NewActiveContractSummary("offer-a", "BOUNTY_HUNTING", domain.ContractInProgress, "def-1")},
		"player-2": {},
		"player-3": {},
	}

	decision := domain.DecideRosterAcceptance(offer, domain.SeasonStateActive, memberActive, policy)
	require.True(t, decision.Accepted, "expected acceptance, got %q", decision.Reason)
}

func TestDecideRosterAcceptance_RejectsWhenSeasonNotActive(t *testing.T) {
	policy := defaultLimitPolicy(2)
	offer := aBountyOffer()

	decision := domain.DecideRosterAcceptance(offer, domain.SeasonStateFinalized, nil, policy)
	require.False(t, decision.Accepted || decision.Reason != domain.ReasonSeasonNotActive)
}
