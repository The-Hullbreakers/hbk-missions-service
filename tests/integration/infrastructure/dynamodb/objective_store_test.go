package dynamodb_test

import (
	"context"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestObjectiveStore_RoundTripAndMissing(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	tableName := "hbk-missions-" + t.Name()
	if err := client.CreateMissionsTable(ctx, tableName); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	store := missiondynamo.NewObjectiveStore(client, tableName)
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)
	if err := store.Save(objective); err != nil {
		t.Fatalf("could not save objective: %v", err)
	}

	loaded, err := store.Get("objective-1")
	if err != nil {
		t.Fatalf("could not get objective: %v", err)
	}
	if loaded.ObjectiveID != "objective-1" {
		t.Fatalf("unexpected objective: %+v", loaded)
	}

	_, err = store.Get("missing")
	if err == nil {
		t.Fatalf("expected an error for missing objective")
	}
}
