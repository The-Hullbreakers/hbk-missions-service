package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeason_FreezeTransitionsFromActiveToFreezing(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	frozen, err := domain.Freeze(season)

	if err != nil {
		t.Fatalf("expected freeze to succeed, got %v", err)
	}
	if frozen.State != domain.SeasonStateFreezing {
		t.Fatalf("expected FREEZING, got %v", frozen.State)
	}
}

func TestSeason_FinalizeTransitionsFromFreezingToFinalized(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)

	finalized, err := domain.Finalize(season)

	if err != nil {
		t.Fatalf("expected finalize to succeed, got %v", err)
	}
	if finalized.State != domain.SeasonStateFinalized {
		t.Fatalf("expected FINALIZED, got %v", finalized.State)
	}
}

func TestSeason_FreezeRejectsNonActive(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFinalized)

	_, err := domain.Freeze(season)

	if err == nil {
		t.Fatalf("expected freeze to be rejected")
	}
	if !errors.Is(err, domain.ErrInvalidSeasonStateTransition) {
		t.Fatalf("expected ErrInvalidSeasonStateTransition, got %v", err)
	}
}

func TestSeason_FinalizeRejectsNonFreezing(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)

	_, err := domain.Finalize(season)

	if err == nil {
		t.Fatalf("expected finalize to be rejected")
	}
	if !errors.Is(err, domain.ErrInvalidSeasonStateTransition) {
		t.Fatalf("expected ErrInvalidSeasonStateTransition, got %v", err)
	}
}

func TestSeason_CanAcceptMeritsDuringFreezing(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()

	accepted := season.CanAcceptMerits(contract)

	if !accepted {
		t.Fatalf("expected FREEZING season to accept in-progress contract")
	}
}

func TestSeason_CanAcceptMeritsRejectsNotStartedDuringFreezing(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFreezing)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})

	accepted := season.CanAcceptMerits(contract)

	if accepted {
		t.Fatalf("expected FREEZING season to reject not-started contract")
	}
}

func TestSeason_CanAcceptMeritsRejectsFinalized(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateFinalized)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()

	accepted := season.CanAcceptMerits(contract)

	if accepted {
		t.Fatalf("expected FINALIZED season to reject any contract")
	}
}
