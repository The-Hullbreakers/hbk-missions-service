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

// Moderation journey E2E (critical path): a member confirms evidence — the
// merit stays staged and the leaderboard is empty — submits for review, loses
// evidence and self-adjudication rights while pending, and a moderator rejects
// the contract. Rejection writes zero ledger credits.
func TestModerationJourney_RejectPaysNothing(t *testing.T) {
	ctx := t.Context()
	client, err := missiondynamo.NewDynamoDBClient(ctx, "us-east-1", dynamodbTestEndpoint())
	require.NoError(t, err)
	err = client.CreateMissionsTable(ctx, "hbk-missions-composition")
	require.NoError(t, err)
	clearMissionsTable(t, ctx, client, "hbk-missions-composition")

	seasonStore := missiondynamo.NewSeasonStore(client, "hbk-missions-composition")
	err = seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive))
	require.NoError(t, err)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "Elimina atacantes", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	err = contract.Start(member)
	require.NoError(t, err)
	contractStore := missiondynamo.NewContractStore(client, "hbk-missions-composition")
	err = contractStore.Save(contract)
	require.NoError(t, err)

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   "hbk-missions-composition",
		DynamoDB:    client,
		UserSession: tokenSession{},
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
	leaderboardTotal := func() int {
		req, _ := http.NewRequest(http.MethodGet,
			ts.URL+"/api/v1/missions/seasons/season-1/leaderboards/BOUNTY_HUNTING?track=INDIVIDUAL", nil)
		req.Header.Set("Authorization", token("player-1"))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var board struct {
			Items []struct {
				Total int `json:"total"`
			} `json:"items"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&board)
		if len(board.Items) == 0 {
			return 0
		}
		return board.Items[0].Total
	}

	player := token("player-1")
	moderator := token("mod-1", "moderator")

	// Evidence is staged, not paid — the leaderboard stays empty until review.
	if status := post(player, "/api/v1/missions/contracts/contract-1/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-1","season_id":"season-1"}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "evidence: expected 201, got %d", status)
	}
	total := leaderboardTotal()
	require.False(t, total != 0)

	// Member submits for review; pending contracts close evidence intake.
	if status := post(player, "/api/v1/missions/contracts/contract-1/submit", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "submit: expected 200, got %d", status)
	}
	if status := post(player, "/api/v1/missions/contracts/contract-1/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-2","season_id":"season-1"}`); status != http.StatusUnprocessableEntity {
		require.Failf(t, "unexpected failure", "pending evidence: expected 422, got %d", status)
	}

	// The member cannot adjudicate their own contract.
	if status := post(player, "/api/v1/missions/contracts/contract-1/approve", `{}`); status != http.StatusForbidden {
		require.Failf(t, "unexpected failure", "self-approve: expected 403, got %d", status)
	}

	// Moderator rejects — the staged credit is discarded, never ledgered.
	if status := post(moderator, "/api/v1/missions/contracts/contract-1/reject", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "reject: expected 200, got %d", status)
	}
	total = leaderboardTotal()
	require.False(t, total != 0)

	// Terminal contract rejects further adjudication.
	if status := post(moderator, "/api/v1/missions/contracts/contract-1/approve", `{}`); status != http.StatusUnprocessableEntity {
		require.Failf(t, "unexpected failure", "terminal approve: expected 422, got %d", status)
	}
}
