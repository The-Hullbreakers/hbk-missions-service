package dynamodb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestDynamoDBClient_NewDynamoDBClientWithoutEndpoint(t *testing.T) {
	ctx := context.Background()
	// This covers the default AWS config branch; no actual network call is made.
	_, _ = missiondynamo.NewDynamoDBClient(ctx, "us-east-1", "")
}

func TestDynamoDBClient_CreateMissionsTableIsIdempotent(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	name := "hbk-missions-create-idempotent"
	_, _ = client.Client().DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	})

	if err := client.CreateMissionsTable(ctx, name); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, name); err != nil {
		t.Fatalf("second create should be idempotent: %v", err)
	}
}
