package domain_test

import (
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// A day-granular schedule exposes day fields and computes day durations.
func TestNewScheduledSeasonDays_DayWindows(t *testing.T) {
	// Arrange
	start := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)

	// Act
	season, err := domain.NewScheduledSeasonDays("season-2026", 1, start, 30, 3, true)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 30, season.Schedule.ActiveDays)
	require.Equal(t, 3, season.Schedule.OffDays)
	require.Equal(t, 30*24*time.Hour, season.Schedule.ActiveDuration())
	require.Equal(t, 33*24*time.Hour, season.Schedule.CycleDuration())
}

// The day schedule drives freeze/finalize at day boundaries.
func TestScheduledSeasonDays_Transitions(t *testing.T) {
	// Arrange
	start := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeasonDays("season-2026", 1, start, 30, 3, true)

	// Assert — not yet past 30 days
	require.False(t, season.ShouldFreeze(start.Add(29*24*time.Hour)), "expected active before day 30")
	require.True(t, season.ShouldFreeze(start.Add(30*24*time.Hour)), "expected freeze at day 30")
}

// The week constructor remains a valid alias producing day-equivalent windows.
func TestNewSeasonScheduleWeeks_AliasToDays(t *testing.T) {
	// Act
	schedule, err := domain.NewSeasonScheduleWeeks("season-2026", 1, 4, 1)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 28, schedule.ActiveDays)
	require.Equal(t, 7, schedule.OffDays)
}

// A day-created season's successor inherits its day windows.
func TestNextScheduledSeason_KeepsDayWindows(t *testing.T) {
	// Arrange
	start := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeasonDays("season", 0, start, 30, 3, true)

	// Act
	next, err := season.NextScheduledSeason()

	// Assert
	require.NoError(t, err)
	require.Equal(t, "season-1", next.SeasonID)
	require.Equal(t, 30, next.Schedule.ActiveDays)
	require.Equal(t, 3, next.Schedule.OffDays)
	require.True(t, next.StartsAt.Equal(start.Add(33*24*time.Hour)), "expected starts_at at cycle end")
}

// Invalid day values are rejected.
func TestNewSeasonScheduleDays_InvalidRejected(t *testing.T) {
	// Act
	_, zeroActive := domain.NewSeasonScheduleDays("season-2026", 1, 0, 3)
	_, negativeOff := domain.NewSeasonScheduleDays("season-2026", 1, 30, -1)

	// Assert
	require.ErrorIs(t, zeroActive, domain.ErrSeasonInvalid)
	require.ErrorIs(t, negativeOff, domain.ErrSeasonInvalid)
}
