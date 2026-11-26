package application_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestCreateScheduledSeasonHandler_CreatesScheduledSeason(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateScheduledSeasonHandler(store, store)

	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, err := handler.Handle(application.CreateScheduledSeasonCommand{
		SeriesPrefix: "season-2026",
		Iteration:    1,
		StartsAt:     start,
		ActiveWeeks:  4,
		OffWeeks:     1,
		Official:     true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if season.SeasonID != "season-2026-1" {
		t.Fatalf("expected season id season-2026-1, got %s", season.SeasonID)
	}
	if !season.StartsAt.Equal(start) {
		t.Fatalf("expected StartsAt %v, got %v", start, season.StartsAt)
	}
	if season.Schedule.ActiveWeeks != 4 || season.Schedule.OffWeeks != 1 {
		t.Fatalf("unexpected schedule: %+v", season.Schedule)
	}
	if len(store.saved) != 1 {
		t.Fatalf("expected one saved season, got %d", len(store.saved))
	}
}

func TestCreateScheduledSeasonHandler_RejectsZeroActiveWeeks(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateScheduledSeasonHandler(store, store)

	_, err := handler.Handle(application.CreateScheduledSeasonCommand{
		SeriesPrefix: "season-2026",
		Iteration:    1,
		StartsAt:     time.Now(),
		ActiveWeeks:  0,
		OffWeeks:     1,
	})
	if err != domain.ErrSeasonInvalid {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
	if len(store.saved) != 0 {
		t.Fatalf("expected no saved season")
	}
}

func TestCreateScheduledSeasonHandler_RejectsWhenActiveSeasonExists(t *testing.T) {
	store := &recordingSeasonStore{
		existing: []domain.Season{
			domain.NewSeason("season-active", domain.SeasonStateActive),
		},
	}
	handler := application.NewCreateScheduledSeasonHandler(store, store)

	_, err := handler.Handle(application.CreateScheduledSeasonCommand{
		SeriesPrefix: "season-2026",
		Iteration:    1,
		StartsAt:     time.Now(),
		ActiveWeeks:  4,
		OffWeeks:     1,
	})
	if err != domain.ErrActiveSeasonAlreadyExists {
		t.Fatalf("expected ErrActiveSeasonAlreadyExists, got %v", err)
	}
}
