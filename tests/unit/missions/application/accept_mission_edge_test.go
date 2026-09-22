package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type failingActiveContractsStore struct{}

func (failingActiveContractsStore) ListActiveByPlayer(string, string, string) ([]domain.MissionContract, error) {
	return nil, errors.New("active contracts unavailable")
}

type failingContractWriter struct{}

func (failingContractWriter) Save(*domain.MissionContract) error {
	return errors.New("contract writer unavailable")
}

func TestAcceptMissionHandler_ReturnsOfferStoreError(t *testing.T) {
	policy := domain.NewActiveLimitPolicy(5, nil)
	handler := application.NewAcceptMissionHandler(policy, failingOfferStore{}, fixedActiveContractsStore{}, &recordingContractWriter{})
	_, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if err == nil {
		t.Fatalf("expected offer store error")
	}
}

func TestAcceptMissionHandler_ReturnsActiveContractsError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, failingActiveContractsStore{}, &recordingContractWriter{})
	_, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if err == nil {
		t.Fatalf("expected active contracts error")
	}
}

func TestAcceptMissionHandler_ReturnsContractWriterError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, failingContractWriter{})
	_, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if err == nil {
		t.Fatalf("expected contract writer error")
	}
}

func TestAcceptMissionHandler_RejectsWhenOfferCapacityExhausted(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 1, "season-1")
	active := domain.NewMissionContract("contract-2", offer, []string{"player-1"})
	active.Start()
	policy := domain.NewActiveLimitPolicy(5, nil)

	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{contracts: []domain.MissionContract{*active}}, &recordingContractWriter{})
	_, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if !errors.Is(err, domain.ErrOfferCapacityExhausted) {
		t.Fatalf("expected ErrOfferCapacityExhausted, got %v", err)
	}
}
