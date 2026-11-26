package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmObjective_AcceptsFreezingWhenContractInProgress(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	startContract(t, contract)
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	freezeTerms(t, contract, objective)

	credit, err := domain.ConfirmObjective(contract, objective.ObjectiveID, "target-1", season)

	if err != nil {
		t.Fatalf("expected freezing season to accept in-progress contract, got %v", err)
	}
	if credit.SeasonID != "season-1" {
		t.Fatalf("expected credit season to be season-1, got %v", credit.SeasonID)
	}
}

func TestConfirmObjective_RejectsFreezingWhenContractNotStarted(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	freezeTerms(t, contract, objective)

	_, err := domain.ConfirmObjective(contract, objective.ObjectiveID, "target-1", season)

	if err == nil {
		t.Fatalf("expected not-started contract to be rejected")
	}
	if !errors.Is(err, domain.ErrContractNotInProgress) {
		t.Fatalf("expected ErrContractNotInProgress, got %v", err)
	}
}
