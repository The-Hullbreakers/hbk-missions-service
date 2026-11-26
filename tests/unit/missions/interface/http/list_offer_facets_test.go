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

func newFacetsHandler(offers application.OfferListReader, seasons application.ActiveSeasonsReader) http.Handler {
	return missionhttp.ListOfferFacetsHandler(application.NewListOfferFacetsHandler(offers, seasons))
}

func getFacets(handler http.Handler, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func facetErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body["code"]
}

// REQ-MB-009 facets: a GET over two offers exposes the distinct, sorted
// category and mode values for the resolved active season.
func TestListOfferFacetsHandler_ExposesDistinctFacets(t *testing.T) {
	// Arrange
	offers := &fakeOfferListReader{offers: []domain.MissionOffer{
		domain.NewMissionOfferWithMode("offer-1", "BOUNTY", domain.MissionModeSolo, "season-2026-q4"),
		domain.NewMissionOfferWithMode("offer-2", "EXPLORATION", domain.MissionModeTeam, "season-2026-q4"),
	}}
	seasons := &fakeActiveSeasons{seasons: []domain.Season{
		domain.NewSeason("season-2026-q4", domain.SeasonStateActive),
	}}
	handler := newFacetsHandler(offers, seasons)

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets")

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	var facets domain.OfferFacets
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&facets))
	require.Equal(t, []string{"BOUNTY", "EXPLORATION"}, facets.Categories)
	require.Equal(t, []domain.MissionMode{domain.MissionModeSolo, domain.MissionModeTeam}, facets.Modes)
}

// An unrecognized status value is a query-parameter rejection, not a domain
// failure: 400 INVALID_QUERY_PARAMETER before any reader is consulted.
func TestListOfferFacetsHandler_InvalidStatusRejects(t *testing.T) {
	// Arrange
	handler := newFacetsHandler(&fakeOfferListReader{}, &fakeActiveSeasons{})

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets?status=INVALID")

	// Assert
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "INVALID_QUERY_PARAMETER", facetErrorCode(t, rec))
}

// No active season surfaces the board's 422 NO_ACTIVE_SEASON contract.
func TestListOfferFacetsHandler_NoActiveSeasonRejects(t *testing.T) {
	// Arrange
	handler := newFacetsHandler(&fakeOfferListReader{}, &fakeActiveSeasons{})

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets")

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.Equal(t, "NO_ACTIVE_SEASON", facetErrorCode(t, rec))
}

// An unclassified offer-reader failure stays a plain 500 INTERNAL_ERROR.
func TestListOfferFacetsHandler_ReaderErrorIsInternal(t *testing.T) {
	// Arrange
	offers := &failingOfferListReader{}
	seasons := &fakeActiveSeasons{seasons: []domain.Season{
		domain.NewSeason("season-2026-q4", domain.SeasonStateActive),
	}}
	handler := newFacetsHandler(offers, seasons)

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets")

	// Assert
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "INTERNAL_ERROR", facetErrorCode(t, rec))
}
