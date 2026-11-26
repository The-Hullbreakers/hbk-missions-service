package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// detailReader is a ContractReader double returning a fixture or error.
type detailReader struct {
	contract *domain.MissionContract
	err      error
	calls    int
}

func (r *detailReader) Get(string) (*domain.MissionContract, error) {
	r.calls++
	return r.contract, r.err
}

// detailServer mirrors the composition wiring for the detail route.
func detailServer(session boundarySession, reader *detailReader) *httptest.Server {
	app := application.NewGetContractHandler(reader)
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/contracts/{contract_id}",
		auth.Authenticate(missionhttp.ContractDetailHandler(app)))
	return httptest.NewServer(mux)
}

// getContractDetail performs the authenticated detail request.
func getContractDetail(t *testing.T, server *httptest.Server, contractID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/v1/missions/contracts/"+contractID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// pendingContractWithMedia submits a contract whose staged credit carries a
// media reference — the reviewer detail must expose it.
func pendingContractWithMedia(t *testing.T) *domain.MissionContract {
	t.Helper()
	contract := lifecycleContractFixture()
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	media, err := domain.NewMediaURI("evidence/contract-1/abc-123", "contract-1")
	require.NoError(t, err)
	_, err = contract.ConfirmWithMedia("objective-1", "unit-1", media, season)
	require.NoError(t, err)
	err = contract.Submit(member)
	require.NoError(t, err)
	return contract
}

// contractDetail is the reviewer-facing detail projection.
type contractDetail struct {
	ContractID     string         `json:"contract_id"`
	OfferID        string         `json:"offer_id"`
	DefinitionCode string         `json:"definition_code"`
	Title          string         `json:"title"`
	State          string         `json:"state"`
	Roster         []string       `json:"roster"`
	SeasonID       string         `json:"season_id"`
	TeamID         string         `json:"team_id"`
	Confirmations  []string       `json:"confirmations"`
	PendingCredits []detailCredit `json:"pending_credits"`
}

type detailCredit struct {
	ObjectiveID string `json:"objective_id"`
	UnitID      string `json:"unit_id"`
	RecipientID string `json:"recipient_id"`
	Amount      int    `json:"amount"`
	MediaURI    string `json:"media_uri"`
}

func TestContractDetail_ReviewerGetsDetail(t *testing.T) {
	reader := &detailReader{contract: pendingContractWithMedia(t)}
	server := detailServer(boundarySession{playerID: "mod-1", roles: []any{"moderator"}}, reader)
	defer server.Close()

	resp := getContractDetail(t, server, "contract-1")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var detail contractDetail
	err := json.NewDecoder(resp.Body).Decode(&detail)
	require.NoError(t, err)
	require.False(t, detail.ContractID != "contract-1" || detail.State != "PENDING_VALIDATION")
	require.Len(t, detail.PendingCredits, 1)
	credit := detail.PendingCredits[0]
	require.False(t, credit.UnitID != "unit-1" || credit.MediaURI != "evidence/contract-1/abc-123")
}

// A roster member reads their contract with a minimized projection — own
// pending credits only and no media references.
func TestContractDetail_MemberGetsMinimizedDetail(t *testing.T) {
	reader := &detailReader{contract: pendingContractWithMedia(t)}
	server := detailServer(boundarySession{playerID: "player-1"}, reader)
	defer server.Close()

	resp := getContractDetail(t, server, "contract-1")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var detail contractDetail
	err := json.NewDecoder(resp.Body).Decode(&detail)
	require.NoError(t, err)
	require.False(t, detail.ContractID != "contract-1" || detail.State != "PENDING_VALIDATION")
	require.False(t, len(detail.PendingCredits) != 1 || detail.PendingCredits[0].RecipientID != "player-1")
	require.Empty(t, detail.PendingCredits[0].MediaURI)
}

func TestContractDetail_OutsiderDenied(t *testing.T) {
	reader := &detailReader{contract: pendingContractWithMedia(t)}
	server := detailServer(boundarySession{playerID: "player-9"}, reader)
	defer server.Close()

	resp := getContractDetail(t, server, "contract-1")
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestContractDetail_UnknownContract(t *testing.T) {
	reader := &detailReader{err: domain.ErrContractNil}
	server := detailServer(boundarySession{playerID: "mod-1", roles: []any{"admin"}}, reader)
	defer server.Close()

	resp := getContractDetail(t, server, "contract-9")
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "CONTRACT_NOT_FOUND", body["code"])
}
