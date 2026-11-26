package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func newInProgressContract() *domain.MissionContract {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	return contract
}

func TestConfirmObjectiveHandler_RosterMemberConfirms(t *testing.T) {
	// Arrange
	contract := newInProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewConfirmObjectiveHandler(
		fixedContractStore{contract: contract}, &atomicWriter{},
		fixedSeasonStore{season: season})
	identity := testIdentity(t, "player-1", nil)

	// Act
	credit, err := handler.Handle(identity, "contract-1", "objective-1", "target-1", "season-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, 100, credit.Amount)
}

func TestConfirmObjectiveHandler_RejectsOutsider(t *testing.T) {
	// Arrange
	contract := newInProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	writer := &atomicWriter{}
	handler := application.NewConfirmObjectiveHandler(
		fixedContractStore{contract: contract}, writer,
		fixedSeasonStore{season: season})
	identity := testIdentity(t, "player-2", nil)

	// Act
	_, err := handler.Handle(identity, "contract-1", "objective-1", "target-1", "season-1")

	// Assert
	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Equal(t, 0, writer.calls)
}

func TestConfirmObjectiveHandler_PrivilegedCallerConfirms(t *testing.T) {
	// Arrange
	contract := newInProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewConfirmObjectiveHandler(
		fixedContractStore{contract: contract}, &atomicWriter{},
		fixedSeasonStore{season: season})
	identity, err := domain.NewCallerIdentity("ops-1", []string{"admin"}, nil)
	require.NoError(t, err)

	// Act
	credit, err := handler.Handle(identity, "contract-1", "objective-1", "target-1", "season-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, 100, credit.Amount)
}
