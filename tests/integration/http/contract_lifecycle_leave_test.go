package http_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func teamContractFixture(t *testing.T) *domain.MissionContract {
	t.Helper()
	offer := domain.NewMissionOffer("offer-t1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewTeamMissionContract("contract-1", offer, []string{"player-1", "player-2"}, "team-9")
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	err := contract.Start(member)
	require.NoError(t, err)
	return contract
}

func TestContractLifecycle_LeaveKeepsContractAlive(t *testing.T) {
	store := &lifecycleContractStore{contract: teamContractFixture(t)}
	server := lifecycleServer("player-1", store)
	defer server.Close()

	resp := postLifecycle(t, server, "leave")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractInProgress)
	require.False(t, !store.saved.IsDeparted("player-1"))
}

func TestContractLifecycle_LeaveRejectsOutsider(t *testing.T) {
	store := &lifecycleContractStore{contract: teamContractFixture(t)}
	server := lifecycleServer("outsider", store)
	defer server.Close()

	resp := postLifecycle(t, server, "leave")
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestContractLifecycle_LeaveLastMemberAbandons(t *testing.T) {
	contract := teamContractFixture(t)
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	err := contract.Leave(member)
	require.NoError(t, err)
	store := &lifecycleContractStore{contract: contract}
	server := lifecycleServer("player-2", store)
	defer server.Close()

	resp := postLifecycle(t, server, "leave")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, store.saved == nil || store.saved.State != domain.ContractAbandoned)
}
