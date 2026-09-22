package composition_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

func TestCompositionServer_ReadyFailsWhenDynamoDBUnavailable(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	tableName := "hbk-missions-readyz"
	if err := client.CreateMissionsTable(ctx, tableName); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   tableName,
		DynamoDB:    client,
		UserSession: session.NewJWTUserSession("", "", nil),
	})
	if err != nil {
		t.Fatalf("could not create server: %v", err)
	}

	ts := httptest.NewServer(server)
	defer ts.Close()

	_, err = client.Client().DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Fatalf("could not delete table: %v", err)
	}

	resp, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("ready request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}
