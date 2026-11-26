package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type failingActiveSeasons struct{ err error }

func (s *failingActiveSeasons) ListActive() ([]domain.Season, error) {
	return nil, s.err
}

// An explicit ?season_id resolves that season without consulting active
// seasons; only offers bound to it contribute facets.
func TestListOfferFacetsHandler_ExplicitSeasonFilters(t *testing.T) {
	// Arrange
	offers := &fakeOfferListReader{offers: []domain.MissionOffer{
		domain.NewMissionOfferWithMode("offer-1", "BOUNTY", domain.MissionModeSolo, "season-2026-q3"),
		domain.NewMissionOfferWithMode("offer-2", "EXPLORATION", domain.MissionModeTeam, "season-2026-q4"),
	}}
	handler := newFacetsHandler(offers, &fakeActiveSeasons{})

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets?season_id=season-2026-q3")

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	var facets domain.OfferFacets
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&facets))
	require.Equal(t, []string{"BOUNTY"}, facets.Categories)
	require.Equal(t, []domain.MissionMode{domain.MissionModeSolo}, facets.Modes)
}

// Multiple active seasons with no explicit season_id is the board's
// ACTIVE_SEASON_EXISTS ambiguity: 422, not a silent pick.
func TestListOfferFacetsHandler_MultipleActiveSeasonsRejects(t *testing.T) {
	// Arrange
	seasons := &fakeActiveSeasons{seasons: []domain.Season{
		domain.NewSeason("season-2026-q3", domain.SeasonStateActive),
		domain.NewSeason("season-2026-q4", domain.SeasonStateActive),
	}}
	handler := newFacetsHandler(&fakeOfferListReader{}, seasons)

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets")

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.Equal(t, "ACTIVE_SEASON_EXISTS", facetErrorCode(t, rec))
}

// A technical failure listing active seasons stays a plain
// 500 INTERNAL_ERROR.
func TestListOfferFacetsHandler_SeasonReaderErrorIsInternal(t *testing.T) {
	// Arrange
	seasons := &failingActiveSeasons{err: errors.New("season reader down")}
	handler := newFacetsHandler(&fakeOfferListReader{}, seasons)

	// Act
	rec := getFacets(handler, "/api/v1/missions/offers/facets")

	// Assert
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "INTERNAL_ERROR", facetErrorCode(t, rec))
}
