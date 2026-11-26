package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type failingActiveContractsStore struct{}

func (failingActiveContractsStore) ListActiveByPlayer(string) ([]domain.MissionContract, error) {
	return nil, errors.New("active contracts unavailable")
}

type failingContractWriter struct{}

func (failingContractWriter) CommitAcceptance(*domain.MissionContract, domain.MissionOffer, int) error {
	return errors.New("contract writer unavailable")
}

func acceptEdgeHandler(offers application.OfferReader, actives application.ActiveContractsReader, writer application.AtomicAcceptanceWriter) *application.AcceptMissionHandler {
	return application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(5, nil),
		offers,
		actives,
		historyContracts{},
		versionedDefinitions{},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)
}

func TestAcceptMissionHandler_ReturnsOfferStoreError(t *testing.T) {
	handler := acceptEdgeHandler(failingOfferStore{}, fixedActiveContractsStore{}, &recordingContractWriter{})
	_, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)
	require.Error(t, err)
}

func TestAcceptMissionHandler_ReturnsActiveContractsError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	handler := acceptEdgeHandler(fixedOfferStore{offer: offer}, failingActiveContractsStore{}, &recordingContractWriter{})
	_, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)
	require.Error(t, err)
}

func TestAcceptMissionHandler_ReturnsContractWriterError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	handler := acceptEdgeHandler(fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, failingContractWriter{})
	_, err := handler.Handle(testIdentity(t, "player-1", nil), "offer-1", "", nil)
	require.Error(t, err)
}
