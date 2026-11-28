package composition_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// blobStoreFake simulates a blob backend: issued upload grants mark the key
// as stored (as a successful client upload would), Exists sees stored keys.
type blobStoreFake struct {
	stored map[string]bool
}

func (f *blobStoreFake) PresignUpload(key, _ string, _ int64) (application.UploadGrant, error) {
	f.stored[key] = true
	return application.UploadGrant{URL: "https://signed.example/upload?key=" + key, MediaURI: domain.MediaURI(key)}, nil
}

func (f *blobStoreFake) Exists(uri domain.MediaURI) (bool, error) {
	return f.stored[string(uri)], nil
}

func (f *blobStoreFake) PresignRead(uri domain.MediaURI) (application.ReadGrant, error) {
	return application.ReadGrant{URL: "https://signed.example/read?key=" + string(uri)}, nil
}

// Evidence media journey (critical path): upload grant → evidence with
// media_uri → reviewer read grant; members stay denied.
func TestEvidenceMediaJourney_EndToEnd(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:        "hbk-missions-composition",
		DynamoDB:         client,
		UserSession:      tokenSession{},
		Evidence:         &blobStoreFake{stored: map[string]bool{}},
		EvidenceMaxBytes: 50 << 20,
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
	reviewer := token("ops-1", "moderator")

	if status, body := post(admin, "/api/v1/seasons", `{"season_id":"season-1","official_statistics":true}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "create season: expected 201, got %d %v", status, body)
	}
	status, body := post(admin, "/api/v1/missions/definitions",
		`{"definition_code":"bounty-media","title":"Caza","description":"d","category":"BOUNTY_HUNTING","completion_reward":0,
		  "objectives":[{"objective_id":"objective-1","name":"o1","merit_per_unit":100,"bonus":0,"unit_limit":1}],
		  "translations":{"en":{"title":"EN","description":"EN","objectives":[{"objective_id":"objective-1","name":"EN"}]}}}`)
	require.Equal(t, http.StatusCreated, status)
	if status, body := post(admin, "/api/v1/missions/definitions/bounty-media/publish", `{"version":"v1"}`); status != http.StatusOK && status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "publish definition: expected success, got %d %v", status, body)
	}
	if status, body := post(admin, "/api/v1/missions/offers", `{"offer_id":"offer-1","definition_code":"bounty-media","definition_version":"v1"}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "publish offer: expected 201, got %d %v", status, body)
	}
	status, body = post(player, "/api/v1/missions/offers/offer-1/accept", `{}`)
	require.Equal(t, http.StatusCreated, status)
	contractID, _ := body["ContractID"].(string)
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/start", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "start: expected 200, got %d %v", status, body)
	}

	// A confirmation without a media reference is rejected before staging.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-0","season_id":"season-1"}`); status != http.StatusUnprocessableEntity || body["code"] != "EVIDENCE_REQUIRED" {
		require.Failf(t, "unexpected failure", "missing media: expected 422 EVIDENCE_REQUIRED, got %d %v", status, body)
	}

	// Request an upload grant.
	status, body = post(player, "/api/v1/missions/contracts/"+contractID+"/evidence/upload-url",
		`{"content_type":"video/mp4","content_length":1048576}`)
	require.Equal(t, http.StatusCreated, status)
	mediaURI, _ := body["media_uri"].(string)
	require.False(t, !strings.HasPrefix(mediaURI, "evidence/"+contractID+"/"))

	// A disallowed type is rejected before any grant.
	if status, _ := post(player, "/api/v1/missions/contracts/"+contractID+"/evidence/upload-url",
		`{"content_type":"application/x-msdownload","content_length":10}`); status != http.StatusBadRequest {
		require.Failf(t, "unexpected failure", "bad type: expected 400, got %d", status)
	}

	// Confirm evidence attaching the uploaded media reference.
	status, body = post(player, "/api/v1/missions/contracts/"+contractID+"/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-1","season_id":"season-1","media_uri":"`+mediaURI+`"}`)
	require.Equal(t, http.StatusCreated, status)
	got, _ := body["MediaURI"].(string)
	require.False(t, got != mediaURI)

	// A phantom reference is rejected.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-2","season_id":"season-1","media_uri":"evidence/`+contractID+`/ghost"}`); status != http.StatusUnprocessableEntity {
		require.Failf(t, "unexpected failure", "phantom media: expected 422, got %d %v", status, body)
	}

	// Reviewer gets a read grant; the roster member does not.
	if status, body := post(reviewer, "/api/v1/missions/contracts/"+contractID+"/evidence/media-url",
		`{"media_uri":"`+mediaURI+`"}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "media-url reviewer: expected 200, got %d %v", status, body)
	}
	if status, _ := post(player, "/api/v1/missions/contracts/"+contractID+"/evidence/media-url",
		`{"media_uri":"`+mediaURI+`"}`); status != http.StatusForbidden {
		require.Failf(t, "unexpected failure", "media-url member: expected 403, got %d", status)
	}

	// The media-backed contract enters review and is adjudicated.
	if status, body := post(player, "/api/v1/missions/contracts/"+contractID+"/submit", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "submit: expected 200, got %d %v", status, body)
	}
	if status, body := post(reviewer, "/api/v1/missions/contracts/"+contractID+"/approve", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "approve: expected 200, got %d %v", status, body)
	}
}
