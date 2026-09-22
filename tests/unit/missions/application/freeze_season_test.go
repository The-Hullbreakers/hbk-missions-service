package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type recordingSeasonStateStore struct {
	seasons map[string]domain.Season
}

func newRecordingSeasonStateStore() *recordingSeasonStateStore {
	return &recordingSeasonStateStore{seasons: make(map[string]domain.Season)}
}

func (s *recordingSeasonStateStore) Get(seasonID string) (domain.Season, error) {
	season, ok := s.seasons[seasonID]
	if !ok {
		return domain.Season{}, domain.ErrSeasonNotFound
	}
	return season, nil
}

func (s *recordingSeasonStateStore) Save(season domain.Season) error {
	s.seasons[season.SeasonID] = season
	return nil
}

type failingSeasonStateReader struct{}

func (failingSeasonStateReader) Get(string) (domain.Season, error) {
	return domain.Season{}, errors.New("season reader unavailable")
}

type failingSeasonStateWriter struct{}

func (failingSeasonStateWriter) Save(domain.Season) error {
	return errors.New("season writer unavailable")
}

func TestFreezeSeasonHandler_TransitionsActiveToFreezing(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewFreezeSeasonHandler(store, store)

	season, err := handler.Handle("season-1")

	if err != nil {
		t.Fatalf("expected freeze to succeed, got %v", err)
	}
	if season.State != domain.SeasonStateFreezing {
		t.Fatalf("expected FREEZING, got %v", season.State)
	}
	if store.seasons["season-1"].State != domain.SeasonStateFreezing {
		t.Fatalf("expected persisted season to be FREEZING")
	}
}

func TestFreezeSeasonHandler_RejectsNonActive(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateFinalized)
	handler := application.NewFreezeSeasonHandler(store, store)

	_, err := handler.Handle("season-1")

	if err != domain.ErrInvalidSeasonStateTransition {
		t.Fatalf("expected ErrInvalidSeasonStateTransition, got %v", err)
	}
	if store.seasons["season-1"].State != domain.SeasonStateFinalized {
		t.Fatalf("expected no change in store")
	}
}

func TestFreezeSeasonHandler_RejectsMissing(t *testing.T) {
	store := newRecordingSeasonStateStore()
	handler := application.NewFreezeSeasonHandler(store, store)

	_, err := handler.Handle("missing")

	if err != domain.ErrSeasonNotFound {
		t.Fatalf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestFreezeSeasonHandler_ReturnsGetError(t *testing.T) {
	handler := application.NewFreezeSeasonHandler(failingSeasonStateReader{}, newRecordingSeasonStateStore())
	_, err := handler.Handle("season-1")
	if err == nil {
		t.Fatalf("expected get error")
	}
}

func TestFreezeSeasonHandler_ReturnsSaveError(t *testing.T) {
	store := newRecordingSeasonStateStore()
	store.seasons["season-1"] = domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewFreezeSeasonHandler(store, failingSeasonStateWriter{})
	_, err := handler.Handle("season-1")
	if err == nil {
		t.Fatalf("expected save error")
	}
}
