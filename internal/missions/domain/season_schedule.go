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
// ActiveDays/OffDays are the canonical storage; ActiveWeeks/OffWeeks remain
// populated for legacy callers that read week fields.
type SeasonSchedule struct {
	ActiveDays  int          `json:"active_days"`
	OffDays     int          `json:"off_days"`
	ActiveWeeks int          `json:"active_weeks,omitempty"`
	OffWeeks    int          `json:"off_weeks,omitempty"`
	Series      SeasonSeries `json:"series"`
}

// NewSeasonScheduleDays creates a valid day-granular schedule.
func NewSeasonScheduleDays(prefix string, iteration, activeDays, offDays int) (SeasonSchedule, error) {
	if prefix == "" || iteration < 0 || activeDays <= 0 || offDays < 0 {
		return SeasonSchedule{}, ErrSeasonInvalid
	}
	return SeasonSchedule{
		ActiveDays: activeDays,
		OffDays:    offDays,
		Series: SeasonSeries{
			Prefix:    prefix,
			Iteration: iteration,
		},
	}, nil
}

// NewSeasonScheduleWeeks creates a schedule from week values (1 week = 7 days).
func NewSeasonScheduleWeeks(prefix string, iteration, activeWeeks, offWeeks int) (SeasonSchedule, error) {
	if iteration <= 0 || activeWeeks <= 0 || offWeeks < 0 {
		return SeasonSchedule{}, ErrSeasonInvalid
	}
	schedule, err := NewSeasonScheduleDays(prefix, iteration, activeWeeks*7, offWeeks*7)
	if err != nil {
		return SeasonSchedule{}, err
	}
	schedule.ActiveWeeks = activeWeeks
	schedule.OffWeeks = offWeeks
	return schedule, nil
}

// NewSeasonSchedule keeps the original week-based signature as an alias.
func NewSeasonSchedule(prefix string, iteration, activeWeeks, offWeeks int) (SeasonSchedule, error) {
	return NewSeasonScheduleWeeks(prefix, iteration, activeWeeks, offWeeks)
}

// ActiveDuration returns the length of the active window.
func (s SeasonSchedule) ActiveDuration() time.Duration {
	return daysToDuration(s.ActiveDays)
}

// CycleDuration returns the length of the active window plus off-season.
func (s SeasonSchedule) CycleDuration() time.Duration {
	return daysToDuration(s.ActiveDays + s.OffDays)
}

// NewScheduledSeason creates an active scheduled season from primitive values.
// The signature keeps weeks for backward compatibility.
func NewScheduledSeason(seriesPrefix string, iteration int, startsAt time.Time, activeWeeks, offWeeks int, official bool) (Season, error) {
	schedule, err := NewSeasonScheduleWeeks(seriesPrefix, iteration, activeWeeks, offWeeks)
	if err != nil {
		return Season{}, err
	}
	return NewScheduledSeasonFromSchedule(schedule, startsAt, official)
}

// NewScheduledSeasonDays creates an active scheduled season from day values.
func NewScheduledSeasonDays(seriesPrefix string, iteration int, startsAt time.Time, activeDays, offDays int, official bool) (Season, error) {
	schedule, err := NewSeasonScheduleDays(seriesPrefix, iteration, activeDays, offDays)
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

func (s Season) activeEnd() time.Time {
	return s.StartsAt.Add(s.Schedule.ActiveDuration())
}

func (s Season) cycleEnd() time.Time {
	return s.StartsAt.Add(s.Schedule.CycleDuration())
}

func daysToDuration(days int) time.Duration {
	return time.Duration(days) * 24 * time.Hour
}
