package domain

import "time"

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
		ActiveDays:  s.Schedule.ActiveDays,
		OffDays:     s.Schedule.OffDays,
		ActiveWeeks: s.Schedule.ActiveWeeks,
		OffWeeks:    s.Schedule.OffWeeks,
		Series:      s.Schedule.Series.Next(),
	}
	return NewScheduledSeasonFromSchedule(nextSchedule, nextStartsAt, s.Official)
}

// SuccessorPending reports whether a finalized scheduled season still owes
// its next-season creation — the marker a later scheduler tick retries.
// Manual seasons and non-finalized states never report pending.
func (s Season) SuccessorPending(existing []Season) bool {
	if s.Schedule == nil || s.State != SeasonStateFinalized {
		return false
	}
	next, err := s.NextScheduledSeason()
	if err != nil {
		return false
	}
	for _, other := range existing {
		if other.SeasonID == next.SeasonID {
			return false
		}
	}
	return true
}
