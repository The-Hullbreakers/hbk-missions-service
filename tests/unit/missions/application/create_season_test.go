package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type recordingSeasonStore struct {
	existing []domain.Season
	saved    []domain.Season
}

func (s *recordingSeasonStore) Save(season domain.Season) error {
	s.saved = append(s.saved, season)
	return nil
}

func (s *recordingSeasonStore) Create(season domain.Season) error {
	s.saved = append(s.saved, season)
	return nil
}

func (s *recordingSeasonStore) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	var out []domain.Season
	for _, season := range s.existing {
		if filter.Match(season) {
			out = append(out, season)
		}
	}
	return out, nil
}

type failingCreateSeasonStore struct{}

func (failingCreateSeasonStore) List(domain.SeasonStatusFilter) ([]domain.Season, error) {
	return nil, errors.New("season list unavailable")
}

func (failingCreateSeasonStore) Create(domain.Season) error {
	return errors.New("season writer unavailable")
}

func TestCreateSeasonHandler_CreatesOfficialSeason(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(store, store)

	season, err := handler.Handle("season-2026-q4", true)

	require.NoError(t, err)
	require.Equal(t, "season-2026-q4", season.SeasonID)
	require.True(t, season.Official)
	require.Len(t, store.saved, 1)
}

func TestCreateSeasonHandler_CreatesUnofficialSeason(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(store, store)

	season, err := handler.Handle("practice-2026-q4", false)

	require.NoError(t, err)
	require.False(t, season.Official)
	require.Len(t, store.saved, 1)
}

func TestCreateSeasonHandler_RejectsEmptySeasonID(t *testing.T) {
	store := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(store, store)

	_, err := handler.Handle("", true)

	require.ErrorIs(t, err, domain.ErrSeasonInvalid)
	require.Empty(t, store.saved)
}

func TestCreateSeasonHandler_RejectsWhenActiveSeasonExists(t *testing.T) {
	store := &recordingSeasonStore{
		existing: []domain.Season{
			domain.NewSeason("season-2026-q4", domain.SeasonStateActive),
		},
	}
	handler := application.NewCreateSeasonHandler(store, store)

	_, err := handler.Handle("season-2026-q5", true)

	require.ErrorIs(t, err, domain.ErrActiveSeasonAlreadyExists)
	require.Empty(t, store.saved)
}

func TestCreateSeasonHandler_ReturnsListActiveError(t *testing.T) {
	handler := application.NewCreateSeasonHandler(failingCreateSeasonStore{}, &recordingSeasonStore{})
	_, err := handler.Handle("season-2026-q5", true)
	require.Error(t, err)
}

func TestCreateSeasonHandler_ReturnsSaveError(t *testing.T) {
	reader := &recordingSeasonStore{}
	handler := application.NewCreateSeasonHandler(reader, failingCreateSeasonStore{})
	_, err := handler.Handle("season-2026-q5", true)
	require.Error(t, err)
}
