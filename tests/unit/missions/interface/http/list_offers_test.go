package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

func TestListOffersHandler_ReturnsAvailableOffers(t *testing.T) {
	handler := newListOffersHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	result := decodeOfferPage(t, rec)
	require.Len(t, result.Offers, 2)
	require.Equal(t, 2, result.Total)
	require.Equal(t, domain.OfferListDefaultLimit, result.Limit)
}

func TestListOffersHandler_ReturnsRulesAndMarkdownDescription(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "PVP_AIR", "season-2026-q4")
	offer.Description = "## Requisitos\n\n- Solo PU/LIVE"
	offer.Rules = []string{"Solo PU/LIVE", "Video continuo"}
	app := application.NewListOffersHandler(
		&fakeOfferListReader{offers: []domain.MissionOffer{offer}},
		&fakePlayerContractsReader{},
		&fakeActiveSeasons{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}},
	)
	handler := missionhttp.ListOffersHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	result := decodeOfferPage(t, rec)
	require.Len(t, result.Offers, 1)
	require.Equal(t, "## Requisitos\n\n- Solo PU/LIVE", result.Offers[0]["Description"])
	require.Equal(t, []any{"Solo PU/LIVE", "Video continuo"}, result.Offers[0]["Rules"])
}

func TestListOffersHandler_FiltersByCategory(t *testing.T) {
	handler := newListOffersHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers?category=BOUNTY", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	result := decodeOfferPage(t, rec)
	require.False(t, len(result.Offers) != 1 || result.Offers[0]["OfferID"] != "offer-1")
}

func TestListOffersHandler_PaginatesWithEnvelope(t *testing.T) {
	handler := newListOffersHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers?limit=1&offset=1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	result := decodeOfferPage(t, rec)
	require.Equal(t, 2, result.Total)
	require.Equal(t, 1, result.Limit)
	require.Equal(t, 1, result.Offset)
	require.Len(t, result.Offers, 1)
}

func TestListOffersHandler_RejectsUnknownStatus(t *testing.T) {
	handler := newListOffersHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers?status=CLOSED", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, "INVALID_QUERY_PARAMETER", body["code"])
}

func TestListOffersHandler_RejectsInvalidSort(t *testing.T) {
	handler := newListOffersHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers?sort=price", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, "INVALID_QUERY_PARAMETER", body["code"])
}
