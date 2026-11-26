package composition_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestCompositionServer_CreateSeason_RejectsWhenActiveSeasonExists(t *testing.T) {
	ctx := t.Context()
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

	first := `{"tenant_id":"tenant-1","season_id":"season-2026-q4","official_statistics":true}`
	r1 := postBody(t, ts.URL+"/api/v1/seasons", first)
	defer r1.Body.Close()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for first season, got %d", r1.StatusCode)
	}

	second := `{"tenant_id":"tenant-1","season_id":"season-2026-q5","official_statistics":true}`
	r2 := postBody(t, ts.URL+"/api/v1/seasons", second)
	defer r2.Body.Close()

	if r2.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", r2.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(r2.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["code"] != "ACTIVE_SEASON_EXISTS" {
		t.Fatalf("expected ACTIVE_SEASON_EXISTS, got %v", result["code"])
	}
}

func postBody(t *testing.T, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create season request failed: %v", err)
	}
	return resp
}
