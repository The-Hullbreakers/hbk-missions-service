package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestMissionContract_Confirm_CreatesCredit(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	startContract(t, contract)
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	freezeTerms(t, contract, objective)

	credit, err := contract.Confirm(objective.ObjectiveID, "target-1", season)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if credit.ConfirmationID != "contract-1~objective-1~target-1~player-1" {
		t.Fatalf("unexpected confirmation id: %s", credit.ConfirmationID)
	}
}

func TestMissionContract_Confirm_RejectsDuplicate(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	startContract(t, contract)
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	freezeTerms(t, contract, objective)

	_, _ = contract.Confirm(objective.ObjectiveID, "target-1", season)
	_, err := contract.Confirm(objective.ObjectiveID, "target-1", season)
	if !errors.Is(err, domain.ErrDuplicateConfirmation) {
		t.Fatalf("expected ErrDuplicateConfirmation, got %v", err)
	}
}

func TestMissionContract_Confirm_RejectsSeasonMismatch(t *testing.T) {
	season := domain.NewSeason("season-2", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	startContract(t, contract)
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	freezeTerms(t, contract, objective)

	_, err := contract.Confirm(objective.ObjectiveID, "target-1", season)
	if !errors.Is(err, domain.ErrSeasonMismatch) {
		t.Fatalf("expected ErrSeasonMismatch, got %v", err)
	}
}

func TestMissionContract_Confirm_Nil(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)

	var contract *domain.MissionContract
	_, err := contract.Confirm(objective.ObjectiveID, "target-1", season)
	if !errors.Is(err, domain.ErrContractNil) {
		t.Fatalf("expected ErrContractNil, got %v", err)
	}
}
