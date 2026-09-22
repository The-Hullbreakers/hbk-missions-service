package dynamodb_test

import (
	"context"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestSeasonStore_List(t *testing.T) {
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
	seasons := []domain.Season{
		domain.NewSeason("season-active", domain.SeasonStateActive),
		domain.NewSeason("season-archived", domain.SeasonStateArchived),
	}
	for _, season := range seasons {
		if err := store.Save(season); err != nil {
			t.Fatalf("could not save season: %v", err)
		}
	}

	active, err := store.List(domain.SeasonStatusFilterActive)
	if err != nil {
		t.Fatalf("could not list active seasons: %v", err)
	}
	if len(active) != 1 || active[0].SeasonID != "season-active" {
		t.Fatalf("expected one active season, got %+v", active)
	}

	inactive, err := store.List(domain.SeasonStatusFilterInactive)
	if err != nil {
		t.Fatalf("could not list inactive seasons: %v", err)
	}
	if len(inactive) != 1 || inactive[0].SeasonID != "season-archived" {
		t.Fatalf("expected one inactive season, got %+v", inactive)
	}

	all, err := store.List(domain.SeasonStatusFilterAll)
	if err != nil {
		t.Fatalf("could not list all seasons: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected two seasons, got %d", len(all))
	}
}
