package main

import (
	"context"
	"fmt"
	"regexp"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

type publishOptions struct {
	endpoint  string
	table     string
	directory string
	apply     bool
	dryRun    bool
}

type reportRow struct {
	Partition string `json:"partition"`
	ID        string `json:"id"`
	Reason    string `json:"reason,omitempty"`
}

type publishReport struct {
	Applied    bool        `json:"applied"`
	Eligible   []reportRow `json:"eligible"`
	Updated    []reportRow `json:"updated"`
	Unresolved []reportRow `json:"unresolved"`
}

func newReport() publishReport {
	return publishReport{
		Eligible: make([]reportRow, 0), Updated: make([]reportRow, 0),
		Unresolved: make([]reportRow, 0),
	}
}

func addRecords(target []reportRow, records []missiondynamo.BackfillRecord) []reportRow {
	for _, record := range records {
		target = append(target, reportRow{
			Partition: record.Partition, ID: record.ID, Reason: record.Reason,
		})
	}
	return target
}

func publishReviewedEnglish(ctx context.Context, options publishOptions) (publishReport, error) {
	result := newReport()
	if err := missiondynamo.ValidateLocalBackfillEndpoint(options.endpoint); err != nil {
		return result, err
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]{3,255}$`).MatchString(options.table) {
		return result, fmt.Errorf("an explicit local DynamoDB table name is required")
	}
	records, err := missiondynamo.LoadReviewedTranslationManifests(options.directory)
	if err != nil {
		return result, err
	}
	if len(records) != 36 {
		return result, fmt.Errorf("expected all 36 reviewed catalog translations, found %d", len(records))
	}
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", options.endpoint)
	if err != nil {
		return result, err
	}
	backfill := missiondynamo.NewTranslationBackfiller(client, options.table)
	for _, reviewed := range records {
		preview, err := backfill.Rehearse(ctx, reviewed, false)
		result.Eligible = addRecords(result.Eligible, preview.Eligible)
		result.Unresolved = addRecords(result.Unresolved, preview.Unresolved)
		if err != nil {
			return result, fmt.Errorf("failed to preview local translation: %w", err)
		}
	}
	if !options.apply {
		return result, nil
	}
	if len(result.Unresolved) != 0 {
		return result, fmt.Errorf("local translation preview has unresolved records; no writes made")
	}
	result.Applied = true
	for _, reviewed := range records {
		applied, err := backfill.Rehearse(ctx, reviewed, true)
		result.Updated = addRecords(result.Updated, applied.Updated)
		result.Unresolved = addRecords(result.Unresolved, applied.Unresolved)
		if err != nil {
			return result, fmt.Errorf("local translation partly applied; inspect updated IDs: %w", err)
		}
		if len(result.Unresolved) != 0 {
			return result, fmt.Errorf("concurrent local translation conflict; inspect updated IDs before retry")
		}
	}
	return result, nil
}
