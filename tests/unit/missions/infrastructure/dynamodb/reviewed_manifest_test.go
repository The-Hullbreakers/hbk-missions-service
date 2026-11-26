package dynamodb_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func reviewedLocalFixture() missiondynamo.ReviewedMissionTranslation {
	return missiondynamo.ReviewedMissionTranslation{
		DefinitionCode: "local-translation", DefinitionVersion: "4",
		SourceSpanish: domain.LocalizedContent{
			Title: "Mision", Description: "Descripcion", Rules: []string{"Regla"},
			Objectives: []domain.LocalizedObjective{
				{ObjectiveID: "objetivo", Name: "Nombre", Description: "Detalle"},
			},
		},
		English: domain.LocalizedContent{
			Title: "Mission", Description: "Description", Rules: []string{"Rule"},
			Objectives: []domain.LocalizedObjective{
				{ObjectiveID: "objetivo", Name: "Name", Description: "Detail"},
			},
		},
	}
}

func TestReviewedTranslationManifestLoadsVersionBoundContent(t *testing.T) {
	t.Parallel()
	// Arrange
	dir := t.TempDir()
	record := reviewedLocalFixture()
	source, err := json.Marshal(map[string]any{"translations": []missiondynamo.ReviewedMissionTranslation{record}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviewed.json"), source, 0600))

	// Act
	loaded, err := missiondynamo.LoadReviewedTranslationManifests(dir)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []missiondynamo.ReviewedMissionTranslation{record}, loaded)
}

func TestReviewedTranslationManifestRejectsMissingDuplicateOrIncompleteRecords(t *testing.T) {
	t.Parallel()
	// Arrange
	valid := reviewedLocalFixture()
	blankTitle := reviewedLocalFixture()
	blankTitle.English.Title = ""
	wrongObjective := reviewedLocalFixture()
	wrongObjective.English.Objectives = append([]domain.LocalizedObjective(nil), wrongObjective.English.Objectives...)
	wrongObjective.English.Objectives[0].ObjectiveID = "wrong"
	cases := []struct {
		name    string
		records []missiondynamo.ReviewedMissionTranslation
	}{
		{"missing", nil},
		{"duplicate", []missiondynamo.ReviewedMissionTranslation{valid, valid}},
		{"blank English", []missiondynamo.ReviewedMissionTranslation{blankTitle}},
		{"wrong objective ID", []missiondynamo.ReviewedMissionTranslation{wrongObjective}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			content, err := json.Marshal(map[string]any{"translations": test.records})
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(dir, "reviewed.json"), content, 0600))

			// Act
			_, err = missiondynamo.LoadReviewedTranslationManifests(dir)

			// Assert
			require.Error(t, err)
		})
	}
}

func TestReviewedTranslationManifestRejectsEconomicFields(t *testing.T) {
	t.Parallel()
	// Arrange
	dir := t.TempDir()
	record := reviewedLocalFixture()
	source, err := json.Marshal(map[string]any{"translations": []map[string]any{{
		"DefinitionCode": record.DefinitionCode, "DefinitionVersion": record.DefinitionVersion,
		"SourceSpanish": record.SourceSpanish,
		"English": map[string]any{
			"Title": record.English.Title, "Description": record.English.Description,
			"Rules": record.English.Rules, "Objectives": record.English.Objectives,
			"CompletionReward": 100,
		},
	}}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "reviewed.json"), source, 0600))

	// Act
	_, err = missiondynamo.LoadReviewedTranslationManifests(dir)

	// Assert
	require.Error(t, err)
}

func TestReviewedTranslationManifestRequiresLoopbackEndpoint(t *testing.T) {
	t.Parallel()
	// Arrange
	allowed := []string{"http://localhost:8000", "http://127.0.0.1:18000", "http://[::1]:18000"}
	rejected := []string{
		"", "https://localhost:8000", "http://example.invalid:18000",
		"http://localhost", "http://localhost:8000/other",
		"http://user:password@localhost:8000", "http://localhost:8000?table=prod",
	}
	for _, endpoint := range allowed {
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()
			// Act
			err := missiondynamo.ValidateLocalBackfillEndpoint(endpoint)
			// Assert
			require.NoError(t, err)
		})
	}
	for _, endpoint := range rejected {
		t.Run("reject "+endpoint, func(t *testing.T) {
			t.Parallel()
			// Act
			err := missiondynamo.ValidateLocalBackfillEndpoint(endpoint)
			// Assert
			require.Error(t, err)
		})
	}
}
