package dynamodb_test

import (
	"context"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestSeasonStore_ScheduledSeasonRoundTrip(t *testing.T) {
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
	startsAt := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	season, err := domain.NewScheduledSeason("season-2026", 1, startsAt, 4, 1, true)
	if err != nil {
		t.Fatalf("could not create scheduled season: %v", err)
	}

	if err := store.Save(season); err != nil {
		t.Fatalf("could not save scheduled season: %v", err)
	}

	loaded, err := store.Get("season-2026-1")
	if err != nil {
		t.Fatalf("could not get scheduled season: %v", err)
	}
	if loaded.SeasonID != "season-2026-1" {
		t.Fatalf("unexpected season id: %s", loaded.SeasonID)
	}
	if !loaded.StartsAt.Equal(startsAt) {
		t.Fatalf("expected starts_at %v, got %v", startsAt, loaded.StartsAt)
	}
	if loaded.Schedule == nil {
		t.Fatalf("expected schedule to be loaded")
	}
	if loaded.Schedule.ActiveWeeks != 4 || loaded.Schedule.OffWeeks != 1 {
		t.Fatalf("unexpected schedule: %+v", loaded.Schedule)
	}
	if loaded.Schedule.Series.Prefix != "season-2026" || loaded.Schedule.Series.Iteration != 1 {
		t.Fatalf("unexpected series: %+v", loaded.Schedule.Series)
	}

	active, err := store.ListActive()
	if err != nil {
		t.Fatalf("could not list active seasons: %v", err)
	}
	if len(active) != 1 || active[0].SeasonID != "season-2026-1" {
		t.Fatalf("unexpected active seasons: %+v", active)
	}
}
