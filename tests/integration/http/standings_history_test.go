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

// bucketReaderDouble is a StandingsBucketReader stub.
type bucketReaderDouble struct {
	buckets []domain.StandingBucket
	err     error
}

func (r *bucketReaderDouble) ListBuckets(string) ([]domain.StandingBucket, error) {
	return r.buckets, r.err
}

// historyServer mirrors the composition wiring for the standings-history route.
func historyServer(session boundarySession, handler *application.ListStandingsHistoryHandler) *httptest.Server {
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/standings/history",
		auth.Authenticate(missionhttp.StandingsHistoryHandler(handler)))
	return httptest.NewServer(mux)
}

func getHistory(t *testing.T, server *httptest.Server, query string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/v1/missions/seasons/season-1/standings/history"+query, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// A consumer reads the finalized season's ranked buckets.
func TestStandingsHistory_FinalizedReturnsBuckets(t *testing.T) {
	// Arrange
	buckets := []domain.StandingBucket{{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual}}
	standings := []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150},
	}
	handler := application.NewListStandingsHistoryHandler(
		&seasonReaderDouble{season: domain.NewSeason("season-1", domain.SeasonStateFinalized)},
		&bucketReaderDouble{buckets: buckets},
		&leaderboardReaderDouble{standings: standings})
	server := historyServer(boundarySession{playerID: "player-1"}, handler)
	defer server.Close()

	// Act
	resp := getHistory(t, server, "")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.False(t, body["SeasonID"] != "season-1" && body["season_id"] != "season-1")
}

// A non-finalized season returns 422 SEASON_STATE_INVALID.
func TestStandingsHistory_ActiveRejected(t *testing.T) {
	// Arrange
	handler := application.NewListStandingsHistoryHandler(
		&seasonReaderDouble{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&bucketReaderDouble{},
		&leaderboardReaderDouble{})
	server := historyServer(boundarySession{playerID: "player-1"}, handler)
	defer server.Close()

	// Act
	resp := getHistory(t, server, "")

	// Assert
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "SEASON_STATE_INVALID", body["code"])
}
