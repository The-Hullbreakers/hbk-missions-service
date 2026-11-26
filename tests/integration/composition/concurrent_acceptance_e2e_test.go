package composition_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// SCN-MB-CONCURRENT-ACCEPTANCE through the public API: concurrent accepts on
// an unlimited offer both succeed.
func TestConcurrentAcceptance_UnlimitedOfferBothSucceed(t *testing.T) {
	client, table := chg0053Table(t, "concurrent-unlimited")

	seasonStore := missiondynamo.NewSeasonStore(client, table)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	err := seasonStore.Save(season)
	require.NoError(t, err)
	offerStore := missiondynamo.NewOfferStore(client, table)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	err = offerStore.Save(offer)
	require.NoError(t, err)

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   table,
		DynamoDB:    client,
		UserSession: tokenSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	accept := func(tok string) (int, string) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/missions/offers/offer-1/accept", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", tok)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		code, _ := body["code"].(string)
		return resp.StatusCode, code
	}

	var wg sync.WaitGroup
	type result struct {
		status int
		code   string
	}
	results := make([]result, 2)
	for i, tok := range []string{token("player-1"), token("player-2")} {
		wg.Add(1)
		go func(tok string, n int) {
			defer wg.Done()
			status, code := accept(tok)
			results[n] = result{status, code}
		}(tok, i)
	}
	wg.Wait()

	created := 0
	for _, r := range results {
		if r.status == http.StatusCreated {
			created++
		}
	}
	require.Equal(t, 2, created)
}

// SCN-MB-SEASON-CLOSED through the public API: an offer bound to a FINALIZED
// season rejects with 422 SEASON_NOT_ACTIVE.
func TestConcurrentAcceptance_ClosedSeasonRejects(t *testing.T) {
	client, table := chg0053Table(t, "concurrent-closed-season")

	seasonStore := missiondynamo.NewSeasonStore(client, table)
	season := domain.NewSeason("season-1", domain.SeasonStateFinalized)
	err := seasonStore.Save(season)
	require.NoError(t, err)
	offerStore := missiondynamo.NewOfferStore(client, table)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	err = offerStore.Save(offer)
	require.NoError(t, err)

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   table,
		DynamoDB:    client,
		UserSession: tokenSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/missions/offers/offer-1/accept", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", token("player-1"))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.False(t, resp.StatusCode != http.StatusUnprocessableEntity || body["code"] != "SEASON_NOT_ACTIVE")
}
