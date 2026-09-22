package application_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmObjectiveHandler_ConfirmsAndCredits(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()

	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)
	domainLedger := domain.NewMeritLedger()

	contracts := fixedContractStore{contract: contract}
	seasons := fixedSeasonStore{season: season}
	objectives := fixedObjectiveStore{objective: objective}
	ledger := &recordingLedger{ledger: domainLedger}

	handler := application.NewConfirmObjectiveHandler(contracts, fixedContractWriter{}, seasons, objectives, ledger)

	// Act
	credit, err := handler.Handle("tenant-1", "contract-1", "objective-1", "target-1", "season-1")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if credit.Amount != 100 {
		t.Fatalf("expected credit amount 100, got %d", credit.Amount)
	}
	if credit.RecipientID != "player-1" {
		t.Fatalf("expected recipient player-1, got %s", credit.RecipientID)
	}
	if domainLedger.TotalFor("tenant-1", "season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1") != 100 {
		t.Fatalf("expected ledger total 100")
	}
}
