package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestMeritLedger_TotalForBucket(t *testing.T) {
	// Arrange
	ledger := domain.NewMeritLedger()
	ledger.Record(domain.MeritCredit{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 100})
	ledger.Record(domain.MeritCredit{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 100})
	ledger.Record(domain.MeritCredit{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Amount: 50})
	ledger.Record(domain.MeritCredit{SeasonID: "season-1", Category: "EXPLORATION", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 30})

	// Act
	total := ledger.TotalFor("season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1")

	// Assert
	if total != 200 {
		t.Fatalf("expected total 200 for player-1 in BOUNTY_HUNTING, got %d", total)
	}
}

func TestStanding_ProjectsTotalFromLedger(t *testing.T) {
	// Arrange
	ledger := domain.NewMeritLedger()
	ledger.Record(domain.MeritCredit{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Amount: 150})

	// Act
	standing := domain.ProjectStanding(ledger, "season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1")

	// Assert
	if standing.Total != 150 {
		t.Fatalf("expected standing total 150, got %d", standing.Total)
	}
}
