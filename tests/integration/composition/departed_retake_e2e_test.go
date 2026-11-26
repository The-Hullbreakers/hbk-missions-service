package composition_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

type retakeOutcome struct {
	leaveStatus   int
	boardStatus   int
	boardOfferIDs []string
	retakeStatus  int
	retakeCode    string
}

// runRetakeJourney drives the leave -> board -> retake sequence and returns
// raw outcomes; assertions live in the test. OfferID decodes the PascalCase
// key emitted by MissionOffer.MarshalJSON.
func runRetakeJourney(baseURL, auth, contractID, offerID string) (retakeOutcome, error) {
	out := retakeOutcome{}
	call := func(method, path string) (*http.Response, error) {
		req, err := http.NewRequest(method, baseURL+path, bytes.NewReader([]byte(`{}`)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth)
		return http.DefaultClient.Do(req)
	}
	leave, err := call(http.MethodPost, "/api/v1/missions/contracts/"+contractID+"/leave")
	if err != nil {
		return out, err
	}
	defer leave.Body.Close()
	out.leaveStatus = leave.StatusCode
	board, err := call(http.MethodGet, "/api/v1/missions/offers")
	if err != nil {
		return out, err
	}
	defer board.Body.Close()
	out.boardStatus = board.StatusCode
	var page struct {
		Offers []struct {
			OfferID string `json:"OfferID"`
		} `json:"offers"`
	}
	if err := json.NewDecoder(board.Body).Decode(&page); err != nil {
		return out, fmt.Errorf("decode board: %w", err)
	}
	for _, o := range page.Offers {
		out.boardOfferIDs = append(out.boardOfferIDs, o.OfferID)
	}
	retake, err := call(http.MethodPost, "/api/v1/missions/offers/"+offerID+"/accept")
	if err != nil {
		return out, err
	}
	defer retake.Body.Close()
	out.retakeStatus = retake.StatusCode
	var retakeBody map[string]any
	if err := json.NewDecoder(retake.Body).Decode(&retakeBody); err != nil {
		return out, fmt.Errorf("decode retake: %w", err)
	}
	out.retakeCode, _ = retakeBody["code"].(string)
	return out, nil
}

// REG-0053-01 journey: after player-1 leaves the live team contract, the same
// definition reappears on their board and the retake succeeds — the departed
// contract must not consume their active slot (limit pinned at 1).
func TestTeamLeaveJourney_DepartedMemberRetakes(t *testing.T) {
	// Arrange
	client, table := chg0053Table(t, "e2e-retake")

	seasonStore := missiondynamo.NewSeasonStore(client, table)
	require.NoError(t, seasonStore.Save(domain.NewSeason("season-1", domain.SeasonStateActive)))
	offer := domain.NewMissionOffer("offer-x", "BOUNTY_HUNTING", "season-1")
	offer.DefinitionCode = "def-x"
	offer.DefinitionVersion = "legacy"
	offerStore := missiondynamo.NewOfferStore(client, table)
	require.NoError(t, offerStore.Save(offer))
	contract := domain.NewTeamMissionContract("contract-1", offer, []string{"player-1", "player-2"}, "team-9")
	member, err := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, err)
	require.NoError(t, contract.Start(member))
	contractStore := missiondynamo.NewContractStore(client, table)
	require.NoError(t, contractStore.Save(contract))
	policyStore := missiondynamo.NewPolicyStore(client, table)
	require.NoError(t, policyStore.SaveActiveLimit(domain.NewActiveLimitPolicy(1, nil)))
	seedPlayerActive(t, client, table, "player-1", "contract-1")
	seedPlayerReservation(t, client, table, "player-1", "def-x", "contract-1")
	seedPlayerReservation(t, client, table, "player-2", "def-x", "contract-1")

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:   table,
		DynamoDB:    client,
		UserSession: tokenSession{},
	})
	require.NoError(t, err)
	ts := httptest.NewServer(server)
	defer ts.Close()

	// Act
	outcome, actErr := runRetakeJourney(ts.URL, token("player-1"), "contract-1", "offer-x")

	// Assert
	require.NoError(t, actErr)
	require.Equal(t, http.StatusOK, outcome.leaveStatus)
	require.Equal(t, http.StatusOK, outcome.boardStatus)
	require.Equal(t, []string{"offer-x"}, outcome.boardOfferIDs)
	require.Equal(t, http.StatusCreated, outcome.retakeStatus, "retake rejected: %s", outcome.retakeCode)
}
