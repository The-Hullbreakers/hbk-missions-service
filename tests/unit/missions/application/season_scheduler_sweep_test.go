package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

type recordingLockLister struct {
	locked []application.LockedSeason
	err    error
	calls  int
}

func (l *recordingLockLister) ListLocked(_ context.Context) ([]application.LockedSeason, error) {
	l.calls++
	return l.locked, l.err
}

type recordingOrphanReporter struct {
	orphans []application.LockedSeason
	err     error
	calls   int
}

func (r *recordingOrphanReporter) Report(_ context.Context, orphans []application.LockedSeason) error {
	r.calls++
	r.orphans = orphans
	return r.err
}

// TEST-APP-SWP-001 — a lock held past the reconcile threshold is surfaced to
// the operator reporter; a fresh lock is not an orphan.
func TestSeasonScheduler_OrphanedLockSweep(t *testing.T) {
	now := time.Now()
	orphan := application.LockedSeason{SeasonID: "season-stuck", AttemptID: "ADJ#season-stuck", Cutoff: now.Add(-25 * time.Hour)}
	fresh := application.LockedSeason{SeasonID: "season-fresh", AttemptID: "ADJ#season-fresh", Cutoff: now.Add(-time.Hour)}

	store := newSchedulerStore()
	lister := &recordingLockLister{locked: []application.LockedSeason{orphan, fresh}}
	reporter := &recordingOrphanReporter{}
	scheduler := store.buildScheduler(now).WithOrphanSweep(lister, reporter, 24*time.Hour)

	err := scheduler.Tick()

	require.NoError(t, err)
	assert.Equal(t, 1, lister.calls)
	require.Equal(t, 1, reporter.calls)
	require.Equal(t, []application.LockedSeason{orphan}, reporter.orphans)
}

func TestSeasonScheduler_LockListerFailureFailsTheTick(t *testing.T) {
	store := newSchedulerStore()
	lister := &recordingLockLister{err: errors.New("scan failed")}
	reporter := &recordingOrphanReporter{}
	scheduler := store.buildScheduler(time.Now()).WithOrphanSweep(lister, reporter, 24*time.Hour)

	err := scheduler.Tick()

	require.Error(t, err)
	assert.Equal(t, 0, reporter.calls)
}
