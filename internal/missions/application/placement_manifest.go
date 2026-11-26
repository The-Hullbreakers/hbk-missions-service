package application

import (
	"context"
	"time"
)

// PlacementManifest is the committed, constant-size record of one
// adjudication attempt's staged fact set. NextIndex is the durable progress
// cursor: how many declared facts have already been marked published.
type PlacementManifest struct {
	ID        string
	SeasonID  string
	AttemptID string
	Cutoff    time.Time
	FactCount int
	FactHash  string
	NextIndex int
}

// ManifestIntentReader lists committed manifests still pending publication.
// Staged-but-uncommitted rows are never returned, so they can never reach
// the bus.
type ManifestIntentReader interface {
	ListPendingManifests(ctx context.Context, limit int) ([]PlacementManifest, error)
}

// StagedFactReader maps the committed manifest's staged facts to their
// canonical pending intents — including the manifest cutoff as occurred_at —
// and returns only the remainder after durable progress.
type StagedFactReader interface {
	ListRemaining(ctx context.Context, manifest PlacementManifest) ([]PendingIntent, error)
}

// ManifestProgressWriter durably records that one fact of the manifest was
// published; implementations tolerate repeats so at-least-once republication
// stays idempotent.
type ManifestProgressWriter interface {
	MarkFactPublished(ctx context.Context, manifestID, eventID string) error
}

// ManifestCompletionWriter marks the manifest fully drained; it is invoked
// only after every declared fact was published.
type ManifestCompletionWriter interface {
	MarkComplete(ctx context.Context, manifestID string) error
}
