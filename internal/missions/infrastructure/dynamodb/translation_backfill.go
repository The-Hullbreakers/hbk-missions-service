package dynamodb

import (
	"context"
	"fmt"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ReviewedMissionTranslation binds English text to its exact Spanish source
// and definition version. Backfill never derives a translation from current
// catalog content.
type ReviewedMissionTranslation struct {
	DefinitionCode    string
	DefinitionVersion string
	SourceSpanish     domain.LocalizedContent
	English           domain.LocalizedContent
}

type BackfillRecord struct {
	Partition string
	ID        string
	Reason    string
}

type TranslationBackfillReport struct {
	Eligible   []BackfillRecord
	Updated    []BackfillRecord
	Unresolved []BackfillRecord
}

// TranslationBackfiller is an offline, explicitly invoked rehearsal adapter.
// No runtime handler or scheduler invokes it.
type TranslationBackfiller struct {
	client *DynamoDBClient
	table  string
}

func NewTranslationBackfiller(client *DynamoDBClient, table string) *TranslationBackfiller {
	return &TranslationBackfiller{client: client, table: table}
}

// Rehearse inspects every candidate for one reviewed definition version.
// apply=false is read-only; apply=true conditionally adds ContentEn only.
func (b *TranslationBackfiller) Rehearse(ctx context.Context, reviewed ReviewedMissionTranslation, apply bool) (TranslationBackfillReport, error) {
	var report TranslationBackfillReport
	if err := validateReviewedTranslation(reviewed); err != nil {
		return report, err
	}
	if b == nil || b.client == nil || b.table == "" {
		return report, fmt.Errorf("translation backfill requires a client and table")
	}
	for _, partition := range []string{"DEFINITION", "OFFER", "CONTRACT"} {
		if err := b.inspectPartition(ctx, partition, reviewed, apply, &report); err != nil {
			return report, err
		}
	}
	return report, nil
}
