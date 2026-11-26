package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// invitationStore is a hand-written double satisfying ContractReader/Writer.
type invitationStore struct {
	contract *domain.MissionContract
	saved    *domain.MissionContract
}

func (s *invitationStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }
func (s *invitationStore) Save(c *domain.MissionContract) error        { s.saved = c; return nil }

func invitedContract(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewTeamMissionOffer("off-1", "BOUNTY_HUNTING", "season-1")
	return domain.NewTeamMissionContract("off-1:player-1:1", offer,
		[]string{"player-1", "player-2"}, "team-1")
}

// SCN-MCF-INVITE-CONFIRM — invitee confirms and the change is persisted.
func TestConfirmInvitationHandler_ConfirmsAndSaves(t *testing.T) {
	store := &invitationStore{contract: invitedContract(t)}
	handler := application.NewConfirmInvitationHandler(store, store)
	identity, _ := domain.NewCallerIdentity("player-2", nil, nil)

	err := handler.Handle(identity, "off-1:player-1:1")
	require.NoError(t, err)
	require.NotNil(t, store.saved)
	require.Equal(t, domain.InvitationConfirmed, store.saved.InvitationState("player-2"))
}

// SCN-MCF-INVITE-DECLINE — invitee declines; the slot is released on save.
func TestDeclineInvitationHandler_DropsMemberAndSaves(t *testing.T) {
	store := &invitationStore{contract: invitedContract(t)}
	handler := application.NewDeclineInvitationHandler(store, store)
	identity, _ := domain.NewCallerIdentity("player-2", nil, nil)

	err := handler.Handle(identity, "off-1:player-1:1")
	require.NoError(t, err)
	require.False(t, store.saved.IncludesPlayer("player-2"))
}

// An outsider is rejected with ErrNotInvited and nothing is persisted.
func TestConfirmInvitationHandler_OutsiderRejected(t *testing.T) {
	store := &invitationStore{contract: invitedContract(t)}
	handler := application.NewConfirmInvitationHandler(store, store)
	identity, _ := domain.NewCallerIdentity("player-9", nil, nil)

	err := handler.Handle(identity, "off-1:player-1:1")
	require.Equal(t, domain.ErrNotInvited, err)
	require.Nil(t, store.saved)
}
