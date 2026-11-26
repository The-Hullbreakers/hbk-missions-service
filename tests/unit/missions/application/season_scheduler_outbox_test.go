package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

type recordingDrainer struct {
	calls  int
	limits []int
	err    error
}

func (d *recordingDrainer) PublishPending(_ context.Context, limit int) error {
	d.calls++
	d.limits = append(d.limits, limit)
	return d.err
}

func TestSeasonScheduler_TickDrainsOutboxIntents(t *testing.T) {
	store := newSchedulerStore()
	drainer := &recordingDrainer{}
	scheduler := application.NewSeasonScheduler(
		store,
		&fixedClock{now: time.Now()},
		application.NewFreezeSeasonHandler(store, store),
		application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, store),
		application.NewCreateScheduledSeasonHandler(store, store),
		drainer,
	)

	err := scheduler.Tick()

	require.NoError(t, err)
	require.Equal(t, 1, drainer.calls)
}

func TestSeasonScheduler_OutboxDrainFailureFailsTheTick(t *testing.T) {
	store := newSchedulerStore()
	drainer := &recordingDrainer{err: errors.New("sns unavailable")}
	scheduler := application.NewSeasonScheduler(
		store,
		&fixedClock{now: time.Now()},
		application.NewFreezeSeasonHandler(store, store),
		application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, store),
		application.NewCreateScheduledSeasonHandler(store, store),
		drainer,
	)

	err := scheduler.Tick()

	require.Error(t, err)
}

func TestSeasonScheduler_StartReportsTickErrors(t *testing.T) {
	store := newSchedulerStore()
	drainer := &recordingDrainer{err: errors.New("sns unavailable")}
	scheduler := application.NewSeasonScheduler(
		store,
		&fixedClock{now: time.Now()},
		application.NewFreezeSeasonHandler(store, store),
		application.NewFinalizeSeasonHandler(store, staticBucketReader{}, staticStandingsReader{}, staticRosterReader{}, store),
		application.NewCreateScheduledSeasonHandler(store, store),
		drainer,
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reported := make(chan error, 1)

	scheduler.Start(ctx, time.Millisecond, func(err error) {
		select {
		case reported <- err:
		default:
		}
	})

	select {
	case err := <-reported:
		require.ErrorContains(t, err, "sns unavailable")
	case <-time.After(2 * time.Second):
		t.Fatal("tick error was not surfaced to the error handler")
	}
}
