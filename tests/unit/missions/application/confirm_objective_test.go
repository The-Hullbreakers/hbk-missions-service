package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type fixedContractStore struct {
	contract *domain.MissionContract
}

func (s fixedContractStore) Get(string) (*domain.MissionContract, error) {
	return s.contract, nil
}

type fixedContractWriter struct{}

func (fixedContractWriter) Save(*domain.MissionContract) error { return nil }

type fixedSeasonStore struct {
	season domain.Season
}

func (s fixedSeasonStore) Get(string) (domain.Season, error) {
	return s.season, nil
}

type fixedObjectiveStore struct {
	objective domain.Objective
}

func (s fixedObjectiveStore) Get(string) (domain.Objective, error) {
	return s.objective, nil
}

type recordingLedger struct {
	ledger *domain.MeritLedger
}

func (r *recordingLedger) Save(credit domain.MeritCredit) error {
	r.ledger.Record(credit)
	return nil
}

type failingContractStore struct{}

func (s failingContractStore) Get(string) (*domain.MissionContract, error) {
	return nil, errors.New("contract not found")
}

type failingSeasonStore struct{}

func (s failingSeasonStore) Get(string) (domain.Season, error) {
	return domain.Season{}, errors.New("season not found")
}

type failingObjectiveStore struct{}

func (s failingObjectiveStore) Get(string) (domain.Objective, error) {
	return domain.Objective{}, errors.New("objective not found")
}

type failingLedger struct{}

func (l failingLedger) Save(domain.MeritCredit) error {
	return errors.New("ledger unavailable")
}

func TestConfirmObjectiveHandler_ReturnsContractStoreError(t *testing.T) {
	handler := application.NewConfirmObjectiveHandler(failingContractStore{}, fixedContractWriter{}, fixedSeasonStore{}, fixedObjectiveStore{}, &recordingLedger{})
	_, err := handler.Handle("tenant-1", "contract-1", "objective-1", "target-1", "season-1")
	if err == nil {
		t.Fatalf("expected contract store error")
	}
}

func TestConfirmObjectiveHandler_ReturnsSeasonStoreError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()

	handler := application.NewConfirmObjectiveHandler(fixedContractStore{contract: contract}, fixedContractWriter{}, failingSeasonStore{}, fixedObjectiveStore{}, &recordingLedger{})
	_, err := handler.Handle("tenant-1", "contract-1", "objective-1", "target-1", "season-1")
	if err == nil {
		t.Fatalf("expected season store error")
	}
}

func TestConfirmObjectiveHandler_ReturnsObjectiveStoreError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	handler := application.NewConfirmObjectiveHandler(fixedContractStore{contract: contract}, fixedContractWriter{}, fixedSeasonStore{season: season}, failingObjectiveStore{}, &recordingLedger{})
	_, err := handler.Handle("tenant-1", "contract-1", "objective-1", "target-1", "season-1")
	if err == nil {
		t.Fatalf("expected objective store error")
	}
}

func TestConfirmObjectiveHandler_ReturnsContractWriterError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	handler := application.NewConfirmObjectiveHandler(fixedContractStore{contract: contract}, failingContractWriter{}, fixedSeasonStore{season: season}, fixedObjectiveStore{objective: objective}, &recordingLedger{})
	_, err := handler.Handle("tenant-1", "contract-1", "objective-1", "target-1", "season-1")
	if err == nil {
		t.Fatalf("expected contract writer error")
	}
}

func TestConfirmObjectiveHandler_ReturnsLedgerError(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	handler := application.NewConfirmObjectiveHandler(fixedContractStore{contract: contract}, fixedContractWriter{}, fixedSeasonStore{season: season}, fixedObjectiveStore{objective: objective}, failingLedger{})
	_, err := handler.Handle("tenant-1", "contract-1", "objective-1", "target-1", "season-1")
	if err == nil {
		t.Fatalf("expected ledger error")
	}
}
