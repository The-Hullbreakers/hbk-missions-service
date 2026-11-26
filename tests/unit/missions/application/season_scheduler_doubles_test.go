package application_test

import (
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

func (s *schedulerStore) SaveFinalized(season domain.Season, _ []domain.PlacementFact) error {
	return s.Save(season)
}

func (s *schedulerStore) Save(season domain.Season) error {
	s.seasons[season.SeasonID] = season
	return nil
}

func (s *schedulerStore) Create(season domain.Season) error {
	if _, exists := s.seasons[season.SeasonID]; exists {
		return domain.ErrSeasonAlreadyExists
	}
	return s.Save(season)
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
		application.NewFinalizeSeasonHandler(s, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, s),
		application.NewCreateScheduledSeasonHandler(s, s),
		nil,
	)
}

type fixedClock struct {
	now time.Time
}

func (c *fixedClock) Now() time.Time { return c.now }
