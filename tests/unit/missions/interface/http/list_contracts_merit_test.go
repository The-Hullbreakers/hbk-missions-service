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

// CHG-0047 / REG-0047 — a listed contract's embedded offer surfaces the
// frozen terms as objectives so merit_summary reports the earnable total.

func listContractsResponse(t *testing.T, contracts []domain.MissionContract) []map[string]any {
	t.Helper()
	reader := &fakePlayerContractsReaderForHTTP{contracts: contracts}
	app := application.NewListPlayerContractsHandler(reader)
	auth := missionhttp.NewAuthMiddleware(memberSession{})
	handler := auth.Authenticate(missionhttp.ListContractsHandler(app))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/contracts", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	return result
}

func TestListContracts_OfferMeritSummaryComesFromFrozenTerms(t *testing.T) {
	// Arrange: terms freeze 450x1 + 50x1 merits; the embedded offer was
	// persisted without objectives (current production shape).
	offer := domain.NewMissionOffer("offer-1", "PVP_AIR", "season-1")
	contract := domain.NewMissionContract("c1", offer, []string{"player-1"})
	terms, err := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("resultados-verificados", "Resultados PvP verificados", "", "PVP_AIR", 450, 0, 1),
		domain.NewObjectiveNamed("revision-de-evidencias", "Revisión de evidencias", "", "PVP_AIR", 50, 0, 1),
	})
	require.NoError(t, err)
	contract.Terms = terms

	// Act
	result := listContractsResponse(t, []domain.MissionContract{*contract})

	// Assert
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

func TestListContracts_OfferObjectivesTakePrecedenceOverTerms(t *testing.T) {
	// Arrange: an offer that already carries objectives keeps them; the
	// summary derives from the offer's own snapshot.
	offer := domain.NewMissionOffer("offer-1", "PVP_AIR", "season-1")
	offer.CompletionReward = 200
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("obj-a", "obj-a", "", "PVP_AIR", 10, 5, 3),
	}
	contract := domain.NewMissionContract("c1", offer, []string{"player-1"})

	// Act
	result := listContractsResponse(t, []domain.MissionContract{*contract})

	// Assert
	require.Len(t, result, 1)
	embedded, ok := result[0]["Offer"].(map[string]any)
	require.True(t, ok)
	objectives, ok := embedded["Objectives"].([]any)
	require.True(t, ok)
	require.Len(t, objectives, 1)
	assert.Equal(t, "obj-a", objectives[0].(map[string]any)["ObjectiveID"])
	summary := embedded["MeritSummary"].(map[string]any)
	assert.Equal(t, float64(45), summary["ObjectiveMerits"])
	assert.Equal(t, float64(245), summary["TotalMerits"])
}

func TestListContracts_ContractWithoutTermsDegradesToReward(t *testing.T) {
	// Arrange: legacy contract with no terms — summary falls back to the
	// frozen completion reward only.
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	offer.CompletionReward = 75
	contract := domain.NewMissionContract("c1", offer, []string{"player-1"})

	// Act
	result := listContractsResponse(t, []domain.MissionContract{*contract})

	// Assert
	require.Len(t, result, 1)
	embedded, ok := result[0]["Offer"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, embedded["Objectives"])
	summary := embedded["MeritSummary"].(map[string]any)
	assert.Equal(t, float64(0), summary["ObjectiveMerits"])
	assert.Equal(t, float64(75), summary["TotalMerits"])
}
