package infrastructure_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure"
)

func TestMeritLedgerRepository_SavesCredit(t *testing.T) {
	ledger := domain.NewMeritLedger()
	repo := infrastructure.NewMeritLedgerRepository(ledger)

	credit := domain.MeritCredit{
		SeasonID:    "season-1",
		Category:    "BOUNTY_HUNTING",
		Track:       domain.TrackIndividual,
		RecipientID: "player-1",
		Amount:      100,
	}

	err := repo.Save(credit)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ledger.TotalFor("season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1") != 100 {
		t.Fatalf("expected ledger total 100")
	}
}
