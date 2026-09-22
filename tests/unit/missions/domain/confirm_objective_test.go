package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmObjective_DuplicateTargetIsRejected(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()

	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	// Act
	_, errFirst := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")
	_, errSecond := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")

	// Assert
	if errFirst != nil {
		t.Fatalf("expected first confirmation to succeed, got %v", errFirst)
	}
	if errSecond == nil {
		t.Fatalf("expected duplicate confirmation to be rejected")
	}
}

func TestConfirmObjective_RejectsNilContract(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	_, err := domain.ConfirmObjective(nil, objective, "target-1", season, "tenant-1")
	if err == nil {
		t.Fatalf("expected nil contract to be rejected")
	}
}

func TestConfirmObjective_RejectsContractNotInProgress(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	_, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")
	if err == nil {
		t.Fatalf("expected non-started contract to be rejected")
	}
}

func TestConfirmObjective_RejectsFinalizedSeason(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateFinalized)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	// Act
	_, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")

	// Assert
	if err == nil {
		t.Fatalf("expected finalized season to be rejected")
	}
	if !errors.Is(err, domain.ErrSeasonNotActive) {
		t.Fatalf("expected ErrSeasonNotActive, got %v", err)
	}
}

func TestConfirmObjective_RejectsCategoryMismatch(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	objective := domain.NewObjective("objective-1", "EXPLORATION", 100)

	_, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")
	if err == nil {
		t.Fatalf("expected category mismatch to be rejected")
	}
}

func TestConfirmObjective_RejectsTerminalContract(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	contract.Abandon()
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	_, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")
	if err == nil {
		t.Fatalf("expected terminal contract to be rejected")
	}
	if !errors.Is(err, domain.ErrContractTerminal) {
		t.Fatalf("expected ErrContractTerminal, got %v", err)
	}
}

func TestConfirmObjective_RejectsEmptyRoster(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{})
	contract.Start()
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	_, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")
	if err == nil {
		t.Fatalf("expected empty roster to be rejected")
	}
}
