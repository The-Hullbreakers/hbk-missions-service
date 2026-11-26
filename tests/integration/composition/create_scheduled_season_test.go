package composition_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestCompositionServer_CreateScheduledSeason(t *testing.T) {
	ctx := context.Background()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions-composition"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: fakeUserSession{},
	})
	if err != nil {
		t.Fatalf("could not create server: %v", err)
	}

	ts := httptest.NewServer(server)
	defer ts.Close()

	body := `{"tenant_id":"tenant-1","official_statistics":true,"starts_at":"2026-09-02T00:00:00Z","schedule":{"active_weeks":4,"off_weeks":1,"series":{"prefix":"season-2026","iteration":1}}}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/seasons", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create season request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["SeasonID"] != "season-2026-1" {
		t.Fatalf("unexpected season in response: %+v", result)
	}
	if result["starts_at"] != "2026-09-02T00:00:00Z" {
		t.Fatalf("expected starts_at, got %v", result["starts_at"])
	}
	schedule, ok := result["schedule"].(map[string]any)
	if !ok {
		t.Fatalf("expected schedule object, got %v", result["schedule"])
	}
	if schedule["active_weeks"] != float64(4) || schedule["off_weeks"] != float64(1) {
		t.Fatalf("unexpected schedule: %+v", schedule)
	}
}
