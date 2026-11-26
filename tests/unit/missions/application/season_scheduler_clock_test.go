package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

func TestNewRealClock_NowReturnsCurrentTime(t *testing.T) {
	clock := application.NewRealClock()
	before := time.Now().Add(-time.Second)
	now := clock.Now()
	after := time.Now().Add(time.Second)
	if now.Before(before) || now.After(after) {
		t.Fatalf("expected real clock near now, got %v", now)
	}
}

func TestSeasonScheduler_StartRunsAndStops(t *testing.T) {
	store := newSchedulerStore()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := store.buildScheduler(time.Now())
	scheduler.Start(ctx, time.Millisecond, nil)
	time.Sleep(2 * time.Millisecond)
	cancel()
}
