package composition_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

func compositionServer(t *testing.T, client *missiondynamo.DynamoDBClient) *httptest.Server {
	t.Helper()
	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: fakeUserSession{},
	})
	require.NoError(t, err)
	return httptest.NewServer(server)
}

func postJSON(t *testing.T, url, body string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var payload map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return resp, payload
}

// SCN-MCF-TEAM-SOLO-REJECT end to end: a forged team_id on a SOLO offer is
// rejected at the real boundary with 422 TEAM_NOT_ALLOWED.
func TestCompositionServer_AcceptSoloOfferWithTeamIDRejected(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")
	err = seasonStore.Create(domain.NewSeason("season-1", domain.SeasonStateActive))
	require.NoError(t, err)
	offerStore := missiondynamo.NewOfferStore(client, "hbk-missions-composition")
	err = offerStore.Save(domain.NewMissionOffer("offer-solo-1", "BOUNTY_HUNTING", "season-1"))
	require.NoError(t, err)

	ts := compositionServer(t, client)
	defer ts.Close()

	resp, payload := postJSON(t, ts.URL+"/api/v1/missions/offers/offer-solo-1/accept", `{"team_id":"team-1"}`)
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.Equal(t, "TEAM_NOT_ALLOWED", payload["code"])
}

// SCN-SCC-CREATE-EXISTS end to end: creating an existing season_id is refused
// and the stored season keeps its original state and Official flag.
func TestCompositionServer_CreateSeasonRejectsExistingID(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")
	err = seasonStore.Create(domain.NewSeasonWithOfficial("season-x", domain.SeasonStateFinalized, true))
	require.NoError(t, err)

	ts := compositionServer(t, client)
	defer ts.Close()

	resp, payload := postJSON(t, ts.URL+"/api/v1/seasons", `{"season_id":"season-x","official_statistics":false}`)
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.Equal(t, "SEASON_ALREADY_EXISTS", payload["code"])

	stored, err := seasonStore.Get("season-x")
	require.NoError(t, err)
	require.False(t, stored.State != domain.SeasonStateFinalized || !stored.Official)
}
