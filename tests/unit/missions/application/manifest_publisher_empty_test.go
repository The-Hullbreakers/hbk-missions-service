package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// TEST-APP-PUB-001 — completion is only honest when the manifest's own
// declared fact count is fully drained.
func TestPlacementFactPublisher_EmptyManifestCompletesWithoutPublishing(t *testing.T) {
	// Arrange
	log := &finCallLog{}
	events := &pubEventSpy{log: log}
	publisher := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finEmptyManifest()}},
		&pubRemainingStub{remaining: nil},
		events, &pubProgressSpy{log: log}, &pubCompleteSpy{log: log})

	// Act
	err := publisher.PublishPending(context.Background(), 10)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"complete:MANIFEST#season-1#ADJ#season-1"}, log.calls,
		"a committed empty manifest (FactCount=0) is completed without publishing")
	assert.Empty(t, events.published)
}

func TestPlacementFactPublisher_IncompleteManifestCannotComplete(t *testing.T) {
	// Arrange — FactCount=2 with NextIndex=0 but the reader reports zero
	// remaining: completing here would permanently lose two facts.
	log := &finCallLog{}
	progress := &pubProgressSpy{log: log}
	completion := &pubCompleteSpy{log: log}
	publisher := application.NewPlacementFactPublisher(
		&pubManifestStub{manifests: []application.PlacementManifest{finManifest()}},
		&pubRemainingStub{remaining: nil},
		&pubEventSpy{log: log}, progress, completion)

	// Act
	err := publisher.PublishPending(context.Background(), 10)

	// Assert
	require.Error(t, err,
		"a committed manifest that never served its declared facts must not complete")
	assert.Empty(t, progress.marked)
	assert.Empty(t, completion.completed)
}
