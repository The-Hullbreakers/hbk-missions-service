package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestObjective_BonusIsAddedToCredit(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	objective := domain.NewObjectiveWithBonus("objective-1", "BOUNTY_HUNTING", 100, 25)

	// Act
	credit, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if credit.Amount != 125 {
		t.Fatalf("expected credit amount 125, got %d", credit.Amount)
	}
}

func TestObjective_UnitLimitIsEnforced(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	objective := domain.NewObjectiveWithBonusAndLimit("objective-1", "BOUNTY_HUNTING", 100, 0, 2)

	// Act
	_, errFirst := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")
	_, errSecond := domain.ConfirmObjective(contract, objective, "target-2", season, "tenant-1")
	_, errThird := domain.ConfirmObjective(contract, objective, "target-3", season, "tenant-1")

	// Assert
	if errFirst != nil {
		t.Fatalf("expected first confirmation to succeed, got %v", errFirst)
	}
	if errSecond != nil {
		t.Fatalf("expected second confirmation to succeed, got %v", errSecond)
	}
	if errThird == nil {
		t.Fatalf("expected third confirmation to exceed unit limit")
	}
	if !errors.Is(errThird, domain.ErrObjectiveLimitReached) {
		t.Fatalf("expected ErrObjectiveLimitReached, got %v", errThird)
	}
}
