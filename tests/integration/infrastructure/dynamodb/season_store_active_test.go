package dynamodb_test

import (
	"context"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestSeasonStore_ListActiveAndGet(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	tableName := "hbk-missions-" + t.Name()
	if err := client.CreateMissionsTable(ctx, tableName); err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	store := missiondynamo.NewSeasonStore(client, tableName)
	season := domain.NewSeason("season-active", domain.SeasonStateActive)
	if err := store.Save(season); err != nil {
		t.Fatalf("could not save season: %v", err)
	}

	active, err := store.ListActive()
	if err != nil {
		t.Fatalf("could not list active seasons: %v", err)
	}
	if len(active) != 1 || active[0].SeasonID != "season-active" {
		t.Fatalf("unexpected active seasons: %+v", active)
	}

	loaded, err := store.Get("season-active")
	if err != nil {
		t.Fatalf("could not get season: %v", err)
	}
	if loaded.SeasonID != "season-active" || loaded.State != domain.SeasonStateActive {
		t.Fatalf("unexpected season: %+v", loaded)
	}

	_, err = store.Get("missing")
	if err == nil {
		t.Fatalf("expected an error for missing season")
	}
}
