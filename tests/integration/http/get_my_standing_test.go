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

// seasonReaderDouble is a SeasonReader stub.
type seasonReaderDouble struct {
	season domain.Season
	err    error
}

func (r *seasonReaderDouble) Get(string) (domain.Season, error) {
	return r.season, r.err
}

// leaderboardReaderDouble is a LeaderboardReader stub.
type leaderboardReaderDouble struct {
	standings []domain.Standing
	err       error
}

func (r *leaderboardReaderDouble) ListStandings(_, _ string, _ domain.Track, _ int) ([]domain.Standing, error) {
	return r.standings, r.err
}

// myStandingServer mirrors the composition wiring for the my-standing route.
func myStandingServer(session boundarySession, handler *application.GetMyStandingHandler) *httptest.Server {
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me",
		auth.Authenticate(missionhttp.GetMyStandingHandler(handler)))
	return httptest.NewServer(mux)
}

func getMyStanding(t *testing.T, server *httptest.Server, query string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/v1/missions/seasons/season-1/leaderboards/BOUNTY_HUNTING/me"+query, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// A member reads their own standing with rank and total.
func TestGetMyStanding_ReturnsCallerEntry(t *testing.T) {
	// Arrange
	standings := []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Total: 200},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
	}
	handler := application.NewGetMyStandingHandler(
		&seasonReaderDouble{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&leaderboardReaderDouble{standings: standings})
	server := myStandingServer(boundarySession{playerID: "player-1"}, handler)
	defer server.Close()

	// Act
	resp := getMyStanding(t, server, "?track=individual")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, float64(2), body["Rank"])
}

// A caller absent from the bucket gets 404 STANDING_NOT_FOUND.
func TestGetMyStanding_AbsentReturns404(t *testing.T) {
	// Arrange
	handler := application.NewGetMyStandingHandler(
		&seasonReaderDouble{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&leaderboardReaderDouble{})
	server := myStandingServer(boundarySession{playerID: "player-9"}, handler)
	defer server.Close()

	// Act
	resp := getMyStanding(t, server, "?track=individual")

	// Assert
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "STANDING_NOT_FOUND", body["code"])
}
