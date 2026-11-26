package application

import (
	"context"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// AdjudicationSnapshot is the strongly consistent read of rank-relevant
// state captured while the adjudication lock is held. It is the only input
// the finalizer uses to derive official placement facts.
type AdjudicationSnapshot struct {
	Buckets []AdjudicationBucket
}

// AdjudicationBucket groups ranked entries of one category/track pair.
type AdjudicationBucket struct {
	Category string
	Track    domain.Track
	Entries  []AdjudicationEntry
}

// AdjudicationEntry is one ranked recipient with its final merit total and
// the contract ids whose credits actually contributed to this bucket —
// the only contracts whose frozen rosters may feed TEAM member_ids.
type AdjudicationEntry struct {
	RecipientID         string
	Total               int
	CreditedContractIDs []string
}

// AdjudicationLock acquires the internal adjudication attempt for a season.
// Acquiring is idempotent: a retry after an orphaned lock resumes the same
// stable attempt (same ID and cutoff) instead of minting a parallel one.
type AdjudicationLock interface {
	Acquire(seasonID string) (domain.AdjudicationAttempt, error)
}

// AdjudicationSnapshotReader reads the fenced snapshot for a held attempt.
type AdjudicationSnapshotReader interface {
	Read(attempt domain.AdjudicationAttempt) (AdjudicationSnapshot, error)
}

// CreditedRosterReader resolves member_ids for a TEAM fact from the frozen
// rosters of exactly the credited contracts — never an uncredited union.
type CreditedRosterReader interface {
	Members(seasonID, category, recipientID string, creditedContractIDs []string) ([]string, error)
}

// PlacementFactStager stages immutable facts under the attempt with
// conditional semantics: identical replays are no-ops, divergent content
// under a staged identity is a conflict, never an overwrite.
type PlacementFactStager interface {
	Stage(attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error
}

// StagedFactVerifier confirms the staged set matches the derived facts
// before the manifest may commit.
type StagedFactVerifier interface {
	Verify(attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error
}

// FinalizationCommitter commits the constant-size guarded transaction:
// FREEZING season to FINALIZED plus the deterministic manifest record. It
// receives the still-FREEZING season because the transition itself is the
// guarded write.
type FinalizationCommitter interface {
	Commit(season domain.Season, attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error
}

// LockedSeason reports one season still holding an adjudication lock, with
// the persisted attempt identity and cutoff minted at lock acquisition.
type LockedSeason struct {
	SeasonID  string
	AttemptID string
	Cutoff    time.Time
}

// AdjudicationLockLister lists every season currently holding an
// adjudication lock, for the internal orphaned-lock sweep.
type AdjudicationLockLister interface {
	ListLocked(ctx context.Context) ([]LockedSeason, error)
}

// OrphanedLockReporter surfaces locks held past the reconcile threshold to
// the operator channel; it never unlocks or mutates them itself.
type OrphanedLockReporter interface {
	Report(ctx context.Context, orphans []LockedSeason) error
}
