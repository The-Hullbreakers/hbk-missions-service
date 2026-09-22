package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmObjective_AcceptsFreezingWhenContractInProgress(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	credit, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")

	if err != nil {
		t.Fatalf("expected freezing season to accept in-progress contract, got %v", err)
	}
	if credit.SeasonID != "season-1" {
		t.Fatalf("expected credit season to be season-1, got %v", credit.SeasonID)
	}
}

func TestConfirmObjective_RejectsFreezingWhenContractNotStarted(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)

	_, err := domain.ConfirmObjective(contract, objective, "target-1", season, "tenant-1")

	if err == nil {
		t.Fatalf("expected not-started contract to be rejected")
	}
	if !errors.Is(err, domain.ErrContractNotInProgress) {
		t.Fatalf("expected ErrContractNotInProgress, got %v", err)
	}
}
