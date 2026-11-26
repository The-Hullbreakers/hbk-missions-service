package composition_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// Abandon journey E2E (SCN-MCF-ABANDON + SCN-MB-TEMPLATE-RELEASED): the
// player abandons a live contract — the contract leaves the ACTIVE bucket,
// the active-slot is freed, and the template's offer reappears on the
// board so a fresh attempt can be accepted.
func TestAbandonJourney_FreesSlotAndRevealsTemplate(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: tokenSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	post := func(tok, path, body string) (int, map[string]any) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+path, bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", tok)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var result map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&result)
		return resp.StatusCode, result
	}
	get := func(tok, path string) []map[string]any {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		req.Header.Set("Authorization", tok)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var result []map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&result)
		return result
	}
	getOffers := func(tok string) []map[string]any {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/missions/offers", nil)
		req.Header.Set("Authorization", tok)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var page struct {
			Offers []map[string]any `json:"offers"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&page)
		return page.Offers
	}

	admin := token("admin-1", "admin")
	player := token("player-1")

	// Admin publishes a template with a live offer.
	if status, body := post(admin, "/api/v1/seasons",
		`{"season_id":"season-1","official_statistics":true}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "create season: expected 201, got %d %v", status, body)
	}
	if status, body := post(admin, "/api/v1/missions/definitions",
		`{"definition_code":"bounty-template","title":"Caza","description":"Elimina","category":"BOUNTY_HUNTING",
		  "completion_reward":200,
		  "objectives":[{"objective_id":"objective-1","name":"Elimina atacantes","merit_per_unit":100,"bonus":25,"unit_limit":2}],
		  "translations":{"en":{"title":"EN","description":"EN","objectives":[{"objective_id":"objective-1","name":"EN"}]}}}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "create definition: expected 201, got %d %v", status, body)
	}
	if status, body := post(admin, "/api/v1/missions/definitions/bounty-template/publish",
		`{"version":"v1"}`); status != http.StatusOK && status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "publish definition: expected success, got %d %v", status, body)
	}
	if status, body := post(admin, "/api/v1/missions/offers",
		`{"offer_id":"offer-1","definition_code":"bounty-template","definition_version":"v1"}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "publish offer: expected 201, got %d %v", status, body)
	}

	// Player accepts and starts the contract — the template hides.
	status, body := post(player, "/api/v1/missions/offers/offer-1/accept", `{}`)
	require.Equal(t, http.StatusCreated, status)
	contractID, _ := body["ContractID"].(string)
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/start", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "start: expected 200, got %d %v", status, body)
	}
	for _, offer := range getOffers(player) {
		require.NotEqual(t, offer["DefinitionCode"], "bounty-template")
	}

	// The player abandons the live contract.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/abandon", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "abandon: expected 200, got %d %v", status, body)
	}

	// Slot released: no ACTIVE contract remains for the caller.
	contracts := get(player, "/api/v1/missions/contracts?status=ACTIVE")
	require.False(t, len(contracts) != 0)
	// The abandoned contract still appears in history with terminal state.
	history := get(player, "/api/v1/missions/contracts?status=ABANDONED")
	require.False(t, len(history) != 1 || history[0]["ContractID"] != contractID)
	// The template is visible again and retakeable.
	visible := false
	for _, offer := range getOffers(player) {
		if offer["DefinitionCode"] == "bounty-template" {
			visible = true
		}
	}
	require.True(t, visible, "expected bounty-template to reappear on the board")
	status, body = post(player, "/api/v1/missions/offers/offer-1/accept", `{}`)
	require.Equal(t, http.StatusCreated, status)
	newID, _ := body["ContractID"].(string)
	require.False(t, newID == "" || newID == contractID)
}
