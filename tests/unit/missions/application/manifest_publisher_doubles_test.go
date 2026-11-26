package application_test

import (
	"context"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

func finManifest() application.PlacementManifest {
	return application.PlacementManifest{
		ID: "MANIFEST#season-1#ADJ#season-1", SeasonID: "season-1",
		AttemptID: "ADJ#season-1",
		Cutoff:    time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC),
		FactCount: 2, FactHash: "digest-abc", NextIndex: 0,
	}
}

func finResumedManifest() application.PlacementManifest {
	m := finManifest()
	m.NextIndex = 1
	return m
}

func finEmptyManifest() application.PlacementManifest {
	m := finManifest()
	m.FactCount = 0
	m.FactHash = "digest-empty"
	return m
}

func pubFact(eventID string) application.PendingIntent {
	return application.PendingIntent{
		EventID: eventID, EventType: "missions.season.placement-awarded",
		Envelope: `{"event_id":"` + eventID +
			`","event_type":"missions.season.placement-awarded","occurred_at":"2026-10-21T12:00:00Z"}`,
	}
}

type pubManifestStub struct {
	manifests []application.PlacementManifest
	gotLimit  int
	err       error
}

func (s *pubManifestStub) ListPendingManifests(_ context.Context, limit int) ([]application.PlacementManifest, error) {
	s.gotLimit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.manifests, nil
}

type pubRemainingStub struct {
	remaining []application.PendingIntent
	got       []application.PlacementManifest
	err       error
}

func (s *pubRemainingStub) ListRemaining(_ context.Context, m application.PlacementManifest) ([]application.PendingIntent, error) {
	s.got = append(s.got, m)
	if s.err != nil {
		return nil, s.err
	}
	return s.remaining, nil
}

type pubEventSpy struct {
	published []application.PendingIntent
	failAt    int
	err       error
	log       *finCallLog
}

func (s *pubEventSpy) Publish(_ context.Context, intent application.PendingIntent) error {
	s.published = append(s.published, intent)
	s.log.add("publish:" + intent.EventID)
	if len(s.published) == s.failAt {
		return s.err
	}
	return nil
}

type pubProgressSpy struct {
	marked [][2]string
	log    *finCallLog
	err    error
}

func (s *pubProgressSpy) MarkFactPublished(_ context.Context, manifestID, eventID string) error {
	s.marked = append(s.marked, [2]string{manifestID, eventID})
	s.log.add("mark:" + eventID)
	return s.err
}

type pubCompleteSpy struct {
	completed []string
	log       *finCallLog
	err       error
}

func (s *pubCompleteSpy) MarkComplete(_ context.Context, manifestID string) error {
	s.completed = append(s.completed, manifestID)
	s.log.add("complete:" + manifestID)
	return s.err
}
