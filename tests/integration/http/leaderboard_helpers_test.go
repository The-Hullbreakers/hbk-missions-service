package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionsession "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type leaderboardFakeSession struct{ tenantID string }

func (s leaderboardFakeSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{"tenant_id": s.tenantID}, nil
}

func (leaderboardFakeSession) HasRole(map[string]any, string) bool { return true }

type rejectingFakeSession struct{}

func (rejectingFakeSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "", nil, errors.New("invalid token")
}

func (rejectingFakeSession) HasRole(map[string]any, string) bool { return false }

type fixedLeaderboardSeasonReader struct {
	season domain.Season
	err    error
}

func (s fixedLeaderboardSeasonReader) Get(string) (domain.Season, error) { return s.season, s.err }

type fixedLeaderboardReader struct{ standings []domain.Standing }

func (r fixedLeaderboardReader) ListStandings(string, string, string, domain.Track, int) ([]domain.Standing, error) {
	return r.standings, nil
}

func newLeaderboardTestServer(standings []domain.Standing, seasonErr error, session missionsession.UserSession) *httptest.Server {
	app := application.NewListLeaderboardHandler(
		fixedLeaderboardSeasonReader{season: domain.NewSeason("season-1", domain.SeasonStateActive), err: seasonErr},
		fixedLeaderboardReader{standings: standings},
	)
	auth := missionhttp.NewAuthMiddleware(session)
	return httptest.NewServer(auth.Authenticate(missionhttp.LeaderboardHandler(app)))
}

func getLeaderboard(t *testing.T, server *httptest.Server, query string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/seasons/season-1/leaderboards/BOUNTY_HUNTING"+query, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}
