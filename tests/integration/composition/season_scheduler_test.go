package composition_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestCompositionServer_SchedulerTransitionsScheduledSeason(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	clearMissionsTable(t, ctx, client, "hbk-missions")

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:       "hbk-missions",
		DynamoDB:        client,
		UserSession:     fakeUserSession{},
		ShutdownContext: ctx,
		SchedulerTick:   100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("could not create server: %v", err)
	}

	ts := httptest.NewServer(server)
	defer ts.Close()

	start := time.Now().UTC().AddDate(0, 0, -10).Truncate(time.Second)
	body := fmt.Sprintf(`{"tenant_id":"tenant-1","official_statistics":true,"starts_at":"%s","schedule":{"active_weeks":1,"off_weeks":0,"series":{"prefix":"season-2026","iteration":1}}}`, start.Format(time.RFC3339))

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/seasons", strings.NewReader(body))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create season request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	time.Sleep(300 * time.Millisecond)

	req, err = http.NewRequest(http.MethodGet, ts.URL+"/api/v1/seasons?status=all", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer any")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list seasons request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	var oldSeason, newSeason map[string]any
	for _, s := range result {
		switch s["SeasonID"] {
		case "season-2026-1":
			oldSeason = s
		case "season-2026-2":
			newSeason = s
		}
	}

	if oldSeason["State"] != "FINALIZED" {
		t.Fatalf("expected season-2026-1 FINALIZED, got %v", oldSeason["State"])
	}
	if newSeason["State"] != "ACTIVE" {
		t.Fatalf("expected season-2026-2 ACTIVE, got %v", newSeason["State"])
	}
}
