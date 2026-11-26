package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// activeSeasonsDouble is an ActiveSeasonsReader stub.
type activeSeasonsDouble struct {
	seasons []domain.Season
	err     error
}

func (s *activeSeasonsDouble) ListActive() ([]domain.Season, error) {
	return s.seasons, s.err
}

// currentSeasonServer mirrors the composition wiring for the current-season route.
func currentSeasonServer(session boundarySession, reader *activeSeasonsDouble) *httptest.Server {
	app := application.NewGetCurrentSeasonHandler(reader)
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/seasons/current",
		auth.Authenticate(missionhttp.GetCurrentSeasonHandler(app)))
	return httptest.NewServer(mux)
}

func getCurrentSeason(t *testing.T, server *httptest.Server) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/v1/missions/seasons/current", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// An authenticated consumer reads the single active season.
func TestGetCurrentSeason_ReturnsActiveSeason(t *testing.T) {
	// Arrange
	reader := &activeSeasonsDouble{seasons: []domain.Season{
		domain.NewSeason("season-1", domain.SeasonStateActive),
	}}
	server := currentSeasonServer(boundarySession{playerID: "player-1"}, reader)
	defer server.Close()

	// Act
	resp := getCurrentSeason(t, server)

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "season-1", body["SeasonID"])
}

// No active season returns 404 NO_ACTIVE_SEASON.
func TestGetCurrentSeason_NoneReturns404(t *testing.T) {
	// Arrange
	server := currentSeasonServer(boundarySession{playerID: "player-1"}, &activeSeasonsDouble{})
	defer server.Close()

	// Act
	resp := getCurrentSeason(t, server)

	// Assert
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "NO_ACTIVE_SEASON", body["code"])
}
