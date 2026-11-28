package composition_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// Solo contract journey E2E (critical path): definition with objectives →
// publish offer → accept freezes terms → start → evidence confirms against
// frozen terms → complete only once every objective is confirmed.
func TestSoloContractJourney_FrozenTermsEndToEnd(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	blobs := &blobStoreFake{stored: map[string]bool{}}
	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: tokenSession{},
		Evidence:    blobs,
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

	admin := token("admin-1", "admin")
	player := token("player-1")

	// Admin creates a season and a definition carrying the reward rules.
	if status, body := post(admin, "/api/v1/seasons",
		`{"season_id":"season-1","official_statistics":true}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "create season: expected 201, got %d %v", status, body)
	}
	status, body := post(admin, "/api/v1/missions/definitions",
		`{"definition_code":"bounty-template","title":"Caza de recompensas","description":"Elimina al objetivo",
		  "category":"BOUNTY_HUNTING","completion_reward":200,
		  "objectives":[{"objective_id":"objective-1","name":"Elimina atacantes","merit_per_unit":100,"bonus":25,"unit_limit":2}],
		  "translations":{"en":{"title":"EN","description":"EN","objectives":[{"objective_id":"objective-1","name":"EN"}]}}}`)
	require.Equal(t, http.StatusCreated, status)
	if status, body := post(admin, "/api/v1/missions/definitions/bounty-template/publish",
		`{"version":"v1"}`); status != http.StatusOK && status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "publish definition: expected success, got %d %v", status, body)
	}
	if status, body := post(admin, "/api/v1/missions/offers",
		`{"offer_id":"offer-1","definition_code":"bounty-template","definition_version":"v1"}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "publish offer: expected 201, got %d %v", status, body)
	}

	// Player accepts — the contract freezes the definition's objectives.
	status, body = post(player, "/api/v1/missions/offers/offer-1/accept", `{}`)
	require.Equal(t, http.StatusCreated, status)
	contractID, _ := body["ContractID"].(string)
	require.NotEmpty(t, contractID)
	terms, _ := body["Terms"].([]any)
	require.Len(t, terms, 1)

	// Start the contract as the roster member.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/start", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "start: expected 200, got %d %v", status, body)
	}

	// Evidence pays merit_per_unit + bonus from frozen terms; a stored media
	// reference is required for every confirmation.
	mediaURI := "evidence/" + contractID + "/e1"
	blobs.stored[mediaURI] = true
	status, body = post(player, "/api/v1/missions/contracts/"+contractID+"/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-1","season_id":"season-1","media_uri":"`+mediaURI+`"}`)
	require.Equal(t, http.StatusCreated, status)
	amount, _ := body["Amount"].(float64)
	require.False(t, amount != 125)

	// An objective outside the frozen terms is rejected.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/evidence",
		`{"objective_id":"not-in-terms","unit_id":"unit-9","season_id":"season-1","media_uri":"`+mediaURI+`"}`); status != http.StatusUnprocessableEntity {
		require.Failf(t, "unexpected failure", "foreign objective: expected 422, got %d %v", status, body)
	}

	// The objective's unit_limit is two: the second unit completes the terms.
	status, body = post(player, "/api/v1/missions/contracts/"+contractID+"/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-2","season_id":"season-1","media_uri":"`+mediaURI+`"}`)
	require.Equal(t, http.StatusCreated, status)

	// All terms confirmed → the member submits for review and the admin approves.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/submit", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "submit: expected 200, got %d %v", status, body)
	}
	if status, body := post(admin, "/api/v1/missions/contracts/"+contractID+"/approve", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "approve: expected 200, got %d %v", status, body)
	}

	// A second approval on a terminal contract is rejected.
	if status, body := post(admin, "/api/v1/missions/contracts/"+contractID+"/approve", `{}`); status != http.StatusUnprocessableEntity {
		require.Failf(t, "unexpected failure", "repeat approve: expected 422, got %d %v", status, body)
	}

	// The season standing reflects two unit credits (125 × 2) + completion
	// reward (200) — the bonus only lands when every unit is confirmed.
	req, _ := http.NewRequest(http.MethodGet,
		ts.URL+"/api/v1/missions/seasons/season-1/leaderboards/BOUNTY_HUNTING?track=INDIVIDUAL", nil)
	req.Header.Set("Authorization", player)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var board struct {
		Items []struct {
			RecipientID string `json:"recipient_id"`
			Total       int    `json:"total"`
		} `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&board)
	require.False(t, len(board.Items) != 1 || board.Items[0].RecipientID != "player-1")
	require.Equal(t, 450, board.Items[0].Total)
}
