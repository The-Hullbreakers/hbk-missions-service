package composition_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/stretchr/testify/require"
)

// Partial-verdict E2E (critical path): a member stages two units of a
// unit-limited objective, the moderator approves while rejecting one — only
// the accepted unit lands on the leaderboard and the completion bonus is
// withheld because the objective never reached its limit.
func TestVerdictJourney_PartialApprovalPaysAcceptedUnitsOnly(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")
	require.NoError(t, seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)))
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.CompletionReward = 50
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "Elimina atacantes", "", "BOUNTY_HUNTING", 100, 0, 2),
	})
	require.NoError(t, err)
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, contract.Start(member))
	contractStore := missiondynamo.NewContractStore(client, "hbk-missions-composition")
	require.NoError(t, contractStore.Save(contract))

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: tokenSession{},
		Evidence:    &blobStoreFake{stored: map[string]bool{"evidence/contract-1/e1": true, "evidence/contract-1/e2": true}},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	post := func(tok, path, body string) int {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+path, bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", tok)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		return resp.StatusCode
	}

	player := token("player-1")
	moderator := token("mod-1", "moderator")

	for i, unit := range []string{"unit-1", "unit-2"} {
		media := fmt.Sprintf("evidence/contract-1/e%d", i+1)
		if status := post(player, "/api/v1/missions/contracts/contract-1/evidence",
			`{"objective_id":"objective-1","unit_id":"`+unit+`","season_id":"season-1","media_uri":"`+media+`"}`); status != http.StatusCreated {
			require.Failf(t, "unexpected failure", "evidence %s: expected 201, got %d", unit, status)
		}
	}
	if status := post(player, "/api/v1/missions/contracts/contract-1/submit", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "submit: expected 200, got %d", status)
	}

	// Moderator approves unit-1 and rejects unit-2 in one verdict.
	if status := post(moderator, "/api/v1/missions/contracts/contract-1/approve",
		`{"rejected_units":[{"objective_id":"objective-1","unit_id":"unit-2"}]}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "approve verdict: expected 200, got %d", status)
	}

	// Only the accepted unit pays: 100 merits, no 50-point completion bonus.
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
	require.Equal(t, 100, board.Items[0].Total)

	// The contract exposes the rejected unit for member audit.
	req, _ = http.NewRequest(http.MethodGet, ts.URL+"/api/v1/missions/contracts/contract-1", nil)
	req.Header.Set("Authorization", player)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var detail struct {
		State         string `json:"state"`
		RejectedUnits []struct {
			ObjectiveID string `json:"objective_id"`
			UnitID      string `json:"unit_id"`
		} `json:"rejected_units"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&detail)
	require.Equal(t, "COMPLETED", detail.State)
	require.Len(t, detail.RejectedUnits, 1)
	require.Equal(t, "unit-2", detail.RejectedUnits[0].UnitID)
}
