package application_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestSeasonScheduler_FreezesActiveWhenWindowEnds(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)

	store := newSchedulerStore()
	_ = store.Save(season)

	scheduler := store.buildScheduler(start.AddDate(0, 0, 28))
	if err := scheduler.Tick(); err != nil {
		t.Fatalf("expected tick to succeed, got %v", err)
	}

	updated := store.seasons["season-2026-1"]
	if updated.State != domain.SeasonStateFreezing {
		t.Fatalf("expected FREEZING, got %v", updated.State)
	}
}

func TestSeasonScheduler_FinalizesAndCreatesNextWhenOffSeasonEnds(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)
	season.State = domain.SeasonStateFreezing

	store := newSchedulerStore()
	_ = store.Save(season)

	scheduler := store.buildScheduler(start.AddDate(0, 0, 35))
	if err := scheduler.Tick(); err != nil {
		t.Fatalf("expected tick to succeed, got %v", err)
	}

	if store.seasons["season-2026-1"].State != domain.SeasonStateFinalized {
		t.Fatalf("expected season-2026-1 FINALIZED, got %v", store.seasons["season-2026-1"].State)
	}

	next, ok := store.seasons["season-2026-2"]
	if !ok {
		t.Fatalf("expected season-2026-2 to be created")
	}
	if next.State != domain.SeasonStateActive {
		t.Fatalf("expected next season ACTIVE, got %v", next.State)
	}
	if !next.StartsAt.Equal(start.AddDate(0, 0, 35)) {
		t.Fatalf("expected next start %v, got %v", start.AddDate(0, 0, 35), next.StartsAt)
	}
}

func TestSeasonScheduler_IgnoresManualSeasons(t *testing.T) {
	store := newSchedulerStore()
	_ = store.Save(domain.NewSeason("manual", domain.SeasonStateActive))

	scheduler := store.buildScheduler(time.Now())
	if err := scheduler.Tick(); err != nil {
		t.Fatalf("expected tick to succeed, got %v", err)
	}

	if store.seasons["manual"].State != domain.SeasonStateActive {
		t.Fatalf("expected manual season to stay active, got %v", store.seasons["manual"].State)
	}
}

func TestSeasonScheduler_CatchesUpMultipleCycles(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)

	store := newSchedulerStore()
	_ = store.Save(season)

	scheduler := store.buildScheduler(start.AddDate(0, 0, 70))
	if err := scheduler.Tick(); err != nil {
		t.Fatalf("expected tick to succeed, got %v", err)
	}

	if store.seasons["season-2026-1"].State != domain.SeasonStateFinalized {
		t.Fatalf("expected first season finalized")
	}
	if store.seasons["season-2026-2"].State != domain.SeasonStateFinalized {
		t.Fatalf("expected second season finalized")
	}
	if store.seasons["season-2026-3"].State != domain.SeasonStateActive {
		t.Fatalf("expected third season active")
	}
}
