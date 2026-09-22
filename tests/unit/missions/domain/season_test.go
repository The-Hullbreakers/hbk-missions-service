package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeason_NewSeasonDefaultsToOfficial(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	if !season.Official {
		t.Fatalf("expected NewSeason to default Official to true, got %v", season.Official)
	}
}

func TestSeason_NewSeasonWithOfficialFalse(t *testing.T) {
	season := domain.NewSeasonWithOfficial("season-2", domain.SeasonStateActive, false)
	if season.Official {
		t.Fatalf("expected Official to be false, got %v", season.Official)
	}
}

func TestSeason_NewSeasonWithOfficialTrue(t *testing.T) {
	season := domain.NewSeasonWithOfficial("season-3", domain.SeasonStateActive, true)
	if !season.Official {
		t.Fatalf("expected Official to be true, got %v", season.Official)
	}
}

func TestValidateCanCreateSeason_RejectsWhenActiveExists(t *testing.T) {
	active := []domain.Season{
		domain.NewSeason("season-2026-q4", domain.SeasonStateActive),
	}

	err := domain.ValidateCanCreateSeason(active)
	if err != domain.ErrActiveSeasonAlreadyExists {
		t.Fatalf("expected ErrActiveSeasonAlreadyExists, got %v", err)
	}
}

func TestSeasonStatusFilter_Matches(t *testing.T) {
	active := domain.NewSeason("season-active", domain.SeasonStateActive)
	finalized := domain.NewSeason("season-finalized", domain.SeasonStateFinalized)

	if !domain.SeasonStatusFilterActive.Match(active) {
		t.Fatalf("expected ACTIVE filter to match active season")
	}
	if domain.SeasonStatusFilterActive.Match(finalized) {
		t.Fatalf("expected ACTIVE filter not to match finalized season")
	}
	if !domain.SeasonStatusFilterInactive.Match(finalized) {
		t.Fatalf("expected INACTIVE filter to match finalized season")
	}
	if domain.SeasonStatusFilterInactive.Match(active) {
		t.Fatalf("expected INACTIVE filter not to match active season")
	}
	if !domain.SeasonStatusFilterAll.Match(active) || !domain.SeasonStatusFilterAll.Match(finalized) {
		t.Fatalf("expected ALL filter to match any season")
	}
}

func TestValidateCanCreateSeason_AllowsWhenNoActive(t *testing.T) {
	active := []domain.Season{
		domain.NewSeason("season-2026-q3", domain.SeasonStateFinalized),
		domain.NewSeason("season-2026-q2", domain.SeasonStateArchived),
	}

	err := domain.ValidateCanCreateSeason(active)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestNewActiveSeason_AcceptsValid(t *testing.T) {
	season, err := domain.NewActiveSeason("season-2026-q4", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if season.SeasonID != "season-2026-q4" || season.State != domain.SeasonStateActive || !season.Official {
		t.Fatalf("unexpected season: %+v", season)
	}
}

func TestNewActiveSeason_RejectsEmpty(t *testing.T) {
	_, err := domain.NewActiveSeason("", true)
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}

func TestParseSeasonStatusFilter_AcceptsEmptyAndKnown(t *testing.T) {
	for _, raw := range []string{"", "active", "inactive", "all"} {
		_, err := domain.ParseSeasonStatusFilter(raw)
		if err != nil {
			t.Fatalf("expected no error for %q, got %v", raw, err)
		}
	}
}

func TestParseSeasonStatusFilter_RejectsUnknown(t *testing.T) {
	_, err := domain.ParseSeasonStatusFilter("closed")
	if !errors.Is(err, domain.ErrInvalidSeasonStatusFilter) {
		t.Fatalf("expected ErrInvalidSeasonStatusFilter, got %v", err)
	}
}

func TestSeasonStatusFilter_DefaultDoesNotMatch(t *testing.T) {
	var filter domain.SeasonStatusFilter
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	if filter.Match(season) {
		t.Fatalf("expected zero-value filter not to match")
	}
}
