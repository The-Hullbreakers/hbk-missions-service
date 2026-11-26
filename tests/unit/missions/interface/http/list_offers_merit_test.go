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

func TestListOffersHandler_ExposesObjectivesAndMeritSummary(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-air", "PVP_AIR", "season-2026-q4")
	offer.DefinitionCode = "pvp-live-air-10"
	offer.Objectives = []domain.Objective{
		domain.NewObjectiveNamed("pvp-kills", "Derribos verificados", "Valida el paquete", "PVP_AIR", 270, 0, 1),
		domain.NewObjectiveNamed("evidence-review", "Revisión de evidencias", "", "PVP_AIR", 30, 0, 1),
	}
	app := application.NewListOffersHandler(
		&fakeOfferListReader{offers: []domain.MissionOffer{offer}},
		&fakePlayerContractsReader{},
		&fakeActiveSeasons{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}},
	)
	handler := missionhttp.ListOffersHandler(app)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	got := decodeOfferPage(t, rec).Offers[0]
	objectives, ok := got["Objectives"].([]any)
	require.True(t, ok, "offer must expose structured objectives")
	require.Len(t, objectives, 2)
	first, ok := objectives[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "pvp-kills", first["ObjectiveID"])
	require.Equal(t, "Derribos verificados", first["Name"])
	require.Equal(t, float64(270), first["MeritPerUnit"])
	require.Equal(t, float64(1), first["UnitLimit"])
	summary, ok := got["MeritSummary"].(map[string]any)
	require.True(t, ok, "offer must expose a merit summary")
	require.Equal(t, float64(300), summary["ObjectiveMerits"])
	require.Equal(t, float64(0), summary["CompletionReward"])
	require.Equal(t, float64(300), summary["TotalMerits"])
}
