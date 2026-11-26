//go:build integration

package dynamodb_test

import (
	"fmt"
	"testing"
	"time"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// manifestStoreFixture names the table with the test identity as well as
// time: parallel tests can share a clock tick, but cannot share a table.
func manifestStoreFixture(t *testing.T) (*missiondynamo.DynamoDBClient, string, error) {
	t.Helper()
	endpoint := dynamodbTestEndpoint()
	if err := missiondynamo.ValidateLocalBackfillEndpoint(endpoint); err != nil {
		return nil, "", err
	}
	client, err := missiondynamo.NewDynamoDBClient(t.Context(), "us-east-1", endpoint)
	if err != nil {
		return nil, "", err
	}
	table := fmt.Sprintf("hbk-chg0051-manifest-%s-%d", t.Name(), time.Now().UnixNano())
	return client, table, client.CreateMissionsTable(t.Context(), table)
}
