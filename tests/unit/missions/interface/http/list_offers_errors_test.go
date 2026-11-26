package http_test

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type failingOfferListReader struct{}

func (r *failingOfferListReader) List(domain.OfferListFilter) ([]domain.MissionOffer, error) {
	return nil, errors.New("store unavailable")
}

func TestListOffersHandler_Returns500OnReaderError(t *testing.T) {
	// Arrange
	app := application.NewListOffersHandler(
		&failingOfferListReader{},
		&fakePlayerContractsReader{},
		&fakeActiveSeasons{seasons: []domain.Season{domain.NewSeason("season-2026-q4", domain.SeasonStateActive)}},
	)
	handler := missionhttp.ListOffersHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// REG-0019-NO-ACTIVE-SEASON: a domain rejection must surface as 422 with its
// public code, not as an internal error.
func TestListOffersHandler_ReturnsNoActiveSeason(t *testing.T) {
	app := application.NewListOffersHandler(
		&fakeOfferListReader{},
		&fakePlayerContractsReader{},
		&fakeActiveSeasons{},
	)
	handler := missionhttp.ListOffersHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/offers?status=AVAILABLE", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var result map[string]any
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "NO_ACTIVE_SEASON", result["code"])
}
