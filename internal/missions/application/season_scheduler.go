package application

import (
	"context"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// Clock provides the current time. It is the only source of time in the scheduler.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

// Now returns the current system time.
func (realClock) Now() time.Time { return time.Now() }

// NewRealClock returns a Clock backed by the system time.
func NewRealClock() Clock { return realClock{} }

// SeasonScheduler transitions scheduled seasons through their lifecycle.
type SeasonScheduler struct {
	reader   SeasonListReader
	clock    Clock
	freeze   *FreezeSeasonHandler
	finalize *FinalizeSeasonHandler
	create   *CreateScheduledSeasonHandler
}

// NewSeasonScheduler creates a scheduler with the required ports.
func NewSeasonScheduler(reader SeasonListReader, clock Clock, freeze *FreezeSeasonHandler, finalize *FinalizeSeasonHandler, create *CreateScheduledSeasonHandler) *SeasonScheduler {
	return &SeasonScheduler{
		reader:   reader,
		clock:    clock,
		freeze:   freeze,
		finalize: finalize,
		create:   create,
	}
}

// Tick evaluates all seasons and transitions those whose schedule windows have elapsed.
// It repeats until a complete pass makes no transitions, allowing the scheduler to catch up.
func (s *SeasonScheduler) Tick() error {
	for {
		seasons, err := s.reader.List(domain.SeasonStatusFilterAll)
		if err != nil {
			return err
		}

		now := s.clock.Now()
		changed := false
		for _, season := range seasons {
			if season.State == domain.SeasonStateActive && season.ShouldFreeze(now) {
				if _, err := s.freeze.Handle(season.SeasonID); err != nil {
					return err
				}
				changed = true
			}
			if season.State == domain.SeasonStateFreezing && season.ShouldFinalize(now) {
				if _, err := s.finalize.Handle(season.SeasonID); err != nil {
					return err
				}
				if season.Schedule != nil {
					next, err := season.NextScheduledSeason()
					if err != nil {
						return err
					}
					if _, err := s.create.Handle(scheduledCommandFromSeason(next)); err != nil {
						return err
					}
				}
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return nil
}

// Start runs the scheduler in a background goroutine.
func (s *SeasonScheduler) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = s.Tick()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func scheduledCommandFromSeason(season domain.Season) CreateScheduledSeasonCommand {
	return CreateScheduledSeasonCommand{
		SeriesPrefix: season.Schedule.Series.Prefix,
		Iteration:    season.Schedule.Series.Iteration,
		StartsAt:     season.StartsAt,
		ActiveWeeks:  season.Schedule.ActiveWeeks,
		OffWeeks:     season.Schedule.OffWeeks,
		Official:     season.Official,
	}
}
