//go:build integration

package composition_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-E2E-LANG-001 — critical journey: a member reads an English offer,
// accepts it, and the resulting contract keeps both frozen locales even
// after the source definition/offer receives a v2 translation edit.
func TestLocalizedAcceptanceE2E_GivenBilingualOffer_WhenAccepted_ThenContractKeepsFrozenLocales(t *testing.T) {
	// Arrange — fresh test-owned table; ACTIVE season; bilingual published
	// definition and offer carrying both locales.
	ctx := t.Context()
	tableName := fmt.Sprintf("hbk-chg0054-e2e-%d", time.Now().UnixNano())
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	require.NoError(t, client.CreateMissionsTable(ctx, tableName))
	t.Logf("created test-owned table (left intact): %s", tableName)

	seasonStore := missiondynamo.NewSeasonStore(client, tableName)
	require.NoError(t, seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)))

	definitionStore := missiondynamo.NewDefinitionStore(client, tableName)
	definition, err := domain.NewMissionDefinition("def-lang", "Caza ES", "Descripción ES", "BOUNTY_HUNTING", domain.MissionModeSolo, 0,
		[]domain.Objective{domain.NewObjectiveNamed("obj-1", "Elimina ES", "", "BOUNTY_HUNTING", 100, 0, 1)}, nil)
	require.NoError(t, err)
	definition = englishize(definition)
	require.NoError(t, definition.Publish("v1"))
	require.NoError(t, definitionStore.Save(definition))

	offerStore := missiondynamo.NewOfferStore(client, tableName)
	offer, err := domain.NewMissionOfferFromDefinition("offer-lang-1", "season-1", definition)
	require.NoError(t, err)
	require.NoError(t, offerStore.Save(offer))

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName: tableName, DynamoDB: client, UserSession: fakeUserSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	// The member reads the English offer.
	status, offerBody := getJSON(t, ts.URL+"/api/v1/missions/offers/offer-lang-1?lang=en")
	require.Equal(t, http.StatusOK, status, "offer read: %v", offerBody)
	assert.Equal(t, "en", offerBody["ContentLanguage"])
	assert.Equal(t, "EN Caza ES", offerBody["Title"])

	// Act — the member accepts the offer.
	status, acceptBody := postJSONForTest(t, ts.URL+"/api/v1/missions/offers/offer-lang-1/accept", `{}`)
	require.Equal(t, http.StatusCreated, status, "accept: %v", acceptBody)
	contractID, _ := acceptBody["ContractID"].(string)
	require.NotEmpty(t, contractID)

	// Assert — the contract serves the frozen English snapshot.
	status, enContract := getJSON(t, ts.URL+"/api/v1/missions/contracts/"+contractID+"?lang=en")
	require.Equal(t, http.StatusOK, status, "en contract: %v", enContract)
	assert.Equal(t, "en", enContract["content_language"])
	assert.Equal(t, "EN Caza ES", enContract["title"])

	status, esContract := getJSON(t, ts.URL+"/api/v1/missions/contracts/"+contractID+"?lang=es")
	require.Equal(t, http.StatusOK, status, "es contract: %v", esContract)
	assert.Equal(t, "es", esContract["content_language"])
	assert.Equal(t, "Caza ES", esContract["title"])

	// A later translation edit on the stored definition and offer must not
	// reach the already-accepted contract.
	definition.ContentEn.Title = "RENAMED EN"
	require.NoError(t, definitionStore.Save(definition))
	edited, err := domain.NewMissionOfferFromDefinition("offer-lang-1", "season-1", definition)
	require.NoError(t, err)
	require.NoError(t, offerStore.Save(edited))

	status, after := getJSON(t, ts.URL+"/api/v1/missions/contracts/"+contractID+"?lang=en")
	require.Equal(t, http.StatusOK, status, "contract after v2: %v", after)
	assert.Equal(t, "EN Caza ES", after["title"], "v2 translation must not rewrite the frozen snapshot")
}

func getJSON(t *testing.T, url string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body
}

func postJSONForTest(t *testing.T, url, payload string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(payload)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body
}
