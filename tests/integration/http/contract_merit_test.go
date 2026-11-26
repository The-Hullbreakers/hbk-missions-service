package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// CHG-0047 / REG-0047 — boundary evidence: the member-facing contract list
// exposes the frozen terms as offer objectives with the earnable
// merit_summary.

type meritContractsReader struct{ contracts []domain.MissionContract }

func (r *meritContractsReader) ListByPlayer(string, domain.ContractListFilter) ([]domain.MissionContract, error) {
	return r.contracts, nil
}

func TestListContractsEndpoint_OfferMeritSummaryComesFromFrozenTerms(t *testing.T) {
	// Arrange: persisted offer has no objectives; terms pay 450x1 + 50x1.
	offer := domain.NewMissionOffer("offer-1", "PVP_AIR", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("resultados-verificados", "Resultados PvP verificados", "", "PVP_AIR", 450, 0, 1),
		domain.NewObjectiveNamed("revision-de-evidencias", "Revisión de evidencias", "", "PVP_AIR", 50, 0, 1),
	})
	require.NoError(t, err)
	contract.Terms = terms

	app := application.NewListPlayerContractsHandler(&meritContractsReader{contracts: []domain.MissionContract{*contract}})
	auth := missionhttp.NewAuthMiddleware(boundarySession{playerID: "player-1"})
	server := httptest.NewServer(auth.Authenticate(missionhttp.ListContractsHandler(app)))
	defer server.Close()

	// Act
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=ALL", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var result []map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Len(t, result, 1)
	embedded, ok := result[0]["Offer"].(map[string]any)
	require.True(t, ok)
	objectives, ok := embedded["Objectives"].([]any)
	require.True(t, ok)
	assert.Len(t, objectives, 2)
	summary, ok := embedded["MeritSummary"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(500), summary["ObjectiveMerits"])
	assert.Equal(t, float64(0), summary["CompletionReward"])
	assert.Equal(t, float64(500), summary["TotalMerits"])
}
