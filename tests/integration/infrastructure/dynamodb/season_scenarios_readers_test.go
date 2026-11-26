//go:build integration

package dynamodb_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-0055-005: the existing use cases can read the persistent local preview.
func TestSeasonScenarioReaders(t *testing.T) {
	// Arrange
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "../../../..")
	endpoint := "http://localhost:8000"
	table := "hbk-missions-preview-mid-season"
	command := exec.Command("python3", "-B", filepath.Join(root, "scripts/seed_season_scenarios.py"),
		"--scenario", "mid-season", "--endpoint-url", endpoint, "--apply")
	command.Dir = root
	command.Env = append(os.Environ(), "AWS_ACCESS_KEY_ID=dummy", "AWS_SECRET_ACCESS_KEY=dummy",
		"AWS_EC2_METADATA_DISABLED=true", "AWS_DEFAULT_REGION=us-east-1", "AWS_SESSION_TOKEN=")
	// Act
	output, runErr := command.CombinedOutput()
	// Assert
	require.NoError(t, runErr, string(output))
	client, err := missiondynamo.NewDynamoDBClient(context.Background(), "us-east-1", endpoint)
	require.NoError(t, err)
	seasons := missiondynamo.NewSeasonStore(client, table)
	ledger := missiondynamo.NewMeritLedgerStore(client, table)
	current, err := application.NewGetCurrentSeasonHandler(seasons).Handle()
	require.NoError(t, err)
	assert.Equal(t, "preview-mid-season", current.SeasonID)
	assert.False(t, current.Official)
	statistics, err := application.NewGetSeasonStatisticsHandler(seasons, seasons, ledger, ledger).Handle("current", "", "", 20)
	require.NoError(t, err)
	require.Len(t, statistics.Buckets, 4)
	for _, bucket := range statistics.Buckets {
		assert.Equal(t, domain.TrackIndividual, bucket.Bucket.Track)
		assert.Greater(t, bucket.Participants, 0)
		assert.Greater(t, bucket.TotalMerits, 0)
	}
	recipient := statistics.Buckets[0].Top[0].RecipientID
	identity, err := domain.NewCallerIdentity(recipient, nil, nil)
	require.NoError(t, err)
	standing, err := application.NewGetMyStandingHandler(seasons, ledger).Handle(identity,
		current.SeasonID, statistics.Buckets[0].Bucket.Category, "INDIVIDUAL")
	require.NoError(t, err)
	assert.Equal(t, recipient, standing.RecipientID)
	contracts, err := application.NewListPlayerContractsHandler(missiondynamo.NewContractStore(client, table)).Handle(
		recipient, domain.ContractListFilter{SeasonID: current.SeasonID}, domain.LanguageSpanish)
	require.NoError(t, err)
	assert.NotEmpty(t, contracts)
}
