package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestMissionOffer_DefaultsToSoloMode(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	if offer.Mode != domain.MissionModeSolo {
		t.Fatalf("expected SOLO mode, got %s", offer.Mode)
	}
}

func TestMissionOffer_TeamMode(t *testing.T) {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	if offer.Mode != domain.MissionModeTeam {
		t.Fatalf("expected TEAM mode, got %s", offer.Mode)
	}
}

func TestMissionOffer_CarriesSeasonID(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-2026-q4")
	if offer.SeasonID != "season-2026-q4" {
		t.Fatalf("expected season season-2026-q4, got %s", offer.SeasonID)
	}
}

func TestNewValidMissionOffer_AcceptsValid(t *testing.T) {
	offer, err := domain.NewValidMissionOffer("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offer.OfferID != "offer-1" {
		t.Fatalf("unexpected offer id: %s", offer.OfferID)
	}
	if offer.SeasonID != "season-1" {
		t.Fatalf("unexpected season id: %s", offer.SeasonID)
	}
}

func TestNewValidMissionOffer_RejectsEmptyOfferID(t *testing.T) {
	_, err := domain.NewValidMissionOffer("", "BOUNTY_HUNTING", domain.MissionModeSolo, "season-1")
	if !errors.Is(err, domain.ErrOfferInvalid) {
		t.Fatalf("expected ErrOfferInvalid, got %v", err)
	}
}

func TestNewValidMissionOffer_RejectsEmptySeasonID(t *testing.T) {
	_, err := domain.NewValidMissionOffer("offer-1", "BOUNTY_HUNTING", domain.MissionModeSolo, "")
	if !errors.Is(err, domain.ErrOfferInvalid) {
		t.Fatalf("expected ErrOfferInvalid, got %v", err)
	}
}

func TestNewValidMissionOffer_RejectsInvalidMode(t *testing.T) {
	_, err := domain.NewValidMissionOffer("offer-1", "BOUNTY_HUNTING", domain.MissionMode("squad"), "season-1")
	if !errors.Is(err, domain.ErrOfferInvalid) {
		t.Fatalf("expected ErrOfferInvalid, got %v", err)
	}
}
