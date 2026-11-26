package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type lifecycleContractStore struct {
	contract *domain.MissionContract
	saved    *domain.MissionContract
}

func (s *lifecycleContractStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }
func (s *lifecycleContractStore) Save(c *domain.MissionContract) error        { s.saved = c; return nil }

func (s *lifecycleContractStore) SaveDeparture(c *domain.MissionContract, _ string, _ domain.ContractState) error {
	return s.Save(c)
}

func lifecycleServer(playerID string, store *lifecycleContractStore) *httptest.Server {
	start := application.NewStartContractHandler(store, store)
	abandon := application.NewAbandonContractHandler(store, store)
	leave := application.NewLeaveContractHandler(store, store)
	auth := missionhttp.NewAuthMiddleware(boundarySession{playerID: playerID})

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/start", auth.Authenticate(missionhttp.StartContractHandler(start)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/abandon", auth.Authenticate(missionhttp.AbandonContractHandler(abandon)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/leave", auth.Authenticate(missionhttp.LeaveContractHandler(leave)))
	return httptest.NewServer(mux)
}

func lifecycleContractFixture() *domain.MissionContract {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)})
	contract.Terms = terms
	return contract
}

func postLifecycle(t *testing.T, server *httptest.Server, action string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost,
		server.URL+"/api/v1/missions/contracts/contract-1/"+action, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestContractLifecycle_StartEndpointTransitionsContract(t *testing.T) {
	store := &lifecycleContractStore{contract: lifecycleContractFixture()}
	server := lifecycleServer("player-1", store)
	defer server.Close()

	resp := postLifecycle(t, server, "start")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractInProgress)
}

func TestContractLifecycle_StartEndpointRejectsOutsider(t *testing.T) {
	store := &lifecycleContractStore{contract: lifecycleContractFixture()}
	server := lifecycleServer("outsider", store)
	defer server.Close()

	resp := postLifecycle(t, server, "start")
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestContractLifecycle_AbandonEndpointTerminates(t *testing.T) {
	store := &lifecycleContractStore{contract: lifecycleContractFixture()}
	server := lifecycleServer("player-1", store)
	defer server.Close()

	resp := postLifecycle(t, server, "abandon")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractAbandoned)
}
