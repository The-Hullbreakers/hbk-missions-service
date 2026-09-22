package domain

import (
	"fmt"
	"time"
)

// SeasonSeries identifies a recurring sequence of scheduled seasons.
type SeasonSeries struct {
	Prefix    string `json:"prefix"`
	Iteration int    `json:"iteration"`
}

// SeasonID derives the season identity from the series.
func (s SeasonSeries) SeasonID() string {
	return fmt.Sprintf("%s-%d", s.Prefix, s.Iteration)
}

// Next returns the following iteration of the same series.
func (s SeasonSeries) Next() SeasonSeries {
	return SeasonSeries{
		Prefix:    s.Prefix,
		Iteration: s.Iteration + 1,
	}
}

// SeasonSchedule defines the active and off-season windows of a season.
type SeasonSchedule struct {
	ActiveWeeks int          `json:"active_weeks"`
	OffWeeks    int          `json:"off_weeks"`
	Series      SeasonSeries `json:"series"`
}

// NewSeasonSchedule creates a valid schedule.
func NewSeasonSchedule(prefix string, iteration, activeWeeks, offWeeks int) (SeasonSchedule, error) {
	if prefix == "" || iteration <= 0 || activeWeeks <= 0 || offWeeks < 0 {
		return SeasonSchedule{}, ErrSeasonInvalid
	}
	return SeasonSchedule{
		ActiveWeeks: activeWeeks,
		OffWeeks:    offWeeks,
		Series: SeasonSeries{
			Prefix:    prefix,
			Iteration: iteration,
		},
	}, nil
}

// ActiveDuration returns the length of the active window.
func (s SeasonSchedule) ActiveDuration() time.Duration {
	return weeksToDuration(s.ActiveWeeks)
}

// CycleDuration returns the length of the active window plus off-season.
func (s SeasonSchedule) CycleDuration() time.Duration {
	return weeksToDuration(s.ActiveWeeks + s.OffWeeks)
}

// NewScheduledSeason creates an active scheduled season from primitive values.
func NewScheduledSeason(seriesPrefix string, iteration int, startsAt time.Time, activeWeeks, offWeeks int, official bool) (Season, error) {
	schedule, err := NewSeasonSchedule(seriesPrefix, iteration, activeWeeks, offWeeks)
	if err != nil {
		return Season{}, err
	}
	return NewScheduledSeasonFromSchedule(schedule, startsAt, official)
}

// NewScheduledSeasonFromSchedule creates an active scheduled season from a schedule.
func NewScheduledSeasonFromSchedule(schedule SeasonSchedule, startsAt time.Time, official bool) (Season, error) {
	if startsAt.IsZero() {
		return Season{}, ErrSeasonInvalid
	}
	return Season{
		SeasonID: schedule.Series.SeasonID(),
		State:    SeasonStateActive,
		Official: official,
		StartsAt: startsAt,
		Schedule: &schedule,
	}, nil
}

// ShouldFreeze reports whether the active window has ended.
func (s Season) ShouldFreeze(now time.Time) bool {
	if s.State != SeasonStateActive || s.Schedule == nil {
		return false
	}
	return !now.Before(s.activeEnd())
}

// ShouldFinalize reports whether the off-season has ended.
func (s Season) ShouldFinalize(now time.Time) bool {
	if s.State != SeasonStateFreezing || s.Schedule == nil {
		return false
	}
	return !now.Before(s.cycleEnd())
}

// NextStartsAt returns the start time of the next season in the series.
func (s Season) NextStartsAt() (time.Time, error) {
	if s.Schedule == nil {
		return time.Time{}, ErrSeasonInvalid
	}
	return s.StartsAt.Add(s.Schedule.CycleDuration()), nil
}

// NextScheduledSeason returns the next season in the series.
func (s Season) NextScheduledSeason() (Season, error) {
	if s.Schedule == nil {
		return Season{}, ErrSeasonInvalid
	}
	nextStartsAt, err := s.NextStartsAt()
	if err != nil {
		return Season{}, err
	}
	nextSchedule := SeasonSchedule{
		ActiveWeeks: s.Schedule.ActiveWeeks,
		OffWeeks:    s.Schedule.OffWeeks,
		Series:      s.Schedule.Series.Next(),
	}
	return NewScheduledSeasonFromSchedule(nextSchedule, nextStartsAt, s.Official)
}

func (s Season) activeEnd() time.Time {
	return s.StartsAt.Add(s.Schedule.ActiveDuration())
}

func (s Season) cycleEnd() time.Time {
	return s.StartsAt.Add(s.Schedule.CycleDuration())
}

func weeksToDuration(weeks int) time.Duration {
	return time.Duration(weeks) * 7 * 24 * time.Hour
}
