package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
	"github.com/stretchr/testify/require"
)

func invitationServer(playerID string, store *lifecycleContractStore) *httptest.Server {
	confirm := application.NewConfirmInvitationHandler(store, store)
	decline := application.NewDeclineInvitationHandler(store, store)
	auth := missionhttp.NewAuthMiddleware(boundarySession{playerID: playerID})

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/confirm-invitation",
		auth.Authenticate(missionhttp.ConfirmInvitationHandler(confirm)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/decline-invitation",
		auth.Authenticate(missionhttp.DeclineInvitationHandler(decline)))
	return httptest.NewServer(mux)
}

func invitedContractFixture() *domain.MissionContract {
	offer := domain.NewTeamMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	return domain.NewTeamMissionContract("contract-1", offer, []string{"player-1", "player-2"}, "team-1")
}

func postInvitation(t *testing.T, server *httptest.Server, action string) (*http.Response, map[string]string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/contracts/contract-1/"+action, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var payload map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	defer resp.Body.Close()
	return resp, payload
}

// SCN-MCF-INVITE-CONFIRM over HTTP: pending invitee confirms → 200.
func TestConfirmInvitationEndpoint_InviteeConfirms(t *testing.T) {
	store := &lifecycleContractStore{contract: invitedContractFixture()}
	server := invitationServer("player-2", store)
	defer server.Close()

	resp, _ := postInvitation(t, server, "confirm-invitation")
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// A non-invitee is rejected with 403 NOT_INVITED.
func TestConfirmInvitationEndpoint_OutsiderForbidden(t *testing.T) {
	store := &lifecycleContractStore{contract: invitedContractFixture()}
	server := invitationServer("player-9", store)
	defer server.Close()

	resp, payload := postInvitation(t, server, "confirm-invitation")
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	require.Equal(t, "NOT_INVITED", payload["code"])
}

// Already-resolved invitations answer 422 INVITATION_NOT_PENDING.
func TestDeclineInvitationEndpoint_AlreadyResolved(t *testing.T) {
	contract := invitedContractFixture()
	identity, _ := domain.NewCallerIdentity("player-2", nil, nil)
	err := contract.DeclineInvitation(identity, time.Now())
	require.NoError(t, err)
	store := &lifecycleContractStore{contract: contract}
	server := invitationServer("player-2", store)
	defer server.Close()

	resp, payload := postInvitation(t, server, "decline-invitation")
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.Equal(t, "INVITATION_NOT_PENDING", payload["code"])
}
