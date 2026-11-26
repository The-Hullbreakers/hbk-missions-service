package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func statsHandlerFor(season domain.Season, err error) *application.GetSeasonStatisticsHandler {
	return application.NewGetSeasonStatisticsHandler(
		&seasonReaderDouble{season: season, err: err},
		&activeSeasonsDouble{},
		&bucketReaderDouble{},
		&leaderboardReaderDouble{})
}

// "current" without an active season returns 404 NO_ACTIVE_SEASON.
func TestSeasonStatistics_CurrentWithoutActiveSeason(t *testing.T) {
	// Arrange
	server := statsServer(boundarySession{playerID: "player-1"},
		statsHandlerFor(domain.Season{}, nil))
	defer server.Close()

	// Act
	resp := getStatistics(t, server, "current", "")

	// Assert
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "NO_ACTIVE_SEASON", body["code"])
}

// An unknown season returns 404 SEASON_NOT_FOUND.
func TestSeasonStatistics_UnknownSeason(t *testing.T) {
	// Arrange
	server := statsServer(boundarySession{playerID: "player-1"},
		statsHandlerFor(domain.Season{}, domain.ErrSeasonNotFound))
	defer server.Close()

	// Act
	resp := getStatistics(t, server, "season-x", "")

	// Assert
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "SEASON_NOT_FOUND", body["code"])
}

// Invalid track and top_n return 400 INVALID_QUERY_PARAMETER.
func TestSeasonStatistics_InvalidQueryParameters(t *testing.T) {
	// Arrange
	server := statsServer(boundarySession{playerID: "player-1"},
		statsHandlerFor(domain.NewSeason("season-1", domain.SeasonStateActive), nil))
	defer server.Close()

	// Act + Assert
	for _, query := range []string{"?track=squad", "?top_n=abc", "?top_n=0", "?top_n=51"} {
		resp := getStatistics(t, server, "season-1", query)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "query %s", query)
		var body map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&body)
		require.Equal(t, "INVALID_QUERY_PARAMETER", body["code"], "query %s", query)
		resp.Body.Close()
	}
}
