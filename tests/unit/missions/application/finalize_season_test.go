package application_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestFinalizeSeasonHandler_TransitionsFreezingToFinalized(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateFreezing)
	handler := application.NewFinalizeSeasonHandler(store, store)

	season, err := handler.Handle("season-1")

	if err != nil {
		t.Fatalf("expected finalize to succeed, got %v", err)
	}
	if season.State != domain.SeasonStateFinalized {
		t.Fatalf("expected FINALIZED, got %v", season.State)
	}
	if store.seasons["season-1"].State != domain.SeasonStateFinalized {
		t.Fatalf("expected persisted season to be FINALIZED")
	}
}

func TestFinalizeSeasonHandler_RejectsNonFreezing(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewFinalizeSeasonHandler(store, store)

	_, err := handler.Handle("season-1")

	if err != domain.ErrInvalidSeasonStateTransition {
		t.Fatalf("expected ErrInvalidSeasonStateTransition, got %v", err)
	}
}

func TestFinalizeSeasonHandler_ReturnsGetError(t *testing.T) {
	handler := application.NewFinalizeSeasonHandler(failingSeasonStateReader{}, newRecordingSeasonStateStore())
	_, err := handler.Handle("season-1")
	if err == nil {
		t.Fatalf("expected get error")
	}
}

func TestFinalizeSeasonHandler_ReturnsSaveError(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateFreezing)
	handler := application.NewFinalizeSeasonHandler(store, failingSeasonStateWriter{})
	_, err := handler.Handle("season-1")
	if err == nil {
		t.Fatalf("expected save error")
	}
}
