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

// statsServer mirrors the composition wiring for the statistics route.
func statsServer(session boundarySession, handler *application.GetSeasonStatisticsHandler) *httptest.Server {
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/statistics",
		auth.Authenticate(missionhttp.SeasonStatisticsHandler(handler)))
	return httptest.NewServer(mux)
}

func getStatistics(t *testing.T, server *httptest.Server, seasonID, query string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/v1/missions/seasons/"+seasonID+"/statistics"+query, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// A consumer reads grouped statistics for an ACTIVE season.
func TestSeasonStatistics_ActiveSeasonReturnsBucketsWithTop(t *testing.T) {
	// Arrange
	buckets := []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
		{Category: "BOUNTY_HUNTING", Track: domain.TrackTeam},
	}
	standings := []domain.Standing{
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-1", Total: 150, CompletedMissions: 2},
		{SeasonID: "season-1", Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual, RecipientID: "player-2", Total: 100},
	}
	handler := application.NewGetSeasonStatisticsHandler(
		&seasonReaderDouble{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		&activeSeasonsDouble{},
		&bucketReaderDouble{buckets: buckets},
		&leaderboardReaderDouble{standings: standings})
	server := statsServer(boundarySession{playerID: "player-1"}, handler)
	defer server.Close()

	// Act
	resp := getStatistics(t, server, "season-1", "")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		SeasonID string `json:"SeasonID"`
		State    string `json:"State"`
		Buckets  []struct {
			Bucket struct {
				Category string `json:"Category"`
				Track    string `json:"Track"`
			} `json:"Bucket"`
			Participants int `json:"Participants"`
			TotalMerits  int `json:"TotalMerits"`
			Top          []struct {
				Rank        int    `json:"Rank"`
				RecipientID string `json:"RecipientID"`
				Total       int    `json:"Total"`
			} `json:"Top"`
		} `json:"Buckets"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, "season-1", body.SeasonID)
	require.Equal(t, "ACTIVE", body.State)
	require.Len(t, body.Buckets, 2)
	require.Equal(t, 2, body.Buckets[0].Participants)
	require.Equal(t, 250, body.Buckets[0].TotalMerits)
	require.Equal(t, "player-1", body.Buckets[0].Top[0].RecipientID)
	require.Equal(t, 1, body.Buckets[0].Top[0].Rank)
}

// "current" resolves to the single ACTIVE season.
func TestSeasonStatistics_CurrentResolvesActiveSeason(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	handler := application.NewGetSeasonStatisticsHandler(
		&seasonReaderDouble{},
		&activeSeasonsDouble{seasons: []domain.Season{season}},
		&bucketReaderDouble{},
		&leaderboardReaderDouble{})
	server := statsServer(boundarySession{playerID: "player-1"}, handler)
	defer server.Close()

	// Act
	resp := getStatistics(t, server, "current", "")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		SeasonID string `json:"SeasonID"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, "season-1", body.SeasonID)
}

// track and category filters narrow the returned buckets.
func TestSeasonStatistics_FiltersNarrowBuckets(t *testing.T) {
	// Arrange
	buckets := []domain.StandingBucket{
		{Category: "BOUNTY_HUNTING", Track: domain.TrackIndividual},
		{Category: "BOUNTY_HUNTING", Track: domain.TrackTeam},
		{Category: "RACING", Track: domain.TrackIndividual},
	}
	handler := application.NewGetSeasonStatisticsHandler(
		&seasonReaderDouble{season: domain.NewSeason("season-1", domain.SeasonStateFinalized)},
		&activeSeasonsDouble{},
		&bucketReaderDouble{buckets: buckets},
		&leaderboardReaderDouble{})
	server := statsServer(boundarySession{playerID: "player-1"}, handler)
	defer server.Close()

	// Act
	resp := getStatistics(t, server, "season-1", "?track=team")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		Buckets []struct {
			Bucket struct {
				Track string `json:"Track"`
			} `json:"Bucket"`
		} `json:"Buckets"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body.Buckets, 1)
	require.Equal(t, "TEAM", body.Buckets[0].Bucket.Track)
}
