package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// TEST-APP-PUB-001 — manifest-driven publisher: committed manifests only,
// per-fact progress after each successful publish, completion last.
func TestPlacementFactPublisher_PublishesRemainingInOrderAndCompletes(t *testing.T) {
	// Arrange
	log := &finCallLog{}
	remaining := &pubRemainingStub{remaining: []application.PendingIntent{pubFact("evt-1"), pubFact("evt-2")}}
	progress := &pubProgressSpy{log: log}
	completion := &pubCompleteSpy{log: log}
	publisher := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finManifest()}},
		remaining, &pubEventSpy{log: log}, progress, completion)

	// Act
	err := publisher.PublishPending(context.Background(), 10)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{
		"publish:evt-1", "mark:evt-1", "publish:evt-2", "mark:evt-2",
		"complete:MANIFEST#season-1#ADJ#season-1"}, log.calls,
		"each fact is marked only after its own publish, completion last")
	require.Len(t, remaining.got, 1)
	assert.Equal(t, finManifest(), remaining.got[0])
	assert.Equal(t, [][2]string{
		{"MANIFEST#season-1#ADJ#season-1", "evt-1"},
		{"MANIFEST#season-1#ADJ#season-1", "evt-2"}}, progress.marked)
}

func TestPlacementFactPublisher_PublishFailureLeavesRemainingAndIncomplete(t *testing.T) {
	// Arrange
	log := &finCallLog{}
	progress := &pubProgressSpy{log: log}
	completion := &pubCompleteSpy{log: log}
	publisher := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finManifest()}},
		&pubRemainingStub{remaining: []application.PendingIntent{pubFact("evt-1"), pubFact("evt-2")}},
		&pubEventSpy{log: log, failAt: 2, err: errors.New("sns throttled")},
		progress, completion)

	// Act
	err := publisher.PublishPending(context.Background(), 10)

	// Assert
	require.Error(t, err)
	assert.Equal(t, []string{"publish:evt-1", "mark:evt-1", "publish:evt-2"}, log.calls)
	assert.Equal(t, [][2]string{{"MANIFEST#season-1#ADJ#season-1", "evt-1"}}, progress.marked,
		"only facts published before the failure are marked")
	assert.Empty(t, completion.completed)
}

func TestPlacementFactPublisher_RetryReemitsSameIntentIdentity(t *testing.T) {
	// Arrange — after a failed drain the manifest resumes at NextIndex=1
	// and the reader returns only the unpublished remainder; evt-2 must
	// replay with identical identity.
	log := &finCallLog{}
	events := &pubEventSpy{log: log}
	progress := &pubProgressSpy{log: log}
	completion := &pubCompleteSpy{log: log}
	publisher := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finResumedManifest()}},
		&pubRemainingStub{remaining: []application.PendingIntent{pubFact("evt-2")}},
		events, progress, completion)

	// Act
	err := publisher.PublishPending(context.Background(), 10)

	// Assert
	require.NoError(t, err)
	require.Len(t, events.published, 1)
	assert.Equal(t, "evt-2", events.published[0].EventID,
		"replay must reuse the persisted intent identity")
	assert.Contains(t, events.published[0].Envelope, "2026-10-21T12:00:00Z",
		"occurred_at stays the adjudication cutoff")
	assert.Equal(t, [][2]string{{"MANIFEST#season-1#ADJ#season-1", "evt-2"}}, progress.marked,
		"the single remaining fact is marked once")
	assert.Equal(t, []string{"MANIFEST#season-1#ADJ#season-1"}, completion.completed,
		"draining the remainder completes the manifest")
}

func TestPlacementFactPublisher_NoPendingManifestsPublishesNothing(t *testing.T) {
	// Arrange — the reader returns only COMMITTED pending manifests, so
	// staged-but-uncommitted rows can never reach the publisher.
	log := &finCallLog{}
	remaining := &pubRemainingStub{}
	events := &pubEventSpy{log: log}
	publisher := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: nil}, remaining, events,
		&pubProgressSpy{log: log}, &pubCompleteSpy{log: log})

	// Act
	err := publisher.PublishPending(context.Background(), 10)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, remaining.got)
	assert.Empty(t, events.published)
	assert.Empty(t, log.calls)
}
