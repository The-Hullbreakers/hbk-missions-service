package composition_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// Team leave journey E2E: a member leaves a live team contract, the contract
// survives, the leaver loses evidence rights, and the survivor confirms and
// completes — the team track still earns merit. The multi-member roster is
// seeded directly because HTTP accept always mints a single-member roster.
func TestTeamLeaveJourney_SurvivorCompletesAfterDeparture(t *testing.T) {
	client, table := chg0053Table(t, "team-leave-survivor")

	// Seed the season and a live 2-member team contract directly.
	seasonStore := missiondynamo.NewSeasonStore(client, table)
	err := seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive))
	require.NoError(t, err)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewTeamMissionContract("contract-1", offer, []string{"player-1", "player-2"}, "team-9")
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "Elimina atacantes", "", "BOUNTY_HUNTING", 100, 0, 0),
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	err = contract.Start(member)
	require.NoError(t, err)
	contractStore := missiondynamo.NewContractStore(client, table)
	err = contractStore.Save(contract)
	require.NoError(t, err)

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   table,
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

	leaver := token("player-1")
	survivor := token("player-2")

	// Member leaves — contract stays alive.
	if status := post(leaver, "/api/v1/missions/contracts/contract-1/leave", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "leave: expected 200, got %d", status)
	}

	// Departed member can no longer submit evidence.
	if status := post(leaver, "/api/v1/missions/contracts/contract-1/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-1","season_id":"season-1"}`); status != http.StatusForbidden {
		require.Failf(t, "unexpected failure", "departed evidence: expected 403, got %d", status)
	}

	// Survivor confirms the objective against frozen terms.
	if status := post(survivor, "/api/v1/missions/contracts/contract-1/evidence",
		`{"objective_id":"objective-1","unit_id":"unit-1","season_id":"season-1"}`); status != http.StatusCreated {
		require.Failf(t, "unexpected failure", "survivor evidence: expected 201, got %d", status)
	}

	// Survivor submits for review once all terms are confirmed; a moderator approves.
	if status := post(survivor, "/api/v1/missions/contracts/contract-1/submit", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "submit: expected 200, got %d", status)
	}
	if status := post(token("mod-1", "moderator"), "/api/v1/missions/contracts/contract-1/approve", `{}`); status != http.StatusOK {
		require.Failf(t, "unexpected failure", "approve: expected 200, got %d", status)
	}

	// The team track earned the evidence merit.
	req, _ := http.NewRequest(http.MethodGet,
		ts.URL+"/api/v1/missions/seasons/season-1/leaderboards/BOUNTY_HUNTING?track=TEAM", nil)
	req.Header.Set("Authorization", survivor)
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
	require.False(t, len(board.Items) != 1 || board.Items[0].RecipientID != "team-9" || board.Items[0].Total != 100)
}
