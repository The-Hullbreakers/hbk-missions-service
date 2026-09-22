package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func confirmBountyTargets(t *testing.T, contract *domain.MissionContract, objective domain.Objective, season domain.Season, targets []string, tenantID string) []domain.MeritCredit {
	t.Helper()
	var credits []domain.MeritCredit
	for _, target := range targets {
		credit, err := domain.ConfirmObjective(contract, objective, target, season, tenantID)
		if err != nil {
			t.Fatalf("unexpected error confirming %s: %v", target, err)
		}
		credits = append(credits, credit)
	}
	return credits
}

func TestMeritCredit_FourBountyTargets(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()

	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)
	targets := []string{"target-1", "target-2", "target-3", "target-4"}

	// Act
	credits := confirmBountyTargets(t, contract, objective, season, targets, "tenant-1")

	// Assert
	total := 0
	for _, c := range credits {
		total += c.Amount
	}
	if total != 400 {
		t.Fatalf("expected 400 merits, got %d", total)
	}
}

func TestParseTrack_AcceptsIndividualAndTeam(t *testing.T) {
	for _, raw := range []string{"individual", "INDIVIDUAL", "  individual  "} {
		track, err := domain.ParseTrack(raw)
		if err != nil {
			t.Fatalf("expected no error for %q, got %v", raw, err)
		}
		if track != domain.TrackIndividual {
			t.Fatalf("expected INDIVIDUAL, got %s", track)
		}
	}
}

func TestParseTrack_RejectsUnknown(t *testing.T) {
	_, err := domain.ParseTrack("squad")
	if !errors.Is(err, domain.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack, got %v", err)
	}
}
