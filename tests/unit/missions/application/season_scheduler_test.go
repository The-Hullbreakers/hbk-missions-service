package application_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type schedulerStore struct {
	seasons map[string]domain.Season
	list    []domain.Season
}

func newSchedulerStore() *schedulerStore {
	return &schedulerStore{seasons: make(map[string]domain.Season)}
}

func (s *schedulerStore) Get(seasonID string) (domain.Season, error) {
	season, ok := s.seasons[seasonID]
	if !ok {
		return domain.Season{}, domain.ErrSeasonNotFound
	}
	return season, nil
}

func (s *schedulerStore) Save(season domain.Season) error {
	s.seasons[season.SeasonID] = season
	return nil
}

func (s *schedulerStore) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	var out []domain.Season
	for _, season := range s.seasons {
		if filter.Match(season) {
			out = append(out, season)
		}
	}
	return out, nil
}

func (s *schedulerStore) ListActive() ([]domain.Season, error) {
	return s.List(domain.SeasonStatusFilterActive)
}

func (s *schedulerStore) buildScheduler(now time.Time) *application.SeasonScheduler {
	return application.NewSeasonScheduler(
		s,
		&fixedClock{now: now},
		application.NewFreezeSeasonHandler(s, s),
		application.NewFinalizeSeasonHandler(s, s),
		application.NewCreateScheduledSeasonHandler(s, s),
	)
}

type fixedClock struct {
	now time.Time
}

func (c *fixedClock) Now() time.Time { return c.now }

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
