package application

import (
	"context"
	"errors"
)

// EventSource is the `source` identity this service stamps on every published
// domain event envelope and SNS filter attribute.
const EventSource = "hbk-missions-service"

// ErrManifestIncomplete is returned when a committed manifest yields fewer
// remaining intents than its declared FactCount requires — completing it
// would silently lose facts.
var ErrManifestIncomplete = errors.New("placement manifest has fewer facts than declared")

// PendingIntent is a persisted publication intent awaiting the bus: the raw
// domain-event envelope plus its identity for progress marking.
type PendingIntent struct {
	EventID   string
	EventType string
	Envelope  string
}

// DomainEventPublisher publishes one intent's envelope to the shared domain
// events bus.
type DomainEventPublisher interface {
	Publish(ctx context.Context, intent PendingIntent) error
}

// PlacementFactPublisher drains committed manifests: for each pending
// manifest it publishes the remaining staged facts in order, marks durable
// progress after every successful publish, and marks the manifest complete
// only when the drain covers the full declared fact count. A failure aborts
// the batch so unpublished facts stay pending for the next drain.
type PlacementFactPublisher struct {
	manifests  ManifestIntentReader
	remaining  StagedFactReader
	publisher  DomainEventPublisher
	progress   ManifestProgressWriter
	completion ManifestCompletionWriter
}

// NewPlacementFactPublisher creates the use case.
func NewPlacementFactPublisher(manifests ManifestIntentReader, remaining StagedFactReader, publisher DomainEventPublisher, progress ManifestProgressWriter, completion ManifestCompletionWriter) *PlacementFactPublisher {
	return &PlacementFactPublisher{
		manifests:  manifests,
		remaining:  remaining,
		publisher:  publisher,
		progress:   progress,
		completion: completion,
	}
}

// PublishPending drains up to limit committed pending manifests.
func (u *PlacementFactPublisher) PublishPending(ctx context.Context, limit int) error {
	manifests, err := u.manifests.ListPendingManifests(ctx, limit)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		if err := u.publishManifest(ctx, manifest); err != nil {
			return err
		}
	}
	return nil
}

func (u *PlacementFactPublisher) publishManifest(ctx context.Context, manifest PlacementManifest) error {
	intents, err := u.remaining.ListRemaining(ctx, manifest)
	if err != nil {
		return err
	}
	for _, intent := range intents {
		if err := u.publisher.Publish(ctx, intent); err != nil {
			return err
		}
		if err := u.progress.MarkFactPublished(ctx, manifest.ID, intent.EventID); err != nil {
			return err
		}
	}
	if manifest.NextIndex+len(intents) != manifest.FactCount {
		return ErrManifestIncomplete
	}
	return u.completion.MarkComplete(ctx, manifest.ID)
}
