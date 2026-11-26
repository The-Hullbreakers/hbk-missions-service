package application_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/stretchr/testify/require"
)

// failOnceSeasonWriter injects a creation failure for the first N saves so a
// scheduler tick fails mid-recovery and a later tick must retry.
type failOnceSeasonWriter struct {
	inner     *schedulerStore
	remaining int
}

func (w *failOnceSeasonWriter) Create(season domain.Season) error {
	if w.remaining > 0 {
		w.remaining--
		return errors.New("injected create failure")
	}
	return w.inner.Create(season)
}

// SCN-SCC-RETRY-NEXT: when next-season creation fails after finalization, the
// failed tick is observable and the following tick retries until it succeeds.
func TestSeasonScheduler_RetriesSuccessorCreationOnNextTick(t *testing.T) {
	start := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, _ := domain.NewScheduledSeason("season-2026", 1, start, 4, 1, true)
	season.State = domain.SeasonStateFreezing

	store := newSchedulerStore()
	_ = store.Save(season)
	flaky := &failOnceSeasonWriter{inner: store, remaining: 1}
	scheduler := application.NewSeasonScheduler(
		store,
		&fixedClock{now: start.AddDate(0, 0, 35)},
		application.NewFreezeSeasonHandler(store, store),
		application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, store),
		application.NewCreateScheduledSeasonHandler(store, flaky),
		nil,
	)

	err := scheduler.Tick()
	require.False(t, err == nil)
	require.Equal(t, domain.SeasonStateFinalized, store.seasons["season-2026-1"].State)
	_, ok := store.seasons["season-2026-2"]
	require.False(t, ok)

	err = scheduler.Tick()
	require.NoError(t, err)
	next, ok := store.seasons["season-2026-2"]
	require.True(t, ok, "expected season-2026-2 created on retry")
	require.Equal(t, domain.SeasonStateActive, next.State)
}
