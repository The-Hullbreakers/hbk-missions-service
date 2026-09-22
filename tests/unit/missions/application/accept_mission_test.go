package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fixedOfferStore struct {
	offer domain.MissionOffer
}

func (s fixedOfferStore) Get(string, string) (domain.MissionOffer, error) {
	return s.offer, nil
}

type failingOfferStore struct{}

func (s failingOfferStore) Get(string, string) (domain.MissionOffer, error) {
	return domain.MissionOffer{}, errors.New("offer not found")
}

type fixedActiveContractsStore struct {
	contracts []domain.MissionContract
}

func (s fixedActiveContractsStore) ListActiveByPlayer(string, string, string) ([]domain.MissionContract, error) {
	return s.contracts, nil
}

type recordingContractWriter struct {
	saved []*domain.MissionContract
}

func (w *recordingContractWriter) Save(contract *domain.MissionContract) error {
	w.saved = append(w.saved, contract)
	return nil
}

func TestAcceptMissionHandler_AcceptsSoloOffer(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	writer := &recordingContractWriter{}

	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, writer)
	contract, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if contract == nil {
		t.Fatalf("expected a contract")
	}
	if contract.TeamID != "" {
		t.Fatalf("expected solo contract")
	}
	if len(writer.saved) != 1 {
		t.Fatalf("expected contract to be saved")
	}
}

func TestAcceptMissionHandler_AcceptsTeamOffer(t *testing.T) {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)
	writer := &recordingContractWriter{}

	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, writer)
	contract, err := handler.Handle("tenant-1", "player-1", "offer-1", "team-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if contract.TeamID != "team-1" {
		t.Fatalf("expected team contract for team-1, got %s", contract.TeamID)
	}
}

func TestAcceptMissionHandler_RequiresTeamIDForTeamOffer(t *testing.T) {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	policy := domain.NewActiveLimitPolicy(5, nil)

	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{}, &recordingContractWriter{})
	_, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if err == nil {
		t.Fatalf("expected team id to be required")
	}
	if !errors.Is(err, domain.ErrTeamIDRequired) {
		t.Fatalf("expected ErrTeamIDRequired, got %v", err)
	}
}

func TestAcceptMissionHandler_RejectsWhenActiveLimitReached(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	activeOffer := domain.NewMissionOffer("offer-2", "BOUNTY_HUNTING", 10, "season-1")
	active := domain.NewMissionContract("contract-2", activeOffer, []string{"player-1"})
	active.Start()
	policy := domain.NewActiveLimitPolicy(1, nil)

	handler := application.NewAcceptMissionHandler(policy, fixedOfferStore{offer: offer}, fixedActiveContractsStore{contracts: []domain.MissionContract{*active}}, &recordingContractWriter{})
	_, err := handler.Handle("tenant-1", "player-1", "offer-1", "")
	if err == nil {
		t.Fatalf("expected active limit to be reached")
	}
	if !errors.Is(err, domain.ErrActiveMissionLimitReached) {
		t.Fatalf("expected ErrActiveMissionLimitReached, got %v", err)
	}
}
