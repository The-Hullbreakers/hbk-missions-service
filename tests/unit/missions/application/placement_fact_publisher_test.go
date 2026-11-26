package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// Legacy publisher coverage migrated to the committed-manifest model:
// pending manifests replace the flat outbox, per-fact progress replaces
// per-intent marks, and completion is the new terminal mark.
func TestPlacementFactPublisher_PublishesPendingAndMarksThem(t *testing.T) {
	t.Parallel()

	// Arrange
	log := &finCallLog{}
	publisher := &pubEventSpy{log: log}
	progress := &pubProgressSpy{log: log}
	useCase := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finManifest()}},
		&pubRemainingStub{remaining: []application.PendingIntent{pubFact("evt-1"), pubFact("evt-2")}},
		publisher, progress, &pubCompleteSpy{log: log})

	// Act
	err := useCase.PublishPending(context.Background(), 10)

	// Assert
	require.NoError(t, err)
	require.Equal(t, []application.PendingIntent{pubFact("evt-1"), pubFact("evt-2")}, publisher.published)
	require.Equal(t, [][2]string{
		{"MANIFEST#season-1#ADJ#season-1", "evt-1"},
		{"MANIFEST#season-1#ADJ#season-1", "evt-2"}}, progress.marked)
}

func TestPlacementFactPublisher_LeavesFailedIntentPending(t *testing.T) {
	t.Parallel()

	// Arrange
	log := &finCallLog{}
	progress := &pubProgressSpy{log: log}
	completion := &pubCompleteSpy{log: log}
	useCase := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finManifest()}},
		&pubRemainingStub{remaining: []application.PendingIntent{pubFact("evt-1")}},
		&pubEventSpy{log: log, failAt: 1, err: errors.New("sns unavailable")},
		progress, completion)

	// Act
	err := useCase.PublishPending(context.Background(), 10)

	// Assert
	require.Error(t, err)
	require.Empty(t, progress.marked)
	require.Empty(t, completion.completed)
}

func TestPlacementFactPublisher_IdempotentRepublish(t *testing.T) {
	t.Parallel()

	// Arrange — the reader returns the same committed manifest twice in
	// one batch (at-least-once delivery) with the same remaining intent;
	// progress and completion writes absorb the repeats.
	log := &finCallLog{}
	publisher := &pubEventSpy{log: log}
	progress := &pubProgressSpy{log: log}
	completion := &pubCompleteSpy{log: log}
	useCase := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finResumedManifest(), finResumedManifest()}},
		&pubRemainingStub{remaining: []application.PendingIntent{pubFact("evt-2")}},
		publisher, progress, completion)

	// Act
	err := useCase.PublishPending(context.Background(), 10)

	// Assert — identical event identity and progress marks per delivery.
	require.NoError(t, err)
	require.Len(t, publisher.published, 2)
	assert.Equal(t, "evt-2", publisher.published[0].EventID)
	assert.Equal(t, "evt-2", publisher.published[1].EventID)
	require.Equal(t, [][2]string{
		{"MANIFEST#season-1#ADJ#season-1", "evt-2"},
		{"MANIFEST#season-1#ADJ#season-1", "evt-2"}}, progress.marked)
	assert.Equal(t, []string{
		"MANIFEST#season-1#ADJ#season-1",
		"MANIFEST#season-1#ADJ#season-1"}, completion.completed,
		"repeated completion writes stay idempotent")
}
