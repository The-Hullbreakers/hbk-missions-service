package application_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// allSeasonsStore serves the full season list so the create-exists invariant
// can see non-ACTIVE seasons too.
type allSeasonsStore struct {
	seasons []domain.Season
	saved   []domain.Season
}

func (s *allSeasonsStore) List(filter domain.SeasonStatusFilter) ([]domain.Season, error) {
	matched := make([]domain.Season, 0, len(s.seasons))
	for _, season := range s.seasons {
		if filter.Match(season) {
			matched = append(matched, season)
		}
	}
	return matched, nil
}

func (s *allSeasonsStore) Save(season domain.Season) error {
	s.saved = append(s.saved, season)
	return nil
}

func (s *allSeasonsStore) Create(season domain.Season) error {
	for _, existing := range s.seasons {
		if existing.SeasonID == season.SeasonID {
			return domain.ErrSeasonAlreadyExists
		}
	}
	return s.Save(season)
}

// REQ-SCC-005 / SCN-SCC-CREATE-EXISTS: creating an id that already exists is
// refused and the stored season is untouched — even when it is FINALIZED.
func TestCreateSeasonHandler_RejectsExistingSeasonID(t *testing.T) {
	store := &allSeasonsStore{
		seasons: []domain.Season{domain.NewSeasonWithOfficial("season-x", domain.SeasonStateFinalized, true)},
	}
	handler := application.NewCreateSeasonHandler(store, store)

	_, err := handler.Handle("season-x", false)
	require.ErrorIs(t, err, domain.ErrSeasonAlreadyExists)
	require.Len(t, store.saved, 0)
}

func TestCreateSeasonHandler_StillRejectsSecondActive(t *testing.T) {
	store := &allSeasonsStore{
		seasons: []domain.Season{domain.NewSeasonWithOfficial("season-active", domain.SeasonStateActive, true)},
	}
	handler := application.NewCreateSeasonHandler(store, store)

	_, err := handler.Handle("season-z", true)
	require.ErrorIs(t, err, domain.ErrActiveSeasonAlreadyExists)
}

func TestCreateSeasonHandler_CreatesWhenIDFreeAndNoActive(t *testing.T) {
	store := &allSeasonsStore{
		seasons: []domain.Season{domain.NewSeasonWithOfficial("season-x", domain.SeasonStateFinalized, true)},
	}
	handler := application.NewCreateSeasonHandler(store, store)

	season, err := handler.Handle("season-z", false)
	require.NoError(t, err)
	require.False(t, season.SeasonID != "season-z" || len(store.saved) != 1)
}

// Same invariant for the scheduler-driven creation path.
func TestCreateScheduledSeasonHandler_RejectsExistingSeasonID(t *testing.T) {
	store := &allSeasonsStore{
		seasons: []domain.Season{domain.NewSeasonWithOfficial("weekly-3", domain.SeasonStateFinalized, true)},
	}
	handler := application.NewCreateScheduledSeasonHandler(store, store)

	_, err := handler.Handle(application.CreateScheduledSeasonCommand{
		SeriesPrefix: "weekly",
		Iteration:    3,
		StartsAt:     time.Now(),
		ActiveWeeks:  2,
		Official:     true,
	})
	require.ErrorIs(t, err, domain.ErrSeasonAlreadyExists)
	require.Len(t, store.saved, 0)
}
