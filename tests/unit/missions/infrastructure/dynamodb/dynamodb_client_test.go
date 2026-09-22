package dynamodb_test

import (
	"context"
	"testing"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestNewDynamoDBClient_ProductionPath(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")

	client, err := missiondynamo.NewDynamoDBClient(context.Background(), "us-east-1", "")
	if err != nil {
		t.Fatalf("expected no error creating production client: %v", err)
	}
	if client.Client() == nil {
		t.Fatalf("expected non-nil AWS client")
	}
}
