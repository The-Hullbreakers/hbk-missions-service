package composition_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func TestCompositionServer_PublishMissionOffer(t *testing.T) {
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
	definitionStore := missiondynamo.NewDefinitionStore(client, "hbk-missions")

	if err := seasonStore.Save(domain.NewSeason("season-2026-q4", domain.SeasonStateActive)); err != nil {
		t.Fatalf("could not seed season: %v", err)
	}

	definition, _ := domain.NewMissionDefinition("bounty-template", "BOUNTY_HUNTING", 10, domain.MissionModeTeam)
	_ = definition.Publish("v1")
	if err := definitionStore.Save(definition); err != nil {
		t.Fatalf("could not seed definition: %v", err)
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

	body := `{"tenant_id":"tenant-1","offer_id":"offer-1","definition_id":"bounty-template","definition_version":"v1","capacity":10}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/missions/offers", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("publish request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if result["OfferID"] != "offer-1" || result["Mode"] != "TEAM" || result["DefinitionID"] != "bounty-template" {
		t.Fatalf("unexpected offer in response: %+v", result)
	}
}
