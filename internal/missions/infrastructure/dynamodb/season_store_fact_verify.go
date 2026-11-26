package dynamodb

import (
	"context"
	"fmt"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// StagedFactVerifier checks the persisted attempt before finalization may
// commit its manifest. Validation never writes or publishes an intent.
type StagedFactVerifier struct {
	store *SeasonStore
}

func NewStagedFactVerifier(client *DynamoDBClient, tableName string) *StagedFactVerifier {
	return &StagedFactVerifier{store: NewSeasonStore(client, tableName)}
}

func (v *StagedFactVerifier) Verify(attempt domain.AdjudicationAttempt, facts []domain.PlacementFact) error {
	valid, err := domain.NewAdjudicationAttempt(attempt.SeasonID, attempt.Cutoff)
	if err != nil {
		return err
	}
	if attempt.ID != valid.ID {
		return domain.ErrAdjudicationAttemptInvalid
	}
	set, err := domain.NewStagedFactSet(facts)
	if err != nil {
		return err
	}
	manifest := application.PlacementManifest{
		ID:       "MANIFEST#" + attempt.SeasonID + "#" + attempt.ID,
		SeasonID: attempt.SeasonID, AttemptID: attempt.ID, Cutoff: attempt.Cutoff,
		FactCount: set.Count(), FactHash: set.Hash(),
	}
	if _, err := v.store.ListRemaining(context.Background(), manifest); err != nil {
		return fmt.Errorf("failed to verify staged facts: %w", err)
	}
	return nil
}
