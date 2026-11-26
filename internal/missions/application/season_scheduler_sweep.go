package application

import (
	"context"
	"time"
)

// WithOrphanSweep enables the internal orphaned-lock sweep: after every tick
// the scheduler lists seasons still holding an adjudication lock and reports
// those whose cutoff predates now-threshold to the operator reporter. The
// sweep never unlocks or mutates seasons — reconcile stays manual.
func (s *SeasonScheduler) WithOrphanSweep(lister AdjudicationLockLister, reporter OrphanedLockReporter, threshold time.Duration) *SeasonScheduler {
	s.lockLister = lister
	s.orphanReporter = reporter
	s.orphanThreshold = threshold
	return s
}

// sweepOrphanedLocks reports locks held longer than the configured
// threshold. A lister failure fails the tick so the outage is visible.
func (s *SeasonScheduler) sweepOrphanedLocks(ctx context.Context) error {
	if s.lockLister == nil || s.orphanReporter == nil {
		return nil
	}
	locked, err := s.lockLister.ListLocked(ctx)
	if err != nil {
		return err
	}
	var orphans []LockedSeason
	cutoff := s.clock.Now().Add(-s.orphanThreshold)
	for _, season := range locked {
		if season.Cutoff.Before(cutoff) {
			orphans = append(orphans, season)
		}
	}
	if len(orphans) == 0 {
		return nil
	}
	return s.orphanReporter.Report(ctx, orphans)
}
