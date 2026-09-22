package composition_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestCompositionServer_FreezeAndFinalize(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	if err := client.CreateMissionsTable(ctx, "hbk-missions"); err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	clearMissionsTable(t, ctx, client, "hbk-missions")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions")
	if err := seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)); err != nil {
		t.Fatalf("could not seed active season: %v", err)
	}
	if err := seasonStore.Save(domain.NewSeason("season-2", domain.SeasonStateFinalized)); err != nil {
		t.Fatalf("could not seed finalized season: %v", err)
	}
	if err := seasonStore.Save(domain.NewSeason("season-3", domain.SeasonStateActive)); err != nil {
		t.Fatalf("could not seed second active season: %v", err)
	}

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions",
		DynamoDB:    client,
		UserSession: fakeUserSession{},
	})
	if err != nil {
		t.Fatalf("could not create server: %v", err)
	}

	ts := httptest.NewServer(server)
	defer ts.Close()

	freeze := func(id string) (int, map[string]any) {
		return postSeason(t, ts, "/api/v1/missions/seasons/"+id+"/freeze")
	}
	finalize := func(id string) (int, map[string]any) {
		return postSeason(t, ts, "/api/v1/missions/seasons/"+id+"/finalize")
	}

	status, body := freeze("season-1")
	if status != http.StatusOK || body["State"] != string(domain.SeasonStateFreezing) {
		t.Fatalf("expected 200 FREEZING, got %d %+v", status, body)
	}

	status, body = freeze("season-2")
	if status != http.StatusUnprocessableEntity || body["code"] != "SEASON_STATE_INVALID" {
		t.Fatalf("expected 422 SEASON_STATE_INVALID, got %d %+v", status, body)
	}

	status, body = finalize("missing")
	if status != http.StatusNotFound || body["code"] != "SEASON_NOT_FOUND" {
		t.Fatalf("expected 404 SEASON_NOT_FOUND, got %d %+v", status, body)
	}

	status, body = finalize("season-3")
	if status != http.StatusUnprocessableEntity || body["code"] != "SEASON_STATE_INVALID" {
		t.Fatalf("expected 422 SEASON_STATE_INVALID, got %d %+v", status, body)
	}

	status, body = finalize("season-1")
	if status != http.StatusOK || body["State"] != string(domain.SeasonStateFinalized) {
		t.Fatalf("expected 200 FINALIZED, got %d %+v", status, body)
	}
}

func postSeason(t *testing.T, ts *httptest.Server, path string) (int, map[string]any) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	return resp.StatusCode, result
}
