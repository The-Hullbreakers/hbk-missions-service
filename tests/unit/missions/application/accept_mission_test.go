package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fixedOfferStore struct {
	offer domain.MissionOffer
}

func (s fixedOfferStore) Get(string) (domain.MissionOffer, error) {
	return s.offer, nil
}

type failingOfferStore struct{}

func (s failingOfferStore) Get(string) (domain.MissionOffer, error) {
	return domain.MissionOffer{}, errors.New("offer not found")
}

type fixedActiveContractsStore struct {
	contracts []domain.MissionContract
}

func (s fixedActiveContractsStore) ListActiveByPlayer(string) ([]domain.MissionContract, error) {
	return s.contracts, nil
}

type recordingContractWriter struct {
	saved []*domain.MissionContract
}

func (w *recordingContractWriter) CommitAcceptance(contract *domain.MissionContract, _ domain.MissionOffer, _ int) error {
	w.saved = append(w.saved, contract)
	return nil
}

func acceptHandler(offer domain.MissionOffer, actives domain.MissionContract, writer *recordingContractWriter, limit int) *application.AcceptMissionHandler {
	active := fixedActiveContractsStore{}
	if actives.ContractID != "" {
		active.contracts = []domain.MissionContract{actives}
	}
	return application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(limit, nil),
		fixedOfferStore{offer: offer},
		active,
		historyContracts{},
		versionedDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)
}

func TestAcceptMissionHandler_AcceptsSoloOffer(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	writer := &recordingContractWriter{}

	contract, err := acceptHandler(offer, domain.MissionContract{}, writer, 5).
		Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)

	require.NoError(t, err)
	require.NotNil(t, contract)
	require.Equal(t, "", contract.TeamID)
	require.Len(t, writer.saved, 1)
}

func TestAcceptMissionHandler_AcceptsTeamOffer(t *testing.T) {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	writer := &recordingContractWriter{}

	contract, err := acceptHandler(offer, domain.MissionContract{}, writer, 5).
		Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "team-1", nil)

	require.NoError(t, err)
	require.Equal(t, "team-1", contract.TeamID)
}

func TestAcceptMissionHandler_RequiresTeamIDForTeamOffer(t *testing.T) {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")

	_, err := acceptHandler(offer, domain.MissionContract{}, &recordingContractWriter{}, 5).
		Handle(testIdentity(t, "player-1", []string{"team-1"}), "offer-1", "", nil)

	require.ErrorIs(t, err, domain.ErrTeamIDRequired)
}

func TestAcceptMissionHandler_RejectsWhenActiveLimitReached(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	activeOffer := domain.NewMissionOffer("offer-2", "BOUNTY_HUNTING", "season-1")
	active := domain.NewMissionContract("contract-2", activeOffer, []string{"player-1"})
	require.NoError(t, active.Start(testIdentity(t, "player-1", nil)))

	_, err := acceptHandler(offer, *active, &recordingContractWriter{}, 1).
		Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)

	require.ErrorIs(t, err, domain.ErrActiveMissionLimitReached)
}
