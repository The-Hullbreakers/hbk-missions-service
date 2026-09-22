package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestNewScheduledSeason_CreatesActiveSeasonWithDerivedID(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)

	season, err := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if season.SeasonID != "season-2026-1" {
		t.Fatalf("expected SeasonID season-2026-1, got %s", season.SeasonID)
	}
	if season.State != domain.SeasonStateActive {
		t.Fatalf("expected state ACTIVE, got %v", season.State)
	}
	if !season.Official {
		t.Fatalf("expected Official true")
	}
	if !season.StartsAt.Equal(start) {
		t.Fatalf("expected StartsAt %v, got %v", start, season.StartsAt)
	}
	if season.Schedule == nil {
		t.Fatalf("expected schedule")
	}
	if season.Schedule.ActiveWeeks != 4 || season.Schedule.OffWeeks != 1 {
		t.Fatalf("unexpected schedule: %+v", season.Schedule)
	}
	if season.Schedule.Series.Prefix != "season-2026" || season.Schedule.Series.Iteration != 1 {
		t.Fatalf("unexpected series: %+v", season.Schedule.Series)
	}
}

func TestNewScheduledSeason_RejectsZeroActiveWeeks(t *testing.T) {
	_, err := domain.NewScheduledSeason("season-2026", 1, time.Now(), 0, 1, true)
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}

func TestNewScheduledSeason_RejectsNegativeOffWeeks(t *testing.T) {
	_, err := domain.NewScheduledSeason("season-2026", 1, time.Now(), 4, -1, true)
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}

func TestNewScheduledSeason_RejectsEmptyPrefix(t *testing.T) {
	_, err := domain.NewScheduledSeason("", 1, time.Now(), 4, 1, true)
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}

func TestNewScheduledSeason_RejectsZeroIteration(t *testing.T) {
	_, err := domain.NewScheduledSeason("season-2026", 0, time.Now(), 4, 1, true)
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}

func TestNewScheduledSeason_RejectsZeroStart(t *testing.T) {
	_, err := domain.NewScheduledSeason("season-2026", 1, time.Time{}, 4, 1, true)
	if !errors.Is(err, domain.ErrSeasonInvalid) {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
}
