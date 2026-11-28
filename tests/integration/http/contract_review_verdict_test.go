package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// postApprove sends the approve call with an optional JSON verdict body.
func postApprove(t *testing.T, server *httptest.Server, body any) *http.Response {
	t.Helper()
	var payload *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		payload = bytes.NewReader(raw)
	} else {
		payload = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/contracts/contract-1/approve", payload)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// pendingUnitsContract stages the given units under a limited objective and
// submits the contract for review.
func pendingUnitsContract(t *testing.T, reward, unitLimit int, units ...string) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	offer.CompletionReward = reward
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 54, 0, unitLimit),
	})
	require.NoError(t, err)
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	require.NoError(t, contract.Start(member))
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	for _, unit := range units {
		media, err := domain.NewMediaURI("evidence/contract-1/"+unit, "contract-1")
		require.NoError(t, err)
		_, err = contract.ConfirmWithMedia("objective-1", unit, media, season)
		require.NoError(t, err)
	}
	require.NoError(t, contract.Submit(member))
	return contract
}

func TestContractReview_ApproveWithRejectedUnitsCommitsSubset(t *testing.T) {
	store := &reviewContractStore{contract: pendingUnitsContract(t, 0, 3, "u1", "u2", "u3")}
	server := reviewServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, store)
	defer server.Close()

	resp := postApprove(t, server, map[string]any{
		"rejected_units": []map[string]string{
			{"objective_id": "objective-1", "unit_id": "u2"},
		},
	})
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var credits []domain.MeritCredit
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&credits))
	require.Len(t, credits, 2)
	for _, credit := range credits {
		require.NotEqual(t, "u2", credit.UnitID)
	}
	require.Equal(t, domain.ContractCompleted, store.saved.State)
	require.Len(t, store.saved.RejectedUnits(), 1)
}

func TestContractReview_ApproveUnknownRejectedUnitReturns422(t *testing.T) {
	store := &reviewContractStore{contract: pendingUnitsContract(t, 0, 3, "u1")}
	server := reviewServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, store)
	defer server.Close()

	resp := postApprove(t, server, map[string]any{
		"rejected_units": []map[string]string{
			{"objective_id": "objective-1", "unit_id": "ghost"},
		},
	})
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, "UNIT_NOT_STAGED", body["code"])
	require.Nil(t, store.saved)
}

func TestContractReview_ApproveNoBodyApprovesAll(t *testing.T) {
	store := &reviewContractStore{contract: pendingUnitsContract(t, 60, 2, "u1", "u2")}
	server := reviewServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, store)
	defer server.Close()

	resp := postApprove(t, server, nil)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var credits []domain.MeritCredit
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&credits))
	require.Len(t, credits, 3)
}

func TestContractReview_DetailExposesRejectedUnits(t *testing.T) {
	contract := pendingUnitsContract(t, 0, 3, "u1", "u2")
	_, err := contract.ApproveWithVerdict([]domain.UnitRef{
		{ObjectiveID: "objective-1", UnitID: "u2"},
	})
	require.NoError(t, err)

	reader := &detailReader{contract: contract}
	server := detailServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, reader)
	defer server.Close()

	resp := getContractDetail(t, server, "contract-1")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var detail struct {
		RejectedUnits []struct {
			ObjectiveID string `json:"objective_id"`
			UnitID      string `json:"unit_id"`
		} `json:"rejected_units"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	require.Len(t, detail.RejectedUnits, 1)
	require.Equal(t, "u2", detail.RejectedUnits[0].UnitID)
}
