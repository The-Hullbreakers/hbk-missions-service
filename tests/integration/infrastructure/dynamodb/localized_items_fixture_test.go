//go:build integration

package dynamodb_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// localizedStoreFixture creates a client and a fresh private table owned by
// one test — never reuses or deletes a shared table. The endpoint is guarded
// to loopback HTTP before any networking happens; errors are returned for
// the caller to assert.
func localizedStoreFixture(t *testing.T) (*missiondynamo.DynamoDBClient, string, error) {
	t.Helper()
	endpoint := dynamodbTestEndpoint()
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", err
	}
	if parsed.Scheme != "http" || (parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1") {
		return nil, "", fmt.Errorf("DYNAMODB_TEST_ENDPOINT must be loopback HTTP")
	}
	client, err := missiondynamo.NewDynamoDBClient(t.Context(), "us-east-1", endpoint)
	if err != nil {
		return nil, "", err
	}
	table := fmt.Sprintf("hbk-missions-locale-%d", time.Now().UnixNano())
	if err := client.CreateMissionsTable(t.Context(), table); err != nil {
		return nil, "", err
	}
	return client, table, nil
}

// localizedDefinitionRequest is one bilingual draft authored through the
// application create contract.
func localizedDefinitionRequest() application.CreateMissionDefinitionRequest {
	return application.CreateMissionDefinitionRequest{
		DefinitionCode: "def-rescue", Title: "Rescate", Description: "Salvar al piloto",
		Category: "BOUNTY_HUNTING", Mode: domain.MissionModeSolo, CompletionReward: 50,
		Objectives: []domain.Objective{
			domain.NewObjectiveNamed("obj-1", "Capturar", "Capturar al piloto", "BOUNTY_HUNTING", 10, 0, 2),
		},
		Rules: []string{"No ayudas"},
		ContentEn: &domain.LocalizedContent{
			Title: "Rescue", Description: "Save the pilot", Rules: []string{"No aids"},
			Objectives: []domain.LocalizedObjective{
				{ObjectiveID: "obj-1", Name: "Capture", Description: "Capture the pilot"},
			},
		},
	}
}
