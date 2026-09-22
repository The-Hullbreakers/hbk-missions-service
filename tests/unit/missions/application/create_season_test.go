package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type recordingSeasonStore struct {
	active []domain.Season
	saved  []domain.Season
}

func (s *recordingSeasonStore) Save(season domain.Season) error {
	s.saved = append(s.saved, season)
	return nil
}

func (s *recordingSeasonStore) ListActive() ([]domain.Season, error) {
	return s.active, nil
}

type failingCreateSeasonStore struct{}

func (failingCreateSeasonStore) ListActive() ([]domain.Season, error) {
	return nil, errors.New("active seasons unavailable")
}

func (failingCreateSeasonStore) Save(domain.Season) error {
	return errors.New("season writer unavailable")
}

func TestCreateSeasonHandler_CreatesOfficialSeason(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(store, store)

	season, err := handler.Handle("tenant-1", "season-2026-q4", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if season.SeasonID != "season-2026-q4" {
		t.Fatalf("expected season id season-2026-q4, got %s", season.SeasonID)
	}
	if !season.Official {
		t.Fatalf("expected Official to be true")
	}
	if len(store.saved) != 1 {
		t.Fatalf("expected one saved season, got %d", len(store.saved))
	}
}

func TestCreateSeasonHandler_CreatesUnofficialSeason(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(store, store)

	season, err := handler.Handle("tenant-1", "practice-2026-q4", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if season.Official {
		t.Fatalf("expected Official to be false")
	}
	if len(store.saved) != 1 {
		t.Fatalf("expected one saved season, got %d", len(store.saved))
	}
}

func TestCreateSeasonHandler_RejectsEmptySeasonID(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(store, store)

	_, err := handler.Handle("tenant-1", "", true)
	if err != domain.ErrSeasonInvalid {
		t.Fatalf("expected ErrSeasonInvalid, got %v", err)
	}
	if len(store.saved) != 0 {
		t.Fatalf("expected no saved season, got %d", len(store.saved))
	}
}

func TestCreateSeasonHandler_RejectsWhenActiveSeasonExists(t *testing.T) {
	store := &recordingSeasonStore{
		active: []domain.Season{
			domain.NewSeason("season-2026-q4", domain.SeasonStateActive),
		},
	}
	handler := application.NewCreateSeasonHandler(store, store)

	_, err := handler.Handle("tenant-1", "season-2026-q5", true)
	if err != domain.ErrActiveSeasonAlreadyExists {
		t.Fatalf("expected ErrActiveSeasonAlreadyExists, got %v", err)
	}
	if len(store.saved) != 0 {
		t.Fatalf("expected no saved season, got %d", len(store.saved))
	}
}

func TestCreateSeasonHandler_ReturnsListActiveError(t *testing.T) {
	handler := application.NewCreateSeasonHandler(failingCreateSeasonStore{}, &recordingSeasonStore{})
	_, err := handler.Handle("tenant-1", "season-2026-q5", true)
	if err == nil {
		t.Fatalf("expected list active error")
	}
}

func TestCreateSeasonHandler_ReturnsSaveError(t *testing.T) {
	reader := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(reader, failingCreateSeasonStore{})
	_, err := handler.Handle("tenant-1", "season-2026-q5", true)
	if err == nil {
		t.Fatalf("expected save error")
	}
}
