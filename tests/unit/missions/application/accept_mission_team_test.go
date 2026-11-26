package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// membershipStore answers whether a player belongs to a team.
type membershipStore struct {
	members map[string]map[string]bool
}

func (s membershipStore) IsMember(teamID, playerID string) (bool, error) {
	return s.members[teamID][playerID], nil
}

// rosterActiveStore serves per-player active contracts.
type rosterActiveStore struct {
	byPlayer map[string][]domain.MissionContract
}

func (s rosterActiveStore) ListActiveByPlayer(playerID string) ([]domain.MissionContract, error) {
	return s.byPlayer[playerID], nil
}

func teamAcceptFixture(t *testing.T, offer domain.MissionOffer, active rosterActiveStore, members map[string]map[string]bool) (*application.AcceptMissionHandler, *recordingContractWriter) {
	t.Helper()
	policy := domain.NewActiveLimitPolicy(2, nil)
	writer := &recordingContractWriter{}
	handler := application.NewAcceptMissionHandler(
		policy,
		fixedOfferStore{offer: offer},
		active,
		historyContracts{},
		versionedDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		membershipStore{members: members},
	)
	return handler, writer
}

// SCN-MCF-TEAM-SOLO-REJECT: a team_id on a SOLO offer is rejected.
func TestAcceptMissionHandler_RejectsTeamOnSoloOffer(t *testing.T) {
	handler, writer := teamAcceptFixture(t,
		domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1"),
		rosterActiveStore{},
		nil,
	)

	_, err := handler.Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", nil)
	require.ErrorIs(t, err, domain.ErrTeamNotAllowed)
	require.Len(t, writer.saved, 0)
}

// SCN-MCF-TEAM-ROSTER: a declared teammate who is not in the team rejects the acceptance.
func TestAcceptMissionHandler_RejectsTeammateNotMember(t *testing.T) {
	handler, writer := teamAcceptFixture(t,
		domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1"),
		rosterActiveStore{},
		map[string]map[string]bool{"team-1": {"player-1": true, "player-2": true}},
	)

	_, err := handler.Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", []string{"player-2", "player-9"})
	require.ErrorIs(t, err, domain.ErrTeammateNotMember)
	require.Len(t, writer.saved, 0)
}

// SCN-MCF-TEAM-ROSTER: a teammate already at their active limit rejects the acceptance.
func TestAcceptMissionHandler_RejectsTeammateAtLimit(t *testing.T) {
	atLimit := domain.MissionContract{State: domain.ContractInProgress}
	atLimit2 := domain.MissionContract{State: domain.ContractAccepted}
	handler, writer := teamAcceptFixture(t,
		domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1"),
		rosterActiveStore{byPlayer: map[string][]domain.MissionContract{"player-2": {atLimit, atLimit2}}},
		map[string]map[string]bool{"team-1": {"player-1": true, "player-2": true}},
	)

	_, err := handler.Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", []string{"player-2"})
	require.ErrorIs(t, err, domain.ErrTeammateLimitReached)
	require.Len(t, writer.saved, 0)
}

// REG-0053-02: a teammate already bound to the same definition rejects the
// acceptance with MISSION_DEFINITION_ALREADY_ACTIVE, not a limit error.
func TestAcceptMissionHandler_RejectsTeammateWithSameDefinitionActive(t *testing.T) {
	// Arrange: team offer of def-x; teammate player-2 already holds def-x.
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	offer.DefinitionVersion = "legacy"
	bound := domain.NewMissionContractFromState("contract-x", offer, []string{"player-2"}, domain.ContractInProgress, true, "", "season-1", nil, nil, nil)
	handler, writer := teamAcceptFixture(t,
		offer,
		rosterActiveStore{byPlayer: map[string][]domain.MissionContract{"player-2": {*bound}}},
		map[string]map[string]bool{"team-1": {"player-1": true, "player-2": true}},
	)

	// Act
	_, err := handler.Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", []string{"player-2"})

	// Assert
	require.ErrorIs(t, err, domain.ErrMissionDefinitionAlreadyActive)
	require.Len(t, writer.saved, 0)
}

// REG-0053-02: a teammate's PENDING_VALIDATION contract still blocks the same
// definition even while their active count has spare slots.
func TestAcceptMissionHandler_RejectsTeammatePendingValidation(t *testing.T) {
	// Arrange: teammate player-2 holds def-x awaiting review; spare limit.
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	offer.DefinitionVersion = "legacy"
	pending := domain.NewMissionContractFromState("contract-x", offer, []string{"player-2"}, domain.ContractPendingValidation, true, "", "season-1", nil, nil, nil)
	handler, writer := teamAcceptFixture(t,
		offer,
		rosterActiveStore{byPlayer: map[string][]domain.MissionContract{"player-2": {*pending}}},
		map[string]map[string]bool{"team-1": {"player-1": true, "player-2": true}},
	)

	// Act
	_, err := handler.Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", []string{"player-2"})

	// Assert
	require.ErrorIs(t, err, domain.ErrMissionDefinitionAlreadyActive)
	require.Len(t, writer.saved, 0)
}

// SCN-MCF-TEAM-CONTRACT: a valid team acceptance creates one contract whose
// roster holds the requester plus the validated teammates.
func TestAcceptMissionHandler_TeamContractRosterCoversEveryMember(t *testing.T) {
	handler, writer := teamAcceptFixture(t,
		domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1"),
		rosterActiveStore{},
		map[string]map[string]bool{"team-1": {"player-1": true, "player-2": true, "player-3": true}},
	)

	contract, err := handler.Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", []string{"player-2", "player-3"})
	require.NoError(t, err)
	roster := contract.Roster()
	require.False(t, len(roster) != 3 || roster[0] != "player-1")
	require.Equal(t, "team-1", contract.TeamID)
	require.Len(t, writer.saved, 1)
}

// Teammates declared without a team_id have no team to validate against.
func TestAcceptMissionHandler_RejectsTeammatesWithoutTeamID(t *testing.T) {
	handler, _ := teamAcceptFixture(t,
		domain.NewMissionOfferWithMode("offer-1", "BOUNTY_HUNTING", domain.MissionModeFlex, "season-1"),
		rosterActiveStore{},
		nil,
	)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", []string{"player-2"})
	require.ErrorIs(t, err, domain.ErrTeamIDRequired)
}
