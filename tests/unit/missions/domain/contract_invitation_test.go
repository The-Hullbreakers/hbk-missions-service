package domain_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

func newTeamContract(roster []string) *domain.MissionContract {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	return domain.NewTeamMissionContract("offer-1:player-1:1", offer, roster, "team-1")
}

// SCN-MCF-INVITE-CONFIRM — invitees start pending; requester is confirmed.
func TestTeamContract_InviteesStartPendingRequesterConfirmed(t *testing.T) {
	c := newTeamContract([]string{"player-1", "player-2", "player-3"})

	require.False(t, c.RosterFrozen(), "roster must not freeze while invitations are pending")
	require.Equal(t, domain.InvitationConfirmed, c.InvitationState("player-1"))
	for _, p := range []string{"player-2", "player-3"} {
		require.Equal(t, domain.InvitationPending, c.InvitationState(p))
	}
}

// SCN-MCF-INVITE-CONFIRM — all confirm → frozen roster, everyone stays.
func TestTeamContract_AllConfirmFreezesRoster(t *testing.T) {
	c := newTeamContract([]string{"player-1", "player-2"})
	now := time.Now()

	err := c.ConfirmInvitation(contractMember(t, "player-2"), now)
	require.NoError(t, err)
	require.True(t, c.RosterFrozen(), "roster must freeze once every member is resolved")
	require.False(t, !c.IncludesPlayer("player-2"))
}

// SCN-MCF-INVITE-DECLINE — decline drops the member and releases the slot.
func TestTeamContract_DeclineDropsMember(t *testing.T) {
	c := newTeamContract([]string{"player-1", "player-2"})
	now := time.Now()

	err := c.DeclineInvitation(contractMember(t, "player-2"), now)
	require.NoError(t, err)
	require.False(t, c.IncludesPlayer("player-2"))
	require.Equal(t, domain.InvitationDeclined, c.InvitationState("player-2"))
	require.True(t, c.RosterFrozen(), "roster must freeze once every member is resolved")
}

// SCN-MCF-INVITE-EXPIRE — past the 7-day window the pending member expires.
func TestTeamContract_PendingInvitationExpires(t *testing.T) {
	c := newTeamContract([]string{"player-1", "player-2"})
	later := time.Now().Add(domain.InvitationWindow + time.Minute)

	err := c.ResolveExpirations(later)
	require.NoError(t, err)
	require.Equal(t, domain.InvitationExpired, c.InvitationState("player-2"))
	require.False(t, c.IncludesPlayer("player-2"))
}

func TestTeamContract_NonRosterCallerCannotRespond(t *testing.T) {
	c := newTeamContract([]string{"player-1", "player-2"})

	err := c.ConfirmInvitation(contractMember(t, "player-9"), time.Now())
	require.Equal(t, domain.ErrNotInvited, err)
}

func TestTeamContract_AlreadyResolvedCannotRespondAgain(t *testing.T) {
	c := newTeamContract([]string{"player-1", "player-2"})
	now := time.Now()

	err := c.ConfirmInvitation(contractMember(t, "player-2"), now)
	require.NoError(t, err)
	err = c.DeclineInvitation(contractMember(t, "player-2"), now)
	require.Equal(t, domain.ErrInvitationNotPending, err)
}

// Persisted invitation state rehydrates and drives expiry after reload.
func TestTeamContract_RestoredInvitationsBehaveLive(t *testing.T) {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	invitedAt := time.Now().Add(-2 * time.Hour)
	restored := domain.NewMissionContractFromState("c-1", offer,
		[]string{"player-1", "player-2"}, domain.ContractAccepted, false,
		"team-1", "season-1", nil, nil, nil)

	restored.RestoreInvitationState(map[string]domain.InvitationState{
		"player-1": domain.InvitationConfirmed,
		"player-2": domain.InvitationPending,
	}, invitedAt)

	require.Equal(t, invitedAt, restored.InvitedAt())
	require.Equal(t, 2, len(restored.Invitations()))
	require.False(t, restored.RosterFrozen(), "restored pending invitation must keep the roster unfrozen")
	require.Equal(t, domain.InvitationPending, restored.InvitationState("player-2"))
}

// A restore with no states is a no-op for contracts without invitations.
func TestSoloContract_RestoreInvitationStateNoOp(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	c := domain.NewMissionContract("c-1", offer, []string{"player-1"})

	c.RestoreInvitationState(nil, time.Now())

	require.True(t, c.InvitedAt().IsZero(), "solo contract must not gain an invitation timestamp")
	require.True(t, c.RosterFrozen(), "solo roster stays frozen")
}

// Solo contracts carry no invitations and freeze at creation.
func TestSoloContract_NoInvitationsFrozenAtCreation(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	c := domain.NewMissionContract("offer-1:player-1:1", offer, []string{"player-1"})

	require.True(t, c.RosterFrozen(), "solo roster must be frozen at creation")
	err := c.ConfirmInvitation(contractMember(t, "player-1"), time.Now())
	require.Equal(t, domain.ErrInvitationNotPending, err)
}
