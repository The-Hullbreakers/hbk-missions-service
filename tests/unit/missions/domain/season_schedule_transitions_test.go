package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeason_ShouldFreeze_AtActiveWindowEnd(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, err := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if season.ShouldFreeze(start.AddDate(0, 0, 27)) {
		t.Fatalf("expected not to freeze one day before window end")
	}
	if !season.ShouldFreeze(start.AddDate(0, 0, 28)) {
		t.Fatalf("expected to freeze exactly at active window end")
	}
}

func TestSeason_ShouldFreeze_IgnoresManualSeasons(t *testing.T) {
	season := domain.NewSeason("manual", domain.SeasonStateActive)
	if season.ShouldFreeze(time.Now()) {
		t.Fatalf("expected manual season not to freeze")
	}
}

func TestSeason_ShouldFreeze_IgnoresNonActive(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)
	season.State = domain.SeasonStateFreezing

	if season.ShouldFreeze(start.AddDate(0, 0, 28)) {
		t.Fatalf("expected non-active scheduled season not to freeze")
	}
}

func TestSeason_ShouldFinalize_AtOffSeasonEnd(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, err := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	season.State = domain.SeasonStateFreezing

	if season.ShouldFinalize(start.AddDate(0, 0, 34)) {
		t.Fatalf("expected not to finalize one day before off-season end")
	}
	if !season.ShouldFinalize(start.AddDate(0, 0, 35)) {
		t.Fatalf("expected to finalize exactly at off-season end")
	}
}

func TestSeason_ShouldFinalize_IgnoresManualSeasons(t *testing.T) {
	season := domain.NewSeason("manual", domain.SeasonStateFreezing)
	if season.ShouldFinalize(time.Now()) {
		t.Fatalf("expected manual freezing season not to finalize")
	}
}

func TestSeason_NextStartsAt(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)

	next, err := season.NextStartsAt()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !next.Equal(start.AddDate(0, 0, 35)) {
		t.Fatalf("expected next start at %v, got %v", start.AddDate(0, 0, 35), next)
	}
}

func TestSeason_NextStartsAt_ReturnsErrorForManualSeason(t *testing.T) {
	season := domain.NewSeason("manual", domain.SeasonStateActive)
	_, err := season.NextStartsAt()
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}

func TestSeason_NextScheduledSeason(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)

	next, err := season.NextScheduledSeason()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if next.SeasonID != "season-2026-2" {
		t.Fatalf("expected next season id season-2026-2, got %s", next.SeasonID)
	}
	if next.State != domain.SeasonStateActive {
		t.Fatalf("expected next season active")
	}
	if !next.StartsAt.Equal(start.AddDate(0, 0, 35)) {
		t.Fatalf("expected next start %v, got %v", start.AddDate(0, 0, 35), next.StartsAt)
	}
	if next.Schedule == nil || next.Schedule.Series.Iteration != 2 {
		t.Fatalf("expected next iteration 2, got %+v", next.Schedule)
	}
	if !next.Official {
		t.Fatalf("expected Official flag preserved")
	}
}

func TestSeason_NextScheduledSeries_GeneratesCorrectIDs(t *testing.T) {
	series := domain.SeasonSeries{Prefix: "season-2026", Iteration: 3}

	if series.SeasonID() != "season-2026-3" {
		t.Fatalf("expected SeasonID season-2026-3, got %s", series.SeasonID())
	}

	next := series.Next()
	if next.Prefix != "season-2026" || next.Iteration != 4 {
		t.Fatalf("unexpected next series: %+v", next)
	}
}
